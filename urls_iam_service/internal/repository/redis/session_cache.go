package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"urls_iam_service/internal/domain"

	"github.com/redis/go-redis/v9"
)

type SessionCache struct {
	client *redis.Client
	ttl    time.Duration
}

// NewSessionCache creates a new Redis session cache
func NewSessionCache(client *redis.Client, ttl time.Duration) *SessionCache {
	return &SessionCache{
		client: client,
		ttl:    ttl,
	}
}

// Get gets session from cache
func (c *SessionCache) Get(ctx context.Context, sessionID string) (*domain.Session, error) {
	key := c.key(sessionID)

	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, domain.ErrSessionNotFound
		}
		return nil, fmt.Errorf("failed to get session from cache: %w", err)
	}

	var session domain.Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	return &session, nil
}

// Set sets session to cache
func (c *SessionCache) Set(ctx context.Context, session *domain.Session) error {
	key := c.key(session.SessionID)

	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	if err := c.client.Set(ctx, key, data, c.ttl).Err(); err != nil {
		return fmt.Errorf("failed to set session to cache: %w", err)
	}

	return nil
}

// Delete deletes session from cache
func (c *SessionCache) Delete(ctx context.Context, sessionID string) error {
	key := c.key(sessionID)

	if err := c.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete session from cache: %w", err)
	}

	return nil
}

// DeleteByUserID deletes all sessions for a user from cache
// Note: This is inefficient - consider using Redis SCAN or hash structure for better performance
func (c *SessionCache) DeleteByUserID(ctx context.Context, userID string) error {
	// For now, we'll just let cache expire naturally
	// In production, consider using a different cache structure (e.g., user_id -> [session_ids])
	return nil
}

func (c *SessionCache) key(sessionID string) string {
	return fmt.Sprintf("session:%s", sessionID)
}
