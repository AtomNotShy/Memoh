package channel

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"
)

// ConfigLister lists channel configs for periodic refresh. Used by connection lifecycle.
type ConfigLister interface {
	ListConfigsByType(ctx context.Context, channelType ChannelType) ([]ChannelConfig, error)
}

// ConfigResolver resolves effective configs and user bindings. Used for outbound sending.
type ConfigResolver interface {
	ResolveEffectiveConfig(ctx context.Context, botID string, channelType ChannelType) (ChannelConfig, error)
	GetUserConfig(ctx context.Context, actorUserID string, channelType ChannelType) (ChannelUserBinding, error)
}

// BindingStore resolves user-channel bindings. Used by identity resolution.
type BindingStore interface {
	ResolveUserBinding(ctx context.Context, channelType ChannelType, criteria BindingCriteria) (string, error)
}

// SessionStore manages channel session lifecycle. Used by identity resolution.
type SessionStore interface {
	GetChannelSession(ctx context.Context, sessionID string) (ChannelSession, error)
	UpsertChannelSession(ctx context.Context, sessionID string, botID string, channelConfigID string, userID string, contactID string, platform string, replyTarget string, threadID string, metadata map[string]any) error
	ListSessionsByBotPlatform(ctx context.Context, botID string, platform string) ([]ChannelSession, error)
}

// ConfigStore is the full persistence interface. Components should depend on smaller
// interfaces above; ConfigStore exists as a convenience for wiring.
type ConfigStore interface {
	ConfigLister
	ConfigResolver
	BindingStore
	SessionStore
	UpsertUserConfig(ctx context.Context, actorUserID string, channelType ChannelType, req UpsertUserConfigRequest) (ChannelUserBinding, error)
}

// Middleware wraps an InboundHandler to add cross-cutting behavior.
type Middleware func(next InboundHandler) InboundHandler

// ManagerStore is the minimal persistence interface required by Manager.
type ManagerStore interface {
	ConfigLister
	ConfigResolver
}

// Manager coordinates channel adapters, connection lifecycle, and message dispatch.
// Connection lifecycle lives in connection.go, inbound dispatch in inbound.go,
// and outbound pipeline in outbound.go.
type Manager struct {
	registry        *Registry
	service         ManagerStore
	processor       InboundProcessor
	refreshInterval time.Duration
	logger          *slog.Logger
	middlewares     []Middleware

	inboundQueue   chan inboundTask
	inboundWorkers int
	inboundOnce    sync.Once
	inboundCtx     context.Context
	inboundCancel  context.CancelFunc
	mu             sync.Mutex
	connections    map[string]*connectionEntry
}

// NewManager creates a Manager with the given logger, registry, config store, and inbound processor.
func NewManager(log *slog.Logger, registry *Registry, service ManagerStore, processor InboundProcessor) *Manager {
	if log == nil {
		log = slog.Default()
	}
	if registry == nil {
		registry = NewRegistry()
	}
	return &Manager{
		registry:        registry,
		service:         service,
		processor:       processor,
		refreshInterval: 30 * time.Second,
		connections:     map[string]*connectionEntry{},
		logger:          log.With(slog.String("component", "channel")),
		middlewares:     []Middleware{},
		inboundQueue:    make(chan inboundTask, 256),
		inboundWorkers:  4,
	}
}

// Registry returns the adapter registry used by this manager.
func (m *Manager) Registry() *Registry {
	return m.registry
}

// Use appends middleware to the inbound processing chain.
func (m *Manager) Use(mw ...Middleware) {
	m.middlewares = append(m.middlewares, mw...)
}

// RegisterAdapter adds an adapter to the registry and logs the registration.
func (m *Manager) RegisterAdapter(adapter Adapter) {
	if adapter == nil {
		return
	}
	if err := m.registry.Register(adapter); err != nil {
		if m.logger != nil {
			m.logger.Warn("adapter registration failed", slog.String("channel", adapter.Type().String()), slog.Any("error", err))
		}
		return
	}
	if m.logger != nil {
		m.logger.Info("adapter registered", slog.String("channel", adapter.Type().String()))
	}
}

// AddAdapter registers an adapter and triggers an immediate refresh for hot-plug support.
func (m *Manager) AddAdapter(ctx context.Context, adapter Adapter) {
	m.RegisterAdapter(adapter)
	if ctx != nil {
		m.refresh(ctx)
	}
}

// RemoveAdapter unregisters an adapter and stops all its active connections.
func (m *Manager) RemoveAdapter(ctx context.Context, channelType ChannelType) {
	if ctx == nil {
		ctx = context.Background()
	}
	m.mu.Lock()
	for id, entry := range m.connections {
		if entry != nil && entry.config.ChannelType == channelType {
			if entry.connection != nil {
				if err := entry.connection.Stop(ctx); err != nil && !errors.Is(err, ErrStopNotSupported) && m.logger != nil {
					m.logger.Warn("adapter stop failed", slog.String("config_id", id), slog.Any("error", err))
				}
			}
			delete(m.connections, id)
		}
	}
	m.mu.Unlock()
	m.registry.Unregister(channelType)
}

// Start begins the periodic config refresh loop and inbound worker pool.
func (m *Manager) Start(ctx context.Context) {
	if m.logger != nil {
		m.logger.Info("manager start")
	}
	m.startInboundWorkers(ctx)
	go func() {
		m.refresh(ctx)
		ticker := time.NewTicker(m.refreshInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				if m.logger != nil {
					m.logger.Info("manager stop")
				}
				m.stopAll(ctx)
				return
			case <-ticker.C:
				m.refresh(ctx)
			}
		}
	}()
}

// Send delivers an outbound message to the specified channel, resolving target and config automatically.
func (m *Manager) Send(ctx context.Context, botID string, channelType ChannelType, req SendRequest) error {
	if m.service == nil {
		return fmt.Errorf("channel manager not configured")
	}
	sender, ok := m.registry.GetSender(channelType)
	if !ok {
		return fmt.Errorf("unsupported channel type: %s", channelType)
	}
	config, err := m.service.ResolveEffectiveConfig(ctx, botID, channelType)
	if err != nil {
		return err
	}
	target := strings.TrimSpace(req.Target)
	if target == "" {
		targetUserID := strings.TrimSpace(req.UserID)
		if targetUserID == "" {
			return fmt.Errorf("target or user_id is required")
		}
		userCfg, err := m.service.GetUserConfig(ctx, targetUserID, channelType)
		if err != nil {
			if m.logger != nil {
				m.logger.Warn("channel binding missing", slog.String("channel", channelType.String()), slog.String("user_id", targetUserID))
			}
			return fmt.Errorf("channel binding required")
		}
		target, err = m.registry.ResolveTargetFromUserConfig(channelType, userCfg.Config)
		if err != nil {
			return err
		}
	}
	if normalized, ok := m.registry.NormalizeTarget(channelType, target); ok {
		target = normalized
	}
	if req.Message.IsEmpty() {
		return fmt.Errorf("message is required")
	}
	if m.logger != nil {
		m.logger.Info("send outbound", slog.String("channel", channelType.String()), slog.String("bot_id", botID))
	}
	policy := m.resolveOutboundPolicy(channelType)
	outbound, err := buildOutboundMessages(OutboundMessage{
		Target:  target,
		Message: req.Message,
	}, policy)
	if err != nil {
		return err
	}
	for _, item := range outbound {
		if err := m.sendWithConfig(ctx, sender, config, item, policy); err != nil {
			if m.logger != nil {
				m.logger.Error("send outbound failed", slog.String("channel", channelType.String()), slog.String("bot_id", botID), slog.Any("error", err))
			}
			return err
		}
	}
	return nil
}

// Shutdown cancels the inbound worker pool and stops all active connections.
func (m *Manager) Shutdown(ctx context.Context) error {
	if m.inboundCancel != nil {
		m.inboundCancel()
	}
	m.stopAll(ctx)
	return nil
}
