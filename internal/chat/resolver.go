package chat

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/memohai/memoh/internal/db/sqlc"
	"github.com/memohai/memoh/internal/models"
)

const defaultMaxContextMinutes = 24 * 60

type Resolver struct {
	modelsService   *models.Service
	queries         *sqlc.Queries
	gatewayBaseURL  string
	timeout         time.Duration
	httpClient      *http.Client
	streamingClient *http.Client
}

func NewResolver(modelsService *models.Service, queries *sqlc.Queries, gatewayBaseURL string, timeout time.Duration) *Resolver {
	if strings.TrimSpace(gatewayBaseURL) == "" {
		gatewayBaseURL = "http://127.0.0.1:8081"
	}
	gatewayBaseURL = strings.TrimRight(gatewayBaseURL, "/")
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	return &Resolver{
		modelsService:  modelsService,
		queries:        queries,
		gatewayBaseURL: gatewayBaseURL,
		timeout:        timeout,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		streamingClient: &http.Client{},
	}
}

func (r *Resolver) Chat(ctx context.Context, req ChatRequest) (ChatResponse, error) {
	if strings.TrimSpace(req.Query) == "" {
		return ChatResponse{}, fmt.Errorf("query is required")
	}
	if strings.TrimSpace(req.UserID) == "" {
		return ChatResponse{}, fmt.Errorf("user id is required")
	}

	chatModel, provider, err := r.selectChatModel(ctx, req)
	if err != nil {
		return ChatResponse{}, err
	}
	clientType, err := normalizeClientType(provider.ClientType)
	if err != nil {
		return ChatResponse{}, err
	}

	messages, err := r.loadHistoryMessages(ctx, req.UserID, req.MaxContextLoadTime)
	if err != nil {
		return ChatResponse{}, err
	}
	if len(req.Messages) > 0 {
		messages = append(messages, req.Messages...)
	}

	payload := agentGatewayRequest{
		APIKey:             provider.ApiKey,
		BaseURL:            provider.BaseUrl,
		Model:              chatModel.ModelID,
		ClientType:         clientType,
		Locale:             req.Locale,
		Language:           req.Language,
		MaxSteps:           req.MaxSteps,
		MaxContextLoadTime: normalizeMaxContextLoad(req.MaxContextLoadTime),
		Platforms:          req.Platforms,
		CurrentPlatform:    req.CurrentPlatform,
		Messages:           messages,
		Query:              req.Query,
	}

	resp, err := r.postChat(ctx, payload)
	if err != nil {
		return ChatResponse{}, err
	}

	if err := r.storeHistory(ctx, req.UserID, req.Query, resp.Messages); err != nil {
		return ChatResponse{}, err
	}

	return ChatResponse{
		Messages: resp.Messages,
		Model:    chatModel.ModelID,
		Provider: provider.ClientType,
	}, nil
}

func (r *Resolver) StreamChat(ctx context.Context, req ChatRequest) (<-chan StreamChunk, <-chan error) {
	chunkChan := make(chan StreamChunk)
	errChan := make(chan error, 1)

	go func() {
		defer close(chunkChan)
		defer close(errChan)

		if strings.TrimSpace(req.Query) == "" {
			errChan <- fmt.Errorf("query is required")
			return
		}
		if strings.TrimSpace(req.UserID) == "" {
			errChan <- fmt.Errorf("user id is required")
			return
		}

		chatModel, provider, err := r.selectChatModel(ctx, req)
		if err != nil {
			errChan <- err
			return
		}
		clientType, err := normalizeClientType(provider.ClientType)
		if err != nil {
			errChan <- err
			return
		}

		messages, err := r.loadHistoryMessages(ctx, req.UserID, req.MaxContextLoadTime)
		if err != nil {
			errChan <- err
			return
		}
		if len(req.Messages) > 0 {
			messages = append(messages, req.Messages...)
		}

		payload := agentGatewayRequest{
			APIKey:             provider.ApiKey,
			BaseURL:            provider.BaseUrl,
			Model:              chatModel.ModelID,
			ClientType:         clientType,
			Locale:             req.Locale,
			Language:           req.Language,
			MaxSteps:           req.MaxSteps,
			MaxContextLoadTime: normalizeMaxContextLoad(req.MaxContextLoadTime),
			Platforms:          req.Platforms,
			CurrentPlatform:    req.CurrentPlatform,
			Messages:           messages,
			Query:              req.Query,
		}

		if err := r.streamChat(ctx, payload, req.UserID, req.Query, chunkChan); err != nil {
			errChan <- err
			return
		}
	}()

	return chunkChan, errChan
}

type agentGatewayRequest struct {
	APIKey             string           `json:"apiKey"`
	BaseURL            string           `json:"baseUrl"`
	Model              string           `json:"model"`
	ClientType         string           `json:"clientType"`
	Locale             string           `json:"locale,omitempty"`
	Language           string           `json:"language,omitempty"`
	MaxSteps           int              `json:"maxSteps,omitempty"`
	MaxContextLoadTime int              `json:"maxContextLoadTime"`
	Platforms          []string         `json:"platforms,omitempty"`
	CurrentPlatform    string           `json:"currentPlatform,omitempty"`
	Messages           []GatewayMessage `json:"messages"`
	Query              string           `json:"query"`
}

type agentGatewayResponse struct {
	Messages []GatewayMessage `json:"messages"`
}

func (r *Resolver) postChat(ctx context.Context, payload agentGatewayRequest) (agentGatewayResponse, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return agentGatewayResponse{}, err
	}
	url := r.gatewayBaseURL + "/chat"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return agentGatewayResponse{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return agentGatewayResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		payload, _ := io.ReadAll(resp.Body)
		return agentGatewayResponse{}, fmt.Errorf("agent gateway error: %s", strings.TrimSpace(string(payload)))
	}

	var parsed agentGatewayResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return agentGatewayResponse{}, err
	}
	return parsed, nil
}

func (r *Resolver) streamChat(ctx context.Context, payload agentGatewayRequest, userID, query string, chunkChan chan<- StreamChunk) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	url := r.gatewayBaseURL + "/chat/stream"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")

	resp, err := r.streamingClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		payload, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("agent gateway error: %s", strings.TrimSpace(string(payload)))
	}

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 2*1024*1024)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || !strings.HasPrefix(line, "data:") {
			continue
		}
		data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if data == "" || data == "[DONE]" {
			continue
		}
		chunkChan <- StreamChunk([]byte(data))

		var envelope struct {
			Type string          `json:"type"`
			Data json.RawMessage `json:"data"`
		}
		if err := json.Unmarshal([]byte(data), &envelope); err != nil {
			continue
		}
		if envelope.Type != "done" || len(envelope.Data) == 0 {
			continue
		}
		var parsed agentGatewayResponse
		if err := json.Unmarshal(envelope.Data, &parsed); err != nil {
			continue
		}
		if err := r.storeHistory(ctx, userID, query, parsed.Messages); err != nil {
			return err
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

func (r *Resolver) loadHistoryMessages(ctx context.Context, userID string, maxContextLoadTime int) ([]GatewayMessage, error) {
	if r.queries == nil {
		return nil, fmt.Errorf("history queries not configured")
	}
	pgUserID, err := parseUUID(userID)
	if err != nil {
		return nil, err
	}
	from := time.Now().UTC().Add(-time.Duration(normalizeMaxContextLoad(maxContextLoadTime)) * time.Minute)
	rows, err := r.queries.ListHistoryByUserSince(ctx, sqlc.ListHistoryByUserSinceParams{
		User: pgUserID,
		Timestamp: pgtype.Timestamptz{
			Time:  from,
			Valid: true,
		},
	})
	if err != nil {
		return nil, err
	}
	messages := make([]GatewayMessage, 0, len(rows))
	for _, row := range rows {
		var batch []GatewayMessage
		if len(row.Messages) == 0 {
			continue
		}
		if err := json.Unmarshal(row.Messages, &batch); err != nil {
			return nil, err
		}
		messages = append(messages, batch...)
	}
	return messages, nil
}

func (r *Resolver) storeHistory(ctx context.Context, userID, query string, responseMessages []GatewayMessage) error {
	if r.queries == nil {
		return fmt.Errorf("history queries not configured")
	}
	if strings.TrimSpace(userID) == "" {
		return fmt.Errorf("user id is required")
	}
	if strings.TrimSpace(query) == "" && len(responseMessages) == 0 {
		return nil
	}
	userMessage := GatewayMessage{
		"role":    "user",
		"content": query,
	}
	messages := append([]GatewayMessage{userMessage}, responseMessages...)
	payload, err := json.Marshal(messages)
	if err != nil {
		return err
	}
	pgUserID, err := parseUUID(userID)
	if err != nil {
		return err
	}
	_, err = r.queries.CreateHistory(ctx, sqlc.CreateHistoryParams{
		Messages: payload,
		Timestamp: pgtype.Timestamptz{
			Time:  time.Now().UTC(),
			Valid: true,
		},
		User: pgUserID,
	})
	return err
}

func (r *Resolver) selectChatModel(ctx context.Context, req ChatRequest) (models.GetResponse, sqlc.LlmProvider, error) {
	if r.modelsService == nil {
		return models.GetResponse{}, sqlc.LlmProvider{}, fmt.Errorf("models service not configured")
	}
	modelID := strings.TrimSpace(req.Model)
	providerFilter := strings.TrimSpace(req.Provider)

	if modelID != "" && providerFilter == "" {
		model, err := r.modelsService.GetByModelID(ctx, modelID)
		if err != nil {
			return models.GetResponse{}, sqlc.LlmProvider{}, err
		}
		if model.Type != models.ModelTypeChat {
			return models.GetResponse{}, sqlc.LlmProvider{}, fmt.Errorf("model is not a chat model")
		}
		provider, err := models.FetchProviderByID(ctx, r.queries, model.LlmProviderID)
		if err != nil {
			return models.GetResponse{}, sqlc.LlmProvider{}, err
		}
		return model, provider, nil
	}

	if providerFilter == "" && modelID == "" {
		defaultModel, err := r.modelsService.GetByEnableAs(ctx, models.EnableAsChat)
		if err == nil {
			provider, err := models.FetchProviderByID(ctx, r.queries, defaultModel.LlmProviderID)
			if err != nil {
				return models.GetResponse{}, sqlc.LlmProvider{}, err
			}
			return defaultModel, provider, nil
		}
	}

	var candidates []models.GetResponse
	var err error
	if providerFilter != "" {
		candidates, err = r.modelsService.ListByClientType(ctx, models.ClientType(providerFilter))
	} else {
		candidates, err = r.modelsService.ListByType(ctx, models.ModelTypeChat)
	}
	if err != nil {
		return models.GetResponse{}, sqlc.LlmProvider{}, err
	}

	filtered := make([]models.GetResponse, 0, len(candidates))
	for _, model := range candidates {
		if model.Type != models.ModelTypeChat {
			continue
		}
		filtered = append(filtered, model)
	}
	if len(filtered) == 0 {
		return models.GetResponse{}, sqlc.LlmProvider{}, fmt.Errorf("no chat models available")
	}

	if modelID != "" {
		for _, model := range filtered {
			if model.ModelID == modelID {
				provider, err := models.FetchProviderByID(ctx, r.queries, model.LlmProviderID)
				if err != nil {
					return models.GetResponse{}, sqlc.LlmProvider{}, err
				}
				return model, provider, nil
			}
		}
		return models.GetResponse{}, sqlc.LlmProvider{}, fmt.Errorf("chat model not found")
	}

	selected := filtered[0]
	provider, err := models.FetchProviderByID(ctx, r.queries, selected.LlmProviderID)
	if err != nil {
		return models.GetResponse{}, sqlc.LlmProvider{}, err
	}
	return selected, provider, nil
}

func normalizeMaxContextLoad(value int) int {
	if value <= 0 {
		return defaultMaxContextMinutes
	}
	return value
}

func normalizeClientType(clientType string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(clientType)) {
	case "openai":
		return "openai", nil
	case "openai-compat":
		return "openai", nil
	case "anthropic":
		return "anthropic", nil
	case "google":
		return "google", nil
	default:
		return "", fmt.Errorf("unsupported agent gateway client type: %s", clientType)
	}
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

