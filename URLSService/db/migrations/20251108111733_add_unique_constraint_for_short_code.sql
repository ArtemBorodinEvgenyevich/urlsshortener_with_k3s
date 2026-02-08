-- +goose Up
-- +goose StatementBegin
CREATE UNIQUE INDEX idx_url_short_code on urls(short_code);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_url_short_code;
-- +goose StatementEnd
