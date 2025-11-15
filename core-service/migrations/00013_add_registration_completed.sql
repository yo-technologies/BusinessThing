-- +goose Up
-- +goose StatementBegin
ALTER TABLE users
ADD COLUMN registration_completed BOOLEAN NOT NULL DEFAULT false;
-- Для существующих пользователей с заполненными first_name и last_name устанавливаем registration_completed = true
UPDATE users
SET registration_completed = true
WHERE first_name IS NOT NULL
    AND first_name != ''
    AND last_name IS NOT NULL
    AND last_name != '';
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
ALTER TABLE users DROP COLUMN IF EXISTS registration_completed;
-- +goose StatementEnd