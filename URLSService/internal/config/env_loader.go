package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

func LoadDatabaseConfigFromEnv() (*DatabaseConfig, error) {
	builder := NewDatabaseConfigBuilder()

	if host := os.Getenv("POSTGRES_HOST"); host != "" {
		builder.WithHost(host)
	}

	if portStr := os.Getenv("POSTGRES_PORT"); portStr != "" {
		port, err := strconv.ParseUint(portStr, 10, 16)
		if err != nil {
			return nil, fmt.Errorf("invalid POSTGRES_PORT: %w", err)
		}
		builder.WithPort(uint16(port))
	}

	if user := os.Getenv("POSTGRES_USER"); user != "" {
		builder.WithUser(user)
	}

	if password := os.Getenv("POSTGRES_PASSWORD"); password != "" {
		builder.WithPassword(password)
	}

	if database := os.Getenv("POSTGRES_DB"); database != "" {
		builder.WithDatabase(database)
	}

	if sslMode := os.Getenv("POSTGRES_SSL_MODE"); sslMode != "" {
		builder.WithSSLMode(sslMode)
	}

	if connectTimeoutStr := os.Getenv("DB_CONNECT_TIMEOUT"); connectTimeoutStr != "" {
		timeout, err := parseDuration(connectTimeoutStr)
		if err != nil {
			return nil, fmt.Errorf("invalid DB_CONNECT_TIMEOUT: %w", err)
		}
		builder.WithConnectTimeout(timeout)
	}

	if queryTimeoutStr := os.Getenv("DB_QUERY_TIMEOUT"); queryTimeoutStr != "" {
		timeout, err := parseDuration(queryTimeoutStr)
		if err != nil {
			return nil, fmt.Errorf("invalid DB_QUERY_TIMEOUT: %w", err)
		}
		builder.WithQueryTimeout(timeout)
	}

	if statementTimeoutStr := os.Getenv("DB_STATEMENT_TIMEOUT"); statementTimeoutStr != "" {
		timeout, err := parseDuration(statementTimeoutStr)
		if err != nil {
			return nil, fmt.Errorf("invalid DB_STATEMENT_TIMEOUT: %w", err)
		}
		builder.WithStatementTimeout(timeout)
	}

	if maxConnsStr := os.Getenv("DB_MAX_CONNS"); maxConnsStr != "" {
		maxConns, err := strconv.ParseInt(maxConnsStr, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid DB_MAX_CONNS: %w", err)
		}
		builder.WithMaxConns(int32(maxConns))
	}

	if minConnsStr := os.Getenv("DB_MIN_CONNS"); minConnsStr != "" {
		minConns, err := strconv.ParseInt(minConnsStr, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid DB_MIN_CONNS: %w", err)
		}
		builder.WithMinConns(int32(minConns))
	}

	if maxLifetimeStr := os.Getenv("DB_MAX_CONN_LIFETIME"); maxLifetimeStr != "" {
		lifetime, err := parseDuration(maxLifetimeStr)
		if err != nil {
			return nil, fmt.Errorf("invalid DB_MAX_CONN_LIFETIME: %w", err)
		}
		builder.WithMaxConnLifetime(lifetime)
	}

	if maxIdleTimeStr := os.Getenv("DB_MAX_CONN_IDLE_TIME"); maxIdleTimeStr != "" {
		idleTime, err := parseDuration(maxIdleTimeStr)
		if err != nil {
			return nil, fmt.Errorf("invalid DB_MAX_CONN_IDLE_TIME: %w", err)
		}
		builder.WithMaxConnIdleTime(idleTime)
	}

	if healthCheckPeriodStr := os.Getenv("DB_HEALTH_CHECK_PERIOD"); healthCheckPeriodStr != "" {
		period, err := parseDuration(healthCheckPeriodStr)
		if err != nil {
			return nil, fmt.Errorf("invalid DB_HEALTH_CHECK_PERIOD: %w", err)
		}
		builder.WithHealthCheckPeriod(period)
	}

	return builder.Build()
}

func LoadRedisConfigFromEnv() (*RedisConfig, error) {
	builder := NewRedisConfigBuilder()

	if host := os.Getenv("REDIS_HOST"); host != "" {
		builder.WithHost(host)
	}

	if port := os.Getenv("REDIS_PORT"); port != "" {
		builder.WithPort(port)
	}

	if password := os.Getenv("REDIS_PASSWORD"); password != "" {
		builder.WithPassword(password)
	}

	if dbStr := os.Getenv("REDIS_DB"); dbStr != "" {
		db, err := strconv.Atoi(dbStr)
		if err != nil {
			return nil, fmt.Errorf("invalid REDIS_DB: %w", err)
		}
		builder.WithDB(db)
	}

	if dialTimeoutStr := os.Getenv("REDIS_DIAL_TIMEOUT"); dialTimeoutStr != "" {
		timeout, err := parseDuration(dialTimeoutStr)
		if err != nil {
			return nil, fmt.Errorf("invalid REDIS_DIAL_TIMEOUT: %w", err)
		}
		builder.WithDialTimeout(timeout)
	}

	if readTimeoutStr := os.Getenv("REDIS_READ_TIMEOUT"); readTimeoutStr != "" {
		timeout, err := parseDuration(readTimeoutStr)
		if err != nil {
			return nil, fmt.Errorf("invalid REDIS_READ_TIMEOUT: %w", err)
		}
		builder.WithReadTimeout(timeout)
	}

	if writeTimeoutStr := os.Getenv("REDIS_WRITE_TIMEOUT"); writeTimeoutStr != "" {
		timeout, err := parseDuration(writeTimeoutStr)
		if err != nil {
			return nil, fmt.Errorf("invalid REDIS_WRITE_TIMEOUT: %w", err)
		}
		builder.WithWriteTimeout(timeout)
	}

	if poolSizeStr := os.Getenv("REDIS_POOL_SIZE"); poolSizeStr != "" {
		poolSize, err := strconv.Atoi(poolSizeStr)
		if err != nil {
			return nil, fmt.Errorf("invalid REDIS_POOL_SIZE: %w", err)
		}
		builder.WithPoolSize(poolSize)
	}

	if minIdleConnsStr := os.Getenv("REDIS_MIN_IDLE_CONNS"); minIdleConnsStr != "" {
		minIdleConns, err := strconv.Atoi(minIdleConnsStr)
		if err != nil {
			return nil, fmt.Errorf("invalid REDIS_MIN_IDLE_CONNS: %w", err)
		}
		builder.WithMinIdleConns(minIdleConns)
	}

	if maxRetriesStr := os.Getenv("REDIS_MAX_RETRIES"); maxRetriesStr != "" {
		maxRetries, err := strconv.Atoi(maxRetriesStr)
		if err != nil {
			return nil, fmt.Errorf("invalid REDIS_MAX_RETRIES: %w", err)
		}
		builder.WithMaxRetries(maxRetries)
	}

	return builder.Build()
}

// parseDuration parses duration, uses seconds as default.
// Ex: "5s", "10", "1m", "500ms"
func parseDuration(s string) (time.Duration, error) {
	// trying to parse as duration
	d, err := time.ParseDuration(s)
	if err == nil {
		return d, nil
	}

	// if cant parse, trying to use as seconds
	seconds, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid duration format: %s (expected format: '5s', '1m', or number of seconds)", s)
	}

	return time.Duration(seconds) * time.Second, nil
}
