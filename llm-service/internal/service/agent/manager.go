package agent

import (
	"fmt"
	"llm-service/internal/domain"
	"llm-service/internal/service/tool"
	"sync"
)

// Manager - менеджер агентов, загружает определения из кодовых реестров
type Manager struct {
	mu     sync.RWMutex
	agents map[string]*domain.AgentDefinition
	tools  map[domain.ToolName]*domain.ToolDefinition
}

// NewManager создает новый менеджер агентов
func NewManager() (*Manager, error) {
	m := &Manager{
		agents: GetAgentsRegistry(),
		tools:  tool.GetToolsRegistry(),
	}

	return m, nil
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
	tools := make([]*domain.ToolDefinition, 0, len(allowedToolNames))

	for _, toolName := range allowedToolNames {
		if tool, exists := m.tools[toolName]; exists {
			tools = append(tools, tool)
		}
	}

	return tools, nil
}
