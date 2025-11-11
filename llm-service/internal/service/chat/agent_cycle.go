package chat

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"llm-service/internal/domain"
	"llm-service/internal/domain/dto"
	"llm-service/internal/llm"
	"llm-service/internal/logger"

	"github.com/guregu/null/v6"
	"github.com/opentracing/opentracing-go"
)

// buildLLMMessages builds [system + history] for the given chat
func (s *Service) buildLLMMessages(ctx context.Context, chat domain.Chat) ([]llm.MessageParam, error) {
	system := s.buildSystemPrompt(ctx, chat.UserID)
	systemMessages := []llm.MessageParam{{Role: llm.RoleSystem, Content: system}}

	history, err := s.chatRepository.ListChatMessages(ctx, chat.ID, chatHistoryLimit, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to load chat history: %w", err)
	}

	msgs := make([]llm.MessageParam, 0, len(systemMessages)+len(history))
	msgs = append(msgs, systemMessages...)
	for _, m := range history {
		p, err := s.chatMessageToLLMParam(m)
		if err != nil {
			return nil, err
		}
		msgs = append(msgs, p)
	}
	return msgs, nil
}

// runAgentCycle runs the assistant loop until:
// - it produces a final assistant message, or
// - proposes a tool that requires confirmation (returns ErrAwaitingConfirmation), or
// - ShouldStop() is signaled.
func (s *Service) runAgentCycle(
	ctx context.Context,
	userID domain.ID,
	chat domain.Chat,
	initial []llm.MessageParam,
	toolDefs []llm.ToolDefinition,
	callbacks dto.ChatStreamCallbacks,
) (dto.ChatCompletionDTO, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.chat.runAgentCycle")
	defer span.Finish()

	messages := make([]llm.MessageParam, len(initial))
	copy(messages, initial)

	agentCtx := domain.NewAgentChatContext(userID, chat.ID)

	if callbacks.OnStatus != nil {
		_ = callbacks.OnStatus("assistant_thinking")
	}

	var (
		totalUsage         dto.ChatUsage
		finalResult        dto.ChatCompletionDTO
		assistantResponded bool
	)

	for range maxChatCompletionLoops {
		stream, err := s.llmClient.CreateCompletionStream(ctx, llm.ChatParams{Messages: messages, Tools: toolDefs, IncludeUsage: true})
		if err != nil {
			if callbacks.OnError != nil {
				_ = callbacks.OnError(err)
			}
			_, _ = s.handleAssistantFailure(ctx, chat.ID, err)
			return dto.ChatCompletionDTO{}, err
		}

		assistantMessage, usage, err := s.runStreamingChatCompletion(ctx, stream, &callbacks)
		if err != nil {
			if callbacks.OnError != nil {
				_ = callbacks.OnError(err)
			}
			_, _ = s.handleAssistantFailure(ctx, chat.ID, err)
			return dto.ChatCompletionDTO{}, err
		}

		totalUsage.PromptTokens += usage.PromptTokens
		totalUsage.CompletionTokens += usage.CompletionTokens
		totalUsage.TotalTokens += usage.TotalTokens

		if len(assistantMessage.ToolCalls) > 0 {
			if err := s.handleAssistantToolCalls(ctx, chat.ID, &messages, agentCtx, assistantMessage, &callbacks); err != nil {
				if errors.Is(err, domain.ErrAwaitingConfirmation) {
					// Tool requires confirmation - return current state for client
					updatedHistory, histErr := s.chatRepository.ListChatMessages(ctx, chat.ID, chatHistoryLimit, 0)
					if histErr != nil {
						return dto.ChatCompletionDTO{}, fmt.Errorf("failed to refresh chat history: %w", histErr)
					}

					result := dto.ChatCompletionDTO{Chat: chat, Messages: updatedHistory}
					if totalUsage.TotalTokens > 0 {
						result.Usage = &totalUsage
					}

					// Send final response so client gets chatId
					if callbacks.OnFinalResponse != nil {
						_ = callbacks.OnFinalResponse(result)
					}

					return result, nil
				}
				return dto.ChatCompletionDTO{}, err
			}
			continue
		}

		// Save final assistant message
		finalRecord, err := domain.NewChatMessage(chat.ID, domain.ChatMessageRoleAssistant, assistantMessage.Content, null.String{}, null.String{}, nil)
		if err != nil {
			return dto.ChatCompletionDTO{}, err
		}
		if totalUsage.TotalTokens > 0 {
			finalRecord.TokenUsage = null.IntFrom(int64(totalUsage.TotalTokens))
		}
		if _, err := s.chatRepository.CreateChatMessage(ctx, finalRecord); err != nil {
			return dto.ChatCompletionDTO{}, fmt.Errorf("failed to save assistant message: %w", err)
		}

		updatedHistory, err := s.chatRepository.ListChatMessages(ctx, chat.ID, chatHistoryLimit, 0)
		if err != nil {
			return dto.ChatCompletionDTO{}, fmt.Errorf("failed to refresh chat history: %w", err)
		}

		finalResult = dto.ChatCompletionDTO{Chat: chat, Messages: updatedHistory}
		if totalUsage.TotalTokens > 0 {
			finalResult.Usage = &totalUsage
		}
		assistantResponded = true

		if callbacks.OnFinalResponse != nil {
			_ = callbacks.OnFinalResponse(finalResult)
		}
		break
	}

	// TODO: handle case where max loops reached without final response
	if !assistantResponded {
		return dto.ChatCompletionDTO{}, fmt.Errorf("assistant did not provide a final response: %w", domain.ErrInternal)
	}

	if callbacks.OnUsage != nil && totalUsage.TotalTokens > 0 {
		_ = callbacks.OnUsage(totalUsage)
	}

	if callbacks.OnStatus != nil {
		_ = callbacks.OnStatus("assistant_completed")
	}

	// Confirm quota by total tokens
	_ = s.quotaService.Confirm(ctx, userID, 1, totalUsage.TotalTokens)

	return finalResult, nil
}

func (s *Service) runStreamingChatCompletion(
	ctx context.Context,
	stream llm.ChatStream,
	callbacks *dto.ChatStreamCallbacks,
) (llm.ChatMessage, dto.ChatUsage, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "service.chat.runStreamingChatCompletion")
	defer span.Finish()

	defer stream.Close()

	var (
		contentBuilder strings.Builder
		totalUsage     dto.ChatUsage
		usageRecorded  bool
	)

	type toolCallAccumulator struct {
		id        string
		name      string
		arguments strings.Builder
	}

	toolCalls := make(map[int]*toolCallAccumulator)

	for stream.Next() {
		chunk := stream.Chunk()

		if !usageRecorded && (chunk.Usage.TotalTokens != 0 || chunk.Usage.PromptTokens != 0 || chunk.Usage.CompletionTokens != 0) {
			totalUsage = dto.ChatUsage{
				PromptTokens:     int(chunk.Usage.PromptTokens),
				CompletionTokens: int(chunk.Usage.CompletionTokens),
				TotalTokens:      int(chunk.Usage.TotalTokens),
			}
			usageRecorded = true
		}

		if chunk.Content != "" {
			contentBuilder.WriteString(chunk.Content)
			if callbacks != nil && callbacks.OnContentDelta != nil {
				if err := callbacks.OnContentDelta(chunk.Content); err != nil {
					return llm.ChatMessage{}, totalUsage, err
				}
			}
		}

		if callbacks != nil && callbacks.ShouldStop != nil && callbacks.ShouldStop() {
			return llm.ChatMessage{}, totalUsage, domain.ErrGenerationStopped
		}

		for _, toolCall := range chunk.ToolCalls {
			acc := toolCalls[toolCall.Index]
			if acc == nil {
				acc = &toolCallAccumulator{}
				toolCalls[toolCall.Index] = acc
			}
			if toolCall.ID != "" {
				acc.id = toolCall.ID
			}
			if toolCall.Name != "" {
				acc.name = toolCall.Name
			}
			if toolCall.Arguments != "" {
				acc.arguments.WriteString(toolCall.Arguments)
			}
		}
	}

	if err := stream.Err(); err != nil {
		return llm.ChatMessage{}, totalUsage, err
	}

	finalMessage := llm.ChatMessage{
		Role:    llm.RoleAssistant,
		Content: contentBuilder.String(),
	}

	if len(toolCalls) > 0 {
		indices := make([]int, 0, len(toolCalls))
		for idx := range toolCalls {
			indices = append(indices, idx)
		}
		sort.Ints(indices)

		finalCalls := make([]llm.ToolCall, 0, len(indices))
		for _, idx := range indices {
			acc := toolCalls[idx]
			finalCalls = append(finalCalls, llm.ToolCall{
				ID:        acc.id,
				Name:      acc.name,
				Arguments: acc.arguments.String(),
			})
		}
		finalMessage.ToolCalls = finalCalls
	}

	return finalMessage, totalUsage, nil
}

func (s *Service) chatMessageToLLMParam(message domain.ChatMessage) (llm.MessageParam, error) {
	switch message.Role {
	case domain.ChatMessageRoleUser:
		return llm.MessageParam{Role: llm.RoleUser, Content: message.Content}, nil
	case domain.ChatMessageRoleAssistant:
		// Build assistant message with optional tool call
		assistant := llm.MessageParam{Role: llm.RoleAssistant, Content: message.Content}
		if message.ToolCallID.Valid && message.ToolName.Valid {
			argsJSON := "{}"
			if len(message.ToolArguments) > 0 {
				raw, err := json.Marshal(message.ToolArguments)
				if err != nil {
					return llm.MessageParam{}, fmt.Errorf("failed to marshal tool arguments: %w", err)
				}
				argsJSON = string(raw)
			}
			assistant.ToolCalls = []llm.ToolCall{{ID: message.ToolCallID.String, Name: message.ToolName.String, Arguments: argsJSON}}
		}
		return assistant, nil
	case domain.ChatMessageRoleTool:
		if !message.ToolCallID.Valid {
			return llm.MessageParam{}, fmt.Errorf("tool message missing tool call id: %w", domain.ErrInvalidArgument)
		}
		content := message.Content
		if strings.TrimSpace(content) == "" && message.Error.Valid {
			content = fmt.Sprintf("error: %s", message.Error.String)
		}
		return llm.MessageParam{Role: llm.RoleTool, Content: content, ToolCallID: message.ToolCallID.String}, nil
	case domain.ChatMessageRoleSystem:
		return llm.MessageParam{Role: llm.RoleSystem, Content: message.Content}, nil
	default:
		return llm.MessageParam{}, fmt.Errorf("unsupported chat message role %s: %w", message.Role, domain.ErrInvalidArgument)
	}
}

func (s *Service) handleAssistantToolCalls(
	ctx context.Context,
	chatID domain.ID,
	messages *[]llm.MessageParam,
	chatCtx domain.AgentChatContext,
	assistantMessage llm.ChatMessage,
	callbacks *dto.ChatStreamCallbacks,
) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.chat.handleAssistantToolCalls")
	defer span.Finish()

	contentForFirst := assistantMessage.Content

	for idx, toolCall := range assistantMessage.ToolCalls {
		toolName := toolCall.Name
		toolArgsJSON := toolCall.Arguments

		var toolArgs map[string]any
		trimmedArgs := strings.TrimSpace(toolArgsJSON)
		if trimmedArgs != "" {
			if err := json.Unmarshal([]byte(trimmedArgs), &toolArgs); err != nil {
				decoder := json.NewDecoder(strings.NewReader(trimmedArgs))
				if decodeErr := decoder.Decode(&toolArgs); decodeErr != nil {
					return fmt.Errorf("failed to parse tool arguments for %s: %w (raw: %q)", toolName, err, trimmedArgs)
				}
				logger.Warnf(ctx, "tool %s had malformed arguments, using first valid JSON object: %v", toolName, toolArgs)
			}
		}

		assistantRecord, err := domain.NewChatMessage(
			chatID,
			domain.ChatMessageRoleAssistant,
			"",
			null.String{},
			null.String{},
			toolArgs,
		)
		if err != nil {
			return err
		}
		if idx == 0 {
			assistantRecord.Content = contentForFirst
		}
		if toolName != "" {
			assistantRecord.ToolName = null.StringFrom(toolName)
		}
		if toolCall.ID != "" {
			assistantRecord.ToolCallID = null.StringFrom(toolCall.ID)
		}

		// If tool requires confirmation, mark as proposed and stop further execution.
		requires := s.toolsService.RequiresConfirmation(toolName)
		if requires {
			assistantRecord.ToolState = domain.ToolStateProposed
		}

		savedAssistant, err := s.chatRepository.CreateChatMessage(ctx, assistantRecord)
		if err != nil {
			return fmt.Errorf("failed to persist assistant tool call: %w", err)
		}
		param, err := s.chatMessageToLLMParam(savedAssistant)
		if err != nil {
			return err
		}
		*messages = append(*messages, param)

		if requires {
			// Notify client that tool is awaiting confirmation
			if callbacks != nil && callbacks.OnStatus != nil {
				_ = callbacks.OnStatus("awaiting_confirmation")
			}
			if callbacks != nil && callbacks.OnToolEvent != nil {
				_ = callbacks.OnToolEvent(dto.ToolEvent{
					ToolName:   toolName,
					ToolCallID: toolCall.ID,
					ArgsJSON:   toolArgsJSON,
					State:      dto.ToolInvoking,
				})
			}

			return domain.ErrAwaitingConfirmation
		}

		if callbacks != nil && callbacks.OnToolEvent != nil {
			_ = callbacks.OnToolEvent(dto.ToolEvent{
				ToolName:   toolName,
				ToolCallID: toolCall.ID,
				ArgsJSON:   toolArgsJSON,
				State:      dto.ToolInvoking,
			})
		}

		result, err := s.toolsService.ExecuteChatAgentTool(ctx, chatCtx, toolName, toolArgsJSON)
		if err != nil {
			toolRecord, _ := domain.NewChatMessage(chatID, domain.ChatMessageRoleTool, "", null.String{}, null.String{}, nil)
			if toolName != "" {
				toolRecord.ToolName = null.StringFrom(toolName)
			}
			if toolCall.ID != "" {
				toolRecord.ToolCallID = null.StringFrom(toolCall.ID)
			}
			toolRecord.Error = null.StringFrom(err.Error())

			savedTool, perr := s.chatRepository.CreateChatMessage(ctx, toolRecord)
			if perr != nil {
				return fmt.Errorf("failed to persist tool error message: %w", perr)
			}
			if toolParam, perr := s.chatMessageToLLMParam(savedTool); perr == nil {
				*messages = append(*messages, toolParam)
			}
			if callbacks != nil && callbacks.OnToolEvent != nil {
				_ = callbacks.OnToolEvent(dto.ToolEvent{ToolName: toolName, ToolCallID: toolCall.ID, ArgsJSON: toolArgsJSON, State: dto.ToolError, Error: err.Error()})
				_ = callbacks.OnToolEvent(dto.ToolEvent{ToolName: toolName, ToolCallID: toolCall.ID, ArgsJSON: toolArgsJSON, State: dto.ToolCompleted})
			}
			continue
		}

		toolRecord, _ := domain.NewChatMessage(chatID, domain.ChatMessageRoleTool, result, null.String{}, null.String{}, nil)
		if toolName != "" {
			toolRecord.ToolName = null.StringFrom(toolName)
		}
		if toolCall.ID != "" {
			toolRecord.ToolCallID = null.StringFrom(toolCall.ID)
		}

		savedTool, err := s.chatRepository.CreateChatMessage(ctx, toolRecord)
		if err != nil {
			return fmt.Errorf("failed to persist tool message: %w", err)
		}

		toolParam, err := s.chatMessageToLLMParam(savedTool)
		if err != nil {
			return err
		}
		*messages = append(*messages, toolParam)

		if callbacks != nil && callbacks.OnToolEvent != nil {
			_ = callbacks.OnToolEvent(dto.ToolEvent{ToolName: toolName, ToolCallID: toolCall.ID, ArgsJSON: toolArgsJSON, State: dto.ToolCompleted})
		}
	}

	return nil
}

func (s *Service) handleAssistantFailure(ctx context.Context, chatID domain.ID, originalErr error) (dto.ChatCompletionDTO, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.chat.handleAssistantFailure")
	defer span.Finish()

	errMsg, _ := domain.NewChatMessage(
		chatID,
		domain.ChatMessageRoleAssistant,
		assistantErrorMessage,
		null.String{},
		null.String{},
		nil,
	)
	errMsg.Error = null.StringFrom(originalErr.Error())

	if _, err := s.chatRepository.CreateChatMessage(ctx, errMsg); err != nil {
		return dto.ChatCompletionDTO{}, fmt.Errorf("failed to persist assistant error message: %w", err)
	}

	return dto.ChatCompletionDTO{}, errors.Join(domain.ErrInternal, originalErr)
}
