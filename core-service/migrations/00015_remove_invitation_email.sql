-- +goose Up
-- +goose StatementBegin
-- Remove email column from invitations table
ALTER TABLE invitations DROP COLUMN IF EXISTS email;
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
-- Add email column back to invitations table
ALTER TABLE invitations
ADD COLUMN email VARCHAR(255);
-- +goose StatementEnd