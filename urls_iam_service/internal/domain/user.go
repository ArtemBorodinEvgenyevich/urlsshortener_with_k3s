package domain

import "time"

type User struct {
	ID         string
	Provider   Provider
	ProviderID string
	Metadata   map[string]string
	CreatedAt  time.Time
	LastSeenAt time.Time
}
