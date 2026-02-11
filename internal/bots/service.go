package bots

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/memohai/memoh/internal/db/sqlc"
)

// Service provides bot CRUD and membership management.
type Service struct {
	queries            *sqlc.Queries
	logger             *slog.Logger
	containerLifecycle ContainerLifecycle
}

var (
	ErrBotNotFound       = errors.New("bot not found")
	ErrBotAccessDenied   = errors.New("bot access denied")
	ErrOwnerUserNotFound = errors.New("owner user not found")
)

// AccessPolicy controls bot access behavior.
type AccessPolicy struct {
	AllowPublicMember bool
}

// NewService creates a new bot service.
func NewService(log *slog.Logger, queries *sqlc.Queries) *Service {
	if log == nil {
		log = slog.Default()
	}
	return &Service{
		queries: queries,
		logger:  log.With(slog.String("service", "bots")),
	}
}

// SetContainerLifecycle registers a container lifecycle handler for bot operations.
func (s *Service) SetContainerLifecycle(lc ContainerLifecycle) {
	s.containerLifecycle = lc
}

// AuthorizeAccess checks whether userID may access the given bot.
func (s *Service) AuthorizeAccess(ctx context.Context, userID, botID string, isAdmin bool, policy AccessPolicy) (Bot, error) {
	if s.queries == nil {
		return Bot{}, fmt.Errorf("bot queries not configured")
	}
	bot, err := s.Get(ctx, botID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Bot{}, ErrBotNotFound
		}
		return Bot{}, err
	}
	if isAdmin || bot.OwnerUserID == userID {
		return bot, nil
	}
	if policy.AllowPublicMember && bot.Type == BotTypePublic {
		if _, err := s.GetMember(ctx, botID, userID); err == nil {
			return bot, nil
		}
	}
	return Bot{}, ErrBotAccessDenied
}

// Create creates a new bot owned by owner user.
func (s *Service) Create(ctx context.Context, ownerUserID string, req CreateBotRequest) (Bot, error) {
	if s.queries == nil {
		return Bot{}, fmt.Errorf("bot queries not configured")
	}
	ownerID := strings.TrimSpace(ownerUserID)
	if ownerID == "" {
		return Bot{}, fmt.Errorf("owner user id is required")
	}
	ownerUUID, err := parseUUID(ownerID)
	if err != nil {
		return Bot{}, err
	}
	if err := s.ensureUserExists(ctx, ownerUUID); err != nil {
		return Bot{}, err
	}
	normalizedType, err := normalizeBotType(req.Type)
	if err != nil {
		return Bot{}, err
	}
	displayName := strings.TrimSpace(req.DisplayName)
	if displayName == "" {
		displayName = "bot-" + uuid.NewString()
	}
	avatarURL := strings.TrimSpace(req.AvatarURL)
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	metadata := req.Metadata
	if metadata == nil {
		metadata = map[string]any{}
	}
	payload, err := json.Marshal(metadata)
	if err != nil {
		return Bot{}, err
	}
	row, err := s.queries.CreateBot(ctx, sqlc.CreateBotParams{
		OwnerUserID: ownerUUID,
		Type:        normalizedType,
		DisplayName: pgtype.Text{String: displayName, Valid: displayName != ""},
		AvatarUrl:   pgtype.Text{String: avatarURL, Valid: avatarURL != ""},
		IsActive:    isActive,
		Metadata:    payload,
	})
	if err != nil {
		return Bot{}, err
	}
	bot, err := toBot(row)
	if err != nil {
		return Bot{}, err
	}
	if s.containerLifecycle != nil {
		if err := s.containerLifecycle.SetupBotContainer(ctx, bot.ID); err != nil {
			s.logger.Error("failed to setup bot container",
				slog.String("bot_id", bot.ID),
				slog.Any("error", err),
			)
		}
	}
	return bot, nil
}

// Get returns a bot by its ID.
func (s *Service) Get(ctx context.Context, botID string) (Bot, error) {
	if s.queries == nil {
		return Bot{}, fmt.Errorf("bot queries not configured")
	}
	botUUID, err := parseUUID(botID)
	if err != nil {
		return Bot{}, err
	}
	row, err := s.queries.GetBotByID(ctx, botUUID)
	if err != nil {
		return Bot{}, err
	}
	return toBot(row)
}

// ListByOwner returns bots owned by the given user.
func (s *Service) ListByOwner(ctx context.Context, ownerUserID string) ([]Bot, error) {
	if s.queries == nil {
		return nil, fmt.Errorf("bot queries not configured")
	}
	ownerUUID, err := parseUUID(ownerUserID)
	if err != nil {
		return nil, err
	}
	rows, err := s.queries.ListBotsByOwner(ctx, ownerUUID)
	if err != nil {
		return nil, err
	}
	items := make([]Bot, 0, len(rows))
	for _, row := range rows {
		item, err := toBot(row)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

// ListByMember returns bots where the user is a member.
func (s *Service) ListByMember(ctx context.Context, channelIdentityID string) ([]Bot, error) {
	if s.queries == nil {
		return nil, fmt.Errorf("bot queries not configured")
	}
	memberUUID, err := parseUUID(channelIdentityID)
	if err != nil {
		return nil, err
	}
	rows, err := s.queries.ListBotsByMember(ctx, memberUUID)
	if err != nil {
		return nil, err
	}
	items := make([]Bot, 0, len(rows))
	for _, row := range rows {
		item, err := toBot(row)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

// ListAccessible returns all bots the user can access (owned or member).
func (s *Service) ListAccessible(ctx context.Context, channelIdentityID string) ([]Bot, error) {
	owned, err := s.ListByOwner(ctx, channelIdentityID)
	if err != nil {
		return nil, err
	}
	members, err := s.ListByMember(ctx, channelIdentityID)
	if err != nil {
		return nil, err
	}
	seen := map[string]Bot{}
	for _, item := range owned {
		seen[item.ID] = item
	}
	for _, item := range members {
		if _, ok := seen[item.ID]; !ok {
			seen[item.ID] = item
		}
	}
	items := make([]Bot, 0, len(seen))
	for _, item := range seen {
		items = append(items, item)
	}
	return items, nil
}

// Update updates bot profile fields.
func (s *Service) Update(ctx context.Context, botID string, req UpdateBotRequest) (Bot, error) {
	if s.queries == nil {
		return Bot{}, fmt.Errorf("bot queries not configured")
	}
	botUUID, err := parseUUID(botID)
	if err != nil {
		return Bot{}, err
	}
	existing, err := s.queries.GetBotByID(ctx, botUUID)
	if err != nil {
		return Bot{}, err
	}
	displayName := strings.TrimSpace(existing.DisplayName.String)
	avatarURL := strings.TrimSpace(existing.AvatarUrl.String)
	isActive := existing.IsActive
	metadata, err := decodeMetadata(existing.Metadata)
	if err != nil {
		return Bot{}, err
	}
	if req.DisplayName != nil {
		displayName = strings.TrimSpace(*req.DisplayName)
	}
	if req.AvatarURL != nil {
		avatarURL = strings.TrimSpace(*req.AvatarURL)
	}
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	if req.Metadata != nil {
		metadata = req.Metadata
	}
	if displayName == "" {
		displayName = "bot-" + uuid.NewString()
	}
	payload, err := json.Marshal(metadata)
	if err != nil {
		return Bot{}, err
	}
	row, err := s.queries.UpdateBotProfile(ctx, sqlc.UpdateBotProfileParams{
		ID:          botUUID,
		DisplayName: pgtype.Text{String: displayName, Valid: displayName != ""},
		AvatarUrl:   pgtype.Text{String: avatarURL, Valid: avatarURL != ""},
		IsActive:    isActive,
		Metadata:    payload,
	})
	if err != nil {
		return Bot{}, err
	}
	return toBot(row)
}

// TransferOwner transfers bot ownership to another user.
func (s *Service) TransferOwner(ctx context.Context, botID string, ownerUserID string) (Bot, error) {
	if s.queries == nil {
		return Bot{}, fmt.Errorf("bot queries not configured")
	}
	botUUID, err := parseUUID(botID)
	if err != nil {
		return Bot{}, err
	}
	ownerUUID, err := parseUUID(ownerUserID)
	if err != nil {
		return Bot{}, err
	}
	if err := s.ensureUserExists(ctx, ownerUUID); err != nil {
		return Bot{}, err
	}
	row, err := s.queries.UpdateBotOwner(ctx, sqlc.UpdateBotOwnerParams{
		ID:          botUUID,
		OwnerUserID: ownerUUID,
	})
	if err != nil {
		return Bot{}, err
	}
	return toBot(row)
}

// Delete removes a bot and its associated resources.
func (s *Service) Delete(ctx context.Context, botID string) error {
	if s.queries == nil {
		return fmt.Errorf("bot queries not configured")
	}
	botUUID, err := parseUUID(botID)
	if err != nil {
		return err
	}
	if _, err := s.queries.GetBotByID(ctx, botUUID); err != nil {
		return err
	}
	if s.containerLifecycle != nil {
		s.logger.Info("cleaning up bot container before deletion", slog.String("bot_id", botID))
		if err := s.containerLifecycle.CleanupBotContainer(ctx, botID); err != nil {
			s.logger.Error("failed to cleanup bot container",
				slog.String("bot_id", botID),
				slog.Any("error", err),
			)
		}
	} else {
		s.logger.Warn("container lifecycle not configured, skipping container cleanup", slog.String("bot_id", botID))
	}
	return s.queries.DeleteBotByID(ctx, botUUID)
}

func (s *Service) ensureUserExists(ctx context.Context, userID pgtype.UUID) error {
	if s.queries == nil {
		return fmt.Errorf("bot queries not configured")
	}
	_, err := s.queries.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrOwnerUserNotFound
		}
		return err
	}
	return nil
}

// UpsertMember creates or updates a bot membership.
func (s *Service) UpsertMember(ctx context.Context, botID string, req UpsertMemberRequest) (BotMember, error) {
	if s.queries == nil {
		return BotMember{}, fmt.Errorf("bot queries not configured")
	}
	botUUID, err := parseUUID(botID)
	if err != nil {
		return BotMember{}, err
	}
	memberUUID, err := parseUUID(req.UserID)
	if err != nil {
		return BotMember{}, err
	}
	role, err := normalizeMemberRole(req.Role)
	if err != nil {
		return BotMember{}, err
	}
	row, err := s.queries.UpsertBotMember(ctx, sqlc.UpsertBotMemberParams{
		BotID:  botUUID,
		UserID: memberUUID,
		Role:   role,
	})
	if err != nil {
		return BotMember{}, err
	}
	return toBotMember(row), nil
}

// ListMembers returns all members of a bot.
func (s *Service) ListMembers(ctx context.Context, botID string) ([]BotMember, error) {
	if s.queries == nil {
		return nil, fmt.Errorf("bot queries not configured")
	}
	botUUID, err := parseUUID(botID)
	if err != nil {
		return nil, err
	}
	rows, err := s.queries.ListBotMembers(ctx, botUUID)
	if err != nil {
		return nil, err
	}
	items := make([]BotMember, 0, len(rows))
	for _, row := range rows {
		items = append(items, toBotMember(row))
	}
	return items, nil
}

// GetMember returns a specific bot member.
func (s *Service) GetMember(ctx context.Context, botID, channelIdentityID string) (BotMember, error) {
	if s.queries == nil {
		return BotMember{}, fmt.Errorf("bot queries not configured")
	}
	botUUID, err := parseUUID(botID)
	if err != nil {
		return BotMember{}, err
	}
	memberUUID, err := parseUUID(channelIdentityID)
	if err != nil {
		return BotMember{}, err
	}
	row, err := s.queries.GetBotMember(ctx, sqlc.GetBotMemberParams{
		BotID:  botUUID,
		UserID: memberUUID,
	})
	if err != nil {
		return BotMember{}, err
	}
	return toBotMember(row), nil
}

// DeleteMember removes a member from a bot.
func (s *Service) DeleteMember(ctx context.Context, botID, channelIdentityID string) error {
	if s.queries == nil {
		return fmt.Errorf("bot queries not configured")
	}
	botUUID, err := parseUUID(botID)
	if err != nil {
		return err
	}
	memberUUID, err := parseUUID(channelIdentityID)
	if err != nil {
		return err
	}
	return s.queries.DeleteBotMember(ctx, sqlc.DeleteBotMemberParams{
		BotID:  botUUID,
		UserID: memberUUID,
	})
}

// UpsertMemberSimple creates or updates a bot membership with a direct channel identity ID and role.
// This satisfies the router.BotMemberService interface.
func (s *Service) UpsertMemberSimple(ctx context.Context, botID, channelIdentityID, role string) error {
	_, err := s.UpsertMember(ctx, botID, UpsertMemberRequest{
		UserID: channelIdentityID,
		Role:   role,
	})
	return err
}

// IsMember checks if a user is a member of a bot.
func (s *Service) IsMember(ctx context.Context, botID, channelIdentityID string) (bool, error) {
	_, err := s.GetMember(ctx, botID, channelIdentityID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func normalizeBotType(raw string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(raw))
	if normalized == "" {
		return BotTypePersonal, nil
	}
	switch normalized {
	case BotTypePersonal, BotTypePublic:
		return normalized, nil
	default:
		return "", fmt.Errorf("invalid bot type: %s", raw)
	}
}

func normalizeMemberRole(raw string) (string, error) {
	role := strings.ToLower(strings.TrimSpace(raw))
	if role == "" {
		return MemberRoleMember, nil
	}
	switch role {
	case MemberRoleOwner, MemberRoleAdmin, MemberRoleMember:
		return role, nil
	default:
		return "", fmt.Errorf("invalid member role: %s", raw)
	}
}

func toBot(row sqlc.Bot) (Bot, error) {
	displayName := ""
	if row.DisplayName.Valid {
		displayName = row.DisplayName.String
	}
	avatarURL := ""
	if row.AvatarUrl.Valid {
		avatarURL = row.AvatarUrl.String
	}
	metadata, err := decodeMetadata(row.Metadata)
	if err != nil {
		return Bot{}, err
	}
	createdAt := time.Time{}
	if row.CreatedAt.Valid {
		createdAt = row.CreatedAt.Time
	}
	updatedAt := time.Time{}
	if row.UpdatedAt.Valid {
		updatedAt = row.UpdatedAt.Time
	}
	return Bot{
		ID:          toUUIDString(row.ID),
		OwnerUserID: toUUIDString(row.OwnerUserID),
		Type:        row.Type,
		DisplayName: displayName,
		AvatarURL:   avatarURL,
		IsActive:    row.IsActive,
		Metadata:    metadata,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}, nil
}

func toBotMember(row sqlc.BotMember) BotMember {
	createdAt := time.Time{}
	if row.CreatedAt.Valid {
		createdAt = row.CreatedAt.Time
	}
	return BotMember{
		BotID:     toUUIDString(row.BotID),
		UserID:    toUUIDString(row.UserID),
		Role:      row.Role,
		CreatedAt: createdAt,
	}
}

func decodeMetadata(payload []byte) (map[string]any, error) {
	if len(payload) == 0 {
		return map[string]any{}, nil
	}
	var data map[string]any
	if err := json.Unmarshal(payload, &data); err != nil {
		return nil, err
	}
	if data == nil {
		data = map[string]any{}
	}
	return data, nil
}

func parseUUID(id string) (pgtype.UUID, error) {
	parsed, err := uuid.Parse(strings.TrimSpace(id))
	if err != nil {
		return pgtype.UUID{}, fmt.Errorf("invalid UUID: %w", err)
	}
	var pgID pgtype.UUID
	pgID.Valid = true
	copy(pgID.Bytes[:], parsed[:])
	return pgID, nil
}

func toUUIDString(value pgtype.UUID) string {
	if !value.Valid {
		return ""
	}
	parsed, err := uuid.FromBytes(value.Bytes[:])
	if err != nil {
		return ""
	}
	return parsed.String()
}
