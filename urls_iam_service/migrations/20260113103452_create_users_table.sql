-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider VARCHAR(50) NOT NULL,
    provider_id VARCHAR(255) NOT NULL,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    last_seen_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT users_provider_provider_id_unique UNIQUE (provider, provider_id)
);

CREATE INDEX idx_users_provider ON users(provider, provider_id);
CREATE INDEX idx_users_last_seen_at ON users(last_seen_at);
CREATE INDEX idx_users_created_at ON users(created_at);

COMMENT ON TABLE users IS 'User accounts table supporting multiple auth providers';
COMMENT ON COLUMN users.provider IS 'Auth provider type: anonymous, google, github, etc';
COMMENT ON COLUMN users.provider_id IS 'User ID in the provider system';
COMMENT ON COLUMN users.metadata IS 'Additional user metadata (IP, user-agent, etc)';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_users_created_at;
DROP INDEX IF EXISTS idx_users_last_seen_at;
DROP INDEX IF EXISTS idx_users_provider;
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
