package repository

import (
	"context"
	"time"
	"urls_iam_service/internal/domain"
)

type UserRepository interface {
	// Create creates a new user
	Create(ctx context.Context, user *domain.User) error
	// GetByID gets user by ID
	GetByID(ctx context.Context, id string) (*domain.User, error)
	// GetByProvider gets user by provider and provider_id
	GetByProvider(ctx context.Context, provider domain.Provider, providerID string) (*domain.User, error)
	// UpdateLastSeen updates user's last_seen_at timestamp
	UpdateLastSeen(ctx context.Context, id string, lastSeen time.Time) error
	// Delete deletes user
	Delete(ctx context.Context, id string) error
}
