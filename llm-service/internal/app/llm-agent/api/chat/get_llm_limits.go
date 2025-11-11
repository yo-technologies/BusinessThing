package chat

import (
	"context"

	"llm-service/internal/app/interceptors"
	desc "llm-service/pkg/agent"

	"github.com/opentracing/opentracing-go"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Service) GetLLMLimits(ctx context.Context, _ *emptypb.Empty) (*desc.GetLLMLimitsResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.chat.GetLLMLimits")
	defer span.Finish()

	userID, err := interceptors.UserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	limits, err := s.chatService.GetLLMLimits(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &desc.GetLLMLimitsResponse{
		DailyLimit: int32(limits.DailyLimit),
		Used:       int32(limits.Used),
		Reserved:   int32(limits.Reserved),
	}, nil
}
