package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"llm-service/internal/domain"
	"llm-service/internal/llm"

	"github.com/opentracing/opentracing-go"
)

type agentToolHandler func(ctx context.Context, chatCtx domain.AgentChatContext, raw json.RawMessage) (string, error)

type agentTool struct {
	name                 string
	handler              agentToolHandler
	desc                 string
	params               map[string]any
	requiresConfirmation bool
}

// service defines dependencies used by Tools.
type service interface {
	AddFact(ctx context.Context, userID domain.ID, content string) (domain.UserMemoryFact, error)
}

// Tools provides chat agent tool definitions and dispatch.
type Tools struct {
	chatTools     map[string]agentTool
	chatToolsOnce sync.Once

	service service
}

func New(service service) *Tools {
	return &Tools{
		service:   service,
		chatTools: make(map[string]agentTool),
	}
}

func (t *Tools) ChatAgentToolDefinitions() []llm.ToolDefinition {
	t.ensureChatTools()
	defs := make([]llm.ToolDefinition, 0, len(t.chatTools))
	for _, tool := range t.chatTools {
		defs = append(defs, llm.ToolDefinition{
			Name:        tool.name,
			Description: tool.desc,
			Parameters:  tool.params,
		})
	}
	return defs
}

func (t *Tools) ExecuteChatAgentTool(ctx context.Context, ctxData domain.AgentChatContext, name string, arguments string) (string, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.ExecuteChatAgentTool")
	defer span.Finish()

	return t.executeTool(ctx, ctxData, name, arguments)
}

func (t *Tools) RequiresConfirmation(toolName string) bool {
	t.ensureChatTools()

	tool, ok := t.chatTools[toolName]
	if !ok {
		return false
	}
	return tool.requiresConfirmation
}

func (t *Tools) ensureChatTools() {
	t.chatToolsOnce.Do(func() {
		t.chatTools = map[string]agentTool{}

		for _, tool := range []agentTool{
			t.newSaveUserFactTool(),
			t.newTestApprovalTool(),
		} {
			if tool.name == "" {
				panic("agent tool definition missing name")
			}
			t.chatTools[tool.name] = tool
		}
	})
}

func (t *Tools) executeTool(ctx context.Context, chatCtx domain.AgentChatContext, name string, arguments string) (string, error) {
	t.ensureChatTools()

	tool, ok := t.chatTools[name]
	if !ok {
		return "", fmt.Errorf("unknown tool: %s", name)
	}

	var raw json.RawMessage
	if arguments != "" {
		raw = json.RawMessage(arguments)
	} else {
		raw = json.RawMessage("{}")
	}

	return tool.handler(ctx, chatCtx, raw)
}
