package repository

import (
	"context"
	"time"
	"urls_iam_service/internal/domain"
)

type SessionRepository interface {
	// Create creates a new session
	Create(ctx context.Context, session *domain.Session) error
	// GetByID gets session by sessionID
	GetByID(ctx context.Context, sessionID string) (*domain.Session, error)
	// GetByUserID gets all sessions for a user
	GetByUserID(ctx context.Context, userID string) ([]*domain.Session, error)
	// UpdateLastActivity updates session's last_activity timestamp
	UpdateLastActivity(ctx context.Context, sessionID string, lastActivity time.Time) error
	// UpdateExpiresAt update session's expires_at timestamp (for refresh)
	UpdateExpiresAt(ctx context.Context, sessionID string, expiresAt time.Time) error
	// Delete deletes session by session_id
	Delete(ctx context.Context, sessionID string) error
	// DeleteByUserID deletes all sessions for a user
	DeleteByUserID(ctx context.Context, userID string) error
	// DeleteExpired deletes all expired sessions (for cleanup jobs)
	DeleteExpired(ctx context.Context) (int64, error)
}
