package template

import (
	"context"
	"core-service/internal/domain"
	"strings"

	"github.com/opentracing/opentracing-go"
)

type repository interface {
	CreateTemplate(ctx context.Context, template domain.ContractTemplate) (domain.ContractTemplate, error)
	GetTemplate(ctx context.Context, id domain.ID) (domain.ContractTemplate, error)
	ListTemplates(ctx context.Context, organizationID domain.ID, limit, offset int) ([]domain.ContractTemplate, int, error)
	UpdateTemplate(ctx context.Context, template domain.ContractTemplate) (domain.ContractTemplate, error)
	DeleteTemplate(ctx context.Context, id domain.ID) error
}

type Service struct {
	repo repository
}

func New(repo repository) *Service {
	return &Service{repo: repo}
}

// CreateTemplate creates a new contract template
func (s *Service) CreateTemplate(ctx context.Context, organizationID domain.ID, name, description, templateType, fieldsSchema, contentTemplate string) (domain.ContractTemplate, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.template.CreateTemplate")
	defer span.Finish()

	name = strings.TrimSpace(name)
	if name == "" {
		return domain.ContractTemplate{}, domain.NewInvalidArgumentError("template name is required")
	}
	if contentTemplate == "" {
		return domain.ContractTemplate{}, domain.NewInvalidArgumentError("template content is required")
	}

	template := domain.NewContractTemplate(organizationID, name, description, templateType, fieldsSchema, contentTemplate)

	created, err := s.repo.CreateTemplate(ctx, template)
	if err != nil {
		return domain.ContractTemplate{}, err
	}

	return created, nil
}

// GetTemplate retrieves a template by ID
func (s *Service) GetTemplate(ctx context.Context, id domain.ID) (domain.ContractTemplate, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.template.GetTemplate")
	defer span.Finish()

	return s.repo.GetTemplate(ctx, id)
}

// ListTemplates retrieves templates with pagination
func (s *Service) ListTemplates(ctx context.Context, organizationID domain.ID, page, pageSize int) ([]domain.ContractTemplate, int, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.template.ListTemplates")
	defer span.Finish()

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	return s.repo.ListTemplates(ctx, organizationID, pageSize, offset)
}

// UpdateTemplate updates an existing template
func (s *Service) UpdateTemplate(ctx context.Context, id domain.ID, name, description, fieldsSchema, contentTemplate *string) (domain.ContractTemplate, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.template.UpdateTemplate")
	defer span.Finish()

	// Get existing template
	template, err := s.repo.GetTemplate(ctx, id)
	if err != nil {
		return domain.ContractTemplate{}, err
	}

	// Update template
	template.Update(name, description, fieldsSchema, contentTemplate)

	updated, err := s.repo.UpdateTemplate(ctx, template)
	if err != nil {
		return domain.ContractTemplate{}, err
	}

	return updated, nil
}

// DeleteTemplate deletes a template
func (s *Service) DeleteTemplate(ctx context.Context, id domain.ID) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.template.DeleteTemplate")
	defer span.Finish()

	return s.repo.DeleteTemplate(ctx, id)
}
