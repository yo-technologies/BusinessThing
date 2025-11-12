package domain

import (
	"time"
)

// ChatStatus - статус чата
type ChatStatus string

const (
	ChatStatusActive    ChatStatus = "active"
	ChatStatusCompleted ChatStatus = "completed"
	ChatStatusFailed    ChatStatus = "failed"
	ChatStatusArchived  ChatStatus = "archived"
)

// Chat - основная сущность чата
type Chat struct {
	Model
	OrganizationID   ID         `db:"organization_id"`
	UserID           ID         `db:"user_id"`
	AgentKey         string     `db:"agent_key"`
	Title            string     `db:"title"`
	Status           ChatStatus `db:"status"`
	ParentChatID     *ID        `db:"parent_chat_id"`      // для субагентов - ссылка на родительский чат
	ParentToolCallID *ID        `db:"parent_tool_call_id"` // связь с tool call родительского агента
}

func NewChat(
	organizationID ID,
	userID ID,
	agentKey string,
	title string,
	parentChatID *ID,
	parentToolCallID *ID,
) *Chat {
	return &Chat{
		Model:            NewModel(),
		OrganizationID:   organizationID,
		UserID:           userID,
		AgentKey:         agentKey,
		Title:            title,
		Status:           ChatStatusActive,
		ParentChatID:     parentChatID,
		ParentToolCallID: parentToolCallID,
	}
}

// IsSubagentChat - проверяет, является ли чат субагентом
func (c *Chat) IsSubagentChat() bool {
	return c.ParentChatID != nil
}

// CanSendMessage - проверяет, можно ли отправлять сообщения в чат
func (c *Chat) CanSendMessage() bool {
	return c.Status == ChatStatusActive
}

// Complete - завершает чат успешно
func (c *Chat) Complete() {
	c.Status = ChatStatusCompleted
	c.UpdatedAt = time.Now()
}

// Fail - завершает чат с ошибкой
func (c *Chat) Fail() {
	c.Status = ChatStatusFailed
	c.UpdatedAt = time.Now()
}

// Archive - архивирует чат
func (c *Chat) Archive() {
	c.Status = ChatStatusArchived
	c.UpdatedAt = time.Now()
}
