// Package channel provides a unified abstraction for multi-platform messaging channels.
// It defines types, interfaces, and a registry for channel adapters such as Telegram and Feishu.
package channel

import (
	"strings"
	"time"
)

// ChannelType identifies a messaging platform (e.g., "telegram", "feishu").
type ChannelType string

// String returns the channel type as a plain string.
func (c ChannelType) String() string {
	return string(c)
}

// Identity represents a sender's identity on a channel.
type Identity struct {
	ExternalID  string
	DisplayName string
	Attributes  map[string]string
}

// Attribute returns the trimmed value for the given key, or empty string if absent.
func (i Identity) Attribute(key string) string {
	if i.Attributes == nil {
		return ""
	}
	return strings.TrimSpace(i.Attributes[key])
}

// Conversation holds metadata about the chat or group context.
type Conversation struct {
	ID       string
	Type     string
	Name     string
	ThreadID string
	Metadata map[string]any
}

// InboundMessage is a message received from an external channel.
type InboundMessage struct {
	Channel      ChannelType
	Message      Message
	BotID        string
	ReplyTarget  string
	SessionKey   string
	Sender       Identity
	Conversation Conversation
	ReceivedAt   time.Time
	Source       string
	Metadata     map[string]any
}

// SessionID returns a stable identifier for the conversation session.
// Format: platform:bot_id:conversation_id[:sender_id].
func (m InboundMessage) SessionID() string {
	if strings.TrimSpace(m.SessionKey) != "" {
		return strings.TrimSpace(m.SessionKey)
	}
	senderID := strings.TrimSpace(m.Sender.ExternalID)
	if senderID == "" {
		senderID = strings.TrimSpace(m.Sender.DisplayName)
	}
	return GenerateSessionID(string(m.Channel), m.BotID, m.Conversation.ID, m.Conversation.Type, senderID)
}

// GenerateSessionID builds a session identifier from platform, bot, conversation, and sender info.
// For group chats, the sender ID is appended to provide per-user context.
func GenerateSessionID(platform, botID, conversationID, conversationType, senderID string) string {
	parts := []string{platform, botID, conversationID}
	ct := strings.ToLower(strings.TrimSpace(conversationType))
	if ct != "" && ct != "p2p" && ct != "private" {
		senderID = strings.TrimSpace(senderID)
		if senderID != "" {
			parts = append(parts, senderID)
		}
	}
	return strings.Join(parts, ":")
}

// OutboundMessage pairs a delivery target with the message content.
type OutboundMessage struct {
	Target  string  `json:"target"`
	Message Message `json:"message"`
}

// MessageFormat indicates how the message text should be rendered.
type MessageFormat string

const (
	MessageFormatPlain    MessageFormat = "plain"
	MessageFormatMarkdown MessageFormat = "markdown"
	MessageFormatRich     MessageFormat = "rich"
)

// MessagePartType identifies the kind of a rich-text message part.
type MessagePartType string

const (
	MessagePartText      MessagePartType = "text"
	MessagePartLink      MessagePartType = "link"
	MessagePartCodeBlock MessagePartType = "code_block"
	MessagePartMention   MessagePartType = "mention"
	MessagePartEmoji     MessagePartType = "emoji"
)

// MessageTextStyle describes inline formatting for a text part.
type MessageTextStyle string

const (
	MessageStyleBold          MessageTextStyle = "bold"
	MessageStyleItalic        MessageTextStyle = "italic"
	MessageStyleStrikethrough MessageTextStyle = "strikethrough"
	MessageStyleCode          MessageTextStyle = "code"
)

// MessagePart is a single element within a rich-text message.
type MessagePart struct {
	Type     MessagePartType    `json:"type"`
	Text     string             `json:"text,omitempty"`
	URL      string             `json:"url,omitempty"`
	Styles   []MessageTextStyle `json:"styles,omitempty"`
	Language string             `json:"language,omitempty"`
	UserID   string             `json:"user_id,omitempty"`
	Emoji    string             `json:"emoji,omitempty"`
	Metadata map[string]any     `json:"metadata,omitempty"`
}

// AttachmentType classifies the kind of binary attachment.
type AttachmentType string

const (
	AttachmentImage AttachmentType = "image"
	AttachmentAudio AttachmentType = "audio"
	AttachmentVideo AttachmentType = "video"
	AttachmentVoice AttachmentType = "voice"
	AttachmentFile  AttachmentType = "file"
	AttachmentGIF   AttachmentType = "gif"
)

// Attachment represents a binary file attached to a message.
type Attachment struct {
	Type         AttachmentType `json:"type"`
	URL          string         `json:"url,omitempty"`
	Name         string         `json:"name,omitempty"`
	Size         int64          `json:"size,omitempty"`
	Mime         string         `json:"mime,omitempty"`
	DurationMs   int64          `json:"duration_ms,omitempty"`
	Width        int            `json:"width,omitempty"`
	Height       int            `json:"height,omitempty"`
	ThumbnailURL string         `json:"thumbnail_url,omitempty"`
	Caption      string         `json:"caption,omitempty"`
	Metadata     map[string]any `json:"metadata,omitempty"`
}

// Action describes an interactive button or link in a message.
type Action struct {
	Type  string `json:"type"`
	Label string `json:"label,omitempty"`
	Value string `json:"value,omitempty"`
	URL   string `json:"url,omitempty"`
}

// ThreadRef references a conversation thread by ID.
type ThreadRef struct {
	ID string `json:"id"`
}

// ReplyRef points to a message being replied to.
type ReplyRef struct {
	Target    string `json:"target,omitempty"`
	MessageID string `json:"message_id,omitempty"`
}

// Message is the unified message structure used across all channels.
type Message struct {
	ID          string         `json:"id,omitempty"`
	Format      MessageFormat  `json:"format,omitempty"`
	Text        string         `json:"text,omitempty"`
	Parts       []MessagePart  `json:"parts,omitempty"`
	Attachments []Attachment   `json:"attachments,omitempty"`
	Actions     []Action       `json:"actions,omitempty"`
	Thread      *ThreadRef     `json:"thread,omitempty"`
	Reply       *ReplyRef      `json:"reply,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

// IsEmpty reports whether the message carries no content.
func (m Message) IsEmpty() bool {
	return strings.TrimSpace(m.Text) == "" &&
		len(m.Parts) == 0 &&
		len(m.Attachments) == 0 &&
		len(m.Actions) == 0
}

// PlainText extracts the plain text representation of the message.
func (m Message) PlainText() string {
	if strings.TrimSpace(m.Text) != "" {
		return strings.TrimSpace(m.Text)
	}
	if len(m.Parts) == 0 {
		return ""
	}
	lines := make([]string, 0, len(m.Parts))
	for _, part := range m.Parts {
		switch part.Type {
		case MessagePartText, MessagePartLink, MessagePartCodeBlock, MessagePartMention, MessagePartEmoji:
			value := strings.TrimSpace(part.Text)
			if value == "" && part.Type == MessagePartLink {
				value = strings.TrimSpace(part.URL)
			}
			if value == "" && part.Type == MessagePartEmoji {
				value = strings.TrimSpace(part.Emoji)
			}
			if value == "" {
				continue
			}
			lines = append(lines, value)
		default:
			continue
		}
	}
	return strings.Join(lines, "\n")
}

// BindingCriteria specifies conditions for matching a user-channel binding.
type BindingCriteria struct {
	ExternalID string
	Attributes map[string]string
}

// Attribute returns the trimmed value for the given key, or empty string if absent.
func (c BindingCriteria) Attribute(key string) string {
	if c.Attributes == nil {
		return ""
	}
	return strings.TrimSpace(c.Attributes[key])
}

// BindingCriteriaFromIdentity creates BindingCriteria from a channel Identity.
func BindingCriteriaFromIdentity(identity Identity) BindingCriteria {
	return BindingCriteria{
		ExternalID: strings.TrimSpace(identity.ExternalID),
		Attributes: identity.Attributes,
	}
}

// ChannelConfig holds the configuration for a bot's channel integration.
type ChannelConfig struct {
	ID               string
	BotID            string
	ChannelType      ChannelType
	Credentials      map[string]any
	ExternalIdentity string
	SelfIdentity     map[string]any
	Routing          map[string]any
	Status           string
	VerifiedAt       time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// ChannelUserBinding represents a user's binding to a specific channel type.
type ChannelUserBinding struct {
	ID          string
	ChannelType ChannelType
	UserID      string
	Config      map[string]any
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// UpsertConfigRequest is the input for creating or updating a channel configuration.
type UpsertConfigRequest struct {
	Credentials      map[string]any `json:"credentials"`
	ExternalIdentity string         `json:"external_identity,omitempty"`
	SelfIdentity     map[string]any `json:"self_identity,omitempty"`
	Routing          map[string]any `json:"routing,omitempty"`
	Status           string         `json:"status,omitempty"`
	VerifiedAt       *time.Time     `json:"verified_at,omitempty"`
}

// UpsertUserConfigRequest is the input for creating or updating a user-channel binding.
type UpsertUserConfigRequest struct {
	Config map[string]any `json:"config"`
}

// ChannelSession tracks an active conversation session on a channel.
type ChannelSession struct {
	SessionID       string
	BotID           string
	ChannelConfigID string
	UserID          string
	ContactID       string
	Platform        string
	ReplyTarget     string
	ThreadID        string
	Metadata        map[string]any
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// SendRequest is the input for sending an outbound message through a channel.
type SendRequest struct {
	Target  string  `json:"target,omitempty"`
	UserID  string  `json:"user_id,omitempty"`
	Message Message `json:"message"`
}
