package domain

import (
	"encoding/json"

	"github.com/guregu/null/v6"
)

type ChatMessageRole string

const (
	ChatMessageRoleUser      ChatMessageRole = "user"
	ChatMessageRoleAssistant ChatMessageRole = "assistant"
	ChatMessageRoleTool      ChatMessageRole = "tool"
	ChatMessageRoleSystem    ChatMessageRole = "system"
)

type ToolState string

const (
	ToolStateUnknown  ToolState = ""
	ToolStateProposed ToolState = "proposed"
	ToolStateRejected ToolState = "rejected"
	ToolStateApproved ToolState = "approved"
)

type ChatMessage struct {
	Model
	ChatID        ID              `db:"chat_id"`
	Role          ChatMessageRole `db:"role"`
	Content       string          `db:"content"`
	ToolName      null.String     `db:"tool_name"`
	ToolCallID    null.String     `db:"tool_call_id"`
	ToolArguments json.RawMessage `db:"tool_arguments"`
	TokenUsage    null.Int        `db:"token_usage"`
	Error         null.String     `db:"error"`
	ToolState     ToolState       `db:"tool_state"`
}

// NewChatMessage builds a message. toolArgs will be marshaled into JSON if provided.
func NewChatMessage(chatID ID, role ChatMessageRole, content string, toolName null.String, toolCallID null.String, toolArgs map[string]any) (ChatMessage, error) {
	var raw json.RawMessage
	if toolArgs != nil {
		b, err := json.Marshal(toolArgs)
		if err != nil {
			return ChatMessage{}, err
		}
		raw = b
	}
	return ChatMessage{
		Model:         NewModel(),
		ChatID:        chatID,
		Role:          role,
		Content:       content,
		ToolName:      toolName,
		ToolCallID:    toolCallID,
		ToolArguments: raw,
		ToolState:     ToolStateUnknown,
	}, nil
}
