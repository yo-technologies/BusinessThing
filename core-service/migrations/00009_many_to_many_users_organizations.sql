-- +goose Up
-- +goose StatementBegin

-- Создаем таблицу связей пользователей и организаций
CREATE TABLE IF NOT EXISTS organization_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    role TEXT NOT NULL DEFAULT 'employee',
    status TEXT NOT NULL DEFAULT 'active',
    joined_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(organization_id, user_id)
);

CREATE INDEX idx_organization_members_organization_id ON organization_members(organization_id);
CREATE INDEX idx_organization_members_user_id ON organization_members(user_id);
CREATE INDEX idx_organization_members_status ON organization_members(status);

-- Мигрируем существующие данные из users в organization_members
INSERT INTO organization_members (organization_id, user_id, email, role, status, joined_at, updated_at)
SELECT organization_id, id, email, role, status, created_at, updated_at
FROM users
WHERE organization_id IS NOT NULL;

-- Удаляем колонки из users (делаем пользователя независимым от организации)
ALTER TABLE users DROP COLUMN IF EXISTS organization_id;
ALTER TABLE users DROP COLUMN IF EXISTS email;
ALTER TABLE users DROP COLUMN IF EXISTS role;
ALTER TABLE users DROP COLUMN IF EXISTS status;

-- Добавляем статус для самого пользователя (не зависящий от организации)
ALTER TABLE users ADD COLUMN is_active BOOLEAN NOT NULL DEFAULT true;

-- Убираем уникальный индекс, так как теперь пользователь может быть в нескольких организациях
-- (индекс уже удалился вместе с колонкой organization_id)

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Возвращаем колонки в users
ALTER TABLE users ADD COLUMN organization_id UUID;
ALTER TABLE users ADD COLUMN email VARCHAR(255);
ALTER TABLE users ADD COLUMN role TEXT DEFAULT 'employee';
ALTER TABLE users ADD COLUMN status TEXT DEFAULT 'active';

-- Восстанавливаем данные из organization_members (берем первую запись для каждого пользователя)
UPDATE users u
SET 
    organization_id = om.organization_id,
    email = om.email,
    role = om.role,
    status = om.status
FROM (
    SELECT DISTINCT ON (user_id) *
    FROM organization_members
    ORDER BY user_id, joined_at
) om
WHERE u.id = om.user_id;

ALTER TABLE users DROP COLUMN IF EXISTS is_active;

-- Удаляем таблицу organization_members
DROP TABLE IF EXISTS organization_members;

-- +goose StatementEnd
