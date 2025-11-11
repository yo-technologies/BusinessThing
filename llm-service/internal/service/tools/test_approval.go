package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"llm-service/internal/domain"
)

func (t *Tools) newTestApprovalTool() agentTool {
	params := map[string]any{
		"type":     "object",
		"required": []string{"message"},
		"properties": map[string]any{
			"message": map[string]any{
				"type":        "string",
				"description": "Сообщение для тестирования",
			},
		},
	}
	return agentTool{
		name: "test_approval",
		desc: `ВАЖНО: Этот инструмент предназначен ТОЛЬКО для тестирования механизма подтверждения пользователем (requireApproval).
Инструмент НЕ выполняет никаких реальных действий и НЕ изменяет данные.
Пользователь может явно попросить вызвать этот инструмент для проверки работы системы подтверждений.
При вызове этого инструмента система запросит у пользователя подтверждение, после чего вернёт тестовое сообщение.
НЕ вызывай этот инструмент самостоятельно без явной просьбы пользователя протестировать механизм подтверждения.`,
		params:               params,
		handler:              t.handleTestApproval,
		requiresConfirmation: true,
	}
}

func (t *Tools) handleTestApproval(ctx context.Context, chatCtx domain.AgentChatContext, raw json.RawMessage) (string, error) {
	var payload struct {
		Message string `json:"message"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	// Этот инструмент ничего не делает, просто возвращает тестовое сообщение
	return fmt.Sprintf("✅ Тестовый инструмент выполнен успешно! Получено сообщение: '%s'. Это тестовый инструмент, никаких действий не было выполнено.", payload.Message), nil
}
