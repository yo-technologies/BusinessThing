package chat

import (
	"context"
	"fmt"
	"llm-service/internal/domain"
	"llm-service/internal/domain/dto"
	"llm-service/internal/repository"

	"github.com/opentracing/opentracing-go"
)

// Manager - менеджер чатов
type Manager struct {
	chatRepo    repository.ChatRepository
	messageRepo repository.MessageRepository
	toolRepo    repository.ToolCallRepository
}

// NewManager создает новый менеджер чатов
func NewManager(
	chatRepo repository.ChatRepository,
	messageRepo repository.MessageRepository,
	toolRepo repository.ToolCallRepository,
) *Manager {
	return &Manager{
		chatRepo:    chatRepo,
		messageRepo: messageRepo,
		toolRepo:    toolRepo,
	}
}

// CreateChat создает новый чат
func (m *Manager) CreateChat(ctx context.Context, req dto.CreateChatDTO) (*domain.Chat, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.chat.CreateChat")
	defer span.Finish()

	chat := domain.NewChat(
		req.OrganizationID,
		req.UserID,
		req.AgentKey,
		req.Title,
		req.ParentChatID,
		req.ParentToolCallID,
	)

	if err := m.chatRepo.CreateChat(ctx, chat); err != nil {
		return nil, domain.NewInternalError("failed to create chat", err)
	}

	return chat, nil
}

// GetChat получает чат по ID
func (m *Manager) GetChat(ctx context.Context, chatID domain.ID) (*domain.Chat, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.chat.GetChat")
	defer span.Finish()

	chat, err := m.chatRepo.GetChatByID(ctx, chatID)
	if err != nil {
		return nil, err
	}

	return chat, nil
}

// ListChats получает список чатов пользователя
func (m *Manager) ListChats(ctx context.Context, organizationID, userID domain.ID, page, pageSize int) ([]*domain.Chat, int, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.chat.ListChats")
	defer span.Finish()

	filter := repository.ChatFilter{
		OrganizationID: &organizationID,
		UserID:         &userID,
	}

	chats, total, err := m.chatRepo.ListChats(ctx, filter, page, pageSize)
	if err != nil {
		return nil, 0, domain.NewInternalError("failed to list chats", err)
	}

	return chats, total, nil
}

// DeleteChat удаляет чат
func (m *Manager) DeleteChat(ctx context.Context, chatID, userID, orgID domain.ID) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.chat.DeleteChat")
	defer span.Finish()

	if err := m.chatRepo.DeleteChat(ctx, chatID, userID, orgID); err != nil {
		return domain.NewInternalError("failed to delete chat", err)
	}

	return nil
}

// GetMessages получает сообщения чата
func (m *Manager) GetMessages(ctx context.Context, chatID, userID, orgID domain.ID, limit, offset int) ([]*domain.Message, int, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.chat.GetMessages")
	defer span.Finish()

	chat, err := m.chatRepo.GetChatByID(ctx, chatID)
	if err != nil {
		return nil, 0, err
	}

	if chat.UserID != userID {
		return nil, 0, domain.ErrNotFound
	}

	if chat.OrganizationID != orgID {
		return nil, 0, domain.ErrNotFound
	}

	messages, total, err := m.messageRepo.ListMessagesWithSubchatsWithToolCalls(ctx, chatID, limit, offset)
	if err != nil {
		return nil, 0, domain.NewInternalError("failed to get messages", err)
	}

	return messages, total, nil
}

// SaveMessage сохраняет сообщение
func (m *Manager) SaveMessage(ctx context.Context, message *domain.Message) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.chat.SaveMessage")
	defer span.Finish()

	if err := m.messageRepo.CreateMessage(ctx, message); err != nil {
		return domain.NewInternalError("failed to save message", err)
	}

	// Сохраняем tool calls если есть
	for _, toolCall := range message.ToolCalls {
		if err := m.toolRepo.CreateToolCall(ctx, message.ID, toolCall); err != nil {
			return domain.NewInternalError(fmt.Sprintf("failed to save tool call %s", toolCall.Name), err)
		}
	}

	return nil
}

// GetChatWithMessages получает чат со всеми сообщениями
func (m *Manager) GetChatWithMessages(ctx context.Context, chatID domain.ID) (*domain.Chat, []*domain.Message, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.chat.GetChatWithMessages")
	defer span.Finish()

	chat, err := m.chatRepo.GetChatByID(ctx, chatID)
	if err != nil {
		return nil, nil, err
	}

	messages, _, err := m.messageRepo.ListMessagesByChatIDWithToolCalls(ctx, chatID, 1000, 0)
	if err != nil {
		return nil, nil, domain.NewInternalError("failed to get messages", err)
	}

	return chat, messages, nil
}

// UpdateChat обновляет чат
func (m *Manager) UpdateChat(ctx context.Context, chat *domain.Chat) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.chat.UpdateChat")
	defer span.Finish()

	if err := m.chatRepo.UpdateChat(ctx, chat); err != nil {
		return domain.NewInternalError("failed to update chat", err)
	}
	return nil
}

// GetActiveChildChat получает активный дочерний чат
func (m *Manager) GetActiveChildChat(ctx context.Context, parentChatID domain.ID) (*domain.Chat, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.chat.GetActiveChildChat")
	defer span.Finish()

	chat, err := m.chatRepo.GetActiveChildChat(ctx, parentChatID)
	if err != nil {
		return nil, err
	}

	return chat, nil
}
