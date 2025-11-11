package memory

import (
	"context"

	"llm-service/internal/app/interceptors"
	"llm-service/internal/domain"
	desc "llm-service/pkg/agent"

	"github.com/opentracing/opentracing-go"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Service) DeleteMemoryFact(ctx context.Context, req *desc.DeleteMemoryFactRequest) (*emptypb.Empty, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.chat.DeleteMemoryFact")
	defer span.Finish()

	userID, err := interceptors.UserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	fid, err := domain.ParseID(req.GetId())
	if err != nil {
		return nil, err
	}

	if err := s.memoryService.DeleteFact(ctx, userID, fid); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}
