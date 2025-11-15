package document

import (
	"context"
	"fmt"

	"docs-processor/internal/domain"
	"docs-processor/internal/service"
	desc "docs-processor/pkg/document"

	"github.com/opentracing/opentracing-go"
)

type Service struct {
	desc.UnimplementedDocumentServiceServer
	searchService     *service.SearchService
	templateProcessor *service.TemplateProcessor
}

func NewService(searchService *service.SearchService, templateProcessor *service.TemplateProcessor) *Service {
	return &Service{
		searchService:     searchService,
		templateProcessor: templateProcessor,
	}
}

func (s *Service) SearchChunks(ctx context.Context, req *desc.SearchChunksRequest) (*desc.SearchChunksResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.document.Service.SearchChunks")
	defer span.Finish()

	organizationID, err := domain.ParseID(req.GetOrganizationId())
	if err != nil {
		return nil, fmt.Errorf("invalid organization_id: %w", err)
	}

	limit := int(req.Limit)
	if limit == 0 {
		limit = 10
	}
	minScore := req.MinScore
	if minScore == 0 {
		minScore = 0.5
	}

	results, err := s.searchService.SearchChunks(ctx, organizationID, req.Query, limit, minScore)
	if err != nil {
		return nil, err
	}

	chunks := make([]*desc.DocumentChunk, len(results))
	for i, result := range results {
		chunks[i] = &desc.DocumentChunk{
			ChunkId:      result.ChunkID.String(),
			DocumentId:   result.DocumentID.String(),
			DocumentName: result.DocumentName,
			Content:      result.Content,
			Position:     int32(result.Position),
			Score:        result.Score,
			Metadata:     result.Metadata,
		}
	}

	return &desc.SearchChunksResponse{
		Chunks: chunks,
	}, nil
}

func (s *Service) SearchTemplates(ctx context.Context, req *desc.SearchTemplatesRequest) (*desc.SearchTemplatesResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.document.Service.SearchTemplates")
	defer span.Finish()

	organizationID, err := domain.ParseID(req.GetOrganizationId())
	if err != nil {
		return nil, fmt.Errorf("invalid organization_id: %w", err)
	}

	limit := int(req.Limit)
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	results, err := s.templateProcessor.SearchTemplates(ctx, organizationID, req.Query, limit)
	if err != nil {
		return nil, err
	}

	templates := make([]*desc.TemplateSearchResult, len(results))
	for i, r := range results {
		templates[i] = &desc.TemplateSearchResult{
			TemplateId:   r.TemplateID.String(),
			Name:         r.Name,
			Description:  r.Description,
			TemplateType: r.TemplateType,
			FieldsCount:  int32(r.FieldsCount),
			Score:        r.Score,
		}
	}

	return &desc.SearchTemplatesResponse{
		Templates: templates,
	}, nil
}
