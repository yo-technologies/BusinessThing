package chat

import (
	"context"
	"fmt"

	"llm-service/internal/domain"
	desc "llm-service/pkg/agent"

	"github.com/opentracing/opentracing-go"
)

// GetChats currently returns empty list (listing not implemented yet)
func (s *Service) GetChats(ctx context.Context, req *desc.GetChatsRequest) (*desc.GetChatsResponse, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "api.chat.GetChats")
	defer span.Finish()

	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %w", err, domain.ErrInvalidArgument)
	}

	return &desc.GetChatsResponse{Chats: []*desc.Chat{}}, nil
}
