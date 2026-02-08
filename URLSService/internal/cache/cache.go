package cache

import (
	"context"
	"time"

	"github.com/ArtemBorodinEvgenyevich/URLSService/internal/domain"
)

type URLCache interface {
	Get(ctx context.Context, shortCode string) (*domain.URL, error)
	Set(ctx context.Context, shortCode string, url *domain.URL, ttl time.Duration) error
	Delete(ctx context.Context, shortCode string) error
	SetNegativeCache(ctx context.Context, shortcode string) error
}
