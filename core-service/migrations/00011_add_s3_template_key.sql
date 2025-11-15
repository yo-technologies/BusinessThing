-- +goose Up
-- +goose StatementBegin
ALTER TABLE contract_templates
ADD COLUMN s3_template_key TEXT;
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
ALTER TABLE contract_templates DROP COLUMN s3_template_key;
-- +goose StatementEnd