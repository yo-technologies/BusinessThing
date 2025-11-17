package contracts

import (
	"context"
	"encoding/json"
	"fmt"

	"llm-service/internal/coreservice"
	docsproc "llm-service/pkg/document"

	"github.com/opentracing/opentracing-go"
)

type SearchService struct {
	docsProcessorClient docsproc.DocumentServiceClient
	coreServiceClient   coreservice.Client
}

func NewSearchService(
	docsProcessorClient docsproc.DocumentServiceClient,
	coreServiceClient coreservice.Client,
) *SearchService {
	return &SearchService{
		docsProcessorClient: docsProcessorClient,
		coreServiceClient:   coreServiceClient,
	}
}

// SearchTemplates ищет подходящие шаблоны договоров
func (s *SearchService) SearchTemplates(ctx context.Context, query string, limit int) ([]*TemplateSearchResult, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "contracts.SearchService.SearchTemplates")
	defer span.Finish()

	// Поиск через docs-processor
	resp, err := s.docsProcessorClient.SearchTemplates(ctx, &docsproc.SearchTemplatesRequest{
		Query: query,
		Limit: int32(limit),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to search templates: %w", err)
	}

	results := make([]*TemplateSearchResult, 0, len(resp.Templates))
	for _, t := range resp.Templates {
		// Получаем полную информацию о шаблоне из core-service
		template, err := s.coreServiceClient.GetTemplate(ctx, t.TemplateId)
		if err != nil {
			// Логируем ошибку, но продолжаем с частичными данными
			results = append(results, &TemplateSearchResult{
				TemplateID:   t.TemplateId,
				Name:         t.Name,
				Description:  t.Description,
				TemplateType: t.TemplateType,
				Fields:       []TemplateField{},
				FieldsCount:  int(t.FieldsCount),
				Score:        t.Score,
			})
			continue
		}

		// Парсим fields_schema
		var fieldsSchema struct {
			Fields []TemplateField `json:"fields"`
		}
		if err := json.Unmarshal([]byte(template.FieldsSchema), &fieldsSchema); err == nil {
			results = append(results, &TemplateSearchResult{
				TemplateID:   t.TemplateId,
				Name:         t.Name,
				Description:  t.Description,
				TemplateType: t.TemplateType,
				Fields:       fieldsSchema.Fields,
				FieldsCount:  int(t.FieldsCount),
				Score:        t.Score,
			})
		} else {
			results = append(results, &TemplateSearchResult{
				TemplateID:   t.TemplateId,
				Name:         t.Name,
				Description:  t.Description,
				TemplateType: t.TemplateType,
				Fields:       []TemplateField{},
				FieldsCount:  int(t.FieldsCount),
				Score:        t.Score,
			})
		}
	}

	return results, nil
}
