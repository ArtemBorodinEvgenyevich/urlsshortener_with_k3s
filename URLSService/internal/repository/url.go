package repository

import (
	"context"
	"errors"

	"github.com/ArtemBorodinEvgenyevich/URLSService/internal/domain"
)

var (
	ErrNotFound  = errors.New("url not found")
	ErrForbidden = errors.New("access denied")
)

type URLRepository interface {
	Create(ctx context.Context, url *domain.URL) error
	GetByShortCode(ctx context.Context, shortCode string) (*domain.URL, error)
	GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*domain.URL, error)
	Delete(ctx context.Context, shortCode string) error
	DeleteByShortCodeAndUserID(ctx context.Context, shortCode string, userID string) error
}
