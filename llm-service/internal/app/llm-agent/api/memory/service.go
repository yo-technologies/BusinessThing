package memory

import (
	"context"

	"llm-service/internal/domain"

	desc "llm-service/pkg/agent"
)

type MemoryService interface {
	ListFacts(ctx context.Context, userID domain.ID) ([]domain.UserMemoryFact, error)
	AddFact(ctx context.Context, userID domain.ID, content string) (domain.UserMemoryFact, error)
	DeleteFact(ctx context.Context, userID domain.ID, factID domain.ID) error
}

type Service struct {
	memoryService MemoryService

	desc.UnimplementedMemoryServiceServer
}

func NewService(memoryService MemoryService) *Service {
	return &Service{
		memoryService: memoryService,
	}
}
