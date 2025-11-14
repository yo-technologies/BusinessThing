package memory

import (
	"context"

	"llm-service/internal/domain"
	desc "llm-service/pkg/agent"

	"github.com/opentracing/opentracing-go"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Service) ListMemoryFacts(ctx context.Context, req *desc.ListMemoryFactsRequest) (*desc.ListMemoryFactsResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.chat.ListMemoryFacts")
	defer span.Finish()

	organizationID, err := domain.ParseID(req.GetOrgId())
	if err != nil {
		return nil, err
	}

	facts, err := s.orgMemoryService.ListFacts(ctx, organizationID)
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
