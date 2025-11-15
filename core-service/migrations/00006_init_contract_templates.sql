-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS contract_templates (
    id UUID PRIMARY KEY,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    template_type VARCHAR(100),
    fields_schema JSONB NOT NULL DEFAULT '{}'::jsonb,
    content_template TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_contract_templates_organization_id ON contract_templates(organization_id);
CREATE INDEX idx_contract_templates_template_type ON contract_templates(template_type);
CREATE INDEX idx_contract_templates_created_at ON contract_templates(created_at DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS contract_templates;
-- +goose StatementEnd
