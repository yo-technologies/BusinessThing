package chat

import (
	"context"
	"encoding/json"
	"fmt"

	"llm-service/internal/domain"
	"llm-service/internal/domain/dto"

	"github.com/guregu/null/v6"
	"github.com/opentracing/opentracing-go"
)

// HandleToolCallDecisionStream unifies approve/reject into one streaming flow.
// If approve=true: executes tool, persists result, then runs agent cycle streaming; marks original assistant as approved.
// If approve=false: marks rejected and persists a tool-result message indicating rejection, then runs agent cycle streaming.
func (s *Service) HandleToolCallDecisionStream(
	ctx context.Context,
	userID domain.ID,
	chatID domain.ID,
	toolCallID string,
	approve bool,
	callbacks dto.ChatStreamCallbacks,
) (dto.ChatCompletionDTO, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.chat.HandleToolCallDecisionStream")
	defer span.Finish()

	chat, err := s.chatRepository.GetChatByID(ctx, chatID)
	if err != nil {
		return dto.ChatCompletionDTO{}, fmt.Errorf("get chat: %w", err)
	}
	if chat.UserID != userID {
		return dto.ChatCompletionDTO{}, fmt.Errorf("forbidden: %w", domain.ErrForbidden)
	}

	msg, err := s.chatRepository.GetMessageByToolCallID(ctx, chatID, toolCallID)
	if err != nil {
		return dto.ChatCompletionDTO{}, fmt.Errorf("get message: %w", err)
	}
	if msg.ToolState != domain.ToolStateProposed {
		return dto.ChatCompletionDTO{}, fmt.Errorf("tool call not in proposed state: current state is %v: %w", msg.ToolState, domain.ErrInvalidArgument)
	}

	if !approve {
		if err := s.chatRepository.UpdateMessageToolState(ctx, msg.ID, domain.ToolStateRejected); err != nil {
			return dto.ChatCompletionDTO{}, fmt.Errorf("update message tool state: %w", err)
		}
		// Persist a tool-result message indicating rejection, linked to tool_call_id (system/tool message)
		toolMsg, _ := domain.NewChatMessage(chatID, domain.ChatMessageRoleTool, "Пользователь отклонил вызов инструмента.", msg.ToolName, msg.ToolCallID, nil)
		_, _ = s.chatRepository.CreateChatMessage(ctx, toolMsg)

		msgs, err := s.buildLLMMessages(ctx, chat)
		if err != nil {
			return dto.ChatCompletionDTO{}, fmt.Errorf("build LLM messages: %w", err)
		}

		return s.runAgentCycle(ctx, userID, chat, msgs, s.toolsService.ChatAgentToolDefinitions(), callbacks)
	}

	// execute tool
	argsJSON := "{}"
	if len(msg.ToolArguments) > 0 {
		var tmp map[string]any
		if json.Unmarshal(msg.ToolArguments, &tmp) == nil {
			argsJSON = string(msg.ToolArguments)
		}
	}
	result, execErr := s.toolsService.ExecuteChatAgentTool(ctx, domain.NewAgentChatContext(userID, chatID), msg.ToolName.String, argsJSON)

	// persist tool message
	toolRecord, _ := domain.NewChatMessage(chatID, domain.ChatMessageRoleTool, "", msg.ToolName, msg.ToolCallID, nil)
	if execErr != nil {
		toolRecord.Error = null.StringFrom(execErr.Error())
	} else {
		toolRecord.Content = result
	}
	if _, err := s.chatRepository.CreateChatMessage(ctx, toolRecord); err != nil {
		return dto.ChatCompletionDTO{}, fmt.Errorf("create tool message: %w", err)
	}

	// reflect final state on the original assistant message
	_ = s.chatRepository.UpdateMessageToolState(ctx, msg.ID, domain.ToolStateApproved)

	msgs, err := s.buildLLMMessages(ctx, chat)
	if err != nil {
		return dto.ChatCompletionDTO{}, fmt.Errorf("build LLM messages: %w", err)
	}
	return s.runAgentCycle(ctx, userID, chat, msgs, s.toolsService.ChatAgentToolDefinitions(), callbacks)
}
