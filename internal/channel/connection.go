package channel

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
)

type connectionEntry struct {
	config     ChannelConfig
	connection Connection
}

func (m *Manager) refresh(ctx context.Context) {
	if m.service == nil {
		return
	}
	configs := make([]ChannelConfig, 0)
	for _, channelType := range m.registry.Types() {
		items, err := m.service.ListConfigsByType(ctx, channelType)
		if err != nil {
			if m.logger != nil {
				m.logger.Error("list configs failed", slog.String("channel", channelType.String()), slog.Any("error", err))
			}
			continue
		}
		configs = append(configs, items...)
	}
	m.reconcile(ctx, configs)
}

func (m *Manager) reconcile(ctx context.Context, configs []ChannelConfig) {
	active := map[string]ChannelConfig{}
	for _, cfg := range configs {
		if cfg.ID == "" {
			continue
		}
		status := strings.ToLower(strings.TrimSpace(cfg.Status))
		if status != "" && status != "active" && status != "verified" {
			continue
		}
		active[cfg.ID] = cfg
		if err := m.ensureConnection(ctx, cfg); err != nil {
			if m.logger != nil {
				m.logger.Error("adapter start failed", slog.String("channel", cfg.ChannelType.String()), slog.String("config_id", cfg.ID), slog.Any("error", err))
			}
		}
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	for id, entry := range m.connections {
		if _, ok := active[id]; ok {
			continue
		}
		if entry != nil && entry.connection != nil {
			if m.logger != nil {
				m.logger.Info("adapter stop", slog.String("channel", entry.config.ChannelType.String()), slog.String("config_id", id))
			}
			if err := entry.connection.Stop(ctx); err != nil && !errors.Is(err, ErrStopNotSupported) && m.logger != nil {
				m.logger.Warn("adapter stop failed", slog.String("config_id", id), slog.Any("error", err))
			}
		}
		delete(m.connections, id)
	}
}

func (m *Manager) ensureConnection(ctx context.Context, cfg ChannelConfig) error {
	_, ok := m.registry.GetReceiver(cfg.ChannelType)
	if !ok {
		return nil
	}

	m.mu.Lock()
	entry := m.connections[cfg.ID]
	if entry != nil && !entry.config.UpdatedAt.Before(cfg.UpdatedAt) {
		m.mu.Unlock()
		return nil
	}
	if entry != nil {
		m.mu.Unlock()
		if m.logger != nil {
			m.logger.Info("adapter restart", slog.String("channel", cfg.ChannelType.String()), slog.String("config_id", cfg.ID))
		}
		if err := entry.connection.Stop(ctx); err != nil {
			if errors.Is(err, ErrStopNotSupported) {
				if m.logger != nil {
					m.logger.Warn("adapter restart skipped", slog.String("channel", cfg.ChannelType.String()), slog.String("config_id", cfg.ID))
				}
				return nil
			}
			return err
		}
		m.mu.Lock()
		delete(m.connections, cfg.ID)
		m.mu.Unlock()
	} else {
		m.mu.Unlock()
	}

	receiver, ok := m.registry.GetReceiver(cfg.ChannelType)
	if !ok {
		return nil
	}

	if m.logger != nil {
		m.logger.Info("adapter start", slog.String("channel", cfg.ChannelType.String()), slog.String("config_id", cfg.ID))
	}
	handler := m.handleInbound
	for i := len(m.middlewares) - 1; i >= 0; i-- {
		handler = m.middlewares[i](handler)
	}
	conn, err := receiver.Connect(ctx, cfg, handler)
	if err != nil {
		return err
	}
	m.mu.Lock()
	m.connections[cfg.ID] = &connectionEntry{
		config:     cfg,
		connection: conn,
	}
	m.mu.Unlock()
	return nil
}

func (m *Manager) stopAll(ctx context.Context) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for id, entry := range m.connections {
		if entry != nil && entry.connection != nil {
			if m.logger != nil {
				m.logger.Info("adapter stop", slog.String("channel", entry.config.ChannelType.String()), slog.String("config_id", id))
			}
			if err := entry.connection.Stop(ctx); err != nil && !errors.Is(err, ErrStopNotSupported) && m.logger != nil {
				m.logger.Warn("adapter stop failed", slog.String("config_id", id), slog.Any("error", err))
			}
		}
		delete(m.connections, id)
	}
}

// Stop terminates the connection identified by the given config ID.
func (m *Manager) Stop(ctx context.Context, configID string) error {
	configID = strings.TrimSpace(configID)
	if configID == "" {
		return fmt.Errorf("config id is required")
	}
	m.mu.Lock()
	entry := m.connections[configID]
	m.mu.Unlock()
	if entry == nil || entry.connection == nil {
		return nil
	}
	return entry.connection.Stop(ctx)
}

// StopByBot terminates all connections belonging to the given bot.
func (m *Manager) StopByBot(ctx context.Context, botID string) error {
	botID = strings.TrimSpace(botID)
	if botID == "" {
		return fmt.Errorf("bot id is required")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for id, entry := range m.connections {
		if entry != nil && entry.config.BotID == botID {
			if entry.connection != nil {
				_ = entry.connection.Stop(ctx)
			}
			delete(m.connections, id)
		}
	}
	return nil
}
