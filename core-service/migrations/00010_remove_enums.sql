-- +goose Up
-- +goose StatementBegin

-- Изменяем колонки с енумов на TEXT в organization_members
ALTER TABLE organization_members 
    ALTER COLUMN role TYPE TEXT,
    ALTER COLUMN status TYPE TEXT;

-- Изменяем колонки с енумов на TEXT в invitations
ALTER TABLE invitations 
    ALTER COLUMN role TYPE TEXT;

-- Изменяем колонки с енумов на TEXT в documents
ALTER TABLE documents 
    ALTER COLUMN status TYPE TEXT;

-- Удаляем енумы
DROP TYPE IF EXISTS user_role;
DROP TYPE IF EXISTS user_status;
DROP TYPE IF EXISTS document_status;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Воссоздаем енумы
CREATE TYPE user_role AS ENUM ('admin', 'employee');
CREATE TYPE user_status AS ENUM ('pending', 'active', 'inactive');
CREATE TYPE document_status AS ENUM ('pending', 'processing', 'indexed', 'failed');

-- Возвращаем енумы в таблицы
ALTER TABLE organization_members 
    ALTER COLUMN role TYPE user_role USING role::user_role,
    ALTER COLUMN status TYPE user_status USING status::user_status;

ALTER TABLE invitations 
    ALTER COLUMN role TYPE user_role USING role::user_role;

ALTER TABLE documents 
    ALTER COLUMN status TYPE document_status USING status::document_status;

-- +goose StatementEnd
