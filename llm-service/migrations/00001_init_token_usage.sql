-- +goose Up
CREATE TABLE IF NOT EXISTS llm_token_usage (
    user_id UUID NOT NULL,
    day DATE NOT NULL,
    used_tokens INTEGER NOT NULL DEFAULT 0,
    reserved_tokens INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, day)
);

-- +goose Down
DROP TABLE IF EXISTS llm_token_usage;
