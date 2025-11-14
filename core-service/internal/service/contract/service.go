package contract

import (
	"context"
	"core-service/internal/domain"
	"strings"

	"github.com/opentracing/opentracing-go"
)

type repository interface {
	CreateContract(ctx context.Context, contract domain.GeneratedContract) (domain.GeneratedContract, error)
	GetContract(ctx context.Context, id domain.ID) (domain.GeneratedContract, error)
	ListContracts(ctx context.Context, organizationID domain.ID, limit, offset int) ([]domain.GeneratedContract, int, error)
	ListContractsByTemplate(ctx context.Context, templateID domain.ID) ([]domain.GeneratedContract, error)
	DeleteContract(ctx context.Context, id domain.ID) error
}

type Service struct {
	repo repository
}

func New(repo repository) *Service {
	return &Service{repo: repo}
}

// RegisterGeneratedContract registers a generated contract (called by LLM Service)
func (s *Service) RegisterGeneratedContract(ctx context.Context, organizationID, templateID domain.ID, name, filledData, s3Key, fileType string) (domain.GeneratedContract, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.contract.RegisterGeneratedContract")
	defer span.Finish()

	name = strings.TrimSpace(name)
	if name == "" {
		return domain.GeneratedContract{}, domain.NewInvalidArgumentError("contract name is required")
	}
	if s3Key == "" {
		return domain.GeneratedContract{}, domain.NewInvalidArgumentError("s3 key is required")
	}

	contract := domain.NewGeneratedContract(organizationID, templateID, name, filledData, s3Key, fileType)

	created, err := s.repo.CreateContract(ctx, contract)
	if err != nil {
		return domain.GeneratedContract{}, err
	}

	return created, nil
}

// GetGeneratedContract retrieves a contract by ID
func (s *Service) GetGeneratedContract(ctx context.Context, id domain.ID) (domain.GeneratedContract, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.contract.GetGeneratedContract")
	defer span.Finish()

	return s.repo.GetContract(ctx, id)
}

// ListContracts retrieves contracts with pagination
func (s *Service) ListContracts(ctx context.Context, organizationID domain.ID, page, pageSize int) ([]domain.GeneratedContract, int, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.contract.ListContracts")
	defer span.Finish()

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	return s.repo.ListContracts(ctx, organizationID, pageSize, offset)
}

// DeleteGeneratedContract deletes a contract
func (s *Service) DeleteGeneratedContract(ctx context.Context, id domain.ID) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.contract.DeleteGeneratedContract")
	defer span.Finish()

	return s.repo.DeleteContract(ctx, id)
}

// ListGeneratedContractsByOrganization retrieves all contracts for an organization (without pagination)
func (s *Service) ListGeneratedContractsByOrganization(ctx context.Context, organizationID domain.ID) ([]domain.GeneratedContract, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.contract.ListGeneratedContractsByOrganization")
	defer span.Finish()

	contracts, _, err := s.repo.ListContracts(ctx, organizationID, 1000, 0)
	return contracts, err
}

// ListGeneratedContractsByTemplate retrieves all contracts for a template
func (s *Service) ListGeneratedContractsByTemplate(ctx context.Context, templateID domain.ID) ([]domain.GeneratedContract, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.contract.ListGeneratedContractsByTemplate")
	defer span.Finish()

	return s.repo.ListContractsByTemplate(ctx, templateID)
}
