package agent

import (
	"context"
	"fmt"
	"llm-service/internal/app/interceptors"
	"llm-service/internal/app/llm-agent/mappers"
	"llm-service/internal/domain"
	desc "llm-service/pkg/agent"

	"github.com/opentracing/opentracing-go"
)

func (s *Service) ListChats(ctx context.Context, req *desc.ListChatsRequest) (*desc.ListChatsResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.agent.ListChats")
	defer span.Finish()

	// Получаем userID и orgID из контекста (они должны быть установлены в auth interceptor)
	userID, err := interceptors.UserIDFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user ID from context: %w", err)
	}

	orgID, err := domain.ParseID(req.GetOrgId())
	if err != nil {
		return nil, fmt.Errorf("invalid organization ID: %w", err)
	}

	page := int(req.Page)
	if page < 1 {
		page = 1
	}

	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	chats, total, err := s.chatManager.ListChats(ctx, orgID, userID, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to list chats: %w", err)
	}

	pbChats := make([]*desc.Chat, 0, len(chats))
	for _, chat := range chats {
		pbChats = append(pbChats, mappers.DomainChatToProto(chat))
	}

	return &desc.ListChatsResponse{
		Chats:    pbChats,
		Total:    int32(total),
		Page:     int32(page),
		PageSize: int32(pageSize),
	}, nil
}
