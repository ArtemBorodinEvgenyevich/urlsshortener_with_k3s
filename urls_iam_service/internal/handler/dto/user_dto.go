package dto

import "time"

type UserResponse struct {
	ID         string            `json:"id"`
	Provider   string            `json:"provider"`
	ProviderID string            `json:"provider_id"`
	Metadata   map[string]string `json:"metadata"`
	CreatedAt  time.Time         `json:"created_at"`
	LastSeenAt time.Time         `json:"last_seen_at"`
}
