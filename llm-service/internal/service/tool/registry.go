package tool

import "llm-service/internal/domain"

// GetToolsRegistry возвращает реестр всех доступных инструментов
func GetToolsRegistry() map[domain.ToolName]*domain.ToolDefinition {
	return map[domain.ToolName]*domain.ToolDefinition{
		// Инструменты поиска
		domain.ToolNameWebSearch: {
			Name:        string(domain.ToolNameWebSearch),
			Description: "Выполнить веб-поиск для получения актуальной информации из интернета",
			Parameters: map[string]interface{}{
				"query": map[string]interface{}{
					"type":        "string",
					"description": "Поисковый запрос для поиска информации в интернете",
				},
				"max_results": map[string]interface{}{
					"type":        "integer",
					"description": "Максимальное количество результатов (по умолчанию 5)",
				},
			},
			Required: []string{"query"},
		},
		// Инструменты памяти
		domain.ToolNameSaveOrganizationNote: {
			Name:        string(domain.ToolNameSaveOrganizationNote),
			Description: "Сохранить важный факт об организации для использования в будущих диалогах. Используй этот инструмент, когда узнаешь важную информацию об организации (ключевые процессы, особенности, предпочтения и т.д.)",
			Parameters: map[string]interface{}{
				"content": map[string]interface{}{
					"type":        "string",
					"description": "Краткий факт об организации (максимум 100 символов)",
				},
			},
			Required: []string{"content"},
		},
		// Системные инструменты
		domain.ToolNameSwitchToSubagent: {
			Name:        string(domain.ToolNameSwitchToSubagent),
			Description: "Переключиться на специализированного субагента для решения задачи",
			Parameters: map[string]interface{}{
				"subagent_key": map[string]interface{}{
					"type":        "string",
					"description": "Ключ субагента (marketing_agent, legal_agent, finance_agent)",
					"enum":        []string{"marketing_agent", "legal_agent", "finance_agent"},
				},
				"task": map[string]interface{}{
					"type":        "string",
					"description": "Описание задачи для субагента",
				},
			},
			Required: []string{"subagent_key", "task"},
		},
		domain.ToolNameFinishSubagent: {
			Name:        string(domain.ToolNameFinishSubagent),
			Description: "Завершить работу субагента и вернуться к основному агенту",
			Parameters: map[string]interface{}{
				"summary": map[string]interface{}{
					"type":        "string",
					"description": "Краткое описание выполненной работы и результатов",
				},
			},
			Required: []string{"summary"},
		},
	}
}
