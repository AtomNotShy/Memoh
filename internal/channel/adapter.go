package channel

import (
	"context"
	"errors"
	"sync/atomic"
)

// ErrStopNotSupported is returned when a connection does not support graceful shutdown.
var ErrStopNotSupported = errors.New("channel connection stop not supported")

// InboundHandler is a callback invoked when a message arrives from a channel.
type InboundHandler func(ctx context.Context, cfg ChannelConfig, msg InboundMessage) error

// ReplySender sends an outbound reply within the scope of a single inbound message.
type ReplySender interface {
	Send(ctx context.Context, msg OutboundMessage) error
}

// Adapter is the base interface every channel adapter must implement.
type Adapter interface {
	Type() ChannelType
	Descriptor() Descriptor
}

// Descriptor holds read-only metadata for a registered channel type.
// It contains no behavior — all behavior is expressed through optional interfaces.
type Descriptor struct {
	Type             ChannelType
	DisplayName      string
	Configless       bool
	Capabilities     ChannelCapabilities
	OutboundPolicy   OutboundPolicy
	ConfigSchema     ConfigSchema
	UserConfigSchema ConfigSchema
	TargetSpec       TargetSpec
}

// ConfigNormalizer validates and normalizes channel and user-binding configurations.
type ConfigNormalizer interface {
	NormalizeConfig(raw map[string]any) (map[string]any, error)
	NormalizeUserConfig(raw map[string]any) (map[string]any, error)
}

// TargetResolver handles delivery target normalization and resolution from user bindings.
type TargetResolver interface {
	NormalizeTarget(raw string) string
	ResolveTarget(userConfig map[string]any) (string, error)
}

// BindingMatcher matches user-channel bindings and constructs binding configs from identities.
type BindingMatcher interface {
	MatchBinding(config map[string]any, criteria BindingCriteria) bool
	BuildUserConfig(identity Identity) map[string]any
}

// Sender is an adapter capable of sending outbound messages.
type Sender interface {
	Send(ctx context.Context, cfg ChannelConfig, msg OutboundMessage) error
}

// Receiver is an adapter capable of establishing a long-lived connection to receive messages.
type Receiver interface {
	Connect(ctx context.Context, cfg ChannelConfig, handler InboundHandler) (Connection, error)
}

// Connection represents an active, long-lived link to a channel platform.
type Connection interface {
	ConfigID() string
	BotID() string
	ChannelType() ChannelType
	Stop(ctx context.Context) error
	Running() bool
}

// BaseConnection is a default Connection implementation backed by a stop function.
type BaseConnection struct {
	configID    string
	botID       string
	channelType ChannelType
	stop        func(ctx context.Context) error
	running     atomic.Bool
}

// NewConnection creates a BaseConnection for the given config and stop function.
func NewConnection(cfg ChannelConfig, stop func(ctx context.Context) error) *BaseConnection {
	conn := &BaseConnection{
		configID:    cfg.ID,
		botID:       cfg.BotID,
		channelType: cfg.ChannelType,
		stop:        stop,
	}
	conn.running.Store(true)
	return conn
}

// ConfigID returns the channel configuration identifier.
func (c *BaseConnection) ConfigID() string {
	return c.configID
}

// BotID returns the bot identifier that owns this connection.
func (c *BaseConnection) BotID() string {
	return c.botID
}

// ChannelType returns the type of channel this connection serves.
func (c *BaseConnection) ChannelType() ChannelType {
	return c.channelType
}

// Stop gracefully shuts down the connection.
func (c *BaseConnection) Stop(ctx context.Context) error {
	if c.stop == nil {
		return ErrStopNotSupported
	}
	c.running.Store(false)
	return c.stop(ctx)
}

// Running reports whether the connection is still active.
func (c *BaseConnection) Running() bool {
	return c.running.Load()
}
