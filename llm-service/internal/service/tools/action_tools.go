package tools

// import (
// 	"context"
// 	"encoding/json"
// 	"fmt"

// 	"llm-service/internal/domain"
// )

// // request_sync - запустить синхронизацию данных
// func (t *Tools) newRequestSyncTool() agentTool {
// 	params := map[string]any{
// 		"type":     "object",
// 		"required": []string{"scope"},
// 		"properties": map[string]any{
// 			"scope": map[string]any{
// 				"type":        "string",
// 				"enum":        []string{"accounts", "balances", "transactions", "products"},
// 				"description": "Что синхронизировать: 'accounts' (список счетов), 'balances' (балансы), 'transactions' (историю) или 'products' (продукты). Обязательно.",
// 			},
// 			"account_id": map[string]any{
// 				"type":        "string",
// 				"description": "Синхронизировать конкретный счет. Опционально. Если не указано, синхронизируются все счета.",
// 			},
// 			"bank_code": map[string]any{
// 				"type":        "string",
// 				"description": "Синхронизировать конкретный банк. Опционально.",
// 			},
// 		},
// 	}
// 	return agentTool{
// 		name: "request_sync",
// 		desc: `Запустить немедленную синхронизацию данных с банками.
// Позволяет получить самые свежие данные вместо ожидания плановой синхронизации.
// ТРЕБУЕТ ПОДТВЕРЖДЕНИЯ ПОЛЬЗОВАТЕЛЯ перед выполнением.`,
// 		params:               params,
// 		handler:              t.handleRequestSync,
// 		requiresConfirmation: true,
// 	}
// }

// func (t *Tools) handleRequestSync(
// 	ctx context.Context,
// 	chatCtx domain.AgentChatContext,
// 	raw json.RawMessage,
// ) (string, error) {
// 	var payload struct {
// 		Scope     string `json:"scope"`
// 		AccountID string `json:"account_id,omitempty"`
// 		BankCode  string `json:"bank_code,omitempty"`
// 	}
// 	if err := json.Unmarshal(raw, &payload); err != nil {
// 		return "", fmt.Errorf("invalid arguments: %w", err)
// 	}

// 	if payload.Scope == "" {
// 		return "", fmt.Errorf("scope is required")
// 	}

// 	// MOCK: simulate sync job creation
// 	// TODO: replace with real Syncer API call

// 	mockResponse := map[string]any{
// 		"job_id":                     "sync_job_" + generateID(),
// 		"status":                     "pending",
// 		"scope":                      payload.Scope,
// 		"account_id":                 payload.AccountID,
// 		"bank_code":                  payload.BankCode,
// 		"started_at":                 "2025-11-08T12:30:00Z",
// 		"estimated_duration_seconds": 30,
// 		"message":                    fmt.Sprintf("Синхронизация '%s' запущена. Проверьте статус по ID", payload.Scope),
// 		"next_check_in":              "2025-11-08T12:31:00Z",
// 	}

// 	data, _ := json.Marshal(mockResponse)
// 	return string(data), nil
// }

// // create_payment_consent - создать согласие на платеж
// func (t *Tools) newCreatePaymentConsentTool() agentTool {
// 	params := map[string]any{
// 		"type":     "object",
// 		"required": []string{"account_id", "recipient_iban", "recipient_name", "amount_limit"},
// 		"properties": map[string]any{
// 			"account_id": map[string]any{
// 				"type":        "string",
// 				"description": "ID счета, с которого будет производиться платеж. Обязательно.",
// 			},
// 			"recipient_iban": map[string]any{
// 				"type":        "string",
// 				"description": "IBAN получателя (например, RU00000000000000000000). Обязательно.",
// 			},
// 			"recipient_name": map[string]any{
// 				"type":        "string",
// 				"description": "ФИО или название организации получателя. Обязательно.",
// 			},
// 			"amount_limit": map[string]any{
// 				"type":        "string",
// 				"description": "Максимальная сумма для этого согласия. Обязательно.",
// 			},
// 			"currency": map[string]any{
// 				"type":        "string",
// 				"description": "Валюта. По умолчанию RUB.",
// 				"default":     "RUB",
// 			},
// 		},
// 	}
// 	return agentTool{
// 		name: "create_payment_consent",
// 		desc: `Создать согласие (мандат) на переводы средств конкретному получателю.
// Согласие ограничено суммой и может быть использовано для повторяющихся платежей.
// ТРЕБУЕТ ПОДТВЕРЖДЕНИЯ ПОЛЬЗОВАТЕЛЯ и последующего подтверждения в приложении банка.`,
// 		params:               params,
// 		handler:              t.handleCreatePaymentConsent,
// 		requiresConfirmation: true,
// 	}
// }

// func (t *Tools) handleCreatePaymentConsent(
// 	ctx context.Context,
// 	chatCtx domain.AgentChatContext,
// 	raw json.RawMessage,
// ) (string, error) {
// 	var payload struct {
// 		AccountID     string `json:"account_id"`
// 		RecipientIBAN string `json:"recipient_iban"`
// 		RecipientName string `json:"recipient_name"`
// 		AmountLimit   string `json:"amount_limit"`
// 		Currency      string `json:"currency"`
// 	}
// 	if err := json.Unmarshal(raw, &payload); err != nil {
// 		return "", fmt.Errorf("invalid arguments: %w", err)
// 	}

// 	if payload.AccountID == "" || payload.RecipientIBAN == "" || payload.RecipientName == "" || payload.AmountLimit == "" {
// 		return "", fmt.Errorf("account_id, recipient_iban, recipient_name, and amount_limit are required")
// 	}

// 	if payload.Currency == "" {
// 		payload.Currency = "RUB"
// 	}

// 	// MOCK: simulate consent creation
// 	// TODO: replace with real Bank Connector API call

// 	mockResponse := map[string]any{
// 		"consent_id":     "consent_" + generateID(),
// 		"status":         "pending",
// 		"account_id":     payload.AccountID,
// 		"recipient_iban": payload.RecipientIBAN,
// 		"recipient_name": payload.RecipientName,
// 		"amount_limit":   payload.AmountLimit,
// 		"currency":       payload.Currency,
// 		"created_at":     "2025-11-08T12:30:00Z",
// 		"approval_url":   "https://bank.example.com/confirm/consent_" + generateID(),
// 		"expires_at":     "2026-11-08T12:30:00Z",
// 		"message":        "Согласие создано. Подтвердите в вашем банке по ссылке выше",
// 		"next_steps":     []string{"Откройте ссылку подтверждения в приложении банка", "Подтвердите двухфакторно", "Используйте ID согласия для платежей"},
// 	}

// 	data, _ := json.Marshal(mockResponse)
// 	return string(data), nil
// }

// // execute_payment - выполнить платеж
// func (t *Tools) newExecutePaymentTool() agentTool {
// 	params := map[string]any{
// 		"type":     "object",
// 		"required": []string{"account_id", "recipient_iban", "recipient_name", "amount"},
// 		"properties": map[string]any{
// 			"account_id": map[string]any{
// 				"type":        "string",
// 				"description": "ID счета-отправителя. Обязательно.",
// 			},
// 			"recipient_iban": map[string]any{
// 				"type":        "string",
// 				"description": "IBAN получателя. Обязательно.",
// 			},
// 			"recipient_name": map[string]any{
// 				"type":        "string",
// 				"description": "ФИО или название получателя. Обязательно.",
// 			},
// 			"amount": map[string]any{
// 				"type":        "string",
// 				"description": "Сумма платежа. Обязательно.",
// 			},
// 			"currency": map[string]any{
// 				"type":        "string",
// 				"description": "Валюта. По умолчанию RUB.",
// 				"default":     "RUB",
// 			},
// 			"description": map[string]any{
// 				"type":        "string",
// 				"description": "Описание платежа (назначение). Опционально.",
// 			},
// 			"consent_id": map[string]any{
// 				"type":        "string",
// 				"description": "ID подтвержденного согласия на платеж. Опционально.",
// 			},
// 		},
// 	}
// 	return agentTool{
// 		name: "execute_payment",
// 		desc: `Выполнить денежный перевод на счет получателя.
// Поддерживает как одноразовые платежи, так и платежи по ранее созданному согласию.
// ТРЕБУЕТ ПОДТВЕРЖДЕНИЯ ПОЛЬЗОВАТЕЛЯ и двухфакторной аутентификации в банке.`,
// 		params:               params,
// 		handler:              t.handleExecutePayment,
// 		requiresConfirmation: true,
// 	}
// }

// func (t *Tools) handleExecutePayment(
// 	ctx context.Context,
// 	chatCtx domain.AgentChatContext,
// 	raw json.RawMessage,
// ) (string, error) {
// 	var payload struct {
// 		AccountID     string `json:"account_id"`
// 		RecipientIBAN string `json:"recipient_iban"`
// 		RecipientName string `json:"recipient_name"`
// 		Amount        string `json:"amount"`
// 		Currency      string `json:"currency"`
// 		Description   string `json:"description,omitempty"`
// 		ConsentID     string `json:"consent_id,omitempty"`
// 	}
// 	if err := json.Unmarshal(raw, &payload); err != nil {
// 		return "", fmt.Errorf("invalid arguments: %w", err)
// 	}

// 	if payload.AccountID == "" || payload.RecipientIBAN == "" || payload.RecipientName == "" || payload.Amount == "" {
// 		return "", fmt.Errorf("account_id, recipient_iban, recipient_name, and amount are required")
// 	}

// 	if payload.Currency == "" {
// 		payload.Currency = "RUB"
// 	}

// 	// MOCK: simulate payment execution
// 	// TODO: replace with real Bank Connector API call

// 	mockResponse := map[string]any{
// 		"payment_id":          "pay_" + generateID(),
// 		"status":              "processing",
// 		"from_account_id":     payload.AccountID,
// 		"to_account":          payload.RecipientIBAN,
// 		"to_name":             payload.RecipientName,
// 		"amount":              payload.Amount,
// 		"currency":            payload.Currency,
// 		"description":         payload.Description,
// 		"commission":          "0.00",
// 		"total_amount":        payload.Amount,
// 		"created_at":          "2025-11-08T12:30:00Z",
// 		"expected_completion": "2025-11-09T12:00:00Z",
// 		"message":             "Платеж выполнен. Он будет доставлен в течение 1-2 дней",
// 		"next_steps":          []string{"Проверьте статус платежа по ID", "Платеж появится в истории транзакций"},
// 		"warning":             "Это пример mock ответа. В реальности платеж будет отправлен в банк",
// 	}

// 	data, _ := json.Marshal(mockResponse)
// 	return string(data), nil
// }

// // Helper function to generate random IDs for mock
// func generateID() string {
// 	// In real implementation, would use proper UUID
// 	return fmt.Sprintf("%d", 100000+1000000%10000)
// }
