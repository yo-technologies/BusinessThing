package chat

import (
	"context"
	"time"

	"llm-service/internal/domain"
	"llm-service/internal/llm"
)

type toolsService interface {
	ChatAgentToolDefinitions() []llm.ToolDefinition
	ExecuteChatAgentTool(ctx context.Context, ctxData domain.AgentChatContext, name string, arguments string) (string, error)
	RequiresConfirmation(toolName string) bool
}

type chatRepository interface {
	CreateChat(ctx context.Context, chat domain.Chat) (domain.Chat, error)
	GetChatByID(ctx context.Context, id domain.ID) (domain.Chat, error)
	GetLatestChat(ctx context.Context, userID domain.ID) (domain.Chat, error)

	CreateChatMessage(ctx context.Context, message domain.ChatMessage) (domain.ChatMessage, error)
	ListChatMessages(ctx context.Context, chatID domain.ID, limit, offset int) ([]domain.ChatMessage, error)
	GetMessageByToolCallID(ctx context.Context, chatID domain.ID, toolCallID string) (domain.ChatMessage, error)
	UpdateMessageToolState(ctx context.Context, messageID domain.ID, state domain.ToolState) error
}

type quotaService interface {
	Reserve(ctx context.Context, userID domain.ID, n int) (bool, error)
	Confirm(ctx context.Context, userID domain.ID, reserved int, actual int) error
	GetLLMDailyUsage(ctx context.Context, userID domain.ID, day time.Time) (used int, reserved int, err error)
	DailyLimit(ctx context.Context, userID domain.ID) int
}

type memoryService interface {
	ListFacts(ctx context.Context, userID domain.ID) ([]domain.UserMemoryFact, error)
}

type Service struct {
	toolsService   toolsService
	chatRepository chatRepository
	llmClient      llm.CompletionProvider
	quotaService   quotaService
	memoryService  memoryService
}

func New(
	toolsService toolsService,
	chatRepository chatRepository,
	llmClient llm.CompletionProvider,
	quotaSvc quotaService,
	memorySvc memoryService,
) *Service {
	return &Service{
		toolsService:   toolsService,
		chatRepository: chatRepository,
		llmClient:      llmClient,
		quotaService:   quotaSvc,
		memoryService:  memorySvc,
	}
}
