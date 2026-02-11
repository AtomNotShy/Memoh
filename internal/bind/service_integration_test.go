//go:build ignore
// +build ignore

package bind_test

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/memohai/memoh/internal/channelidentities"
	"github.com/memohai/memoh/internal/bind"
	"github.com/memohai/memoh/internal/db"
	"github.com/memohai/memoh/internal/db/sqlc"
)

func setupBindIntegrationTest(t *testing.T) (*sqlc.Queries, *channelidentities.Service, *bind.Service, func()) {
	t.Helper()

	dsn := os.Getenv("TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("skip integration test: TEST_POSTGRES_DSN is not set")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Skipf("skip integration test: cannot connect to database: %v", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Skipf("skip integration test: database ping failed: %v", err)
	}

	queries := sqlc.New(pool)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	channelIdentitySvc := channelidentities.NewService(logger, queries)
	bindSvc := bind.NewService(logger, pool, queries)

	return queries, channelIdentitySvc, bindSvc, func() { pool.Close() }
}

func createUser(ctx context.Context, queries *sqlc.Queries) (string, error) {
	row, err := queries.CreateUser(ctx, sqlc.CreateUserParams{
		IsActive: true,
		Metadata: []byte("{}"),
	})
	if err != nil {
		return "", err
	}
	return db.UUIDToString(row.ID), nil
}

func createBot(ctx context.Context, queries *sqlc.Queries, ownerUserID string) (string, error) {
	pgOwnerID, err := db.ParseUUID(ownerUserID)
	if err != nil {
		return "", err
	}
	meta, _ := json.Marshal(map[string]any{"source": "bind-integration-test"})
	row, err := queries.CreateBot(ctx, sqlc.CreateBotParams{
		OwnerUserID: pgOwnerID,
		Type:        "personal",
		DisplayName: pgtype.Text{String: "bind-test-bot", Valid: true},
		IsActive:    true,
		Metadata:    meta,
	})
	if err != nil {
		return "", err
	}
	return db.UUIDToString(row.ID), nil
}

func TestIntegrationConsumeBindCodeSuccessAndSingleUse(t *testing.T) {
	queries, channelIdentitySvc, bindSvc, cleanup := setupBindIntegrationTest(t)
	defer cleanup()

	ctx := context.Background()
	ownerUserID, err := createUser(ctx, queries)
	if err != nil {
		t.Fatalf("create owner user failed: %v", err)
	}
	sourceChannelIdentity, err := channelIdentitySvc.Create(ctx, channelidentities.KindChannel)
	if err != nil {
		t.Fatalf("create source channelIdentity failed: %v", err)
	}
	botID, err := createBot(ctx, queries, ownerUserID)
	if err != nil {
		t.Fatalf("create bot failed: %v", err)
	}

	code, err := bindSvc.Issue(ctx, botID, ownerUserID, 10*time.Minute)
	if err != nil {
		t.Fatalf("issue bind code failed: %v", err)
	}
	if err := bindSvc.Consume(ctx, code, sourceChannelIdentity.ID); err != nil {
		t.Fatalf("consume bind code failed: %v", err)
	}

	after, err := bindSvc.Get(ctx, code.Token)
	if err != nil {
		t.Fatalf("get bind code failed: %v", err)
	}
	if after.UsedAt.IsZero() {
		t.Fatal("expected used_at to be set after consume")
	}
	if after.UsedByChannelIdentityID != sourceChannelIdentity.ID {
		t.Fatalf("expected used_by_channel_identity_id=%s, got %s", sourceChannelIdentity.ID, after.UsedByChannelIdentityID)
	}

	linkedUserID, err := channelIdentitySvc.GetLinkedUserID(ctx, sourceChannelIdentity.ID)
	if err != nil {
		t.Fatalf("get linked user failed: %v", err)
	}
	if linkedUserID != ownerUserID {
		t.Fatalf("expected linked user=%s, got %s", ownerUserID, linkedUserID)
	}

	if err := bindSvc.Consume(ctx, code, sourceChannelIdentity.ID); !errors.Is(err, bind.ErrCodeUsed) {
		t.Fatalf("expected ErrCodeUsed on second consume, got %v", err)
	}
}

func TestIntegrationConsumeBindCodeRollbackOnLinkConflict(t *testing.T) {
	queries, channelIdentitySvc, bindSvc, cleanup := setupBindIntegrationTest(t)
	defer cleanup()

	ctx := context.Background()
	ownerUserID, err := createUser(ctx, queries)
	if err != nil {
		t.Fatalf("create owner user failed: %v", err)
	}
	otherUserID, err := createUser(ctx, queries)
	if err != nil {
		t.Fatalf("create other user failed: %v", err)
	}
	sourceChannelIdentity, err := channelIdentitySvc.Create(ctx, channelidentities.KindChannel)
	if err != nil {
		t.Fatalf("create source channelIdentity failed: %v", err)
	}
	if err := channelIdentitySvc.LinkChannelIdentityToUser(ctx, sourceChannelIdentity.ID, otherUserID); err != nil {
		t.Fatalf("pre-link source channelIdentity failed: %v", err)
	}
	botID, err := createBot(ctx, queries, ownerUserID)
	if err != nil {
		t.Fatalf("create bot failed: %v", err)
	}

	code, err := bindSvc.Issue(ctx, botID, ownerUserID, 10*time.Minute)
	if err != nil {
		t.Fatalf("issue bind code failed: %v", err)
	}
	if err := bindSvc.Consume(ctx, code, sourceChannelIdentity.ID); !errors.Is(err, bind.ErrLinkConflict) {
		t.Fatalf("expected ErrLinkConflict, got %v", err)
	}

	after, err := bindSvc.Get(ctx, code.Token)
	if err != nil {
		t.Fatalf("get bind code failed: %v", err)
	}
	if !after.UsedAt.IsZero() {
		t.Fatal("expected used_at to remain empty when consume fails")
	}
}
package bind_test

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/memohai/memoh/internal/channelidentities"
	"github.com/memohai/memoh/internal/bind"
	"github.com/memohai/memoh/internal/db"
	"github.com/memohai/memoh/internal/db/sqlc"
)

func setupBindIntegrationTest(t *testing.T) (*sqlc.Queries, *channelidentities.Service, *bind.Service, func()) {
	t.Helper()

	dsn := os.Getenv("TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("skip integration test: TEST_POSTGRES_DSN is not set")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Skipf("skip integration test: cannot connect to database: %v", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Skipf("skip integration test: database ping failed: %v", err)
	}

	queries := sqlc.New(pool)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	channelIdentitySvc := channelidentities.NewService(logger, queries)
	bindSvc := bind.NewService(logger, pool, queries)

	cleanup := func() {
		pool.Close()
	}
	return queries, channelIdentitySvc, bindSvc, cleanup
}

func createUserForBindTest(ctx context.Context, queries *sqlc.Queries) (string, error) {
	row, err := queries.CreateUser(ctx, sqlc.CreateUserParams{
		IsActive: true,
		Metadata: []byte("{}"),
	})
	if err != nil {
		return "", err
	}
	return db.UUIDToString(row.ID), nil
}

func createBotForBindTest(ctx context.Context, queries *sqlc.Queries, ownerUserID string) (string, error) {
	pgOwnerID, err := db.ParseUUID(ownerUserID)
	if err != nil {
		return "", err
	}
	meta, _ := json.Marshal(map[string]any{"source": "bind-integration-test"})
	row, err := queries.CreateBot(ctx, sqlc.CreateBotParams{
		OwnerUserID: pgOwnerID,
		Type:        "personal",
		DisplayName: pgtype.Text{String: "bind-test-bot", Valid: true},
		AvatarUrl:   pgtype.Text{},
		IsActive:    true,
		Metadata:    meta,
	})
	if err != nil {
		return "", err
	}
	return db.UUIDToString(row.ID), nil
}

func TestIntegrationConsumeBindCodeSuccessAndSingleUse(t *testing.T) {
	queries, channelIdentitySvc, bindSvc, cleanup := setupBindIntegrationTest(t)
	defer cleanup()

	ctx := context.Background()
	ownerUserID, err := createUserForBindTest(ctx, queries)
	if err != nil {
		t.Fatalf("create owner user failed: %v", err)
	}
	sourceChannelIdentity, err := channelIdentitySvc.Create(ctx, channelidentities.KindChannel)
	if err != nil {
		t.Fatalf("create source channelIdentity failed: %v", err)
	}
	botID, err := createBotForBindTest(ctx, queries, ownerUserID)
	if err != nil {
		t.Fatalf("create bot failed: %v", err)
	}

	code, err := bindSvc.Issue(ctx, botID, ownerUserID, 10*time.Minute)
	if err != nil {
		t.Fatalf("issue bind code failed: %v", err)
	}
	if err := bindSvc.Consume(ctx, code, sourceChannelIdentity.ID); err != nil {
		t.Fatalf("consume bind code failed: %v", err)
	}

	after, err := bindSvc.Get(ctx, code.Token)
	if err != nil {
		t.Fatalf("get bind code failed: %v", err)
	}
	if after.UsedAt.IsZero() {
		t.Fatal("expected used_at to be set after successful consume")
	}
	if after.UsedByChannelIdentityID != sourceChannelIdentity.ID {
		t.Fatalf("expected used_by_channel_identity_id=%s, got %s", sourceChannelIdentity.ID, after.UsedByChannelIdentityID)
	}

	linkedUserID, err := channelIdentitySvc.GetLinkedUserID(ctx, sourceChannelIdentity.ID)
	if err != nil {
		t.Fatalf("get linked user failed: %v", err)
	}
	if linkedUserID != ownerUserID {
		t.Fatalf("expected linked user=%s, got %s", ownerUserID, linkedUserID)
	}

	if err := bindSvc.Consume(ctx, code, sourceChannelIdentity.ID); !errors.Is(err, bind.ErrCodeUsed) {
		t.Fatalf("expected ErrCodeUsed on second consume, got %v", err)
	}
}

func TestIntegrationConsumeBindCodeRollbackOnLinkConflict(t *testing.T) {
	queries, channelIdentitySvc, bindSvc, cleanup := setupBindIntegrationTest(t)
	defer cleanup()

	ctx := context.Background()
	ownerUserID, err := createUserForBindTest(ctx, queries)
	if err != nil {
		t.Fatalf("create owner user failed: %v", err)
	}
	otherUserID, err := createUserForBindTest(ctx, queries)
	if err != nil {
		t.Fatalf("create other user failed: %v", err)
	}
	sourceChannelIdentity, err := channelIdentitySvc.Create(ctx, channelidentities.KindChannel)
	if err != nil {
		t.Fatalf("create source channelIdentity failed: %v", err)
	}
	if err := channelIdentitySvc.LinkChannelIdentityToUser(ctx, sourceChannelIdentity.ID, otherUserID); err != nil {
		t.Fatalf("pre-link source channelIdentity failed: %v", err)
	}
	botID, err := createBotForBindTest(ctx, queries, ownerUserID)
	if err != nil {
		t.Fatalf("create bot failed: %v", err)
	}

	code, err := bindSvc.Issue(ctx, botID, ownerUserID, 10*time.Minute)
	if err != nil {
		t.Fatalf("issue bind code failed: %v", err)
	}
	if err := bindSvc.Consume(ctx, code, sourceChannelIdentity.ID); !errors.Is(err, bind.ErrLinkConflict) {
		t.Fatalf("expected ErrLinkConflict, got %v", err)
	}

	after, err := bindSvc.Get(ctx, code.Token)
	if err != nil {
		t.Fatalf("get bind code failed: %v", err)
	}
	if !after.UsedAt.IsZero() {
		t.Fatal("expected used_at to remain empty when consume fails")
	}
}
package bind_test

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/memohai/memoh/internal/channelidentities"
	"github.com/memohai/memoh/internal/bind"
	"github.com/memohai/memoh/internal/db"
	"github.com/memohai/memoh/internal/db/sqlc"
)

func setupBindIntegrationTest(t *testing.T) (*sqlc.Queries, *channelidentities.Service, *bind.Service, func()) {
	t.Helper()

	dsn := os.Getenv("TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("skip integration test: TEST_POSTGRES_DSN is not set")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Skipf("skip integration test: cannot connect to database: %v", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Skipf("skip integration test: database ping failed: %v", err)
	}

	queries := sqlc.New(pool)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	channelIdentitySvc := channelidentities.NewService(logger, queries)
	bindSvc := bind.NewService(logger, pool, queries)

	cleanup := func() {
		pool.Close()
	}
	return queries, channelIdentitySvc, bindSvc, cleanup
}

func createBotForBindTest(ctx context.Context, queries *sqlc.Queries, ownerChannelIdentityID string) (string, error) {
	pgOwnerID, err := db.ParseUUID(ownerChannelIdentityID)
	if err != nil {
		return "", err
	}
	meta, _ := json.Marshal(map[string]any{"source": "bind-integration-test"})
	row, err := queries.CreateBot(ctx, sqlc.CreateBotParams{
		OwnerChannelIdentityID: pgOwnerID,
		Type:         "personal",
		DisplayName:  pgtype.Text{String: "bind-test-bot", Valid: true},
		AvatarUrl:    pgtype.Text{},
		IsActive:     true,
		Metadata:     meta,
	})
	if err != nil {
		return "", err
	}
	return db.UUIDToString(row.ID), nil
}

func createChatForBindTest(ctx context.Context, queries *sqlc.Queries, botID, channelIdentityID string) (string, error) {
	pgBotID, err := db.ParseUUID(botID)
	if err != nil {
		return "", err
	}
	pgChannelIdentityID, err := db.ParseUUID(channelIdentityID)
	if err != nil {
		return "", err
	}
	row, err := queries.CreateChat(ctx, sqlc.CreateChatParams{
		BotID:        pgBotID,
		Kind:         "direct",
		ParentChatID: pgtype.UUID{},
		Title:        pgtype.Text{},
		CreatedBy:    pgChannelIdentityID,
		Metadata:     []byte("{}"),
	})
	if err != nil {
		return "", err
	}
	return db.UUIDToString(row.ID), nil
}

func TestIntegrationConsumeBindCodeSuccessAndSingleUse(t *testing.T) {
	queries, channelIdentitySvc, bindSvc, cleanup := setupBindIntegrationTest(t)
	defer cleanup()

	ctx := context.Background()
	human, err := channelIdentitySvc.Create(ctx, channelidentities.KindHuman)
	if err != nil {
		t.Fatalf("create human failed: %v", err)
	}
	shadow, err := channelIdentitySvc.Create(ctx, channelidentities.KindShadow)
	if err != nil {
		t.Fatalf("create shadow failed: %v", err)
	}
	botID, err := createBotForBindTest(ctx, queries, human.ID)
	if err != nil {
		t.Fatalf("create bot failed: %v", err)
	}
	chatID, err := createChatForBindTest(ctx, queries, botID, human.ID)
	if err != nil {
		t.Fatalf("create chat failed: %v", err)
	}
	pgChatID, err := db.ParseUUID(chatID)
	if err != nil {
		t.Fatalf("parse chat id failed: %v", err)
	}
	pgShadowID, err := db.ParseUUID(shadow.ID)
	if err != nil {
		t.Fatalf("parse shadow id failed: %v", err)
	}
	pgHumanID, err := db.ParseUUID(human.ID)
	if err != nil {
		t.Fatalf("parse human id failed: %v", err)
	}
	if _, err := queries.AddChatParticipant(ctx, sqlc.AddChatParticipantParams{
		ChatID:  pgChatID,
		ChannelIdentityID: pgShadowID,
		Role:    "member",
	}); err != nil {
		t.Fatalf("add shadow participant failed: %v", err)
	}

	code, err := bindSvc.Issue(ctx, botID, human.ID, 10*time.Minute)
	if err != nil {
		t.Fatalf("issue bind code failed: %v", err)
	}
	if err := bindSvc.Consume(ctx, code, shadow.ID); err != nil {
		t.Fatalf("consume bind code failed: %v", err)
	}

	after, err := bindSvc.Get(ctx, code.Token)
	if err != nil {
		t.Fatalf("get bind code failed: %v", err)
	}
	if after.UsedAt.IsZero() {
		t.Fatal("expected used_at to be set after successful consume")
	}
	if after.UsedByChannelIdentityID != shadow.ID {
		t.Fatalf("expected used_by_channel_identity_id=%s, got %s", shadow.ID, after.UsedByChannelIdentityID)
	}

	canonical, err := channelIdentitySvc.Canonicalize(ctx, shadow.ID)
	if err != nil {
		t.Fatalf("canonicalize failed: %v", err)
	}
	if canonical != human.ID {
		t.Fatalf("expected canonical=%s, got %s", human.ID, canonical)
	}
	if _, err := queries.GetChatParticipant(ctx, sqlc.GetChatParticipantParams{
		ChatID:  pgChatID,
		ChannelIdentityID: pgHumanID,
	}); err != nil {
		t.Fatalf("expected human participant after bind, got error: %v", err)
	}
	if _, err := queries.GetChatParticipant(ctx, sqlc.GetChatParticipantParams{
		ChatID:  pgChatID,
		ChannelIdentityID: pgShadowID,
	}); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("expected shadow participant removed after bind, got %v", err)
	}

	if err := bindSvc.Consume(ctx, code, shadow.ID); !errors.Is(err, bind.ErrCodeUsed) {
		t.Fatalf("expected ErrCodeUsed on second consume, got %v", err)
	}
}

func TestIntegrationConsumeBindCodeRollbackOnLinkConflict(t *testing.T) {
	queries, channelIdentitySvc, bindSvc, cleanup := setupBindIntegrationTest(t)
	defer cleanup()

	ctx := context.Background()
	humanA, err := channelIdentitySvc.Create(ctx, channelidentities.KindHuman)
	if err != nil {
		t.Fatalf("create humanA failed: %v", err)
	}
	humanB, err := channelIdentitySvc.Create(ctx, channelidentities.KindHuman)
	if err != nil {
		t.Fatalf("create humanB failed: %v", err)
	}
	shadow, err := channelIdentitySvc.Create(ctx, channelidentities.KindShadow)
	if err != nil {
		t.Fatalf("create shadow failed: %v", err)
	}
	botID, err := createBotForBindTest(ctx, queries, humanA.ID)
	if err != nil {
		t.Fatalf("create bot failed: %v", err)
	}

	// Pre-link shadow to another user so bind consume hits link conflict.
	if err := channelIdentitySvc.LinkChannelIdentityToUser(ctx, shadow.ID, humanB.ID); err != nil {
		t.Fatalf("pre link shadow->humanB failed: %v", err)
	}

	code, err := bindSvc.Issue(ctx, botID, humanA.ID, 10*time.Minute)
	if err != nil {
		t.Fatalf("issue bind code failed: %v", err)
	}
	if err := bindSvc.Consume(ctx, code, shadow.ID); !errors.Is(err, bind.ErrLinkConflict) {
		t.Fatalf("expected ErrLinkConflict, got %v", err)
	}

	after, err := bindSvc.Get(ctx, code.Token)
	if err != nil {
		t.Fatalf("get bind code failed: %v", err)
	}
	if !after.UsedAt.IsZero() {
		t.Fatal("expected used_at to remain empty when consume fails")
	}
}
