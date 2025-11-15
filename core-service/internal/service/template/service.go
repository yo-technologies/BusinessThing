package template

import (
	"context"
	"core-service/internal/db"
	"core-service/internal/domain"
	"fmt"
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

type queuePublisher interface {
	PublishMessage(ctx context.Context, message interface{}) error
}

type Service struct {
	repo  repository
	queue queuePublisher
	tx    *db.ContextManager
}

func New(repo repository, queue queuePublisher, tx *db.ContextManager) *Service {
	return &Service{
		repo:  repo,
		queue: queue,
		tx:    tx,
	}
}

// CreateTemplate creates a new contract template
func (s *Service) CreateTemplate(ctx context.Context, organizationID domain.ID, name, description, templateType, fieldsSchema, s3TemplateKey string) (domain.ContractTemplate, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.template.CreateTemplate")
	defer span.Finish()

	name = strings.TrimSpace(name)
	if name == "" {
		return domain.ContractTemplate{}, domain.NewInvalidArgumentError("template name is required")
	}
	if s3TemplateKey == "" {
		return domain.ContractTemplate{}, domain.NewInvalidArgumentError("s3_template_key is required")
	}

	template := domain.NewContractTemplate(organizationID, name, description, templateType, fieldsSchema, "", s3TemplateKey)

	var created domain.ContractTemplate
	err := s.tx.Do(ctx, func(txCtx context.Context) error {
		var err error
		created, err = s.repo.CreateTemplate(ctx, template)
		if err != nil {
			return fmt.Errorf("failed to create template: %w", err)
		}

		// Подсчитать количество полей из fieldsSchema
		fieldsCount := countFieldsInSchema(fieldsSchema)

		// Отправить задачу на индексацию
		indexJob := domain.NewTemplateIndexJob(created.ID, created.OrganizationID, created.Name, created.Description, created.TemplateType, fieldsCount)
		err = s.queue.PublishMessage(ctx, indexJob)
		if err != nil {
			return fmt.Errorf("failed to publish template index job: %w", err)
		}

		return nil
	})
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

// ListTemplates retrieves templates for an organization
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

// ListTemplatesByOrganization retrieves all templates for an organization (without pagination)
func (s *Service) ListTemplatesByOrganization(ctx context.Context, organizationID domain.ID) ([]domain.ContractTemplate, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.template.ListTemplatesByOrganization")
	defer span.Finish()

	templates, _, err := s.repo.ListTemplates(ctx, organizationID, 1000, 0)
	return templates, err
}

// UpdateTemplate updates an existing template
func (s *Service) UpdateTemplate(ctx context.Context, id domain.ID, name, description, fieldsSchema, s3TemplateKey *string) (domain.ContractTemplate, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.template.UpdateTemplate")
	defer span.Finish()

	// Get existing template
	template, err := s.repo.GetTemplate(ctx, id)
	if err != nil {
		return domain.ContractTemplate{}, err
	}

	// Update template
	template.Update(name, description, fieldsSchema, nil, s3TemplateKey)

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

	err := s.tx.Do(ctx, func(txCtx context.Context) error {
		err := s.repo.DeleteTemplate(ctx, id)
		if err != nil {
			return fmt.Errorf("failed to delete template: %w", err)
		}

		// Отправить задачу на удаление из индекса
		deleteJob := domain.NewTemplateDeleteJob(id)
		if err := s.queue.PublishMessage(ctx, deleteJob); err != nil {
			return fmt.Errorf("failed to publish template delete job: %w", err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

// countFieldsInSchema подсчитывает количество полей в JSON схеме
func countFieldsInSchema(fieldsSchema string) int {
	// Простой подсчет через поиск "name": в JSON
	// Для более точного подсчета можно распарсить JSON
	if fieldsSchema == "" {
		return 0
	}
	// Примерная оценка - можно улучшить
	return strings.Count(fieldsSchema, `"name"`)
}
