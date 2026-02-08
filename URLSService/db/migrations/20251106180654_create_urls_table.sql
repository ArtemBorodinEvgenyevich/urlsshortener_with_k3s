-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS urls (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    short_code VARCHAR(16) NOT NULL,
    original_url TEXT NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT url_not_empty CHECK ( length(original_url) > 0 ),
    CONSTRAINT valid_expiration CHECK ( expires_at > created_at )

);

CREATE INDEX idx_urls_expires_at ON urls(expires_at);
CREATE INDEX idx_urls_created_at ON urls(created_at);

COMMENT ON TABLE urls IS 'Table to store shortened url data';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS urls;
-- +goose StatementEnd
