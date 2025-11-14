package repository

import (
	"context"
	"llm-service/internal/domain"
)

// ChatRepository - репозиторий для работы с чатами
type ChatRepository interface {
	// CreateChat создает новый чат
	CreateChat(ctx context.Context, chat *domain.Chat) error

	// GetChatByID получает чат по ID
	GetChatByID(ctx context.Context, id domain.ID) (*domain.Chat, error)

	// ListChats получает список чатов с фильтрацией и пагинацией
	ListChats(ctx context.Context, filter ChatFilter, page, pageSize int) ([]*domain.Chat, int, error)

	// UpdateChat обновляет чат
	UpdateChat(ctx context.Context, chat *domain.Chat) error

	// DeleteChat удаляет чат
	DeleteChat(ctx context.Context, id, userID, orgID domain.ID) error

	// GetChatByParentToolCallID получает дочерний чат по ID tool call родителя
	GetChatByParentToolCallID(ctx context.Context, parentToolCallID domain.ID) (*domain.Chat, error)

	// GetActiveChildChat получает активный дочерний чат для родительского чата
	GetActiveChildChat(ctx context.Context, parentChatID domain.ID) (*domain.Chat, error)
}

// ChatFilter - фильтр для поиска чатов
type ChatFilter struct {
	OrganizationID *domain.ID
	UserID         *domain.ID
	Status         *domain.ChatStatus
	ParentChatID   *domain.ID
}

// MessageRepository - репозиторий для работы с сообщениями
type MessageRepository interface {
	// CreateMessage создает новое сообщение
	CreateMessage(ctx context.Context, message *domain.Message) error

	// ListMessagesByChatID получает список сообщений чата
	ListMessagesByChatID(ctx context.Context, chatID domain.ID, limit, offset int) ([]*domain.Message, int, error)

	// ListMessagesByChatIDWithToolCalls получает сообщения вместе с tool calls
	ListMessagesByChatIDWithToolCalls(ctx context.Context, chatID domain.ID, limit, offset int) ([]*domain.Message, int, error)

	// ListMessagesWithSubchatsWithToolCalls получает сообщения родительского чата и всех его субчатов с tool calls
	ListMessagesWithSubchatsWithToolCalls(ctx context.Context, parentChatID domain.ID, limit, offset int) ([]*domain.Message, int, error)

	// DeleteMessage удаляет сообщение
	DeleteMessage(ctx context.Context, id domain.ID) error
}

// ToolCallRepository - репозиторий для работы с вызовами инструментов
type ToolCallRepository interface {
	// CreateToolCall создает новый tool call
	CreateToolCall(ctx context.Context, messageID domain.ID, toolCall *domain.ToolCall) error

	// GetToolCallByID получает tool call по ID
	GetToolCallByID(ctx context.Context, id domain.ID) (*domain.ToolCall, error)

	// ListToolCallsByMessageID получает tool calls для сообщения
	ListToolCallsByMessageID(ctx context.Context, messageID domain.ID) ([]*domain.ToolCall, error)

	// UpdateToolCall обновляет tool call
	UpdateToolCall(ctx context.Context, toolCall *domain.ToolCall) error

	// UpdateToolCallStatus обновляет статус tool call
	UpdateToolCallStatus(ctx context.Context, id domain.ID, status domain.ToolCallStatus) error

	// CompleteToolCall завершает tool call с результатом
	CompleteToolCall(ctx context.Context, id domain.ID, result []byte) error
}

// SubagentSessionRepository - репозиторий для работы с сессиями субагентов
type SubagentSessionRepository interface {
	// Create создает новую сессию субагента
	Create(ctx context.Context, session *SubagentSession) error

	// GetByChildChatID получает сессию по ID дочернего чата
	GetByChildChatID(ctx context.Context, childChatID domain.ID) (*SubagentSession, error)

	// GetByParentToolCallID получает сессию по ID tool call родителя
	GetByParentToolCallID(ctx context.Context, parentToolCallID domain.ID) (*SubagentSession, error)

	// Update обновляет сессию
	Update(ctx context.Context, session *SubagentSession) error

	// Complete завершает сессию с summary
	Complete(ctx context.Context, childChatID domain.ID, summary string) error

	// GetActiveByParentChatID получает активную сессию субагента для родительского чата
	GetActiveByParentChatID(ctx context.Context, parentChatID domain.ID) (*SubagentSession, error)
}

// SubagentSession - сессия работы субагента
type SubagentSession struct {
	ID               domain.ID
	ParentChatID     domain.ID
	ParentToolCallID domain.ID
	ChildChatID      domain.ID
	TaskDescription  string
	Summary          *string
	CreatedAt        interface{}
	CompletedAt      *interface{}
}
