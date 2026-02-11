package preauth

import "time"

// Key represents a bot pre-authorization key.
type Key struct {
	ID                        string
	BotID                     string
	Token                     string
	IssuedByChannelIdentityID string
	ExpiresAt                 time.Time
	UsedAt                    time.Time
	CreatedAt                 time.Time
}
