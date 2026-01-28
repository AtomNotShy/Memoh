package chat

import "encoding/json"

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type GatewayMessage map[string]interface{}

type ChatRequest struct {
	UserID             string           `json:"-"`
	Query              string           `json:"query"`
	Model              string           `json:"model,omitempty"`
	Provider           string           `json:"provider,omitempty"`
	MaxContextLoadTime int              `json:"max_context_load_time,omitempty"`
	Locale             string           `json:"locale,omitempty"`
	Language           string           `json:"language,omitempty"`
	MaxSteps           int              `json:"max_steps,omitempty"`
	Platforms          []string         `json:"platforms,omitempty"`
	CurrentPlatform    string           `json:"current_platform,omitempty"`
	Messages           []GatewayMessage `json:"messages,omitempty"`
}

type ChatResponse struct {
	Messages []GatewayMessage `json:"messages"`
	Model    string           `json:"model,omitempty"`
	Provider string           `json:"provider,omitempty"`
}

type StreamChunk = json.RawMessage
