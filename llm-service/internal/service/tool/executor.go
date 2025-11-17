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
	agentManager             service.AgentManager
	subagentManager          service.SubagentManager
	websearchClient          service.WebSearchClient
	orgMemoryService         service.OrganizationMemoryService
	mcpClient                service.MCPClient
	contractSearchService    service.ContractSearchService
	contractGeneratorService service.ContractGeneratorService
}

// NewExecutor создает новый executor для инструментов
func NewExecutor(
	agentManager service.AgentManager,
	subagentManager service.SubagentManager,
	websearchClient service.WebSearchClient,
	orgMemoryService service.OrganizationMemoryService,
	mcpClient service.MCPClient,
	contractSearchService service.ContractSearchService,
	contractGeneratorService service.ContractGeneratorService,
) *Executor {
	return &Executor{
		agentManager:             agentManager,
		subagentManager:          subagentManager,
		websearchClient:          websearchClient,
		orgMemoryService:         orgMemoryService,
		mcpClient:                mcpClient,
		contractSearchService:    contractSearchService,
		contractGeneratorService: contractGeneratorService,
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

	// Проверяем, является ли инструмент MCP инструментом
	if domain.IsMCPTool(toolName) {
		// Извлекаем имя инструмента без префикса
		mcpToolName := domain.GetMCPToolName(toolName)
		return e.executeMCPTool(ctx, mcpToolName, arguments)
	}

	// Роутинг по типу инструмента
	switch domain.ToolName(toolName) {
	case domain.ToolNameWebSearch:
		return e.executeWebSearch(ctx, arguments)
	case domain.ToolNameSaveOrganizationNote:
		return e.executeSaveOrganizationNote(ctx, arguments, execCtx)
	case domain.ToolNameSwitchToSubagent:
		return e.executeSwitchToSubagent(ctx, arguments, execCtx, toolCallID)
	case domain.ToolNameFinishSubagent:
		return e.executeFinishSubagent(ctx, arguments, execCtx)
	case domain.ToolNameSearchContractTemplates:
		return e.executeSearchContractTemplates(ctx, arguments, execCtx)
	case domain.ToolNameGenerateContract:
		return e.executeGenerateContract(ctx, arguments, execCtx)
	case domain.ToolNameListGeneratedContracts:
		return e.executeListGeneratedContracts(ctx, arguments, execCtx)
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

// executeWebSearch выполняет веб-поиск
func (e *Executor) executeWebSearch(
	ctx context.Context,
	arguments map[string]interface{},
) (interface{}, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "tool.Executor.executeWebSearch")
	defer span.Finish()

	// Парсим аргументы
	query, ok := arguments["query"].(string)
	if !ok || query == "" {
		return nil, domain.NewInvalidArgumentError("query is required and must be a non-empty string")
	}

	maxResults := 5 // default
	if mr, ok := arguments["max_results"].(float64); ok {
		maxResults = int(mr)
	}

	// Выполняем поиск
	searchResult, err := e.websearchClient.Search(ctx, query, maxResults)
	if err != nil {
		return nil, err
	}

	return searchResult, nil
}

// executeSaveOrganizationNote сохраняет заметку об организации
func (e *Executor) executeSaveOrganizationNote(
	ctx context.Context,
	arguments map[string]interface{},
	execCtx *domain.ExecutionContext,
) (interface{}, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "tool.Executor.executeSaveOrganizationNote")
	defer span.Finish()

	// Парсим аргументы
	content, ok := arguments["content"].(string)
	if !ok || content == "" {
		return nil, domain.NewInvalidArgumentError("content is required and must be a non-empty string")
	}

	// Сохраняем факт об организации
	fact, err := e.orgMemoryService.AddFact(ctx, execCtx.OrganizationID, content)
	if err != nil {
		return nil, err
	}

	// Возвращаем результат
	result := map[string]interface{}{
		"status":  "success",
		"message": "Факт об организации успешно сохранён",
		"fact_id": fact.ID.String(),
	}

	return result, nil
}

// executeMCPTool выполняет вызов MCP инструмента
func (e *Executor) executeMCPTool(
	ctx context.Context,
	toolName string,
	arguments map[string]interface{},
) (interface{}, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "tool.Executor.executeMCPTool")
	defer span.Finish()

	// Проверяем, что MCP клиент доступен
	if e.mcpClient == nil {
		return nil, domain.NewInternalError("MCP client is not initialized", nil)
	}

	// Вызываем инструмент через MCP клиент
	result, err := e.mcpClient.CallTool(ctx, toolName, arguments)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// executeSearchContractTemplates выполняет поиск шаблонов контрактов
func (e *Executor) executeSearchContractTemplates(
	ctx context.Context,
	arguments map[string]interface{},
	execCtx *domain.ExecutionContext,
) (interface{}, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "tool.Executor.executeSearchContractTemplates")
	defer span.Finish()

	// Парсим аргументы
	query, ok := arguments["query"].(string)
	if !ok || query == "" {
		return nil, domain.NewInvalidArgumentError("query is required and must be a non-empty string")
	}

	limit := 5
	if limitArg, ok := arguments["limit"].(float64); ok {
		limit = int(limitArg)
	}

	// Выполняем поиск
	result, err := e.contractSearchService.SearchTemplates(ctx, query, limit)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// executeGenerateContract выполняет генерацию контракта из шаблона
func (e *Executor) executeGenerateContract(
	ctx context.Context,
	arguments map[string]interface{},
	execCtx *domain.ExecutionContext,
) (interface{}, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "tool.Executor.executeGenerateContract")
	defer span.Finish()

	// Парсим аргументы
	templateID, ok := arguments["template_id"].(string)
	if !ok || templateID == "" {
		return nil, domain.NewInvalidArgumentError("template_id is required")
	}

	contractName, ok := arguments["contract_name"].(string)
	if !ok || contractName == "" {
		return nil, domain.NewInvalidArgumentError("contract_name is required")
	}

	filledData, ok := arguments["filled_data"].(map[string]interface{})
	if !ok {
		return nil, domain.NewInvalidArgumentError("filled_data is required and must be an object")
	}

	// Генерируем контракт
	result, err := e.contractGeneratorService.GenerateContract(
		ctx,
		execCtx.OrganizationID.String(),
		templateID,
		contractName,
		filledData,
	)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// executeListGeneratedContracts получает список сгенерированных контрактов
func (e *Executor) executeListGeneratedContracts(
	ctx context.Context,
	arguments map[string]interface{},
	execCtx *domain.ExecutionContext,
) (interface{}, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "tool.Executor.executeListGeneratedContracts")
	defer span.Finish()

	// Парсим аргументы
	limit := 20
	if limitArg, ok := arguments["limit"].(float64); ok {
		limit = int(limitArg)
	}

	offset := 0
	if offsetArg, ok := arguments["offset"].(float64); ok {
		offset = int(offsetArg)
	}

	// Получаем список контрактов
	contracts, total, err := e.contractGeneratorService.ListContracts(ctx, execCtx.OrganizationID.String(), limit, offset)
	if err != nil {
		return nil, err
	}

	// Возвращаем в виде map для удобства использования в tools
	return map[string]interface{}{
		"contracts": contracts,
		"total":     total,
		"limit":     limit,
		"offset":    offset,
	}, nil
}
