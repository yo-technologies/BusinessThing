package domain

// AgentChatContext carries minimal context for tool execution.
type AgentChatContext struct {
	UserID ID
	ChatID ID
}

func NewAgentChatContext(userID, chatID ID) AgentChatContext {
	return AgentChatContext{UserID: userID, ChatID: chatID}
}
