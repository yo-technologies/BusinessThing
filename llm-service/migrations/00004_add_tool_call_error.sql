-- +goose Up
-- Добавляем колонку error в таблицу tool_calls для хранения сообщений об ошибках
ALTER TABLE tool_calls
ADD COLUMN error TEXT;
-- +goose Down
ALTER TABLE tool_calls DROP COLUMN error;