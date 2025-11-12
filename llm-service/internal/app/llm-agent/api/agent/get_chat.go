package agent

import (
	"context"
	"fmt"
	"llm-service/internal/app/interceptors"
	"llm-service/internal/app/llm-agent/mappers"
	"llm-service/internal/domain"
	"llm-service/internal/logger"
	desc "llm-service/pkg/agent"

	"github.com/opentracing/opentracing-go"
)

func (s *Service) GetChat(ctx context.Context, req *desc.GetChatRequest) (*desc.GetChatResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.agent.GetChat")
	defer span.Finish()

	_, err := interceptors.UserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	_, err = domain.ParseID(req.GetOrgId())
	if err != nil {
		return nil, fmt.Errorf("invalid organization ID: %w", err)
	}

	chatID, err := domain.ParseID(req.ChatId)
	if err != nil {
		return nil, fmt.Errorf("invalid chat ID: %w", err)
	}

	chat, err := s.chatManager.GetChat(ctx, chatID)
	if err != nil {
		logger.Error(ctx, "failed to get chat", "error", err)
		return nil, err
	}

	return &desc.GetChatResponse{
		Chat: mappers.DomainChatToProto(chat),
	}, nil
}
