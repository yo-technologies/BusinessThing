package domain

import (
	"slices"
	"strings"

	"github.com/samber/lo"
)

type ToolName string

const (
	// Системные инструменты
	ToolNameSwitchToSubagent ToolName = "switch_to_subagent"
	ToolNameFinishSubagent   ToolName = "finish_subagent"
	// Инструменты поиска
	ToolNameWebSearch ToolName = "web_search"
)

var AllowedToolNames = []ToolName{
	ToolNameWebSearch,
}

// AgentDefinition - декларативное определение агента из конфигурации
type AgentDefinition struct {
	Key              string     `yaml:"key" json:"key"`
	Name             string     `yaml:"name" json:"name"`
	Description      string     `yaml:"description" json:"description"`
	SystemPrompt     string     `yaml:"system_prompt" json:"system_prompt"`
	AllowedTools     []ToolName `yaml:"allowed_tools" json:"allowed_tools"`
	CanCallSubagents bool       `yaml:"can_call_subagents" json:"can_call_subagents"`
	IsSubagent       bool       `yaml:"is_subagent" json:"is_subagent"`
}

// GetSystemPrompt - возвращает system prompt для агента
func (ad *AgentDefinition) GetSystemPrompt() string {
	var promptBuilder strings.Builder
	promptBuilder.WriteString(ad.SystemPrompt)

	// Если агент - субагент, добавляем специальный инструктаж
	if ad.IsSubagent {
		promptBuilder.WriteString("\n\n")
		promptBuilder.WriteString("Вы являетесь субагентом и выполняете задачу, поставленную родительским агентом. ")
		promptBuilder.WriteString("Пожалуйста, сосредоточьтесь на выполнении этой задачи и по завеоршении предоставьте саммари, используя специальный инструмент.")
	}

	return promptBuilder.String()
}

// GetAllowedToolNames - возвращает список разрешенных инструментов
func (ad *AgentDefinition) GetAllowedToolNames() []ToolName {
	tools := lo.Filter(ad.AllowedTools, func(t ToolName, _ int) bool {
		return slices.Contains(AllowedToolNames, t)
	})

	// Если агент - субагент, добавляем специальный tool для завершения
	if ad.IsSubagent {
		tools = append(tools, ToolNameFinishSubagent)
	}

	// Если агент может вызывать субагентов, добавляем соответствующий tool
	if ad.CanCallSubagents {
		tools = append(tools, ToolNameSwitchToSubagent)
	}

	return tools
}

// CanUseTool - проверяет, может ли агент использовать указанный инструмент
func (ad *AgentDefinition) CanUseTool(toolName ToolName) bool {
	allowedTools := ad.GetAllowedToolNames()
	return slices.Contains(allowedTools, toolName)
}

// ExecutionContext - контекст выполнения агента
type ExecutionContext struct {
	OrganizationID    ID
	UserID            ID
	ChatID            ID
	AgentKey          string
	TaskDescription   string // для субагентов - описание задачи от родителя
	AdditionalContext map[string]any
}

// IsSubagentContext - проверяет, является ли контекст субагентом
func (ec *ExecutionContext) IsSubagentContext() bool {
	return ec.TaskDescription != ""
}
