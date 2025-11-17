-- +goose Up
-- +goose StatementBegin
ALTER TABLE contract_templates DROP CONSTRAINT IF EXISTS contract_templates_organization_id_fkey;
DROP INDEX IF EXISTS idx_contract_templates_organization_id;
ALTER TABLE contract_templates DROP COLUMN IF EXISTS organization_id;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE contract_templates
    ADD COLUMN IF NOT EXISTS organization_id UUID REFERENCES organizations(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_contract_templates_organization_id ON contract_templates(organization_id);
-- +goose StatementEnd
