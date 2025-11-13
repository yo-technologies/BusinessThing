package subagent

import (
	"context"
	"errors"
	"fmt"
	"llm-service/internal/domain"
	"llm-service/internal/domain/dto"
	"llm-service/internal/service"

	"github.com/opentracing/opentracing-go"
)

// Manager - менеджер субагентов
type Manager struct {
	chatManager  service.ChatManager
	agentManager service.AgentManager
}

// NewManager создает новый менеджер субагентов
func NewManager(
	chatManager service.ChatManager,
	agentManager service.AgentManager,
) *Manager {
	return &Manager{
		chatManager:  chatManager,
		agentManager: agentManager,
	}
}

// SwitchToSubagent переключает на субагента
func (m *Manager) SwitchToSubagent(ctx context.Context, parentChatID domain.ID, subagentKey string, task string, parentToolCallID *domain.ID) (*domain.Chat, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "subagent.Manager.SwitchToSubagent")
	defer span.Finish()

	// Проверяем, что субагент существует
	subagentDef, err := m.agentManager.GetAgent(subagentKey)
	if err != nil {
		return nil, err
	}

	if !subagentDef.IsSubagent {
		return nil, domain.NewInvalidArgumentError(fmt.Sprintf("agent %s is not a subagent", subagentKey))
	}

	// Получаем родительский чат
	parentChat, err := m.chatManager.GetChat(ctx, parentChatID)
	if err != nil {
		return nil, err
	}

	// Создаем и сохраняем новый чат для субагента в БД
	childChat, err := m.chatManager.CreateChat(ctx, dto.CreateChatDTO{
		OrganizationID:   parentChat.OrganizationID,
		UserID:           parentChat.UserID,
		AgentKey:         subagentKey,
		Title:            fmt.Sprintf("%s: %s", subagentDef.Name, task),
		ParentChatID:     &parentChatID,
		ParentToolCallID: parentToolCallID,
	})
	if err != nil {
		return nil, err
	}

	return childChat, nil
}

// FinishSubagent завершает работу субагента и возвращает summary
func (m *Manager) FinishSubagent(ctx context.Context, subagentChatID domain.ID, summary string) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "subagent.Manager.FinishSubagent")
	defer span.Finish()

	// Получаем чат субагента
	subagentChat, err := m.chatManager.GetChat(ctx, subagentChatID)
	if err != nil {
		return err
	}

	// Проверяем, что это действительно субагент
	if subagentChat.ParentChatID == nil {
		return domain.NewInvalidArgumentError("chat is not a subagent chat")
	}

	// Завершаем чат
	subagentChat.Complete()
	if err := m.chatManager.UpdateChat(ctx, subagentChat); err != nil {
		return domain.NewInternalError("failed to update chat status", err)
	}

	return nil
}

// GetParentChat получает родительский чат для субагента
func (m *Manager) GetParentChat(ctx context.Context, subagentChatID domain.ID) (*domain.Chat, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "subagent.Manager.GetParentChat")
	defer span.Finish()

	subagentChat, err := m.chatManager.GetChat(ctx, subagentChatID)
	if err != nil {
		return nil, err
	}

	if subagentChat.ParentChatID == nil {
		return nil, domain.NewInvalidArgumentError("chat is not a subagent chat")
	}

	parentChat, err := m.chatManager.GetChat(ctx, *subagentChat.ParentChatID)
	if err != nil {
		return nil, err
	}

	return parentChat, nil
}

// GetActiveChatID возвращает ID активного чата (может быть субагент)
func (m *Manager) GetActiveChatID(ctx context.Context, chatID domain.ID) (domain.ID, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "subagent.Manager.GetActiveChatID")
	defer span.Finish()

	// Ищем активный дочерний чат
	activeChild, err := m.chatManager.GetActiveChildChat(ctx, chatID)
	if errors.Is(err, domain.ErrNotFound) {
		// Если дочернего чата нет - возвращаем исходный chatID
		return chatID, nil
	}
	if err != nil {
		return chatID, err
	}

	// Если есть активный дочерний чат - возвращаем его ID
	return activeChild.ID, nil
}
