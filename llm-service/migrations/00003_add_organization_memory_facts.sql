-- +goose Up
CREATE TABLE IF NOT EXISTS llm_organization_memory_facts (
    id UUID PRIMARY KEY,
    organization_id UUID NOT NULL,
    content VARCHAR(500) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_org_content UNIQUE (organization_id, content)
);

CREATE INDEX IF NOT EXISTS idx_llm_org_memory_facts_org ON llm_organization_memory_facts(organization_id, created_at);

-- +goose Down
DROP TABLE IF EXISTS llm_organization_memory_facts;
