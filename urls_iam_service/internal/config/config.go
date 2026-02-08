package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	App      AppConfig
	Server   ServerConfig
	Postgres PostgresConfig
	Redis    RedisConfig
	Session  SessionConfig
	Cookie   CookieConfig
}

type AppConfig struct {
	Env      string
	LogLevel string
}

type ServerConfig struct {
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

type PostgresConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	Database        string
	SSLMode         string
	MaxConns        int
	MinConns        int
	MaxConnLifetime time.Duration
	MaxConnIdleTime time.Duration
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	Database int
	PoolSize int
}

type SessionConfig struct {
	TTL      time.Duration
	IDLength int
	CacheTTL time.Duration
}

type CookieConfig struct {
	Name     string
	Secure   bool
	HttpOnly bool
	SameSite string
	Domain   string
	Path     string
}

func Load() (*Config, error) {
	cfg := &Config{
		App: AppConfig{
			Env:      getEnv("APP_ENV", "development"),
			LogLevel: getEnv("LOG_LEVEL", "info"),
		},
		Server: ServerConfig{
			Port:         getEnvAsInt("HTTP_PORT", 8080),
			ReadTimeout:  getEnvAsDuration("HTTP_READ_TIMEOUT", 10*time.Second),
			WriteTimeout: getEnvAsDuration("HTTP_WRITE_TIMEOUT", 10*time.Second),
			IdleTimeout:  getEnvAsDuration("HTTP_IDLE_TIMEOUT", 60*time.Second),
		},
		Postgres: PostgresConfig{
			Host:            getEnv("POSTGRES_HOST", "localhost"),
			Port:            getEnvAsInt("POSTGRES_PORT", 5433),
			User:            getEnv("POSTGRES_USER", "iam_user"),
			Password:        getEnv("POSTGRES_PASSWORD", "iam_secure_password"),
			Database:        getEnv("POSTGRES_DB", "iam_db"),
			SSLMode:         getEnv("POSTGRES_SSL_MODE", "disable"),
			MaxConns:        getEnvAsInt("POSTGRES_MAX_CONNS", 25),
			MinConns:        getEnvAsInt("POSTGRES_MIN_CONNS", 5),
			MaxConnLifetime: getEnvAsDuration("POSTGRES_MAX_CONN_LIFETIME", 1*time.Hour),
			MaxConnIdleTime: getEnvAsDuration("POSTGRES_MAX_CONN_IDLE_TIME", 30*time.Minute),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnvAsInt("REDIS_PORT", 6380),
			Password: getEnv("REDIS_PASSWORD", "iam_redis_password"),
			Database: getEnvAsInt("REDIS_DB", 0),
			PoolSize: getEnvAsInt("REDIS_POOL_SIZE", 10),
		},
		Session: SessionConfig{
			TTL:      getEnvAsDuration("SESSION_TTL", 720*time.Hour), // 30 days
			IDLength: getEnvAsInt("SESSION_ID_LENGTH", 32),
			CacheTTL: getEnvAsDuration("SESSION_CACHE_TTL", 30*time.Second),
		},
		Cookie: CookieConfig{
			Name:     getEnv("COOKIE_NAME", "session_id"),
			Secure:   getEnvAsBool("COOKIE_SECURE", false),
			HttpOnly: getEnvAsBool("COOKIE_HTTP_ONLY", true),
			SameSite: getEnv("COOKIE_SAME_SITE", "Lax"),
			Domain:   getEnv("COOKIE_DOMAIN", ""),
			Path:     getEnv("COOKIE_PATH", "/"),
		},
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

func (c *Config) validate() error {
	if c.Postgres.User == "" {
		return fmt.Errorf("POSTGRES_USER is required")
	}
	if c.Postgres.Database == "" {
		return fmt.Errorf("POSTGRES_DB is required")
	}
	if c.Session.IDLength < 16 {
		return fmt.Errorf("SESSION_ID_LENGTH must be at least 16")
	}
	return nil
}

func (c *PostgresConfig) DatabaseConnectionString() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Database, c.SSLMode,
	)
}

func (c *RedisConfig) RedisConnectionAddress() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

func getEnv(key, defaultVal string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultVal
}

func getEnvAsInt(key string, defaultVal int) int {
	valueStr := os.Getenv(key)
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultVal
}

func getEnvAsBool(key string, defaultVal bool) bool {
	valueStr := os.Getenv(key)
	if value, err := strconv.ParseBool(valueStr); err == nil {
		return value
	}
	return defaultVal
}

func getEnvAsDuration(key string, defaultVal time.Duration) time.Duration {
	valueStr := os.Getenv(key)
	if value, err := time.ParseDuration(valueStr); err == nil {
		return value
	}
	return defaultVal
}
