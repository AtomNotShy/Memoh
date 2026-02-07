// Package chat orchestrates conversations with the agent gateway, including
// synchronous and streaming chat, scheduled triggers, history, and memory storage.
package chat

import (
	"encoding/json"
	"strings"
)

// ModelMessage is the canonical message format exchanged with the agent gateway.
// Aligned with Vercel AI SDK ModelMessage structure.
type ModelMessage struct {
	Role       string          `json:"role"`
	Content    json.RawMessage `json:"content,omitempty"`
	ToolCalls  []ToolCall      `json:"tool_calls,omitempty"`
	ToolCallID string          `json:"tool_call_id,omitempty"`
	Name       string          `json:"name,omitempty"`
}

// TextContent extracts the plain text from the message content.
// If content is a string, it returns it directly.
// If content is an array of parts, it joins all text-type parts.
func (m ModelMessage) TextContent() string {
	if len(m.Content) == 0 {
		return ""
	}
	var s string
	if err := json.Unmarshal(m.Content, &s); err == nil {
		return s
	}
	var parts []ContentPart
	if err := json.Unmarshal(m.Content, &parts); err == nil {
		texts := make([]string, 0, len(parts))
		for _, p := range parts {
			if strings.TrimSpace(p.Text) != "" {
				texts = append(texts, p.Text)
			}
		}
		return strings.Join(texts, "\n")
	}
	return ""
}

// ContentParts parses the content as an array of ContentPart.
// Returns nil if the content is a plain string or not parseable.
func (m ModelMessage) ContentParts() []ContentPart {
	if len(m.Content) == 0 {
		return nil
	}
	var parts []ContentPart
	if err := json.Unmarshal(m.Content, &parts); err != nil {
		return nil
	}
	return parts
}

// HasContent reports whether the message carries non-empty content or tool calls.
func (m ModelMessage) HasContent() bool {
	if strings.TrimSpace(m.TextContent()) != "" {
		return true
	}
	if len(m.ContentParts()) > 0 {
		return true
	}
	return len(m.ToolCalls) > 0
}

// NewTextContent creates a json.RawMessage from a plain string.
func NewTextContent(text string) json.RawMessage {
	data, _ := json.Marshal(text)
	return data
}

// ContentPart represents one element of a multi-part message content.
type ContentPart struct {
	Type     string         `json:"type"`
	Text     string         `json:"text,omitempty"`
	URL      string         `json:"url,omitempty"`
	Styles   []string       `json:"styles,omitempty"`
	Language string         `json:"language,omitempty"`
	UserID   string         `json:"user_id,omitempty"`
	Emoji    string         `json:"emoji,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

// HasValue reports whether the content part carries a meaningful value.
func (p ContentPart) HasValue() bool {
	return strings.TrimSpace(p.Text) != "" ||
		strings.TrimSpace(p.URL) != "" ||
		strings.TrimSpace(p.Emoji) != ""
}

// ToolCall represents a function/tool invocation in an assistant message.
type ToolCall struct {
	ID       string           `json:"id,omitempty"`
	Type     string           `json:"type"`
	Function ToolCallFunction `json:"function"`
}

// ToolCallFunction holds the name and serialized arguments of a tool call.
type ToolCallFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// ChatRequest is the input for Chat and StreamChat.
type ChatRequest struct {
	BotID          string `json:"-"`
	SessionID      string `json:"-"`
	Token          string `json:"-"`
	UserID         string `json:"-"`
	ContainerID    string `json:"-"`
	ContactID      string `json:"-"`
	ContactName    string `json:"-"`
	ContactAlias   string `json:"-"`
	ReplyTarget    string `json:"-"`
	SessionToken   string `json:"-"`

	Query              string         `json:"query"`
	Model              string         `json:"model,omitempty"`
	Provider           string         `json:"provider,omitempty"`
	MaxContextLoadTime int            `json:"max_context_load_time,omitempty"`
	Language           string         `json:"language,omitempty"`
	Channels           []string       `json:"channels,omitempty"`
	CurrentChannel     string         `json:"current_channel,omitempty"`
	Messages           []ModelMessage `json:"messages,omitempty"`
	Skills             []string       `json:"skills,omitempty"`
	AllowedActions     []string       `json:"allowed_actions,omitempty"`
}

// ChatResponse is the output of a non-streaming chat call.
type ChatResponse struct {
	Messages []ModelMessage `json:"messages"`
	Skills   []string       `json:"skills,omitempty"`
	Model    string         `json:"model,omitempty"`
	Provider string         `json:"provider,omitempty"`
}

// StreamChunk is a raw JSON chunk from the streaming response.
type StreamChunk = json.RawMessage

// AssistantOutput holds extracted assistant content for downstream consumers.
type AssistantOutput struct {
	Content string
	Parts   []ContentPart
}
