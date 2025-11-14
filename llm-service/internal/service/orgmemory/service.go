package orgmemory

import (
	"context"
	"strings"

	"llm-service/internal/domain"

	"github.com/opentracing/opentracing-go"
)

const (
	// MaxFactsPerOrganization ограничение количества фактов на организацию
	MaxFactsPerOrganization = 50
	// MaxFactLength максимальная длина одного факта (символов)
	MaxFactLength = 100
)

type repository interface {
	CreateOrganizationMemoryFact(ctx context.Context, fact domain.OrganizationMemoryFact) (domain.OrganizationMemoryFact, error)
	ListOrganizationMemoryFacts(ctx context.Context, organizationID domain.ID, limit int) ([]domain.OrganizationMemoryFact, error)
	CountOrganizationMemoryFacts(ctx context.Context, organizationID domain.ID) (int, error)
	DeleteOrganizationMemoryFact(ctx context.Context, organizationID domain.ID, factID domain.ID) error
}

type Service struct {
	repo repository
}

func New(repo repository) *Service {
	return &Service{repo: repo}
}

// AddFact добавляет короткий факт для организации с валидацией и ограничениями.
func (s *Service) AddFact(ctx context.Context, organizationID domain.ID, content string) (domain.OrganizationMemoryFact, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.orgmemory.AddFact")
	defer span.Finish()

	content = strings.TrimSpace(content)
	if content == "" {
		return domain.OrganizationMemoryFact{}, domain.ErrInvalidArgument
	}
	if len([]rune(content)) > MaxFactLength {
		content = string([]rune(content)[:MaxFactLength])
	}

	cnt, err := s.repo.CountOrganizationMemoryFacts(ctx, organizationID)
	if err != nil {
		return domain.OrganizationMemoryFact{}, err
	}
	if cnt >= MaxFactsPerOrganization {
		return domain.OrganizationMemoryFact{}, domain.ErrTooManyRequests
	}

	fact := domain.NewOrganizationMemoryFact(organizationID, content)

	created, err := s.repo.CreateOrganizationMemoryFact(ctx, fact)
	if err != nil {
		return domain.OrganizationMemoryFact{}, err
	}
	if created.ID == (domain.ID{}) {
		return created, nil
	}

	return created, nil
}

// ListFacts возвращает все факты организации с ограничением сверху.
func (s *Service) ListFacts(ctx context.Context, organizationID domain.ID) ([]domain.OrganizationMemoryFact, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.orgmemory.ListFacts")
	defer span.Finish()

	return s.repo.ListOrganizationMemoryFacts(ctx, organizationID, MaxFactsPerOrganization)
}

// DeleteFact удаляет факт организации по ID.
func (s *Service) DeleteFact(ctx context.Context, organizationID domain.ID, factID domain.ID) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.orgmemory.DeleteFact")
	defer span.Finish()

	return s.repo.DeleteOrganizationMemoryFact(ctx, organizationID, factID)
}
