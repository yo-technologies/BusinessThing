package dto

import (
	"llm-service/internal/domain"
)

// CreateChatDTO - DTO для создания чата
type CreateChatDTO struct {
	OrganizationID   ID
	UserID           ID
	AgentKey         string
	Title            string
	ParentChatID     *ID
	ParentToolCallID *ID
}

// SendMessageDTO - DTO для отправки сообщения
type SendMessageDTO struct {
	ChatID  *ID
	UserID  ID
	OrgID   ID
	Content string
}

// ChatUsageDTO - DTO статистики использования токенов
type ChatUsageDTO struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

// StreamEventDTO - DTO события стриминга
type StreamEventDTO struct {
	Type    string
	Content interface{}
}

// ExecuteAgentDTO - DTO для запуска агента
type ExecuteAgentDTO struct {
	AgentKey       string
	OrganizationID ID
	UserID         ID
	Task           string
	ChatID         *ID
	Context        map[string]interface{}
}

// ID - тип для идентификаторов
type ID = domain.ID
