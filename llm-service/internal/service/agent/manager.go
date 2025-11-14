package agent

import (
	"context"
	"fmt"
	"llm-service/internal/domain"
	"llm-service/internal/service"
	"llm-service/internal/service/tool"
	"sync"
)

// Manager - менеджер агентов, загружает определения из кодовых реестров
type Manager struct {
	mu        sync.RWMutex
	agents    map[string]*domain.AgentDefinition
	tools     map[domain.ToolName]*domain.ToolDefinition
	mcpClient service.MCPClient
}

// NewManager создает новый менеджер агентов
func NewManager() (*Manager, error) {
	m := &Manager{
		agents: GetAgentsRegistry(),
		tools:  tool.GetToolsRegistry(),
	}

	return m, nil
}

// SetMCPClient устанавливает MCP клиент для динамического получения инструментов
func (m *Manager) SetMCPClient(ctx context.Context, mcpClient service.MCPClient) error {
	m.mu.Lock()
	m.mcpClient = mcpClient
	m.mu.Unlock()

	// Синхронизируем инструменты сразу после установки клиента
	return m.syncMCPTools(ctx)
}

// syncMCPTools синхронизирует инструменты из MCP клиента
func (m *Manager) syncMCPTools(ctx context.Context) error {
	if m.mcpClient == nil {
		return nil
	}

	mcpTools, err := m.mcpClient.GetTools(ctx)
	if err != nil {
		return fmt.Errorf("failed to get MCP tools: %w", err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Добавляем MCP инструменты с префиксом
	for _, mcpTool := range mcpTools {
		m.tools[domain.ToolName(mcpTool.Name)] = mcpTool
	}

	return nil
}

// GetAgent получает определение агента по ключу
func (m *Manager) GetAgent(agentKey string) (*domain.AgentDefinition, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	agent, exists := m.agents[agentKey]
	if !exists {
		return nil, domain.NewNotFoundError(fmt.Sprintf("agent %s not found", agentKey))
	}

	return agent, nil
}

// ListAgents получает список всех доступных агентов
func (m *Manager) ListAgents() []*domain.AgentDefinition {
	m.mu.RLock()
	defer m.mu.RUnlock()

	agents := make([]*domain.AgentDefinition, 0, len(m.agents))
	for _, agent := range m.agents {
		agents = append(agents, agent)
	}

	return agents
}

// GetTool получает определение инструмента по имени
func (m *Manager) GetTool(toolName domain.ToolName) (*domain.ToolDefinition, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tool, exists := m.tools[toolName]
	if !exists {
		return nil, domain.NewNotFoundError(fmt.Sprintf("tool %s not found", toolName))
	}

	return tool, nil
}

// ListTools получает список всех доступных инструментов
func (m *Manager) ListTools() []*domain.ToolDefinition {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tools := make([]*domain.ToolDefinition, 0, len(m.tools))
	for _, tool := range m.tools {
		tools = append(tools, tool)
	}

	return tools
}

// GetAgentTools получает инструменты для конкретного агента
func (m *Manager) GetAgentTools(agentKey string) ([]*domain.ToolDefinition, error) {
	agent, err := m.GetAgent(agentKey)
	if err != nil {
		return nil, err
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	allowedToolNames := agent.GetAllowedToolNames()
	tools := make([]*domain.ToolDefinition, 0)

	// Проверяем каждый инструмент из реестра
	for toolName, tool := range m.tools {
		// Проверяем прямое совпадение или паттерн
		for _, allowedPattern := range allowedToolNames {
			if domain.MatchesToolPattern(string(toolName), string(allowedPattern)) {
				tools = append(tools, tool)
				break
			}
		}
	}

	return tools, nil
}
