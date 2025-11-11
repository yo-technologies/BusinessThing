package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"llm-service/internal/domain"
)

func (t *Tools) newSaveUserFactTool() agentTool {
	params := map[string]any{
		"type":     "object",
		"required": []string{"content"},
		"properties": map[string]any{
			"content": map[string]any{
				"type":        "string",
				"description": "Краткий факт (<=200 символов)",
			},
		},
	}
	return agentTool{
		name:    "save_user_fact",
		desc:    "Сохранить короткий факт о пользователе для будущего контекста (до 200 символов).",
		params:  params,
		handler: t.handleSaveUserFact,
	}
}

func (t *Tools) handleSaveUserFact(ctx context.Context, chatCtx domain.AgentChatContext, raw json.RawMessage) (string, error) {
	var payload struct {
		Content string `json:"content"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}
	if t.service == nil {
		return "", domain.ErrInternal
	}
	fact, err := t.service.AddFact(ctx, chatCtx.UserID, payload.Content)
	if err != nil {
		return "", err
	}
	if fact.ID == (domain.ID{}) {
		return "Факт уже сохранён ранее.", nil
	}
	return "Факт сохранён.", nil
}
