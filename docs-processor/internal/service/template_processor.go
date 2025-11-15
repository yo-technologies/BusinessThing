package service

import (
	"context"
	"fmt"
	"time"

	"docs-processor/internal/domain"
	"docs-processor/internal/embeddings"
	"docs-processor/internal/logger"
	"docs-processor/internal/vectordb"

	"github.com/opentracing/opentracing-go"
	"github.com/samber/lo"
)

type TemplateProcessor struct {
	embeddingsClient *embeddings.Client
	templatesDB      *vectordb.TemplatesClient
}

func NewTemplateProcessor(
	embeddingsClient *embeddings.Client,
	templatesDB *vectordb.TemplatesClient,
) *TemplateProcessor {
	return &TemplateProcessor{
		embeddingsClient: embeddingsClient,
		templatesDB:      templatesDB,
	}
}

func (p *TemplateProcessor) IndexTemplate(ctx context.Context, job *domain.ProcessingJob) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.TemplateProcessor.IndexTemplate")
	defer span.Finish()

	if job.TemplateID == nil || job.TemplateName == nil || job.Description == nil {
		return fmt.Errorf("invalid template index job: missing required fields")
	}

	logger.Info(ctx, "Indexing template",
		"template_id", job.TemplateID.String(),
		"organization_id", job.OrganizationID.String(),
		"name", *job.TemplateName,
	)

	// Векторизируем название + описание
	text := *job.TemplateName
	if *job.Description != "" {
		text += " " + *job.Description
	}

	embedding, err := p.embeddingsClient.GenerateEmbedding(ctx, text)
	if err != nil {
		logger.Error(ctx, "Failed to get embedding", "error", err)
		return fmt.Errorf("failed to get embedding: %w", err)
	}

	template := &domain.Template{
		ID:             *job.TemplateID,
		OrganizationID: job.OrganizationID,
		Name:           *job.TemplateName,
		Description:    *job.Description,
		TemplateType:   lo.FromPtr(job.TemplateType),
		FieldsCount:    lo.FromPtr(job.FieldsCount),
		CreatedAt:      time.Now(),
	}

	if err := p.templatesDB.IndexTemplate(ctx, template, embedding); err != nil {
		logger.Error(ctx, "Failed to index template", "error", err)
		return fmt.Errorf("failed to index template: %w", err)
	}

	logger.Info(ctx, "Template indexed successfully",
		"template_id", job.TemplateID.String(),
	)

	return nil
}

func (p *TemplateProcessor) DeleteTemplate(ctx context.Context, job *domain.ProcessingJob) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.TemplateProcessor.DeleteTemplate")
	defer span.Finish()

	if job.TemplateID == nil {
		return fmt.Errorf("invalid template delete job: missing template_id")
	}

	logger.Info(ctx, "Deleting template from index",
		"template_id", job.TemplateID.String(),
	)

	if err := p.templatesDB.DeleteTemplate(ctx, *job.TemplateID); err != nil {
		logger.Error(ctx, "Failed to delete template", "error", err)
		return fmt.Errorf("failed to delete template: %w", err)
	}

	logger.Info(ctx, "Template deleted successfully",
		"template_id", job.TemplateID.String(),
	)

	return nil
}

func (p *TemplateProcessor) SearchTemplates(ctx context.Context, organizationID domain.ID, query string, limit int) ([]*vectordb.TemplateSearchResult, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.TemplateProcessor.SearchTemplates")
	defer span.Finish()

	embedding, err := p.embeddingsClient.GenerateEmbedding(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get query embedding: %w", err)
	}

	results, err := p.templatesDB.SearchTemplates(ctx, organizationID, embedding, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search templates: %w", err)
	}

	return results, nil
}
