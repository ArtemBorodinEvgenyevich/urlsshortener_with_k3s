package domain

import "time"

type Session struct {
	SessionID    string
	UserID       string
	CreatedAt    time.Time
	ExpiresAt    time.Time
	LastActivity time.Time
}
