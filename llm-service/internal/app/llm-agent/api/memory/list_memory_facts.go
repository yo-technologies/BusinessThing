package memory

import (
	"context"

	"llm-service/internal/app/interceptors"
	desc "llm-service/pkg/agent"

	"github.com/opentracing/opentracing-go"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Service) ListMemoryFacts(ctx context.Context, _ *emptypb.Empty) (*desc.ListMemoryFactsResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.chat.ListMemoryFacts")
	defer span.Finish()

	userID, err := interceptors.UserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	facts, err := s.memoryService.ListFacts(ctx, userID)
	if err != nil {
		return nil, err
	}

	res := &desc.ListMemoryFactsResponse{Facts: make([]*desc.MemoryFact, 0, len(facts))}
	for _, f := range facts {
		res.Facts = append(res.Facts, &desc.MemoryFact{
			Id:        f.ID.String(),
			Content:   f.Content,
			CreatedAt: timestamppb.New(f.CreatedAt),
		})
	}
	return res, nil
}
