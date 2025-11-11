-- +goose Up
CREATE TABLE IF NOT EXISTS llm_chat_messages (
    id UUID PRIMARY KEY,
    chat_id UUID NOT NULL REFERENCES llm_chats(id) ON DELETE CASCADE,
    role TEXT NOT NULL,
    content TEXT NOT NULL DEFAULT '',
    tool_name TEXT NULL,
    tool_call_id TEXT NULL,
    tool_arguments JSONB NULL,
    token_usage INTEGER NULL,
    error TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_llm_chat_messages_chat_id ON llm_chat_messages(chat_id);
CREATE INDEX IF NOT EXISTS idx_llm_chat_messages_created_at ON llm_chat_messages(created_at);

-- +goose Down
DROP TABLE IF EXISTS llm_chat_messages;
