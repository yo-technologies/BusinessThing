-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS generated_contracts (
    id UUID PRIMARY KEY,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    template_id UUID NOT NULL REFERENCES contract_templates(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    filled_data JSONB NOT NULL DEFAULT '{}'::jsonb,
    s3_key VARCHAR(500) NOT NULL,
    file_type VARCHAR(50),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_generated_contracts_organization_id ON generated_contracts(organization_id);
CREATE INDEX idx_generated_contracts_template_id ON generated_contracts(template_id);
CREATE INDEX idx_generated_contracts_created_at ON generated_contracts(created_at DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS generated_contracts;
-- +goose StatementEnd
