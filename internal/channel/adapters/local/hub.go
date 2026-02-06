package local

import (
	"sync"

	"github.com/google/uuid"

	"github.com/memohai/memoh/internal/channel"
)

// SessionHub is a pub/sub hub that routes outbound messages to CLI/Web session subscribers.
type SessionHub struct {
	mu       sync.RWMutex
	sessions map[string]map[string]chan channel.OutboundMessage
}

// NewSessionHub creates an empty SessionHub.
func NewSessionHub() *SessionHub {
	return &SessionHub{
		sessions: map[string]map[string]chan channel.OutboundMessage{},
	}
}

// Subscribe registers a new stream for the given session and returns a stream ID,
// a read-only channel for messages, and a cancel function to unsubscribe.
func (h *SessionHub) Subscribe(sessionID string) (string, <-chan channel.OutboundMessage, func()) {
	streamID := uuid.NewString()
	ch := make(chan channel.OutboundMessage, 32)

	h.mu.Lock()
	streams, ok := h.sessions[sessionID]
	if !ok {
		streams = map[string]chan channel.OutboundMessage{}
		h.sessions[sessionID] = streams
	}
	streams[streamID] = ch
	h.mu.Unlock()

	cancel := func() {
		h.mu.Lock()
		streams := h.sessions[sessionID]
		if streams != nil {
			if current, ok := streams[streamID]; ok {
				delete(streams, streamID)
				close(current)
			}
			if len(streams) == 0 {
				delete(h.sessions, sessionID)
			}
		}
		h.mu.Unlock()
	}

	return streamID, ch, cancel
}

// Publish delivers a message to all subscribers of the given session.
// Slow receivers are silently dropped.
func (h *SessionHub) Publish(sessionID string, msg channel.OutboundMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, ch := range h.sessions[sessionID] {
		select {
		case ch <- msg:
		default:
			// Drop if receiver is slow.
		}
	}
}
