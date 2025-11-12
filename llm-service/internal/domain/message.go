package domain

// MessageRole - роль сообщения в диалоге
type MessageRole string

const (
	MessageRoleSystem    MessageRole = "system"
	MessageRoleUser      MessageRole = "user"
	MessageRoleAssistant MessageRole = "assistant"
	MessageRoleTool      MessageRole = "tool"
)

// Message - сообщение в чате
type Message struct {
	Model
	ChatID     ID
	Role       MessageRole
	Content    string
	Sender     *string // null для user, agent_key для агентов
	ToolCalls  []*ToolCall
	ToolCallID *ID // для tool результатов
}

// IsUserMessage - проверяет, является ли сообщение пользовательским
func (m *Message) IsUserMessage() bool {
	return m.Role == MessageRoleUser
}

// IsAssistantMessage - проверяет, является ли сообщение от ассистента
func (m *Message) IsAssistantMessage() bool {
	return m.Role == MessageRoleAssistant
}

// IsToolResult - проверяет, является ли сообщение результатом tool
func (m *Message) IsToolResult() bool {
	return m.Role == MessageRoleTool
}

// HasToolCalls - проверяет, есть ли в сообщении вызовы инструментов
func (m *Message) HasToolCalls() bool {
	return len(m.ToolCalls) > 0
}
