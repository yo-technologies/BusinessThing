-- +goose Up
-- +goose StatementBegin
ALTER TABLE documents
ALTER COLUMN file_type TYPE VARCHAR(100);
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
ALTER TABLE documents
ALTER COLUMN file_type TYPE VARCHAR(50);
-- +goose StatementEnd