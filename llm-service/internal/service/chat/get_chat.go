package chat

import (
	"context"
	"fmt"
	"llm-service/internal/domain"
	"llm-service/internal/domain/dto"

	"github.com/opentracing/opentracing-go"
)

func (s *Service) GetChat(ctx context.Context, userID domain.ID, chatID domain.ID) (dto.GetChatDTO, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.GetChat")
	defer span.Finish()

	// Get chat by id
	chat, err := s.chatRepository.GetChatByID(ctx, chatID)
	if err != nil {
		return dto.GetChatDTO{}, fmt.Errorf("failed to get chat by id: %w", err)
	}

	// Проверяем, что пользователь имеет доступ к этому чату
	if chat.UserID != userID {
		return dto.GetChatDTO{}, domain.ErrForbidden
	}

	messages, err := s.chatRepository.ListChatMessages(ctx, chat.ID, 1000, 0) // limit 1000, offset 0
	if err != nil {
		return dto.GetChatDTO{}, fmt.Errorf("failed to list chat messages: %w", err)
	}

	return dto.GetChatDTO{
		Chat:     chat,
		Messages: messages,
	}, nil
}
