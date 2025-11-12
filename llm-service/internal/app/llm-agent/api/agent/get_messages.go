package agent

import (
	"context"
	"llm-service/internal/app/interceptors"
	"llm-service/internal/app/llm-agent/mappers"
	"llm-service/internal/domain"
	"llm-service/internal/logger"
	desc "llm-service/pkg/agent"

	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Service) GetMessages(ctx context.Context, req *desc.GetMessagesRequest) (*desc.GetMessagesResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.agent.GetMessages")
	defer span.Finish()

	userID, err := interceptors.UserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	chatID, err := domain.ParseID(req.ChatId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid chat_id")
	}

	orgID, err := domain.ParseID(req.GetOrgId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid organization_id")
	}

	limit := int(req.Limit)
	if limit < 1 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	offset := int(req.Offset)
	if offset < 0 {
		offset = 0
	}

	messages, total, err := s.chatManager.GetMessages(ctx, chatID, userID, orgID, limit, offset)
	if err != nil {
		logger.Error(ctx, "failed to get messages", "error", err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	pbMessages := make([]*desc.Message, 0, len(messages))
	for _, msg := range messages {
		pbMessages = append(pbMessages, mappers.DomainMessageToProto(msg))
	}

	return &desc.GetMessagesResponse{
		Messages: pbMessages,
		Total:    int32(total),
	}, nil
}
