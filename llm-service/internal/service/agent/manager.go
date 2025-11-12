package agent

import (
	"fmt"
	"llm-service/internal/config"
	"llm-service/internal/domain"
	"sync"

	"github.com/samber/lo"
)

// Manager - менеджер агентов, загружает определения из конфига
type Manager struct {
	mu     sync.RWMutex
	agents map[string]*domain.AgentDefinition
	tools  map[domain.ToolName]*domain.ToolDefinition
}

// NewManager создает новый менеджер агентов
func NewManager(cfg *config.Config) (*Manager, error) {
	m := &Manager{
		agents: make(map[string]*domain.AgentDefinition),
		tools:  make(map[domain.ToolName]*domain.ToolDefinition),
	}

	// Загружаем агентов из конфига
	if err := m.loadAgents(cfg); err != nil {
		return nil, fmt.Errorf("failed to load agents: %w", err)
	}

	// Загружаем инструменты из конфига
	if err := m.loadTools(cfg); err != nil {
		return nil, fmt.Errorf("failed to load tools: %w", err)
	}

	return m, nil
}

// loadAgents загружает определения агентов из конфига
func (m *Manager) loadAgents(cfg *config.Config) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for key, agentCfg := range cfg.Agents {
		allowedTools := lo.FilterMap(agentCfg.AllowedTools, func(t string, _ int) (domain.ToolName, bool) {
			toolName := domain.ToolName(t)
			if lo.Contains(domain.AllowedToolNames, toolName) {
				return toolName, true
			}
			return "", false
		})

		agent := &domain.AgentDefinition{
			Key:              agentCfg.Key,
			Name:             agentCfg.Name,
			Description:      agentCfg.Description,
			SystemPrompt:     agentCfg.SystemPrompt,
			Model:            agentCfg.Model,
			Temperature:      agentCfg.Temperature,
			AllowedTools:     allowedTools,
			CanCallSubagents: agentCfg.CanCallSubagents,
			IsSubagent:       agentCfg.IsSubagent,
		}

		m.agents[key] = agent
	}

	return nil
}

// loadTools загружает определения инструментов из конфига
func (m *Manager) loadTools(cfg *config.Config) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for name, toolCfg := range cfg.Tools {
		tool := &domain.ToolDefinition{
			Name:        toolCfg.Name,
			Description: toolCfg.Description,
			Parameters:  toolCfg.Parameters,
			Required:    toolCfg.Required,
		}

		m.tools[domain.ToolName(name)] = tool
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
	tools := make([]*domain.ToolDefinition, 0, len(allowedToolNames))

	for _, toolName := range allowedToolNames {
		if tool, exists := m.tools[toolName]; exists {
			tools = append(tools, tool)
		}
	}

	return tools, nil
}

// Reload перезагружает агентов и инструменты из конфига
func (m *Manager) Reload(cfg *config.Config) error {
	// Создаем временные мапы
	newAgents := make(map[string]*domain.AgentDefinition)
	newTools := make(map[domain.ToolName]*domain.ToolDefinition)

	// Загружаем в временные мапы
	for key, agentCfg := range cfg.Agents {
		allowedTools := lo.FilterMap(agentCfg.AllowedTools, func(t string, _ int) (domain.ToolName, bool) {
			toolName := domain.ToolName(t)
			if lo.Contains(domain.AllowedToolNames, toolName) {
				return toolName, true
			}
			return "", false
		})

		newAgents[key] = &domain.AgentDefinition{
			Key:              agentCfg.Key,
			Name:             agentCfg.Name,
			Description:      agentCfg.Description,
			SystemPrompt:     agentCfg.SystemPrompt,
			Model:            agentCfg.Model,
			Temperature:      agentCfg.Temperature,
			AllowedTools:     allowedTools,
			CanCallSubagents: agentCfg.CanCallSubagents,
			IsSubagent:       agentCfg.IsSubagent,
		}
	}

	for name, toolCfg := range cfg.Tools {
		newTools[domain.ToolName(name)] = &domain.ToolDefinition{
			Name:        toolCfg.Name,
			Description: toolCfg.Description,
			Parameters:  toolCfg.Parameters,
			Required:    toolCfg.Required,
		}
	}

	// Атомарно заменяем мапы
	m.mu.Lock()
	m.agents = newAgents
	m.tools = newTools
	m.mu.Unlock()

	return nil
}
