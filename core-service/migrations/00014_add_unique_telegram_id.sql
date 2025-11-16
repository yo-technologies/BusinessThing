-- +goose Up
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_users_telegram_id;
CREATE UNIQUE INDEX idx_users_telegram_id_unique ON users(telegram_id)
WHERE telegram_id IS NOT NULL;
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_users_telegram_id_unique;
CREATE INDEX idx_users_telegram_id ON users(telegram_id)
WHERE telegram_id IS NOT NULL;
-- +goose StatementEnd