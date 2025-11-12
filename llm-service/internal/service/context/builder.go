package context

import (
	"context"
	"llm-service/internal/domain"
)

// Builder - построитель контекста для LLM
type Builder struct {
	// TODO: добавить зависимости для RAG, vector search, memory service
}

// NewBuilder создает новый builder контекста
func NewBuilder() *Builder {
	return &Builder{}
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
	// TODO: Реализовать интеграцию с vector search service
	// 1. Отправить запрос в vector search
	// 2. Получить релевантные фрагменты документов
	// 3. Форматировать для включения в контекст

	// Заглушка
	results := make([]string, 0)

	return results, nil
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
