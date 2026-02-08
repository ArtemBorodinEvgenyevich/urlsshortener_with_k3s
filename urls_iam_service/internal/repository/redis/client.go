package redis

import (
	"context"
	"fmt"
	"urls_iam_service/internal/config"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func NewClient(ctx context.Context, cfg config.RedisConfig, logger *zap.Logger) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisConnectionAddress(),
		Password: cfg.Password,
		DB:       cfg.Database,
		PoolSize: cfg.PoolSize,
	})

	// Ping to verify connection
	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	logger.Info("Redis client created",
		zap.String("addr", cfg.RedisConnectionAddress()),
		zap.Int("db", cfg.Database),
		zap.Int("pool_size", cfg.PoolSize),
	)

	return client, nil
}
