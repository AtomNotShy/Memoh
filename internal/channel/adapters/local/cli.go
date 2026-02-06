package local

import (
	"context"
	"fmt"
	"strings"

	"github.com/memohai/memoh/internal/channel"
)

// CLIAdapter implements channel.Sender for the local CLI channel.
type CLIAdapter struct {
	hub *SessionHub
}

// NewCLIAdapter creates a CLIAdapter backed by the given session hub.
func NewCLIAdapter(hub *SessionHub) *CLIAdapter {
	return &CLIAdapter{hub: hub}
}

// Type returns the CLI channel type.
func (a *CLIAdapter) Type() channel.ChannelType {
	return CLIType
}

// Descriptor returns the CLI channel metadata.
func (a *CLIAdapter) Descriptor() channel.Descriptor {
	return channel.Descriptor{
		Type:        CLIType,
		DisplayName: "CLI",
		Configless:  true,
		Capabilities: channel.ChannelCapabilities{
			Text:        true,
			Reply:       true,
			Attachments: true,
		},
		TargetSpec: channel.TargetSpec{
			Format: "session_id",
			Hints: []channel.TargetHint{
				{Label: "Session ID", Example: "cli:uuid"},
			},
		},
	}
}

// Send publishes an outbound message to the CLI session hub.
func (a *CLIAdapter) Send(ctx context.Context, cfg channel.ChannelConfig, msg channel.OutboundMessage) error {
	if a.hub == nil {
		return fmt.Errorf("cli hub not configured")
	}
	target := strings.TrimSpace(msg.Target)
	if target == "" {
		return fmt.Errorf("cli target is required")
	}
	if msg.Message.IsEmpty() {
		return fmt.Errorf("message is required")
	}
	a.hub.Publish(target, msg)
	return nil
}
