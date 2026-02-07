package local

import (
	"context"
	"fmt"
	"strings"

	"github.com/memohai/memoh/internal/channel"
)

// WebAdapter implements channel.Sender for the local Web channel.
type WebAdapter struct {
	hub *SessionHub
}

// NewWebAdapter creates a WebAdapter backed by the given session hub.
func NewWebAdapter(hub *SessionHub) *WebAdapter {
	return &WebAdapter{hub: hub}
}

// Type returns the Web channel type.
func (a *WebAdapter) Type() channel.ChannelType {
	return WebType
}

// Descriptor returns the Web channel metadata.
func (a *WebAdapter) Descriptor() channel.Descriptor {
	return channel.Descriptor{
		Type:        WebType,
		DisplayName: "Web",
		Configless:  true,
		Capabilities: channel.ChannelCapabilities{
			Text:        true,
			Reply:       true,
			Attachments: true,
		},
		TargetSpec: channel.TargetSpec{
			Format: "session_id",
			Hints: []channel.TargetHint{
				{Label: "Session ID", Example: "web:uuid"},
			},
		},
	}
}

// Send publishes an outbound message to the Web session hub.
func (a *WebAdapter) Send(ctx context.Context, cfg channel.ChannelConfig, msg channel.OutboundMessage) error {
	if a.hub == nil {
		return fmt.Errorf("web hub not configured")
	}
	target := strings.TrimSpace(msg.Target)
	if target == "" {
		return fmt.Errorf("web target is required")
	}
	if msg.Message.IsEmpty() {
		return fmt.Errorf("message is required")
	}
	a.hub.Publish(target, msg)
	return nil
}
