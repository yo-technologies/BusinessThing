package chat

import (
	"context"

	"llm-service/internal/domain"
	"llm-service/internal/domain/dto"
	desc "llm-service/pkg/agent"
)

type ChatService interface {
	GetChat(ctx context.Context, userID domain.ID, chatID domain.ID) (dto.GetChatDTO, error)
	GetLatestChat(ctx context.Context, userID domain.ID) (dto.GetChatDTO, error)
	GetLLMLimits(ctx context.Context, userID domain.ID) (domain.LLMLimits, error)
	HandleToolCallDecisionStream(ctx context.Context, userID domain.ID, chatID domain.ID, toolCallID string, approve bool, callbacks dto.ChatStreamCallbacks) (dto.ChatCompletionDTO, error)
	SendChatMessageStream(ctx context.Context, userID domain.ID, req dto.SendChatMessageRequest, callbacks dto.ChatStreamCallbacks) (dto.ChatCompletionDTO, error)
}

// Service is the gRPC transport for ChatService
type Service struct {
	chatService ChatService

	desc.UnimplementedChatServiceServer
}

func NewService(chatService ChatService) *Service {
	return &Service{
		chatService: chatService,
	}
}
