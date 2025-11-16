package service

import (
	"context"
	"llm-service/internal/contracts"
	"llm-service/internal/domain"
	"llm-service/internal/domain/dto"
)

// ChatManager - сервис для управления чатами
type ChatManager interface {
	// CreateChat создает новый чат
	CreateChat(ctx context.Context, req dto.CreateChatDTO) (*domain.Chat, error)

	// GetChat получает чат по ID
	GetChat(ctx context.Context, chatID domain.ID) (*domain.Chat, error)

	// ListChats получает список чатов пользователя
	ListChats(ctx context.Context, organizationID, userID domain.ID, page, pageSize int) ([]*domain.Chat, int, error)

	// DeleteChat удаляет чат
	DeleteChat(ctx context.Context, chatID, userID, orgID domain.ID) error

	// UpdateChat обновляет чат
	UpdateChat(ctx context.Context, chat *domain.Chat) error

	// GetMessages получает сообщения чата
	GetMessages(ctx context.Context, chatID, userID, orgID domain.ID, limit, offset int) ([]*domain.Message, int, error)

	// SaveMessage сохраняет сообщение
	SaveMessage(ctx context.Context, message *domain.Message) error

	// GetChatWithMessages получает чат со всеми сообщениями
	GetChatWithMessages(ctx context.Context, chatID domain.ID) (*domain.Chat, []*domain.Message, error)

	// GetActiveChildChat получает активный дочерний чат
	GetActiveChildChat(ctx context.Context, parentChatID domain.ID) (*domain.Chat, error)
}

// AgentManager - сервис для управления агентами
type AgentManager interface {
	// GetAgent получает определение агента по ключу
	GetAgent(agentKey string) (*domain.AgentDefinition, error)

	// ListAgents получает список всех доступных агентов
	ListAgents() []*domain.AgentDefinition

	// GetTool получает определение инструмента по имени
	GetTool(toolName domain.ToolName) (*domain.ToolDefinition, error)

	// ListTools получает список всех доступных инструментов
	ListTools() []*domain.ToolDefinition

	// GetAgentTools получает инструменты для конкретного агента
	GetAgentTools(agentKey string) ([]*domain.ToolDefinition, error)
}

// ContextBuilder - сервис для построения контекста для LLM
type ContextBuilder interface {
	// EnrichWithRAG обогащает контекст данными из RAG (векторный поиск)
	EnrichWithRAG(ctx context.Context, organizationID domain.ID, query string, limit int) (string, error)

	// EnrichWithOrganizationFacts добавляет факты об организации в контекст
	EnrichWithOrganizationFacts(ctx context.Context, organizationID domain.ID) (string, error)
}

// AgentExecutor - основной сервис для выполнения агентов
type AgentExecutor interface {
	// ExecuteStream выполняет агента с потоковой передачей результатов
	ExecuteStream(ctx context.Context, req dto.ExecuteAgentDTO, stream ExecutionStream) error

	// SendMessageStream отправляет сообщение с потоковым ответом
	SendMessageStream(ctx context.Context, req dto.SendMessageDTO, stream MessageStream) error
}

// ExecutionStream - интерфейс для потоковой передачи результатов выполнения
type ExecutionStream interface {
	SendChunk(content string) error
	SendMessage(message *domain.Message) error
	SendToolCall(toolCall *domain.ToolCall) error
	SendError(err error) error
}

// MessageStream - интерфейс для потоковой передачи сообщений
type MessageStream interface {
	SendChunk(content string) error
	SendMessage(message *domain.Message) error
	SendToolCall(toolCall *domain.ToolCall) error
	SendUsage(usage *dto.ChatUsageDTO) error
	SendError(err error) error
	SendChat(chat *domain.Chat) error
	SendFinal(chat *domain.Chat, messages []*domain.Message) error
}

// ToolExecutor - сервис для выполнения инструментов
type ToolExecutor interface {
	// Execute выполняет инструмент с заданными параметрами
	Execute(ctx context.Context, toolName string, arguments map[string]interface{}, execCtx *domain.ExecutionContext, toolCallID *domain.ID) (interface{}, error)

	// CanExecute проверяет, может ли инструмент быть выполнен
	CanExecute(toolName string, agentKey string) bool
}

// MCPClient - интерфейс для работы с MCP серверами
type MCPClient interface {
	// Initialize инициализирует соединение с MCP сервером
	Initialize(ctx context.Context) error

	// GetTools возвращает список инструментов от MCP сервера (с кешированием)
	GetTools(ctx context.Context) ([]*domain.ToolDefinition, error)

	// CallTool выполняет вызов инструмента на MCP сервере
	CallTool(ctx context.Context, toolName string, arguments map[string]interface{}) (interface{}, error)

	// HasTool проверяет, доступен ли инструмент на MCP сервере
	HasTool(ctx context.Context, toolName string) bool

	// GetTool получает определение конкретного инструмента
	GetTool(ctx context.Context, toolName string) (*domain.ToolDefinition, error)

	// RefreshTools принудительно обновляет кеш инструментов
	RefreshTools(ctx context.Context) error

	// Close закрывает соединение с MCP сервером
	Close() error
}

// SubagentManager - сервис для управления субагентами
type SubagentManager interface {
	// SwitchToSubagent переключает на субагента
	SwitchToSubagent(ctx context.Context, parentChatID domain.ID, subagentKey string, task string, parentToolCallID *domain.ID) (*domain.Chat, error)

	// FinishSubagent завершает работу субагента и возвращает summary
	FinishSubagent(ctx context.Context, subagentChatID domain.ID, summary string) error

	// GetParentChat получает родительский чат для субагента
	GetParentChat(ctx context.Context, subagentChatID domain.ID) (*domain.Chat, error)

	// GetActiveChatID возвращает ID активного чата (может быть субагент)
	GetActiveChatID(ctx context.Context, chatID domain.ID) (domain.ID, error)
}

// WebSearchClient - клиент для веб-поиска
type WebSearchClient interface {
	// Search выполняет поиск и возвращает результаты
	Search(ctx context.Context, query string, maxResults int) (interface{}, error)
}

// OrganizationMemoryService - сервис для управления фактами об организации
type OrganizationMemoryService interface {
	// AddFact добавляет факт об организации
	AddFact(ctx context.Context, organizationID domain.ID, content string) (domain.OrganizationMemoryFact, error)

	// ListFacts возвращает все факты организации
	ListFacts(ctx context.Context, organizationID domain.ID) ([]domain.OrganizationMemoryFact, error)

	// DeleteFact удаляет факт организации
	DeleteFact(ctx context.Context, organizationID domain.ID, factID domain.ID) error
}

// ContractSearchService - сервис для поиска шаблонов контрактов
type ContractSearchService interface {
	// SearchTemplates ищет подходящие шаблоны договоров
	SearchTemplates(ctx context.Context, organizationID, query string, limit int) ([]*contracts.TemplateSearchResult, error)
}

// ContractGeneratorService - сервис для генерации контрактов
type ContractGeneratorService interface {
	// GenerateContract генерирует договор из шаблона с заполненными данными
	GenerateContract(ctx context.Context, organizationID, templateID, contractName string, filledData map[string]interface{}) (*contracts.GeneratedContract, error)

	// ListContracts получает список сгенерированных контрактов
	ListContracts(ctx context.Context, organizationID string, limit, offset int) ([]*contracts.ContractListItem, int, error)
}

// QuotaService - сервис для управления квотами использования токенов
type QuotaService interface {
	// Confirm подтверждает использование токенов и сохраняет в БД
	Confirm(ctx context.Context, userID domain.ID, reserved int, actual int) error

	// GetLimits получает лимиты пользователя
	GetLimits(ctx context.Context, userID domain.ID) (domain.LLMLimits, error)
}
