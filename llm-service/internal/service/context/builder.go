package context

import (
	"context"
	"llm-service/internal/domain"
	"llm-service/internal/rag"
	"strings"
)

// Builder - построитель контекста для LLM
type Builder struct {
	ragClient        *rag.Client
	orgMemoryService OrganizationMemoryService
}

// OrganizationMemoryService - интерфейс для работы с фактами об организации
type OrganizationMemoryService interface {
	ListFacts(ctx context.Context, organizationID domain.ID) ([]domain.OrganizationMemoryFact, error)
}

// NewBuilder создает новый builder контекста
func NewBuilder(ragClient *rag.Client, orgMemoryService OrganizationMemoryService) *Builder {
	return &Builder{
		ragClient:        ragClient,
		orgMemoryService: orgMemoryService,
	}
}

// EnrichWithRAG обогащает контекст данными из RAG (векторный поиск)
func (b *Builder) EnrichWithRAG(
	ctx context.Context,
	organizationID domain.ID,
	query string,
	limit int,
) (string, error) {
	if b.ragClient == nil {
		return "", nil
	}

	chunks, err := b.ragClient.SearchRelevantChunks(ctx, organizationID, query, limit, 0.55)
	if err != nil {
		return "", err
	}

	var result strings.Builder
	result.WriteString("Релевантные фрагменты документов:\n")
	for _, chunk := range chunks {
		result.WriteString("- ")
		result.WriteString(chunk)
		result.WriteString("\n")
	}

	return result.String(), nil
}

// EnrichWithOrganizationFacts добавляет факты об организации в контекст
func (b *Builder) EnrichWithOrganizationFacts(
	ctx context.Context,
	organizationID domain.ID,
) (string, error) {
	if b.orgMemoryService == nil {
		return "", nil
	}

	facts, err := b.orgMemoryService.ListFacts(ctx, organizationID)
	if err != nil {
		return "", err
	}

	var result strings.Builder
	result.WriteString("Факты об организации:\n")
	for _, fact := range facts {
		result.WriteString("- ")
		result.WriteString(fact.Content)
		result.WriteString("\n")
	}

	return result.String(), nil
}
