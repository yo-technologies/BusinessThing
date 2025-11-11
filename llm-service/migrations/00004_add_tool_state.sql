-- +goose Up
ALTER TABLE llm_chat_messages
    ADD COLUMN IF NOT EXISTS tool_state TEXT NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_llm_chat_messages_tool_state ON llm_chat_messages(tool_state);

-- +goose Down
DROP INDEX IF EXISTS idx_llm_chat_messages_tool_state;
ALTER TABLE llm_chat_messages DROP COLUMN IF EXISTS tool_state;
