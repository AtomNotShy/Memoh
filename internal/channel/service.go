package channel

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/memohai/memoh/internal/db"
	"github.com/memohai/memoh/internal/db/sqlc"
)

// Service provides CRUD operations for channel configurations, user bindings, and sessions.
type Service struct {
	queries  *sqlc.Queries
	registry *Registry
}

// NewService creates a Service backed by the given database queries and adapter registry.
func NewService(queries *sqlc.Queries, registry *Registry) *Service {
	if registry == nil {
		registry = NewRegistry()
	}
	return &Service{queries: queries, registry: registry}
}

// UpsertConfig creates or updates a bot's channel configuration.
func (s *Service) UpsertConfig(ctx context.Context, botID string, channelType ChannelType, req UpsertConfigRequest) (ChannelConfig, error) {
	if s.queries == nil {
		return ChannelConfig{}, fmt.Errorf("channel queries not configured")
	}
	if channelType == "" {
		return ChannelConfig{}, fmt.Errorf("channel type is required")
	}
	normalized, err := s.registry.NormalizeConfig(channelType, req.Credentials)
	if err != nil {
		return ChannelConfig{}, err
	}
	credentialsPayload, err := json.Marshal(normalized)
	if err != nil {
		return ChannelConfig{}, err
	}
	botUUID, err := db.ParseUUID(botID)
	if err != nil {
		return ChannelConfig{}, err
	}
	selfIdentity := req.SelfIdentity
	if selfIdentity == nil {
		selfIdentity = map[string]any{}
	}
	selfPayload, err := json.Marshal(selfIdentity)
	if err != nil {
		return ChannelConfig{}, err
	}
	routing := req.Routing
	if routing == nil {
		routing = map[string]any{}
	}
	routingPayload, err := json.Marshal(routing)
	if err != nil {
		return ChannelConfig{}, err
	}
	status := strings.TrimSpace(req.Status)
	if status == "" {
		status = "pending"
	}
	verifiedAt := pgtype.Timestamptz{Valid: false}
	if req.VerifiedAt != nil {
		verifiedAt = pgtype.Timestamptz{Time: req.VerifiedAt.UTC(), Valid: true}
	}
	externalIdentity := strings.TrimSpace(req.ExternalIdentity)
	row, err := s.queries.UpsertBotChannelConfig(ctx, sqlc.UpsertBotChannelConfigParams{
		BotID:       botUUID,
		ChannelType: channelType.String(),
		Credentials: credentialsPayload,
		ExternalIdentity: pgtype.Text{
			String: externalIdentity,
			Valid:  externalIdentity != "",
		},
		SelfIdentity: selfPayload,
		Routing:      routingPayload,
		Capabilities: []byte("{}"),
		Status:       status,
		VerifiedAt:   verifiedAt,
	})
	if err != nil {
		return ChannelConfig{}, err
	}
	return normalizeChannelConfig(row)
}

// UpsertUserConfig creates or updates a user's channel binding.
func (s *Service) UpsertUserConfig(ctx context.Context, actorUserID string, channelType ChannelType, req UpsertUserConfigRequest) (ChannelUserBinding, error) {
	if s.queries == nil {
		return ChannelUserBinding{}, fmt.Errorf("channel queries not configured")
	}
	if channelType == "" {
		return ChannelUserBinding{}, fmt.Errorf("channel type is required")
	}
	normalized, err := s.registry.NormalizeUserConfig(channelType, req.Config)
	if err != nil {
		return ChannelUserBinding{}, err
	}
	payload, err := json.Marshal(normalized)
	if err != nil {
		return ChannelUserBinding{}, err
	}
	pgUserID, err := db.ParseUUID(actorUserID)
	if err != nil {
		return ChannelUserBinding{}, err
	}
	row, err := s.queries.UpsertUserChannelBinding(ctx, sqlc.UpsertUserChannelBindingParams{
		UserID:      pgUserID,
		ChannelType: channelType.String(),
		Config:      payload,
	})
	if err != nil {
		return ChannelUserBinding{}, err
	}
	return normalizeChannelUserBindingRow(row)
}

// ResolveEffectiveConfig returns the active channel configuration for a bot.
// For configless channel types, a synthetic config is returned.
func (s *Service) ResolveEffectiveConfig(ctx context.Context, botID string, channelType ChannelType) (ChannelConfig, error) {
	if s.queries == nil {
		return ChannelConfig{}, fmt.Errorf("channel queries not configured")
	}
	if channelType == "" {
		return ChannelConfig{}, fmt.Errorf("channel type is required")
	}
	if s.registry.IsConfigless(channelType) {
		return ChannelConfig{
			ID:          channelType.String() + ":" + strings.TrimSpace(botID),
			BotID:       strings.TrimSpace(botID),
			ChannelType: channelType,
		}, nil
	}
	botUUID, err := db.ParseUUID(botID)
	if err != nil {
		return ChannelConfig{}, err
	}
	row, err := s.queries.GetBotChannelConfig(ctx, sqlc.GetBotChannelConfigParams{
		BotID:       botUUID,
		ChannelType: channelType.String(),
	})
	if err == nil {
		return normalizeChannelConfig(row)
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return ChannelConfig{}, err
	}
	return ChannelConfig{}, fmt.Errorf("channel config not found")
}

// ListConfigsByType returns all channel configurations of the given type.
func (s *Service) ListConfigsByType(ctx context.Context, channelType ChannelType) ([]ChannelConfig, error) {
	if s.queries == nil {
		return nil, fmt.Errorf("channel queries not configured")
	}
	if s.registry.IsConfigless(channelType) {
		return []ChannelConfig{}, nil
	}
	rows, err := s.queries.ListBotChannelConfigsByType(ctx, channelType.String())
	if err != nil {
		return nil, err
	}
	items := make([]ChannelConfig, 0, len(rows))
	for _, row := range rows {
		item, err := normalizeChannelConfig(row)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

// GetUserConfig returns the user's channel binding for the given channel type.
func (s *Service) GetUserConfig(ctx context.Context, actorUserID string, channelType ChannelType) (ChannelUserBinding, error) {
	if s.queries == nil {
		return ChannelUserBinding{}, fmt.Errorf("channel queries not configured")
	}
	if channelType == "" {
		return ChannelUserBinding{}, fmt.Errorf("channel type is required")
	}
	pgUserID, err := db.ParseUUID(actorUserID)
	if err != nil {
		return ChannelUserBinding{}, err
	}
	row, err := s.queries.GetUserChannelBinding(ctx, sqlc.GetUserChannelBindingParams{
		UserID:      pgUserID,
		ChannelType: channelType.String(),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ChannelUserBinding{}, fmt.Errorf("channel user config not found")
		}
		return ChannelUserBinding{}, err
	}
	config, err := DecodeConfigMap(row.Config)
	if err != nil {
		return ChannelUserBinding{}, err
	}
	return ChannelUserBinding{
		ID:          db.UUIDToString(row.ID),
		ChannelType: ChannelType(row.ChannelType),
		UserID:      db.UUIDToString(row.UserID),
		Config:      config,
		CreatedAt:   db.TimeFromPg(row.CreatedAt),
		UpdatedAt:   db.TimeFromPg(row.UpdatedAt),
	}, nil
}

// ListUserConfigsByType returns all user bindings for the given channel type.
func (s *Service) ListUserConfigsByType(ctx context.Context, channelType ChannelType) ([]ChannelUserBinding, error) {
	if s.queries == nil {
		return nil, fmt.Errorf("channel queries not configured")
	}
	rows, err := s.queries.ListUserChannelBindingsByType(ctx, channelType.String())
	if err != nil {
		return nil, err
	}
	items := make([]ChannelUserBinding, 0, len(rows))
	for _, row := range rows {
		item, err := normalizeChannelUserBindingRow(row)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

// GetChannelSession returns the session with the given ID, or an empty session if not found.
func (s *Service) GetChannelSession(ctx context.Context, sessionID string) (ChannelSession, error) {
	if s.queries == nil {
		return ChannelSession{}, fmt.Errorf("channel queries not configured")
	}
	row, err := s.queries.GetChannelSessionByID(ctx, sessionID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ChannelSession{}, nil
		}
		return ChannelSession{}, err
	}
	return normalizeChannelSession(row)
}

// ListSessionsByBotPlatform returns all sessions for the given bot and platform.
func (s *Service) ListSessionsByBotPlatform(ctx context.Context, botID, platform string) ([]ChannelSession, error) {
	if s.queries == nil {
		return nil, fmt.Errorf("channel queries not configured")
	}
	botID = strings.TrimSpace(botID)
	platform = strings.TrimSpace(platform)
	if botID == "" {
		return nil, fmt.Errorf("bot id is required")
	}
	if platform == "" {
		return nil, fmt.Errorf("platform is required")
	}
	pgBotID, err := db.ParseUUID(botID)
	if err != nil {
		return nil, err
	}
	rows, err := s.queries.ListChannelSessionsByBotPlatform(ctx, sqlc.ListChannelSessionsByBotPlatformParams{
		BotID:    pgBotID,
		Platform: platform,
	})
	if err != nil {
		return nil, err
	}
	items := make([]ChannelSession, 0, len(rows))
	for _, row := range rows {
		item, err := normalizeChannelSession(row)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

// UpsertChannelSession creates or updates a channel session record.
func (s *Service) UpsertChannelSession(ctx context.Context, sessionID string, botID string, channelConfigID string, userID string, contactID string, platform string, replyTarget string, threadID string, metadata map[string]any) error {
	if s.queries == nil {
		return fmt.Errorf("channel queries not configured")
	}
	pgUserID := pgtype.UUID{Valid: false}
	if strings.TrimSpace(userID) != "" {
		parsed, err := db.ParseUUID(userID)
		if err != nil {
			return err
		}
		pgUserID = parsed
	}
	botUUID, err := db.ParseUUID(botID)
	if err != nil {
		return err
	}
	var channelUUID pgtype.UUID
	if strings.TrimSpace(channelConfigID) != "" {
		channelUUID, err = db.ParseUUID(channelConfigID)
		if err != nil {
			return err
		}
	}
	pgContactID := pgtype.UUID{Valid: false}
	if strings.TrimSpace(contactID) != "" {
		parsed, err := db.ParseUUID(contactID)
		if err != nil {
			return err
		}
		pgContactID = parsed
	}
	payload := metadata
	if payload == nil {
		payload = map[string]any{}
	}
	metaBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	_, err = s.queries.UpsertChannelSession(ctx, sqlc.UpsertChannelSessionParams{
		SessionID:       sessionID,
		BotID:           botUUID,
		ChannelConfigID: channelUUID,
		UserID:          pgUserID,
		ContactID:       pgContactID,
		Platform:        platform,
		ReplyTarget: pgtype.Text{
			String: strings.TrimSpace(replyTarget),
			Valid:  strings.TrimSpace(replyTarget) != "",
		},
		ThreadID: pgtype.Text{
			String: strings.TrimSpace(threadID),
			Valid:  strings.TrimSpace(threadID) != "",
		},
		Metadata: metaBytes,
	})
	return err
}

// ResolveUserBinding finds the user ID whose channel binding matches the given criteria.
func (s *Service) ResolveUserBinding(ctx context.Context, channelType ChannelType, criteria BindingCriteria) (string, error) {
	rows, err := s.ListUserConfigsByType(ctx, channelType)
	if err != nil {
		return "", err
	}
	if _, ok := s.registry.Get(channelType); !ok {
		return "", fmt.Errorf("unsupported channel type: %s", channelType)
	}
	for _, row := range rows {
		if s.registry.MatchUserBinding(channelType, row.Config, criteria) {
			return row.UserID, nil
		}
	}
	return "", fmt.Errorf("channel user binding not found")
}

func normalizeChannelConfig(row sqlc.BotChannelConfig) (ChannelConfig, error) {
	credentials, err := DecodeConfigMap(row.Credentials)
	if err != nil {
		return ChannelConfig{}, err
	}
	selfIdentity, err := DecodeConfigMap(row.SelfIdentity)
	if err != nil {
		return ChannelConfig{}, err
	}
	routing, err := DecodeConfigMap(row.Routing)
	if err != nil {
		return ChannelConfig{}, err
	}
	verifiedAt := time.Time{}
	if row.VerifiedAt.Valid {
		verifiedAt = row.VerifiedAt.Time
	}
	externalIdentity := ""
	if row.ExternalIdentity.Valid {
		externalIdentity = strings.TrimSpace(row.ExternalIdentity.String)
	}
	return ChannelConfig{
		ID:               db.UUIDToString(row.ID),
		BotID:            db.UUIDToString(row.BotID),
		ChannelType:      ChannelType(row.ChannelType),
		Credentials:      credentials,
		ExternalIdentity: externalIdentity,
		SelfIdentity:     selfIdentity,
		Routing: routing,
		Status:  strings.TrimSpace(row.Status),
		VerifiedAt:       verifiedAt,
		CreatedAt:        db.TimeFromPg(row.CreatedAt),
		UpdatedAt:        db.TimeFromPg(row.UpdatedAt),
	}, nil
}

func normalizeChannelUserBindingRow(row sqlc.UserChannelBinding) (ChannelUserBinding, error) {
	config, err := DecodeConfigMap(row.Config)
	if err != nil {
		return ChannelUserBinding{}, err
	}
	return ChannelUserBinding{
		ID:          db.UUIDToString(row.ID),
		ChannelType: ChannelType(row.ChannelType),
		UserID:      db.UUIDToString(row.UserID),
		Config:      config,
		CreatedAt:   db.TimeFromPg(row.CreatedAt),
		UpdatedAt:   db.TimeFromPg(row.UpdatedAt),
	}, nil
}

func normalizeChannelSession(row sqlc.ChannelSession) (ChannelSession, error) {
	metadata, err := DecodeConfigMap(row.Metadata)
	if err != nil {
		return ChannelSession{}, err
	}
	return ChannelSession{
		SessionID:       row.SessionID,
		BotID:           db.UUIDToString(row.BotID),
		ChannelConfigID: db.UUIDToString(row.ChannelConfigID),
		UserID:          db.UUIDToString(row.UserID),
		ContactID:       db.UUIDToString(row.ContactID),
		Platform:        row.Platform,
		ReplyTarget:     strings.TrimSpace(row.ReplyTarget.String),
		ThreadID:        strings.TrimSpace(row.ThreadID.String),
		Metadata:        metadata,
		CreatedAt:       db.TimeFromPg(row.CreatedAt),
		UpdatedAt:       db.TimeFromPg(row.UpdatedAt),
	}, nil
}



