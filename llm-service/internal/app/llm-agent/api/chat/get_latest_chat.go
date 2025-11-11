package chat

import (
	"context"

	"llm-service/internal/app/interceptors"
	"llm-service/internal/app/mappers"
	desc "llm-service/pkg/agent"

	"github.com/opentracing/opentracing-go"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Service) GetLatestChat(ctx context.Context, _ *emptypb.Empty) (*desc.GetChatResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.chat.GetLatestChat")
	defer span.Finish()

	userID, err := interceptors.UserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	result, err := s.chatService.GetLatestChat(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &desc.GetChatResponse{
		Chat:     mappers.ToProtoChat(result.Chat),
		Messages: mappers.ToProtoMessages(result.Messages),
	}, nil
}
