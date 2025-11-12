-- +goose Up
CREATE TABLE IF NOT EXISTS llm_user_memory_facts (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    content VARCHAR(200) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_user_content UNIQUE (user_id, content)
);

CREATE INDEX IF NOT EXISTS idx_llm_user_memory_facts_user ON llm_user_memory_facts(user_id, created_at);

-- +goose Down
DROP TABLE IF EXISTS llm_user_memory_facts;
