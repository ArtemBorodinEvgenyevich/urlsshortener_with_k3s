package config

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisConfig contains all parameters for Redis connection.
type RedisConfig struct {
	// Basic connection params
	host     string
	port     string
	password string
	db       int

	// Timeout params
	dialTimeout  time.Duration
	readTimeout  time.Duration
	writeTimeout time.Duration

	// Pool params
	poolSize     int
	minIdleConns int
	maxRetries   int
}

// RedisConfigBuilder builds RedisConfig with validation on each step.
type RedisConfigBuilder struct {
	config RedisConfig
	errors []error
}

// NewRedisConfigBuilder creates new builder with default values.
func NewRedisConfigBuilder() *RedisConfigBuilder {
	return &RedisConfigBuilder{
		config: RedisConfig{
			host:         "localhost",
			port:         "6379",
			password:     "",
			db:           0,
			dialTimeout:  5 * time.Second,
			readTimeout:  3 * time.Second,
			writeTimeout: 3 * time.Second,
			poolSize:     10,
			minIdleConns: 5,
			maxRetries:   3,
		},
		errors: make([]error, 0),
	}
}

// WithHost sets Redis host.
func (b *RedisConfigBuilder) WithHost(host string) *RedisConfigBuilder {
	if host == "" {
		b.errors = append(b.errors, fmt.Errorf("redis host cannot be empty"))
		return b
	}
	b.config.host = host
	return b
}

// WithPort sets Redis port.
func (b *RedisConfigBuilder) WithPort(port string) *RedisConfigBuilder {
	if port == "" {
		b.errors = append(b.errors, fmt.Errorf("redis port cannot be empty"))
		return b
	}
	b.config.port = port
	return b
}

// WithPassword sets Redis password.
func (b *RedisConfigBuilder) WithPassword(password string) *RedisConfigBuilder {
	b.config.password = password
	return b
}

// WithDB sets Redis database number.
func (b *RedisConfigBuilder) WithDB(db int) *RedisConfigBuilder {
	if db < 0 {
		b.errors = append(b.errors, fmt.Errorf("redis database number cannot be negative, got %d", db))
		return b
	}
	b.config.db = db
	return b
}

// WithDialTimeout sets timeout for establishing connection.
func (b *RedisConfigBuilder) WithDialTimeout(timeout time.Duration) *RedisConfigBuilder {
	if timeout <= 0 {
		b.errors = append(b.errors, fmt.Errorf("dial timeout must be positive, got %v", timeout))
		return b
	}
	b.config.dialTimeout = timeout
	return b
}

// WithReadTimeout sets timeout for read operations.
func (b *RedisConfigBuilder) WithReadTimeout(timeout time.Duration) *RedisConfigBuilder {
	if timeout <= 0 {
		b.errors = append(b.errors, fmt.Errorf("read timeout must be positive, got %v", timeout))
		return b
	}
	b.config.readTimeout = timeout
	return b
}

// WithWriteTimeout sets timeout for write operations.
func (b *RedisConfigBuilder) WithWriteTimeout(timeout time.Duration) *RedisConfigBuilder {
	if timeout <= 0 {
		b.errors = append(b.errors, fmt.Errorf("write timeout must be positive, got %v", timeout))
		return b
	}
	b.config.writeTimeout = timeout
	return b
}

// WithPoolSize sets maximum number of connections.
func (b *RedisConfigBuilder) WithPoolSize(poolSize int) *RedisConfigBuilder {
	if poolSize <= 0 {
		b.errors = append(b.errors, fmt.Errorf("pool size must be positive, got %d", poolSize))
		return b
	}
	b.config.poolSize = poolSize
	return b
}

// WithMinIdleConns sets minimum number of idle connections.
func (b *RedisConfigBuilder) WithMinIdleConns(minIdleConns int) *RedisConfigBuilder {
	if minIdleConns < 0 {
		b.errors = append(b.errors, fmt.Errorf("min idle connections cannot be negative, got %d", minIdleConns))
		return b
	}
	b.config.minIdleConns = minIdleConns
	return b
}

// WithMaxRetries sets maximum number of retries for failed commands.
func (b *RedisConfigBuilder) WithMaxRetries(maxRetries int) *RedisConfigBuilder {
	if maxRetries < 0 {
		b.errors = append(b.errors, fmt.Errorf("max retries cannot be negative, got %d", maxRetries))
		return b
	}
	b.config.maxRetries = maxRetries
	return b
}

// Build creates RedisConfig with checking for errors.
func (b *RedisConfigBuilder) Build() (*RedisConfig, error) {
	if len(b.errors) > 0 {
		return nil, fmt.Errorf("configuration errors: %v", b.errors)
	}

	// Complex validation rules
	if b.config.minIdleConns > b.config.poolSize {
		return nil, fmt.Errorf("min idle connections (%d) cannot be greater than pool size (%d)",
			b.config.minIdleConns, b.config.poolSize)
	}

	return &b.config, nil
}

// CreateClient creates configured redis.Client.
func (c *RedisConfig) CreateClient(ctx context.Context) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", c.host, c.port),
		Password:     c.password,
		DB:           c.db,
		DialTimeout:  c.dialTimeout,
		ReadTimeout:  c.readTimeout,
		WriteTimeout: c.writeTimeout,
		PoolSize:     c.poolSize,
		MinIdleConns: c.minIdleConns,
		MaxRetries:   c.maxRetries,
	})

	// Verify connection
	pingCtx, cancel := context.WithTimeout(ctx, c.dialTimeout)
	defer cancel()

	if err := client.Ping(pingCtx).Err(); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	return client, nil
}
