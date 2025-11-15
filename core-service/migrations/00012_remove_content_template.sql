-- +goose Up
ALTER TABLE contract_templates DROP COLUMN content_template;

-- +goose Down
ALTER TABLE contract_templates ADD COLUMN content_template TEXT NOT NULL DEFAULT '';
