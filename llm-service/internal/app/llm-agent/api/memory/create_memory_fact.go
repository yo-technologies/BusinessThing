package memory

import (
	"context"
	"llm-service/internal/domain"
	desc "llm-service/pkg/agent"

	"github.com/opentracing/opentracing-go"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Service) CreateMemoryFact(ctx context.Context, req *desc.CreateMemoryFactRequest) (*desc.CreateMemoryFactResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.chat.CreateMemoryFact")
	defer span.Finish()

	organizationID, err := domain.ParseID(req.GetOrgId())
	if err != nil {
		return nil, err
	}

	fact, err := s.orgMemoryService.AddFact(ctx, organizationID, req.GetContent())
	if err != nil {
		return nil, err
	}

	return &desc.CreateMemoryFactResponse{Fact: &desc.MemoryFact{
		Id:        fact.ID.String(),
		Content:   fact.Content,
		CreatedAt: timestamppb.New(fact.CreatedAt),
	}}, nil
}
