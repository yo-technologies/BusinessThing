-- +goose Up
CREATE TABLE IF NOT EXISTS llm_chats (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    title TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_llm_chats_user_id ON llm_chats(user_id);
CREATE INDEX IF NOT EXISTS idx_llm_chats_created_at ON llm_chats(created_at);

-- +goose Down
DROP TABLE IF EXISTS llm_chats;
