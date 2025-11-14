package context

import (
	"context"
	"llm-service/internal/domain"
	"llm-service/internal/rag"
)

// Builder - построитель контекста для LLM
type Builder struct {
	ragClient *rag.Client
}

// NewBuilder создает новый builder контекста
func NewBuilder(ragClient *rag.Client) *Builder {
	return &Builder{
		ragClient: ragClient,
	}
}

// BuildContext строит контекст для агента на основе истории чата и дополнительных данных
func (b *Builder) BuildContext(
	ctx context.Context,
	chat *domain.Chat,
	messages []*domain.Message,
) ([]interface{}, error) {
	// TODO: Реализовать построение контекста
	// 1. Базовая история сообщений
	// 2. Обогащение через RAG
	// 3. Добавление фактов из памяти

	context := make([]interface{}, 0)

	// Пока просто возвращаем пустой контекст
	// Логика будет добавлена при интеграции с vector search и memory service

	return context, nil
}

// EnrichWithRAG обогащает контекст данными из RAG (векторный поиск)
func (b *Builder) EnrichWithRAG(
	ctx context.Context,
	organizationID domain.ID,
	query string,
	limit int,
) ([]string, error) {
	if b.ragClient == nil {
		return []string{}, nil
	}

	chunks, err := b.ragClient.SearchRelevantChunks(ctx, organizationID, query, limit, 0.9)
	if err != nil {
		return nil, err
	}

	return chunks, nil
}

// EnrichWithMemoryFacts добавляет факты из памяти пользователя
func (b *Builder) EnrichWithMemoryFacts(
	ctx context.Context,
	userID domain.ID,
) ([]string, error) {
	// TODO: Реализовать получение фактов из memory service
	// 1. Запросить факты пользователя
	// 2. Форматировать для включения в system prompt

	// Заглушка
	facts := make([]string, 0)

	return facts, nil
}
