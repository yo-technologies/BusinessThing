package dto

import (
	"llm-service/internal/domain"

	"github.com/guregu/null/v6"
)

type GetChatRequest struct {
	ChatID null.String `json:"chat_id"`
}

type GetChatDTO struct {
	Chat     domain.Chat          `json:"chat"`
	Messages []domain.ChatMessage `json:"messages"`
}

type SendChatMessageRequest struct {
	ChatID  null.String `json:"chat_id"`
	Content string      `json:"content"`
}

type ChatUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type ChatCompletionDTO struct {
	Chat     domain.Chat          `json:"chat"`
	Messages []domain.ChatMessage `json:"messages"`
	Usage    *ChatUsage           `json:"usage,omitempty"`
}

// Streaming callbacks and tool events
type ToolEventState string

const (
	ToolInvoking  ToolEventState = "invoking"
	ToolCompleted ToolEventState = "completed"
	ToolError     ToolEventState = "error"
)

type ToolEvent struct {
	ToolName   string         `json:"tool_name"`
	ToolCallID string         `json:"tool_call_id"`
	ArgsJSON   string         `json:"args_json"`
	State      ToolEventState `json:"state"`
	Error      string         `json:"error,omitempty"`
}

type ChatStreamCallbacks struct {
	// Tokens/Status
	OnStatus        func(status string) error
	OnUsage         func(usage ChatUsage) error
	OnContentDelta  func(delta string) error
	OnFinalResponse func(result ChatCompletionDTO) error
	OnError         func(err error) error

	// Tooling UX
	OnToolEvent func(event ToolEvent) error

	// User stop/abort
	ShouldStop func() bool
}
