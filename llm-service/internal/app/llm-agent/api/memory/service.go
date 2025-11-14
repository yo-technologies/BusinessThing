package memory

import (
	"context"

	"llm-service/internal/domain"

	desc "llm-service/pkg/agent"
)

type OrganizationMemoryService interface {
	ListFacts(ctx context.Context, organizationID domain.ID) ([]domain.OrganizationMemoryFact, error)
	AddFact(ctx context.Context, organizationID domain.ID, content string) (domain.OrganizationMemoryFact, error)
	DeleteFact(ctx context.Context, organizationID domain.ID, factID domain.ID) error
}

type Service struct {
	orgMemoryService OrganizationMemoryService

	desc.UnimplementedMemoryServiceServer
}

func NewService(orgMemoryService OrganizationMemoryService) *Service {
	return &Service{
		orgMemoryService: orgMemoryService,
	}
}
