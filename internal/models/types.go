package models

import (
	"errors"

	"github.com/google/uuid"
)

type ModelType string

const (
	ModelTypeChat      ModelType = "chat"
	ModelTypeEmbedding ModelType = "embedding"
)

type EnableAs string

const (
	EnableAsChat      EnableAs = "chat"
	EnableAsMemory    EnableAs = "memory"
	EnableAsEmbedding EnableAs = "embedding"
)

type ClientType string

const (
	ClientTypeOpenAI    ClientType = "openai"
	ClientTypeAnthropic ClientType = "anthropic"
	ClientTypeGoogle    ClientType = "google"
	ClientTypeBedrock   ClientType = "bedrock"
	ClientTypeOllama    ClientType = "ollama"
	ClientTypeAzure     ClientType = "azure"
	ClientTypeDashscope ClientType = "dashscope"
	ClientTypeOther     ClientType = "other"
)

type Model struct {
	ModelID       string    `json:"model_id"`
	Name          string    `json:"name"`
	LlmProviderID string    `json:"llm_provider_id"`
	IsMultimodal  bool      `json:"is_multimodal"`
	Type          ModelType `json:"type"`
	Dimensions    int       `json:"dimensions"`
	EnableAs      *EnableAs `json:"enable_as,omitempty"`
}

func (m *Model) Validate() error {
	if m.ModelID == "" {
		return errors.New("model ID is required")
	}
	if m.LlmProviderID == "" {
		return errors.New("llm provider ID is required")
	}
	if _, err := uuid.Parse(m.LlmProviderID); err != nil {
		return errors.New("llm provider ID must be a valid UUID")
	}
	if m.Type != ModelTypeChat && m.Type != ModelTypeEmbedding {
		return errors.New("invalid model type")
	}
	if m.Type == ModelTypeEmbedding && m.Dimensions <= 0 {
		return errors.New("dimensions must be greater than 0")
	}
	
	// Validate enable_as based on type
	if m.EnableAs != nil {
		switch m.Type {
		case ModelTypeEmbedding:
			if *m.EnableAs != EnableAsEmbedding {
				return errors.New("embedding models can only have enable_as set to 'embedding'")
			}
		case ModelTypeChat:
			if *m.EnableAs != EnableAsChat && *m.EnableAs != EnableAsMemory {
				return errors.New("chat models can only have enable_as set to 'chat' or 'memory'")
			}
		}
	}
	
	return nil
}

type AddRequest Model

type AddResponse struct {
	ID      string `json:"id"`
	ModelID string `json:"model_id"`
}

type GetRequest struct {
	ID string `json:"id"`
}

type GetResponse struct {
	ModelId string `json:"model_id"`
	Model
}

type UpdateRequest Model

type ListRequest struct {
	Type       ModelType  `json:"type,omitempty"`
	ClientType ClientType `json:"client_type,omitempty"`
}

type DeleteRequest struct {
	ID      string `json:"id,omitempty"`
	ModelID string `json:"model_id,omitempty"`
}

type DeleteResponse struct {
	Message string `json:"message"`
}

type CountResponse struct {
	Count int64 `json:"count"`
}
