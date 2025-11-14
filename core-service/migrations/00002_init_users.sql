-- +goose Up
-- +goose StatementBegin
CREATE TYPE user_role AS ENUM ('admin', 'employee');
CREATE TYPE user_status AS ENUM ('pending', 'active', 'inactive');

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    telegram_id VARCHAR(100),
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    role user_role NOT NULL DEFAULT 'employee',
    status user_status NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(organization_id, email)
);

CREATE INDEX idx_users_organization_id ON users(organization_id);
CREATE INDEX idx_users_telegram_id ON users(telegram_id) WHERE telegram_id IS NOT NULL;
CREATE INDEX idx_users_status ON users(status);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users;
DROP TYPE IF EXISTS user_status;
DROP TYPE IF EXISTS user_role;
-- +goose StatementEnd
