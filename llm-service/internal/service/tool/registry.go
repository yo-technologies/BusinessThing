package tool

import "llm-service/internal/domain"

// GetToolsRegistry возвращает реестр всех доступных инструментов
func GetToolsRegistry() map[domain.ToolName]*domain.ToolDefinition {
	return map[domain.ToolName]*domain.ToolDefinition{
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
