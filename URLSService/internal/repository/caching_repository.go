package repository

import (
	"context"
	"errors"
	"time"

	"github.com/ArtemBorodinEvgenyevich/URLSService/internal/cache"
	"github.com/ArtemBorodinEvgenyevich/URLSService/internal/cache/redis"
	"github.com/ArtemBorodinEvgenyevich/URLSService/internal/domain"
	"github.com/ArtemBorodinEvgenyevich/URLSService/internal/logger"
	"go.uber.org/zap"
)

type cachingRepository struct {
	repo  URLRepository
	cache cache.URLCache
}

func NewCachingRepository(repo URLRepository, cache cache.URLCache) URLRepository {
	return &cachingRepository{
		repo:  repo,
		cache: cache,
	}
}

func (r *cachingRepository) Create(ctx context.Context, url *domain.URL) error {
	err := r.repo.Create(ctx, url)
	if err != nil {
		return err
	}

	cacheTTL := time.Until(url.ExpiresAt)
	if cacheTTL > 0 {
		if err = r.cache.Set(ctx, url.ShortCode, url, cacheTTL); err != nil {
			return err
		}
	}

	return nil
}

func (r *cachingRepository) GetByShortCode(ctx context.Context, shortCode string) (*domain.URL, error) {
	url, err := r.cache.Get(ctx, shortCode)
	if err == nil {
		return url, nil
	}
	if errors.Is(err, redis.ErrNegativeCached) {
		logger.RedisLogInfoCtx(ctx, "Key not found, get negative cache")
		return nil, ErrNotFound
	}
	if !errors.Is(err, redis.ErrCacheMiss) {
		logger.RedisLogErrorCtx(ctx, "Cache error:", zap.Error(err))
	}

	url, err = r.repo.GetByShortCode(ctx, shortCode)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			logger.RedisLogInfoCtx(ctx, "Key not found, set negative cache")
			_ = r.cache.SetNegativeCache(ctx, shortCode)
		}
		return nil, err
	}

	cacheTTL := time.Until(url.ExpiresAt)
	if cacheTTL > 0 {
		if err = r.cache.Set(ctx, url.ShortCode, url, cacheTTL); err != nil {
			return nil, err
		}
	}

	return url, nil
}

func (r *cachingRepository) GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*domain.URL, error) {
	return r.repo.GetByUserID(ctx, userID, limit, offset)
}

func (r *cachingRepository) Delete(ctx context.Context, shortCode string) error {
	err := r.repo.Delete(ctx, shortCode)
	if err != nil {
		return err
	}

	if err = r.cache.Delete(ctx, shortCode); err != nil {
		return err
	}

	return nil
}

func (r *cachingRepository) DeleteByShortCodeAndUserID(ctx context.Context, shortCode string, userID string) error {
	err := r.repo.DeleteByShortCodeAndUserID(ctx, shortCode, userID)
	if err != nil {
		return err
	}

	if err = r.cache.Delete(ctx, shortCode); err != nil {
		logger.RedisLogErrorCtx(ctx, "Failed to delete from cache:", zap.Error(err))
	}

	return nil
}
