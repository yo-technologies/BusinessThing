package domain

import (
	"encoding/json"
	"llm-service/internal/llm"
	"time"
)

// ToolCallStatus - статус выполнения tool call
type ToolCallStatus string

const (
	ToolCallStatusPending   ToolCallStatus = "pending"
	ToolCallStatusExecuting ToolCallStatus = "executing"
	ToolCallStatusCompleted ToolCallStatus = "completed"
	ToolCallStatusFailed    ToolCallStatus = "failed"
)

// ToolCall - вызов инструмента агентом
type ToolCall struct {
	Model
	Name        string
	Arguments   json.RawMessage
	Result      json.RawMessage
	Status      ToolCallStatus
	CompletedAt *time.Time
}

// IsPending - проверяет, ожидает ли tool call выполнения
func (tc *ToolCall) IsPending() bool {
	return tc.Status == ToolCallStatusPending
}

// IsCompleted - проверяет, завершен ли tool call
func (tc *ToolCall) IsCompleted() bool {
	return tc.Status == ToolCallStatusCompleted
}

// IsFailed - проверяет, завершился ли tool call с ошибкой
func (tc *ToolCall) IsFailed() bool {
	return tc.Status == ToolCallStatusFailed
}

// MarkExecuting - помечает tool call как выполняющийся
func (tc *ToolCall) MarkExecuting() {
	tc.Status = ToolCallStatusExecuting
}

// Complete - завершает tool call с результатом
func (tc *ToolCall) Complete(result json.RawMessage) {
	tc.Status = ToolCallStatusCompleted
	tc.Result = result
	now := time.Now()
	tc.CompletedAt = &now
}

// Fail - завершает tool call с ошибкой
func (tc *ToolCall) Fail() {
	tc.Status = ToolCallStatusFailed
	now := time.Now()
	tc.CompletedAt = &now
}

// ToolDefinition - определение инструмента для агента
type ToolDefinition struct {
	Name        string                 `yaml:"name" json:"name"`
	Description string                 `yaml:"description" json:"description"`
	Parameters  map[string]interface{} `yaml:"parameters" json:"parameters"`
	Required    []string               `yaml:"required" json:"required"`
}

// ToLLMObject - конвертирует определение инструмента в формат LLM
func (td *ToolDefinition) ToLLMObject() llm.ToolDefinition {
	return llm.ToolDefinition{
		Name:        td.Name,
		Description: td.Description,
		Parameters: map[string]interface{}{
			"type":       "object",
			"properties": td.Parameters,
			"required":   td.Required,
		},
	}
}
