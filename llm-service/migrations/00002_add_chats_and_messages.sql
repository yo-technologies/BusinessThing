-- +goose Up
-- Таблица чатов
CREATE TABLE IF NOT EXISTS chats (
    id UUID PRIMARY KEY,
    organization_id UUID NOT NULL,
    user_id UUID NOT NULL,
    agent_key VARCHAR(255) NOT NULL,
    title VARCHAR(500),
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    parent_chat_id UUID REFERENCES chats(id) ON DELETE CASCADE,
    parent_tool_call_id UUID,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Индексы для чатов
CREATE INDEX idx_chats_organization_user ON chats(organization_id, user_id);
CREATE INDEX idx_chats_agent_key ON chats(agent_key);
CREATE INDEX idx_chats_status ON chats(status);
CREATE INDEX idx_chats_parent_chat_id ON chats(parent_chat_id);
CREATE INDEX idx_chats_created_at ON chats(created_at DESC);

-- Таблица сообщений
CREATE TABLE IF NOT EXISTS messages (
    id UUID PRIMARY KEY,
    chat_id UUID NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
    role VARCHAR(50) NOT NULL,
    content TEXT NOT NULL,
    sender VARCHAR(255),
    -- NULL для user, agent_key для агентов
    tool_call_id UUID,
    -- для tool результатов
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Индексы для сообщений
CREATE INDEX idx_messages_chat_id ON messages(chat_id);
CREATE INDEX idx_messages_created_at ON messages(created_at);
CREATE INDEX idx_messages_role ON messages(role);
CREATE INDEX idx_messages_tool_call_id ON messages(tool_call_id);

-- Таблица вызовов инструментов
CREATE TABLE IF NOT EXISTS tool_calls (
    id UUID PRIMARY KEY,
    message_id UUID NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    arguments JSONB NOT NULL,
    result JSONB,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP
);

-- Индексы для tool_calls
CREATE INDEX idx_tool_calls_message_id ON tool_calls(message_id);
CREATE INDEX idx_tool_calls_status ON tool_calls(status);
CREATE INDEX idx_tool_calls_created_at ON tool_calls(created_at);

-- +goose Down
DROP FUNCTION IF EXISTS update_chat_updated_at();
DROP TABLE IF EXISTS subagent_sessions;
DROP TABLE IF EXISTS tool_calls;
DROP TABLE IF EXISTS messages;
DROP TABLE IF EXISTS chats;