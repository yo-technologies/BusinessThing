package chat

import (
	"context"
	"fmt"

	"llm-service/internal/domain"
	"llm-service/internal/domain/dto"

	"github.com/opentracing/opentracing-go"
)

func (s *Service) GetLatestChat(ctx context.Context, userID domain.ID) (dto.GetChatDTO, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.GetLatestChat")
	defer span.Finish()

	// Get latest chat by user id
	chat, err := s.chatRepository.GetLatestChat(ctx, userID)
	if err != nil {
		return dto.GetChatDTO{}, fmt.Errorf("failed to get latest chat: %w", err)
	}

	messages, err := s.chatRepository.ListChatMessages(ctx, chat.ID, 1000, 0)
	if err != nil {
		return dto.GetChatDTO{}, fmt.Errorf("failed to list chat messages: %w", err)
	}

	return dto.GetChatDTO{
		Chat:     chat,
		Messages: messages,
	}, nil
}
