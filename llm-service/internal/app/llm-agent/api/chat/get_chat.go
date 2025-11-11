package chat

import (
	"context"
	"fmt"

	"llm-service/internal/app/interceptors"
	"llm-service/internal/app/mappers"
	"llm-service/internal/domain"
	desc "llm-service/pkg/agent"

	"github.com/opentracing/opentracing-go"
)

func (s *Service) GetChat(ctx context.Context, req *desc.GetChatRequest) (*desc.GetChatResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.chat.GetChat")
	defer span.Finish()

	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %w", err, domain.ErrInvalidArgument)
	}

	userID, err := interceptors.UserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	chatID, err := domain.ParseID(req.GetChatId())
	if err != nil {
		return nil, domain.ErrInvalidArgument
	}

	result, err := s.chatService.GetChat(ctx, userID, chatID)
	if err != nil {
		return nil, err
	}

	return &desc.GetChatResponse{
		Chat:     mappers.ToProtoChat(result.Chat),
		Messages: mappers.ToProtoMessages(result.Messages),
	}, nil
}
