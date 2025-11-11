package memory

import (
	"context"
	"llm-service/internal/app/interceptors"
	"llm-service/internal/app/mappers"
	desc "llm-service/pkg/agent"

	"github.com/opentracing/opentracing-go"
)

func (s *Service) CreateMemoryFact(ctx context.Context, req *desc.CreateMemoryFactRequest) (*desc.CreateMemoryFactResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.chat.CreateMemoryFact")
	defer span.Finish()

	userID, err := interceptors.UserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	fact, err := s.memoryService.AddFact(ctx, userID, req.GetContent())
	if err != nil {
		return nil, err
	}

	return &desc.CreateMemoryFactResponse{Fact: &desc.MemoryFact{
		Id:        fact.ID.String(),
		Content:   fact.Content,
		CreatedAt: mappers.ToProtoTimestamp(fact.CreatedAt),
	}}, nil
}
