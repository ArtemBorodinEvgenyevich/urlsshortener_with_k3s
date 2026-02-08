package dto

import "time"

type CreateSessionRequest struct {
	Metadata map[string]string `json:"metadata"`
}

// CreateSessionResponse - response при создании сессии
type CreateSessionResponse struct {
	UserID    string    `json:"user_id"`
	SessionID string    `json:"session_id"`
	IsNewUser bool      `json:"is_new_user"`
	ExpiresAt time.Time `json:"expires_at"`
	Provider  string    `json:"provider"`
}

// RefreshSessionResponse - response при обновлении сессии
type RefreshSessionResponse struct {
	ExpiresAt time.Time `json:"expires_at"`
}

// ValidateResponse - response для ForwardAuth (не используется в JSON, только headers)
type ValidateResponse struct {
	UserID   string `json:"user_id"`
	Provider string `json:"provider"`
}
