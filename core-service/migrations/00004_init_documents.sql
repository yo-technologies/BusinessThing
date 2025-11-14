-- +goose Up
-- +goose StatementBegin
CREATE TYPE document_status AS ENUM ('pending', 'processing', 'indexed', 'failed');

CREATE TABLE IF NOT EXISTS documents (
    id UUID PRIMARY KEY,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    s3_key VARCHAR(500) NOT NULL,
    file_type VARCHAR(50),
    file_size BIGINT NOT NULL,
    status document_status NOT NULL DEFAULT 'pending',
    error_message TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_documents_organization_id ON documents(organization_id);
CREATE INDEX idx_documents_status ON documents(status);
CREATE INDEX idx_documents_created_at ON documents(created_at DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS documents;
DROP TYPE IF EXISTS document_status;
-- +goose StatementEnd
