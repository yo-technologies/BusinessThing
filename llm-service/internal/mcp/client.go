package mcp

import (
	"context"
	"fmt"
	"maps"
	"sync"
	"time"

	"llm-service/internal/domain"
	"llm-service/internal/logger"

	mcpclient "github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/opentracing/opentracing-go"
)

// Client - клиент для работы с MCP серверами (например, AmoCRM)
type Client struct {
	mcpClient *mcpclient.Client
	address   string

	// Кеш определений инструментов
	mu    sync.RWMutex
	tools map[string]*domain.ToolDefinition
	// Время последнего обновления кеша
	lastUpdate time.Time
	// TTL кеша (по умолчанию 5 минут)
	cacheTTL time.Duration
}

// NewSSEClient создает новый MCP клиент через SSE транспорт
// baseURL - базовый URL MCP сервера (например, "http://localhost:3000")
func NewSSEClient(baseURL string) (*Client, error) {
	// Для MCP используем SSE транспорт
	mcpClient, err := mcpclient.NewSSEMCPClient(baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create SSE MCP client: %w", err)
	}

	client := &Client{
		mcpClient:  mcpClient,
		address:    baseURL,
		tools:      make(map[string]*domain.ToolDefinition),
		cacheTTL:   5 * time.Minute,
		lastUpdate: time.Time{},
	}

	return client, nil
}

// Close закрывает соединение с MCP сервером
func (c *Client) Close() error {
	if c.mcpClient != nil {
		return c.mcpClient.Close()
	}
	return nil
}

// Initialize инициализирует соединение с MCP сервером
func (c *Client) Initialize(ctx context.Context) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "mcp.Client.Initialize")
	defer span.Finish()

	// Инициализируем соединение
	err := c.mcpClient.Start(ctx)
	if err != nil {
		return fmt.Errorf("failed to start MCP client: %w", err)
	}

	_, err = c.mcpClient.Initialize(ctx, mcp.InitializeRequest{
		Params: mcp.InitializeParams{
			ClientInfo: mcp.Implementation{
				Name:    "llm-service",
				Version: "1.0.0",
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to initialize MCP connection: %w", err)
	}

	logger.Info(ctx, "MCP client initialized successfully")
	return nil
}

// GetTools возвращает список инструментов от MCP сервера (с кешированием)
func (c *Client) GetTools(ctx context.Context) ([]*domain.ToolDefinition, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "mcp.Client.GetTools")
	defer span.Finish()

	c.mu.RLock()
	// Проверяем, актуален ли кеш
	if time.Since(c.lastUpdate) < c.cacheTTL && len(c.tools) > 0 {
		tools := make([]*domain.ToolDefinition, 0, len(c.tools))
		for _, tool := range c.tools {
			tools = append(tools, tool)
		}
		c.mu.RUnlock()
		logger.Debugf(ctx, "Returning cached MCP tools, count: %d", len(tools))
		return tools, nil
	}
	c.mu.RUnlock()

	// Кеш устарел или пуст, запрашиваем инструменты
	logger.Info(ctx, "Fetching tools from MCP server")
	tools, err := c.fetchTools(ctx)
	if err != nil {
		return nil, err
	}

	// Обновляем кеш
	c.mu.Lock()
	c.tools = make(map[string]*domain.ToolDefinition)
	for _, tool := range tools {
		c.tools[tool.Name] = tool
	}
	c.lastUpdate = time.Now()
	c.mu.Unlock()

	logger.Infof(ctx, "MCP tools cache updated, count: %d", len(tools))
	return tools, nil
}

// fetchTools запрашивает список инструментов с MCP сервера
func (c *Client) fetchTools(ctx context.Context) ([]*domain.ToolDefinition, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "mcp.Client.fetchTools")
	defer span.Finish()

	// Запрашиваем список инструментов через MCP протокол
	resp, err := c.mcpClient.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		return nil, fmt.Errorf("failed to list MCP tools: %w", err)
	}

	// Конвертируем MCP инструменты в доменные модели
	tools := make([]*domain.ToolDefinition, 0, len(resp.Tools))
	for _, mcpTool := range resp.Tools {
		tool := convertMCPToolToDomain(&mcpTool)
		tools = append(tools, tool)
	}

	return tools, nil
}

// CallTool выполняет вызов инструмента на MCP сервере
func (c *Client) CallTool(ctx context.Context, toolName string, arguments map[string]interface{}) (interface{}, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "mcp.Client.CallTool")
	defer span.Finish()

	span.SetTag("tool_name", toolName)

	logger.Infof(ctx, "Calling MCP tool: %s", toolName)

	// Вызываем инструмент через MCP протокол
	resp, err := c.mcpClient.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      toolName,
			Arguments: arguments,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to call MCP tool: %w", err)
	}

	// Обрабатываем результат
	if resp.IsError {
		return nil, fmt.Errorf("MCP tool returned error")
	}

	// Формируем результат из контента
	result := formatMCPToolResult(resp)
	logger.Infof(ctx, "MCP tool %s executed successfully", toolName)

	return result, nil
}

// HasTool проверяет, доступен ли инструмент на MCP сервере
func (c *Client) HasTool(ctx context.Context, toolName string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	_, exists := c.tools[toolName]
	return exists
}

// GetTool получает определение конкретного инструмента
func (c *Client) GetTool(ctx context.Context, toolName string) (*domain.ToolDefinition, error) {
	c.mu.RLock()
	tool, exists := c.tools[toolName]
	c.mu.RUnlock()

	if !exists {
		// Пробуем обновить кеш
		if _, err := c.GetTools(ctx); err != nil {
			return nil, fmt.Errorf("failed to fetch tools: %w", err)
		}

		c.mu.RLock()
		tool, exists = c.tools[toolName]
		c.mu.RUnlock()

		if !exists {
			return nil, domain.NewNotFoundError(fmt.Sprintf("MCP tool %s not found", toolName))
		}
	}

	return tool, nil
}

// InvalidateCache принудительно очищает кеш инструментов
func (c *Client) InvalidateCache() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.tools = make(map[string]*domain.ToolDefinition)
	c.lastUpdate = time.Time{}
	logger.Info(context.Background(), "MCP tools cache invalidated")
}

// RefreshTools принудительно обновляет кеш инструментов
func (c *Client) RefreshTools(ctx context.Context) error {
	c.InvalidateCache()
	_, err := c.GetTools(ctx)
	return err
}

// convertMCPToolToDomain конвертирует MCP определение инструмента в доменную модель
func convertMCPToolToDomain(mcpTool *mcp.Tool) *domain.ToolDefinition {
	// Парсим required поля из JSON schema
	required := []string{}
	if mcpTool.InputSchema.Required != nil {
		required = mcpTool.InputSchema.Required
	}

	// Извлекаем properties из inputSchema
	properties := make(map[string]interface{})
	if mcpTool.InputSchema.Properties != nil {
		maps.Copy(properties, mcpTool.InputSchema.Properties)
	}

	return &domain.ToolDefinition{
		Name:        "ammo-crm-" + mcpTool.Name,
		Description: mcpTool.Description,
		Parameters:  properties,
		Required:    required,
	}
}

// formatMCPToolResult форматирует результат вызова MCP инструмента
func formatMCPToolResult(resp *mcp.CallToolResult) interface{} {
	// Если есть контент, собираем его в строку
	if len(resp.Content) > 0 {
		result := make(map[string]interface{})

		// Собираем текстовое содержимое
		textParts := []string{}
		for _, content := range resp.Content {
			// Пытаемся привести к TextContent
			if tc, ok := mcp.AsTextContent(content); ok {
				if tc != nil {
					textParts = append(textParts, tc.Text)
				}
			}
		}

		if len(textParts) > 0 {
			result["text"] = ""
			for i, part := range textParts {
				if i > 0 {
					result["text"] = result["text"].(string) + "\n"
				}
				result["text"] = result["text"].(string) + part
			}
		}

		// Если есть другие типы контента, добавляем их отдельно
		result["content"] = resp.Content

		return result
	}

	return map[string]interface{}{
		"status": "success",
	}
}
