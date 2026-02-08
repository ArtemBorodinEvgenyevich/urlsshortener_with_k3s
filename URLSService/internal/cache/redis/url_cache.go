package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/ArtemBorodinEvgenyevich/URLSService/internal/cache"
	"github.com/ArtemBorodinEvgenyevich/URLSService/internal/domain"
	"github.com/redis/go-redis/v9"
)

const (
	URLCacheKey = iota
	NotFoundCacheKey
	MaxCacheKey // Should always be the last one
)

const (
	urlKeyPrefix      = "url:"
	notFoundKeyPrefix = "notfound:"
	defaultTTL        = 1 * time.Hour
	notFoundTTL       = 5 * time.Minute
)

var ErrCacheMiss = errors.New("cache miss")
var ErrNegativeCached = errors.New("url doesn't exist")

type urlCache struct {
	client *redis.Client
}

func NewURLCache(client *redis.Client) cache.URLCache {
	return &urlCache{client: client}
}

func (u *urlCache) Get(ctx context.Context, shortCode string) (*domain.URL, error) {
	keys := createCacheKeys(shortCode)
	keyVals := cacheKeysValues(keys)

	rdbMGetRes, err := u.client.MGet(ctx, keyVals...).Result()
	if err != nil {
		return nil, fmt.Errorf("redis mget error: %w", err)
	}
	if rdbMGetRes[NotFoundCacheKey] != nil {
		return nil, ErrNegativeCached
	}
	if rdbMGetRes[URLCacheKey] == nil {
		return nil, ErrCacheMiss
	}

	data, ok := rdbMGetRes[URLCacheKey].(string)
	if !ok {
		return nil, fmt.Errorf("unexpected cache data type")
	}

	var url domain.URL
	if err = json.Unmarshal([]byte(data), &url); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cached URL: %w", err)
	}

	return &url, nil
}

func (u *urlCache) Set(ctx context.Context, shortCode string, url *domain.URL, ttl time.Duration) error {
	keys := createCacheKeys(shortCode)
	if ttl == 0 {
		ttl = defaultTTL
	}

	data, err := json.Marshal(url)
	if err != nil {
		return fmt.Errorf("failed to marshal cached URL: %w", err)
	}

	pipe := u.client.Pipeline()
	pipe.Del(ctx, keys[NotFoundCacheKey])
	pipe.Set(ctx, keys[URLCacheKey], data, ttl)

	if _, err = pipe.Exec(ctx); err != nil {
		return fmt.Errorf("redis pipeline error: %w", err)
	}

	return nil
}

func (u *urlCache) Delete(ctx context.Context, shortCode string) error {
	key := createCacheKeys(shortCode)
	keyVals := cacheKeysValues(key)

	if err := u.client.Del(ctx, keyVals...).Err(); err != nil {
		return fmt.Errorf("redis delete error: %w", err)
	}

	return nil
}

func (u *urlCache) SetNegativeCache(ctx context.Context, shortCode string) error {
	key := createCacheKeys(shortCode)[NotFoundCacheKey]

	if err := u.client.Set(ctx, key, "", notFoundTTL).Err(); err != nil {
		return fmt.Errorf("redis set negative cache error: %w", err)
	}

	return nil
}

func (u *urlCache) cacheKeyExists(ctx context.Context, key string) (bool, error) {
	exists, err := u.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("redis exists error: %w", err)
	}

	return exists > 0, nil
}

func createCacheKeys(shortCode string) map[int]string {
	keys := map[int]string{
		URLCacheKey:      urlKeyPrefix + shortCode,
		NotFoundCacheKey: notFoundKeyPrefix + shortCode,
	}

	return keys
}

func cacheKeysValues(keys map[int]string) []string {
	values := make([]string, 0, MaxCacheKey)
	for i := 0; i < MaxCacheKey; i++ {
		if key, exists := keys[i]; exists {
			values = append(values, key)
		}
	}

	return values
}
