package config

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DatabaseConfig params for PostgreSQL connection.
type DatabaseConfig struct {
	// Basic connection params
	host     string
	port     uint16
	user     string
	password string
	database string
	sslMode  string

	// Timeout params
	connectTimeout   time.Duration
	queryTimeout     time.Duration
	statementTimeout time.Duration

	// Pool params
	maxConns          int32
	minConns          int32
	maxConnLifetime   time.Duration
	maxConnIdleTime   time.Duration
	healthCheckPeriod time.Duration
}

func (c *DatabaseConfig) QueryTimeout() time.Duration {
	return c.queryTimeout
}

// DatabaseConfigBuilder builds DatabaseConfig with validation on each step.
type DatabaseConfigBuilder struct {
	config DatabaseConfig
	errors []error
}

// NewDatabaseConfigBuilder creates new builder with default values.
func NewDatabaseConfigBuilder() *DatabaseConfigBuilder {
	return &DatabaseConfigBuilder{
		config: DatabaseConfig{
			host:              "localhost",
			port:              5432,
			sslMode:           "disable",
			connectTimeout:    10 * time.Second,
			queryTimeout:      5 * time.Second,
			statementTimeout:  5 * time.Second,
			maxConns:          25,
			minConns:          5,
			maxConnLifetime:   1 * time.Hour,
			maxConnIdleTime:   30 * time.Minute,
			healthCheckPeriod: 1 * time.Minute,
		},
		errors: make([]error, 0),
	}
}

func (b *DatabaseConfigBuilder) WithHost(host string) *DatabaseConfigBuilder {
	if host == "" {
		b.errors = append(b.errors, fmt.Errorf("database host cannot be empty"))
		return b
	}
	b.config.host = host
	return b
}

func (b *DatabaseConfigBuilder) WithPort(port uint16) *DatabaseConfigBuilder {
	if port == 0 {
		b.errors = append(b.errors, fmt.Errorf("database port cannot be zero"))
		return b
	}
	b.config.port = port
	return b
}

func (b *DatabaseConfigBuilder) WithUser(user string) *DatabaseConfigBuilder {
	if user == "" {
		b.errors = append(b.errors, fmt.Errorf("database user cannot be empty"))
		return b
	}
	b.config.user = user
	return b
}

func (b *DatabaseConfigBuilder) WithPassword(password string) *DatabaseConfigBuilder {
	b.config.password = password
	return b
}

func (b *DatabaseConfigBuilder) WithDatabase(database string) *DatabaseConfigBuilder {
	if database == "" {
		b.errors = append(b.errors, fmt.Errorf("database name cannot be empty"))
		return b
	}
	b.config.database = database
	return b
}

func (b *DatabaseConfigBuilder) WithSSLMode(sslMode string) *DatabaseConfigBuilder {
	validModes := map[string]bool{
		"disable":     true,
		"require":     true,
		"verify-ca":   true,
		"verify-full": true,
	}
	if !validModes[sslMode] {
		b.errors = append(
			b.errors,
			fmt.Errorf("invalid SSL mode: %s (valid: disable, require, verify-ca, verify-full)", sslMode),
		)
		return b
	}
	b.config.sslMode = sslMode
	return b
}

// WithConnectTimeout sets timeout for db connection.
func (b *DatabaseConfigBuilder) WithConnectTimeout(timeout time.Duration) *DatabaseConfigBuilder {
	if timeout <= 0 {
		b.errors = append(b.errors, fmt.Errorf("connect timeout must be positive, got %v", timeout))
		return b
	}
	b.config.connectTimeout = timeout
	return b
}

// WithQueryTimeout sets timeout for application requests.
func (b *DatabaseConfigBuilder) WithQueryTimeout(timeout time.Duration) *DatabaseConfigBuilder {
	if timeout <= 0 {
		b.errors = append(b.errors, fmt.Errorf("query timeout must be positive, got %v", timeout))
		return b
	}
	b.config.queryTimeout = timeout
	return b
}

// WithStatementTimeout sets timeout on db statement processing for PostgreSQL.
func (b *DatabaseConfigBuilder) WithStatementTimeout(timeout time.Duration) *DatabaseConfigBuilder {
	if timeout <= 0 {
		b.errors = append(b.errors, fmt.Errorf("statement timeout must be positive, got %v", timeout))
		return b
	}
	b.config.statementTimeout = timeout
	return b
}

// WithMaxConns sets max possible connections in pool.
func (b *DatabaseConfigBuilder) WithMaxConns(maxConns int32) *DatabaseConfigBuilder {
	if maxConns <= 0 {
		b.errors = append(b.errors, fmt.Errorf("max connections must be positive, got %d", maxConns))
		return b
	}
	b.config.maxConns = maxConns
	return b
}

// WithMinConns sets min possible connections in pool.
func (b *DatabaseConfigBuilder) WithMinConns(minConns int32) *DatabaseConfigBuilder {
	if minConns < 0 {
		b.errors = append(b.errors, fmt.Errorf("min connections cannot be negative, got %d", minConns))
		return b
	}
	b.config.minConns = minConns
	return b
}

// WithMaxConnLifetime sets max connection ttl.
func (b *DatabaseConfigBuilder) WithMaxConnLifetime(lifetime time.Duration) *DatabaseConfigBuilder {
	if lifetime <= 0 {
		b.errors = append(b.errors, fmt.Errorf("max connection lifetime must be positive, got %v", lifetime))
		return b
	}
	b.config.maxConnLifetime = lifetime
	return b
}

// WithMaxConnIdleTime sets max connection idle timeout.
func (b *DatabaseConfigBuilder) WithMaxConnIdleTime(idleTime time.Duration) *DatabaseConfigBuilder {
	if idleTime <= 0 {
		b.errors = append(b.errors, fmt.Errorf("max connection idle time must be positive, got %v", idleTime))
		return b
	}
	b.config.maxConnIdleTime = idleTime
	return b
}

// WithHealthCheckPeriod sets connections healthcheck period.
func (b *DatabaseConfigBuilder) WithHealthCheckPeriod(period time.Duration) *DatabaseConfigBuilder {
	if period <= 0 {
		b.errors = append(b.errors, fmt.Errorf("health check period must be positive, got %v", period))
		return b
	}
	b.config.healthCheckPeriod = period
	return b
}

// Build creates DatabaseConfig with checking for errors.
func (b *DatabaseConfigBuilder) Build() (*DatabaseConfig, error) {
	if len(b.errors) > 0 {
		return nil, fmt.Errorf("configuration errors: %v", b.errors)
	}

	if b.config.minConns > b.config.maxConns {
		return nil, fmt.Errorf("min connections (%d) cannot be greater than max connections (%d)",
			b.config.minConns, b.config.maxConns)
	}

	return &b.config, nil
}

// BuildPoolConfig creates pgxpool.Config.
// Uses a ParseConfig (is pgx requirement - manual creation of Config invokes panic).
func (c *DatabaseConfig) BuildPoolConfig() (*pgxpool.Config, error) {
	// Creates base DSN string for ParseConfig
	// pgx requires to use ParseConfig - manual Config creation is prohibited
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.user,
		c.password,
		c.host,
		c.port,
		c.database,
		c.sslMode,
	)

	// Config parse
	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// sets additional params
	poolConfig.ConnConfig.ConnectTimeout = c.connectTimeout

	// seta statement_timeout on PostgreSQL level
	if poolConfig.ConnConfig.RuntimeParams == nil {
		poolConfig.ConnConfig.RuntimeParams = make(map[string]string)
	}
	poolConfig.ConnConfig.RuntimeParams["statement_timeout"] = fmt.Sprintf(
		"%d",
		c.statementTimeout.Milliseconds(),
	)

	poolConfig.MaxConns = c.maxConns
	poolConfig.MinConns = c.minConns
	poolConfig.MaxConnLifetime = c.maxConnLifetime
	poolConfig.MaxConnIdleTime = c.maxConnIdleTime
	poolConfig.HealthCheckPeriod = c.healthCheckPeriod

	return poolConfig, nil
}

// CreatePool creates parametrised pgxpool.Pool.
func (c *DatabaseConfig) CreatePool(ctx context.Context) (*pgxpool.Pool, error) {
	poolConfig, err := c.BuildPoolConfig()
	if err != nil {
		return nil, err
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// checking connection with db
	pingCtx, cancel := context.WithTimeout(ctx, c.connectTimeout)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return pool, nil
}
