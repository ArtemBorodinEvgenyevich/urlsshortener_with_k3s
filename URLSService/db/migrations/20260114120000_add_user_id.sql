-- +goose Up
-- +goose StatementBegin
ALTER TABLE urls ADD COLUMN user_id VARCHAR(64);
CREATE INDEX idx_urls_user_id ON urls(user_id);

COMMENT ON COLUMN urls.user_id IS 'Owner user ID from IAM service';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_urls_user_id;
ALTER TABLE urls DROP COLUMN IF EXISTS user_id;
-- +goose StatementEnd
