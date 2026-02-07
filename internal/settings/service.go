package settings

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/memohai/memoh/internal/db/sqlc"
)

type Service struct {
	queries *sqlc.Queries
	logger  *slog.Logger
}

func NewService(log *slog.Logger, queries *sqlc.Queries) *Service {
	return &Service{
		queries: queries,
		logger:  log.With(slog.String("service", "settings")),
	}
}

func (s *Service) Get(ctx context.Context, userID string) (Settings, error) {
	pgID, err := parseUUID(userID)
	if err != nil {
		return Settings{}, err
	}
	row, err := s.queries.GetSettingsByUserID(ctx, pgID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Settings{
				ChatModelID:        "",
				MemoryModelID:      "",
				EmbeddingModelID:   "",
				MaxContextLoadTime: DefaultMaxContextLoadTime,
				Language:           DefaultLanguage,
			}, nil
		}
		return Settings{}, err
	}
	return normalizeUserSetting(row), nil
}

func (s *Service) Upsert(ctx context.Context, userID string, req UpsertRequest) (Settings, error) {
	if s.queries == nil {
		return Settings{}, fmt.Errorf("settings queries not configured")
	}
	pgID, err := parseUUID(userID)
	if err != nil {
		return Settings{}, err
	}

	current := Settings{
		ChatModelID:        "",
		MemoryModelID:      "",
		EmbeddingModelID:   "",
		MaxContextLoadTime: DefaultMaxContextLoadTime,
		Language:           DefaultLanguage,
	}
	existing, err := s.queries.GetSettingsByUserID(ctx, pgID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return Settings{}, err
	}
	if err == nil {
		current = normalizeUserSetting(existing)
	}

	if value := strings.TrimSpace(req.ChatModelID); value != "" {
		current.ChatModelID = value
	}
	if value := strings.TrimSpace(req.MemoryModelID); value != "" {
		current.MemoryModelID = value
	}
	if value := strings.TrimSpace(req.EmbeddingModelID); value != "" {
		current.EmbeddingModelID = value
	}
	if req.MaxContextLoadTime != nil && *req.MaxContextLoadTime > 0 {
		current.MaxContextLoadTime = *req.MaxContextLoadTime
	}
	if strings.TrimSpace(req.Language) != "" {
		current.Language = strings.TrimSpace(req.Language)
	}

	_, err = s.queries.UpsertUserSettings(ctx, sqlc.UpsertUserSettingsParams{
		UserID:             pgID,
		ChatModelID:        pgtype.Text{String: current.ChatModelID, Valid: current.ChatModelID != ""},
		MemoryModelID:      pgtype.Text{String: current.MemoryModelID, Valid: current.MemoryModelID != ""},
		EmbeddingModelID:   pgtype.Text{String: current.EmbeddingModelID, Valid: current.EmbeddingModelID != ""},
		MaxContextLoadTime: int32(current.MaxContextLoadTime),
		Language:           current.Language,
	})
	if err != nil {
		return Settings{}, err
	}
	return current, nil
}

func (s *Service) GetBot(ctx context.Context, botID string) (Settings, error) {
	pgID, err := parseUUID(botID)
	if err != nil {
		return Settings{}, err
	}
	row, err := s.queries.GetSettingsByBotID(ctx, pgID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			settings := Settings{
				MaxContextLoadTime: DefaultMaxContextLoadTime,
				Language:           DefaultLanguage,
				AllowGuest:         false,
			}
			if err := s.attachBotModelConfig(ctx, pgID, &settings); err != nil {
				return Settings{}, err
			}
			return settings, nil
		}
		return Settings{}, err
	}
	settings := normalizeBotSetting(row)
	if err := s.attachBotModelConfig(ctx, pgID, &settings); err != nil {
		return Settings{}, err
	}
	return settings, nil
}

func (s *Service) UpsertBot(ctx context.Context, botID string, req UpsertRequest) (Settings, error) {
	if s.queries == nil {
		return Settings{}, fmt.Errorf("settings queries not configured")
	}
	pgID, err := parseUUID(botID)
	if err != nil {
		return Settings{}, err
	}

	current := Settings{
		MaxContextLoadTime: DefaultMaxContextLoadTime,
		Language:           DefaultLanguage,
		AllowGuest:         false,
	}
	existing, err := s.queries.GetSettingsByBotID(ctx, pgID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return Settings{}, err
	}
	if err == nil {
		current = normalizeBotSetting(existing)
	}
	if req.MaxContextLoadTime != nil && *req.MaxContextLoadTime > 0 {
		current.MaxContextLoadTime = *req.MaxContextLoadTime
	}
	if strings.TrimSpace(req.Language) != "" {
		current.Language = strings.TrimSpace(req.Language)
	}
	if req.AllowGuest != nil {
		current.AllowGuest = *req.AllowGuest
	}

	_, err = s.queries.UpsertBotSettings(ctx, sqlc.UpsertBotSettingsParams{
		BotID:              pgID,
		MaxContextLoadTime: int32(current.MaxContextLoadTime),
		Language:           current.Language,
		AllowGuest:         current.AllowGuest,
	})
	if err != nil {
		return Settings{}, err
	}
	if err := s.upsertBotModelConfig(ctx, pgID, req); err != nil {
		return Settings{}, err
	}
	if err := s.attachBotModelConfig(ctx, pgID, &current); err != nil {
		return Settings{}, err
	}
	return current, nil
}

func (s *Service) Delete(ctx context.Context, botID string) error {
	if s.queries == nil {
		return fmt.Errorf("settings queries not configured")
	}
	pgID, err := parseUUID(botID)
	if err != nil {
		return err
	}
	return s.queries.DeleteSettingsByBotID(ctx, pgID)
}

func normalizeUserSetting(row sqlc.UserSetting) Settings {
	settings := Settings{
		ChatModelID:        strings.TrimSpace(row.ChatModelID.String),
		MemoryModelID:      strings.TrimSpace(row.MemoryModelID.String),
		EmbeddingModelID:   strings.TrimSpace(row.EmbeddingModelID.String),
		MaxContextLoadTime: int(row.MaxContextLoadTime),
		Language:           strings.TrimSpace(row.Language),
	}
	if settings.MaxContextLoadTime <= 0 {
		settings.MaxContextLoadTime = DefaultMaxContextLoadTime
	}
	if settings.Language == "" {
		settings.Language = DefaultLanguage
	}
	return settings
}

func normalizeBotSetting(row sqlc.BotSetting) Settings {
	settings := Settings{
		MaxContextLoadTime: int(row.MaxContextLoadTime),
		Language:           strings.TrimSpace(row.Language),
		AllowGuest:         row.AllowGuest,
	}
	if settings.MaxContextLoadTime <= 0 {
		settings.MaxContextLoadTime = DefaultMaxContextLoadTime
	}
	if settings.Language == "" {
		settings.Language = DefaultLanguage
	}
	return settings
}

func (s *Service) attachBotModelConfig(ctx context.Context, botID pgtype.UUID, target *Settings) error {
	if s.queries == nil || target == nil {
		return nil
	}
	row, err := s.queries.GetBotModelConfigByBotID(ctx, botID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return err
	}
	target.ChatModelID = strings.TrimSpace(row.ChatModelID.String)
	target.MemoryModelID = strings.TrimSpace(row.MemoryModelID.String)
	target.EmbeddingModelID = strings.TrimSpace(row.EmbeddingModelID.String)
	return nil
}

func (s *Service) upsertBotModelConfig(ctx context.Context, botID pgtype.UUID, req UpsertRequest) error {
	if s.queries == nil {
		return fmt.Errorf("settings queries not configured")
	}
	params := sqlc.UpsertBotModelConfigParams{
		BotID: botID,
	}
	hasUpdate := false
	if value := strings.TrimSpace(req.ChatModelID); value != "" {
		modelID, err := s.resolveModelUUID(ctx, value)
		if err != nil {
			return err
		}
		params.ChatModelID = modelID
		hasUpdate = true
	}
	if value := strings.TrimSpace(req.MemoryModelID); value != "" {
		modelID, err := s.resolveModelUUID(ctx, value)
		if err != nil {
			return err
		}
		params.MemoryModelID = modelID
		hasUpdate = true
	}
	if value := strings.TrimSpace(req.EmbeddingModelID); value != "" {
		modelID, err := s.resolveModelUUID(ctx, value)
		if err != nil {
			return err
		}
		params.EmbeddingModelID = modelID
		hasUpdate = true
	}
	if !hasUpdate {
		return nil
	}
	_, err := s.queries.UpsertBotModelConfig(ctx, params)
	return err
}

func (s *Service) resolveModelUUID(ctx context.Context, modelID string) (pgtype.UUID, error) {
	if strings.TrimSpace(modelID) == "" {
		return pgtype.UUID{}, fmt.Errorf("model_id is required")
	}
	row, err := s.queries.GetModelByModelID(ctx, modelID)
	if err != nil {
		return pgtype.UUID{}, err
	}
	return row.ID, nil
}

func parseUUID(id string) (pgtype.UUID, error) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return pgtype.UUID{}, fmt.Errorf("invalid UUID: %w", err)
	}
	var pgID pgtype.UUID
	pgID.Valid = true
	copy(pgID.Bytes[:], parsed[:])
	return pgID, nil
}
