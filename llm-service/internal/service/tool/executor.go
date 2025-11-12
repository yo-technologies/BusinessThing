package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"llm-service/internal/domain"
	"llm-service/internal/service"

	"github.com/opentracing/opentracing-go"
)

// Executor - исполнитель инструментов
type Executor struct {
	agentManager    service.AgentManager
	subagentManager service.SubagentManager
	// Здесь будут добавляться конкретные исполнители для каждого типа tools
}

// NewExecutor создает новый executor для инструментов
func NewExecutor(
	agentManager service.AgentManager,
	subagentManager service.SubagentManager,
) *Executor {
	return &Executor{
		agentManager:    agentManager,
		subagentManager: subagentManager,
	}
}

// Execute выполняет инструмент с заданными параметрами
func (e *Executor) Execute(
	ctx context.Context,
	toolName string,
	arguments map[string]interface{},
	execCtx *domain.ExecutionContext,
	toolCallID *domain.ID,
) (interface{}, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "tool.Executor.Execute")
	defer span.Finish()

	// Проверяем разрешения
	if !e.CanExecute(toolName, execCtx.AgentKey) {
		return nil, domain.NewForbiddenError(fmt.Sprintf("agent %s is not allowed to use tool %s", execCtx.AgentKey, toolName))
	}

	// Роутинг по типу инструмента
	switch domain.ToolName(toolName) {
	case domain.ToolNameSwitchToSubagent:
		return e.executeSwitchToSubagent(ctx, arguments, execCtx, toolCallID)
	case domain.ToolNameFinishSubagent:
		return e.executeFinishSubagent(ctx, arguments, execCtx)
	default:
		return nil, domain.NewInvalidArgumentError(fmt.Sprintf("unknown tool: %s", toolName))
	}
}

// CanExecute проверяет, может ли инструмент быть выполнен агентом
func (e *Executor) CanExecute(toolName string, agentKey string) bool {
	// Получаем инструменты агента
	tools, err := e.agentManager.GetAgentTools(agentKey)
	if err != nil {
		return false
	}

	// Проверяем, есть ли такой инструмент у агента
	for _, tool := range tools {
		if tool.Name == toolName {
			return true
		}
	}

	return false
}

// executeSwitchToSubagent выполняет переключение на субагента
func (e *Executor) executeSwitchToSubagent(
	ctx context.Context,
	arguments map[string]interface{},
	execCtx *domain.ExecutionContext,
	toolCallID *domain.ID,
) (interface{}, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "tool.Executor.executeSwitchToSubagent")
	defer span.Finish()

	// Парсим аргументы
	subagentKey, ok := arguments["subagent_key"].(string)
	if !ok {
		return nil, domain.NewInvalidArgumentError("subagent_key is required")
	}

	task, ok := arguments["task"].(string)
	if !ok {
		return nil, domain.NewInvalidArgumentError("task is required")
	}

	// Переключаемся на субагента
	childChat, err := e.subagentManager.SwitchToSubagent(ctx, execCtx.ChatID, subagentKey, task, toolCallID)
	if err != nil {
		return nil, err
	}

	// Возвращаем информацию о созданном субагенте
	result := map[string]interface{}{
		"status":       "subagent_started",
		"subagent_key": subagentKey,
		"chat_id":      childChat.ID, // возвращаем domain.ID напрямую
		"message":      fmt.Sprintf("Субагент %s начал работу над задачей", subagentKey),
	}

	return result, nil
}

// executeFinishSubagent завершает работу субагента
func (e *Executor) executeFinishSubagent(
	ctx context.Context,
	arguments map[string]interface{},
	execCtx *domain.ExecutionContext,
) (interface{}, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "tool.Executor.executeFinishSubagent")
	defer span.Finish()

	// Парсим аргументы
	summary, ok := arguments["summary"].(string)
	if !ok {
		return nil, domain.NewInvalidArgumentError("summary is required")
	}

	// Завершаем работу субагента
	if err := e.subagentManager.FinishSubagent(ctx, execCtx.ChatID, summary); err != nil {
		return nil, err
	}

	// Возвращаем summary
	result := map[string]interface{}{
		"status":  "completed",
		"summary": summary,
	}

	return result, nil
}

// MarshalToolResult преобразует результат tool в JSON
func MarshalToolResult(result interface{}) (json.RawMessage, error) {
	data, err := json.Marshal(result)
	if err != nil {
		return nil, domain.NewInternalError("failed to marshal tool result", err)
	}
	return data, nil
}
