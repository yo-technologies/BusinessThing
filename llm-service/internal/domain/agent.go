package domain

import (
	"slices"
	"strings"
)

type ToolName string

const (
	// Системные инструменты
	ToolNameSwitchToSubagent ToolName = "switch_to_subagent"
	ToolNameFinishSubagent   ToolName = "finish_subagent"
	// Инструменты поиска
	ToolNameWebSearch ToolName = "web_search"
	// Инструменты памяти
	ToolNameSaveOrganizationNote ToolName = "save_organization_note"
	// MCP инструменты (префикс для динамических инструментов)
)

const AmoCRMMCPToolPrefix = "ammo-crm-"

var mcpPrefixes = []string{
	AmoCRMMCPToolPrefix,
}

var AllowedToolNames = []ToolName{
	ToolNameWebSearch,
	ToolNameSaveOrganizationNote,
}

// IsMCPTool проверяет, является ли инструмент MCP инструментом
func IsMCPTool(toolName string) bool {
	for _, prefix := range mcpPrefixes {
		if strings.HasPrefix(toolName, prefix) {
			return true
		}
	}
	return false
}

// GetMCPToolName извлекает имя инструмента без префикса MCP
func GetMCPToolName(toolName string) string {
	if IsMCPTool(toolName) {
		for _, prefix := range mcpPrefixes {
			if strings.HasPrefix(toolName, prefix) {
				return toolName[len(prefix):]
			}
		}
	}
	return toolName
}

// MatchesToolPattern проверяет, соответствует ли имя инструмента паттерну
// Поддерживает паттерны с * в конце (например, "ammo-crm-*" для всех MCP инструментов)
func MatchesToolPattern(toolName, pattern string) bool {
	if pattern == string(toolName) {
		return true
	}

	// Проверяем паттерн со звездочкой
	if strings.HasSuffix(pattern, "*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(string(toolName), prefix)
	}

	return false
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
	tools := make([]ToolName, 0)

	// Добавляем разрешенные инструменты, включая MCP инструменты и паттерны
	for _, t := range ad.AllowedTools {
		// Проверяем, что это либо известный инструмент, либо MCP инструмент/паттерн
		if slices.Contains(AllowedToolNames, t) || IsMCPTool(string(t)) {
			tools = append(tools, t)
		}
	}

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
