package domain

import "time"

type URL struct {
	ShortCode   string
	OriginalURL string
	UserID      *string
	ExpiresAt   time.Time
	CreatedAt   time.Time
}
