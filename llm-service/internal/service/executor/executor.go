package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"llm-service/internal/config"
	"llm-service/internal/domain"
	"llm-service/internal/domain/dto"
	"llm-service/internal/llm"
	"llm-service/internal/logger"
	"llm-service/internal/service"
	"strings"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/samber/lo"
)

const (
	maxIterations = 20
)

// Executor - основной исполнитель агентов
type Executor struct {
	chatManager     service.ChatManager
	agentManager    service.AgentManager
	contextBuilder  service.ContextBuilder
	toolExecutor    service.ToolExecutor
	subagentManager service.SubagentManager
	llmProvider     llm.CompletionProvider
	cfg             *config.Config
}

// NewExecutor создает новый executor
func NewExecutor(
	chatManager service.ChatManager,
	agentManager service.AgentManager,
	contextBuilder service.ContextBuilder,
	toolExecutor service.ToolExecutor,
	subagentManager service.SubagentManager,
	llmProvider llm.CompletionProvider,
	cfg *config.Config,
) *Executor {
	return &Executor{
		chatManager:     chatManager,
		agentManager:    agentManager,
		contextBuilder:  contextBuilder,
		toolExecutor:    toolExecutor,
		subagentManager: subagentManager,
		llmProvider:     llmProvider,
		cfg:             cfg,
	}
}

// ExecuteStream выполняет агента с потоковой передачей результатов
func (e *Executor) ExecuteStream(ctx context.Context, req dto.ExecuteAgentDTO, stream service.ExecutionStream) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "executor.ExecuteStream")
	defer span.Finish()

	// Получаем определение агента
	agentDef, err := e.agentManager.GetAgent(req.AgentKey)
	if err != nil {
		return stream.SendError(err)
	}

	// Получаем или создаем чат
	var chat *domain.Chat
	if req.ChatID != nil {
		chat, err = e.chatManager.GetChat(ctx, *req.ChatID)
		if err != nil {
			return stream.SendError(err)
		}
	} else {
		chat, err = e.chatManager.CreateChat(ctx, dto.CreateChatDTO{
			OrganizationID: req.OrganizationID,
			UserID:         req.UserID,
			AgentKey:       req.AgentKey,
			Title:          req.Task,
		})
		if err != nil {
			return stream.SendError(err)
		}

		// Сохраняем system message с RAG при создании нового чата
		systemPrompt, err := e.buildSystemPromptWithRAG(ctx, chat, agentDef, req.Task)
		if err != nil {
			return stream.SendError(err)
		}

		systemMessage := &domain.Message{
			Model:   domain.NewModel(),
			ChatID:  chat.ID,
			Role:    domain.MessageRoleSystem,
			Content: systemPrompt,
		}
		if err := e.chatManager.SaveMessage(ctx, systemMessage); err != nil {
			return stream.SendError(err)
		}
	}

	// Создаем execution context
	execCtx := &domain.ExecutionContext{
		OrganizationID:    req.OrganizationID,
		UserID:            req.UserID,
		ChatID:            chat.ID,
		AgentKey:          req.AgentKey,
		TaskDescription:   req.Task,
		AdditionalContext: req.Context,
	}

	// Добавляем сообщение пользователя
	userMessage := &domain.Message{
		Model:   domain.NewModel(),
		ChatID:  chat.ID,
		Role:    domain.MessageRoleUser,
		Content: req.Task,
	}
	if err := e.chatManager.SaveMessage(ctx, userMessage); err != nil {
		return stream.SendError(err)
	}

	// Отправляем сообщение пользователя
	if err := stream.SendMessage(userMessage); err != nil {
		return err
	}

	// Выполняем стриминг цикл агента
	return e.runAgentLoopStream(ctx, chat, agentDef, execCtx, stream)
}

// SendMessageStream отправляет сообщение с потоковым ответом
func (e *Executor) SendMessageStream(ctx context.Context, req dto.SendMessageDTO, stream service.MessageStream) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "executor.SendMessageStream")
	defer span.Finish()

	logger.Infof(ctx, "SendMessageStream: started with chatID=%v, userID=%s, orgID=%s, content_len=%d",
		req.ChatID, req.UserID, req.OrgID, len(req.Content))

	// Получаем чат (всегда родительский - пользователь пишет в него)
	// Если нет, создаем новый
	var (
		chat *domain.Chat
		err  error
	)
	if req.ChatID == nil {
		logger.Info(ctx, "SendMessageStream: no chatID provided, creating new chat")
		title, genErr := e.generateChatTitle(ctx, req.Content)
		if genErr != nil {
			logger.Errorf(ctx, "failed to generate chat title: %v", genErr)
		}
		chat, err = e.chatManager.CreateChat(ctx, dto.CreateChatDTO{
			OrganizationID: req.OrgID,
			UserID:         req.UserID,
			AgentKey:       "main", // основной агент по умолчанию
			Title:          lo.Ternary(title != "", title, "New Chat"),
		})
		if err != nil {
			logger.Errorf(ctx, "SendMessageStream: failed to create chat: %v", err)
			return stream.SendError(err)
		}

		logger.Infof(ctx, "SendMessageStream: created new chat with ID=%s, title=%s", chat.ID, chat.Title)

		// Получаем агента и сохраняем system message с RAG
		agentDef, err := e.agentManager.GetAgent("main")
		if err != nil {
			logger.Errorf(ctx, "SendMessageStream: failed to get agent: %v", err)
			return stream.SendError(err)
		}

		logger.Info(ctx, "SendMessageStream: building system prompt with RAG")
		systemPrompt, err := e.buildSystemPromptWithRAG(ctx, chat, agentDef, req.Content)
		if err != nil {
			logger.Errorf(ctx, "SendMessageStream: failed to build system prompt: %v", err)
			return stream.SendError(err)
		}

		systemMessage := &domain.Message{
			Model:   domain.NewModel(),
			ChatID:  chat.ID,
			Role:    domain.MessageRoleSystem,
			Content: systemPrompt,
		}
		if err := e.chatManager.SaveMessage(ctx, systemMessage); err != nil {
			logger.Errorf(ctx, "SendMessageStream: failed to save system message: %v", err)
			return stream.SendError(err)
		}

		logger.Info(ctx, "SendMessageStream: sending chat event to client")
		stream.SendChat(chat)
		req.ChatID = &chat.ID
	} else {
		logger.Infof(ctx, "SendMessageStream: loading existing chat with ID=%s", *req.ChatID)
		chat, err = e.chatManager.GetChat(ctx, *req.ChatID)
		if err != nil {
			logger.Errorf(ctx, "SendMessageStream: failed to get chat: %v", err)
			return stream.SendError(err)
		}
		logger.Infof(ctx, "SendMessageStream: loaded chat, agentKey=%s, status=%v", chat.AgentKey, chat.Status)
	}

	// Определяем активный чат: если есть активная сессия субагента, используем её
	activeChat := chat
	activeChatID, err := e.getActiveChatID(ctx, chat.ID)
	if err != nil {
		logger.Errorf(ctx, "SendMessageStream: failed to get active chat ID: %v", err)
		return stream.SendError(err)
	}

	logger.Infof(ctx, "SendMessageStream: activeChatID=%s (parent chatID=%s)", activeChatID, chat.ID)

	if activeChatID != chat.ID {
		activeChat, err = e.chatManager.GetChat(ctx, activeChatID)
		if err != nil {
			logger.Errorf(ctx, "SendMessageStream: failed to get active chat: %v", err)
			return stream.SendError(err)
		}
		logger.Infof(ctx, "SendMessageStream: using subagent chat, agentKey=%s", activeChat.AgentKey)
	}

	// Получаем определение активного агента
	agentDef, err := e.agentManager.GetAgent(activeChat.AgentKey)
	if err != nil {
		logger.Errorf(ctx, "SendMessageStream: failed to get agent definition: %v", err)
		return stream.SendError(err)
	}

	logger.Infof(ctx, "SendMessageStream: using agent '%s'(%s)", agentDef.Name, agentDef.Key)

	// Сохраняем сообщение пользователя в активный чат
	userMessage := &domain.Message{
		Model:   domain.NewModel(),
		ChatID:  activeChatID,
		Role:    domain.MessageRoleUser,
		Content: req.Content,
	}
	if err := e.chatManager.SaveMessage(ctx, userMessage); err != nil {
		logger.Errorf(ctx, "SendMessageStream: failed to save user message: %v", err)
		return stream.SendError(err)
	}

	logger.Infof(ctx, "SendMessageStream: saved user message with ID=%s", userMessage.ID)

	// Создаем execution context для активного чата
	execCtx := &domain.ExecutionContext{
		OrganizationID: chat.OrganizationID,
		UserID:         req.UserID,
		ChatID:         activeChat.ID,
		AgentKey:       activeChat.AgentKey,
	}

	logger.Infof(ctx, "SendMessageStream: starting agent loop for chatID=%s, agentKey=%s",
		activeChat.ID, activeChat.AgentKey)

	// Выполняем стриминг с активным чатом
	err = e.runAgentLoopStream(ctx, activeChat, agentDef, execCtx, stream)
	if err != nil {
		logger.Errorf(ctx, "SendMessageStream: agent loop failed: %v", err)
		return err
	}

	logger.Info(ctx, "SendMessageStream: agent loop completed successfully")

	// Отправляем финальное состояние чата
	finalChat, err := e.chatManager.GetChat(ctx, chat.ID)
	if err != nil {
		return stream.SendError(err)
	}

	finalMessages, _, err := e.chatManager.GetMessages(ctx, chat.ID, req.UserID, req.OrgID, 1000, 0)
	if err != nil {
		return stream.SendError(err)
	}

	if err := stream.SendFinal(finalChat, finalMessages); err != nil {
		return err
	}

	return nil
}

// getActiveChatID определяет ID активного чата (может быть субагент)
func (e *Executor) getActiveChatID(ctx context.Context, chatID domain.ID) (domain.ID, error) {
	return e.subagentManager.GetActiveChatID(ctx, chatID)
}

// runAgentLoopStream - цикл агента со стримингом
func (e *Executor) runAgentLoopStream(
	ctx context.Context,
	chat *domain.Chat,
	agentDef *domain.AgentDefinition,
	execCtx *domain.ExecutionContext,
	stream service.ExecutionStream,
) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "executor.runAgentLoopStream")
	defer span.Finish()

	// Текущий контекст выполнения (может меняться при переходах к субагентам)
	currentChat := chat
	currentAgent := agentDef
	currentExecCtx := execCtx

	for range maxIterations {
		// Получаем актуальную историю текущего чата
		_, messages, err := e.chatManager.GetChatWithMessages(ctx, currentChat.ID)
		if err != nil {
			return stream.SendError(err)
		}

		// Строим контекст для LLM - просто конвертируем messages из БД
		llmMessages, err := e.buildLLMMessages(messages)
		if err != nil {
			return stream.SendError(err)
		}

		// Получаем инструменты текущего агента
		tools, err := e.buildLLMTools(currentAgent)
		if err != nil {
			return stream.SendError(err)
		}

		// Вызываем LLM стрим
		params := llm.ChatParams{
			Messages:     llmMessages,
			Tools:        tools,
			IncludeUsage: true,
		}

		llmStream, err := e.llmProvider.CreateCompletionStream(ctx, params)
		if err != nil {
			return stream.SendError(domain.NewInternalError("failed to create LLM stream", err))
		}

		// Собираем контент из стрима
		var contentBuilder strings.Builder
		var toolCalls []*domain.ToolCall
		toolCallsMap := make(map[int]*domain.ToolCall)

		for llmStream.Next() {
			chunk := llmStream.Chunk()
			content := chunk.Content

			// Отправляем текстовый контент
			if content != "" {
				contentBuilder.WriteString(content)
				if err := stream.SendChunk(content); err != nil {
					llmStream.Close()
					return err
				}
			}

			// Обрабатываем tool calls
			for _, tcDelta := range chunk.ToolCalls {
				tc, exists := toolCallsMap[tcDelta.Index]
				if !exists {
					tc = &domain.ToolCall{
						Model:  domain.NewModel(),
						Status: domain.ToolCallStatusPending,
					}
					toolCallsMap[tcDelta.Index] = tc
					toolCalls = append(toolCalls, tc)
				}

				if tcDelta.Name != "" {
					tc.Name = tcDelta.Name
				}
				if tcDelta.Arguments != "" {
					var args json.RawMessage
					if tc.Arguments != nil {
						args = tc.Arguments
					}
					args = append(args, []byte(tcDelta.Arguments)...)
					tc.Arguments = args
				}
			}
		}

		llmStream.Close()

		if err := llmStream.Err(); err != nil {
			return stream.SendError(err)
		}

		// Сохраняем сообщение ассистента в текущий чат
		sender := currentChat.AgentKey
		content := contentBuilder.String()

		assistantMessage := &domain.Message{
			Model:     domain.NewModel(),
			ChatID:    currentChat.ID,
			Role:      domain.MessageRoleAssistant,
			Content:   content,
			Sender:    &sender,
			ToolCalls: toolCalls,
		}

		if err := e.chatManager.SaveMessage(ctx, assistantMessage); err != nil {
			return stream.SendError(err)
		}

		// Отправляем финальное сообщение
		if err := stream.SendMessage(assistantMessage); err != nil {
			return err
		}

		// Если нет tool calls - завершаем цикл
		if len(toolCalls) == 0 {
			return nil
		}

		// Выполняем все tool calls
		hasActiveTools := false
		for _, toolCall := range toolCalls {
			// Отправляем событие о вызове tool
			if err := stream.SendToolCall(toolCall); err != nil {
				return err
			}

			// Парсим аргументы
			var arguments map[string]interface{}
			if err := json.Unmarshal(toolCall.Arguments, &arguments); err != nil {
				return stream.SendError(domain.NewInternalError("failed to parse tool arguments", err))
			}

			// Выполняем tool
			toolCall.MarkExecuting()
			result, err := e.toolExecutor.Execute(ctx, toolCall.Name, arguments, currentExecCtx, &toolCall.ID)

			var resultJSON []byte
			if err != nil {
				toolCall.Fail()
				resultJSON, _ = json.Marshal(map[string]interface{}{
					"error": err.Error(),
				})
			} else {
				resultJSON, _ = json.Marshal(result)
				toolCall.Complete(resultJSON)
			}

			// Специальная обработка для switch_to_subagent
			if toolCall.Name == string(domain.ToolNameSwitchToSubagent) && err == nil {
				// Извлекаем chat_id из результата безопасно (поддерживаем как domain.ID, так и string)
				subagentChatID, err := extractChatIDFromResult(result)
				if err != nil {
					return stream.SendError(err)
				}

				// Извлекаем subagent_key и task из аргументов с проверкой типов
				subagentKey, ok := arguments["subagent_key"].(string)
				if !ok || subagentKey == "" {
					return stream.SendError(domain.NewInvalidArgumentError("missing or invalid subagent_key argument"))
				}
				task, _ := arguments["task"].(string)

				// Получаем чат субагента
				subagentChat, err := e.chatManager.GetChat(ctx, subagentChatID)
				if err != nil {
					return stream.SendError(err)
				}

				// Получаем определение субагента
				subagentDef, err := e.agentManager.GetAgent(subagentKey)
				if err != nil {
					return stream.SendError(err)
				}

				// Создаем новый execution context для субагента
				subagentExecCtx := &domain.ExecutionContext{
					OrganizationID:    currentExecCtx.OrganizationID,
					UserID:            currentExecCtx.UserID,
					ChatID:            subagentChat.ID,
					AgentKey:          subagentKey,
					TaskDescription:   task,
					AdditionalContext: currentExecCtx.AdditionalContext,
				}

				// ПЕРЕКЛЮЧАЕМ КОНТЕКСТ на субагента
				currentChat = subagentChat
				currentAgent = subagentDef
				currentExecCtx = subagentExecCtx

				// Сохраняем system message с RAG для субагента
				subagentSystemPrompt, err := e.buildSystemPromptWithRAG(ctx, currentChat, currentAgent, task)
				if err != nil {
					return stream.SendError(err)
				}

				systemMessage := &domain.Message{
					Model:   domain.NewModel(),
					ChatID:  currentChat.ID,
					Role:    domain.MessageRoleSystem,
					Content: subagentSystemPrompt,
				}
				taskSystemMessage := &domain.Message{
					Model:   domain.NewModel(),
					ChatID:  currentChat.ID,
					Role:    domain.MessageRoleSystem,
					Content: fmt.Sprintf("Задача субагента: %s", task),
				}
				if err := e.chatManager.SaveMessage(ctx, systemMessage); err != nil {
					return stream.SendError(err)
				}
				if err := e.chatManager.SaveMessage(ctx, taskSystemMessage); err != nil {
					return stream.SendError(err)
				}

				hasActiveTools = true
				continue
			}

			// Специальная обработка для finish_subagent
			if toolCall.Name == string(domain.ToolNameFinishSubagent) && err == nil {
				// Проверяем, что мы действительно в субагенте
				if currentChat.ParentChatID == nil {
					return stream.SendError(domain.NewInvalidArgumentError("cannot finish_subagent from main agent"))
				}

				// Получаем родительский чат
				parentChat, err := e.chatManager.GetChat(ctx, *currentChat.ParentChatID)
				if err != nil {
					return stream.SendError(err)
				}

				// Получаем определение родительского агента
				parentAgentDef, err := e.agentManager.GetAgent(parentChat.AgentKey)
				if err != nil {
					return stream.SendError(err)
				}

				// Восстанавливаем execution context родителя
				parentExecCtx := &domain.ExecutionContext{
					OrganizationID:    currentExecCtx.OrganizationID,
					UserID:            currentExecCtx.UserID,
					ChatID:            parentChat.ID,
					AgentKey:          parentChat.AgentKey,
					TaskDescription:   "",
					AdditionalContext: currentExecCtx.AdditionalContext,
				}

				// Сохраняем результат субагента в родительский чат как tool result
				toolResultMessage := &domain.Message{
					Model:      domain.NewModel(),
					ChatID:     parentChat.ID,
					Role:       domain.MessageRoleTool,
					Content:    string(resultJSON),
					ToolCallID: currentChat.ParentToolCallID, // ссылка на tool call родителя
				}

				if err := e.chatManager.SaveMessage(ctx, toolResultMessage); err != nil {
					return stream.SendError(err)
				}

				// ПЕРЕКЛЮЧАЕМ КОНТЕКСТ обратно на родителя
				currentChat = parentChat
				currentAgent = parentAgentDef
				currentExecCtx = parentExecCtx

				hasActiveTools = true
				continue
			}

			// Обычные tools - сохраняем результат в текущий чат
			toolResultMessage := &domain.Message{
				Model:      domain.NewModel(),
				ChatID:     currentChat.ID,
				Role:       domain.MessageRoleTool,
				Content:    string(resultJSON),
				ToolCallID: &toolCall.ID,
			}

			if err := e.chatManager.SaveMessage(ctx, toolResultMessage); err != nil {
				return stream.SendError(err)
			}

			hasActiveTools = true
		}

		// Если были активные tools, продолжаем цикл с текущим контекстом
		if !hasActiveTools {
			break
		}
	}

	return nil
}

// extractChatIDFromResult безопасно извлекает ID чата из результата tool switch_to_subagent
func extractChatIDFromResult(result interface{}) (domain.ID, error) {
	// Ожидаем map[string]interface{}
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		return domain.ID{}, domain.NewInternalError("invalid switch_to_subagent result", nil)
	}

	// Пробуем как domain.ID (когда результат от внутреннего сервиса без marshaling)
	if id, ok := resultMap["chat_id"].(domain.ID); ok {
		return id, nil
	}

	// Пробуем как string (если где-то был marshal/unmarshal)
	if s, ok := resultMap["chat_id"].(string); ok && s != "" {
		id, err := domain.ParseID(s)
		if err != nil {
			return domain.ID{}, domain.NewInvalidArgumentError(fmt.Sprintf("invalid chat_id: %v", err))
		}
		return id, nil
	}

	return domain.ID{}, domain.NewInvalidArgumentError("missing or invalid chat_id in switch_to_subagent result")
}

// buildSystemPromptWithRAG строит system prompt с RAG
func (e *Executor) buildSystemPromptWithRAG(
	ctx context.Context,
	chat *domain.Chat,
	agentDef *domain.AgentDefinition,
	query string,
) (string, error) {
	systemPrompt := fmt.Sprintf("Текущее время: %s\n\n", time.Now().Format("2006-01-02 15:04:05"))
	systemPrompt += agentDef.GetSystemPrompt()

	// Обогащаем контекст фактами об организации
	orgContext, err := e.contextBuilder.EnrichWithOrganizationFacts(ctx, chat.OrganizationID)
	if err == nil && orgContext != "" {
		systemPrompt += orgContext
	}

	// Обогащаем контекст через RAG если есть query
	if query != "" {
		ragContext, err := e.contextBuilder.EnrichWithRAG(
			ctx,
			chat.OrganizationID,
			query,
			5, // топ-5 релевантных фрагментов
		)
		if err == nil && ragContext != "" {
			systemPrompt += ragContext
		}
	}

	return systemPrompt, nil
}

// buildLLMMessages строит сообщения для LLM из БД
func (e *Executor) buildLLMMessages(
	messages []*domain.Message,
) ([]llm.MessageParam, error) {
	llmMessages := make([]llm.MessageParam, 0, len(messages))

	// Добавляем историю сообщений
	for _, msg := range messages {
		llmMsg := llm.MessageParam{
			Role:    e.mapMessageRole(msg.Role),
			Content: msg.Content,
		}

		// Добавляем tool calls для assistant сообщений
		if msg.HasToolCalls() {
			llmMsg.ToolCalls = make([]llm.ToolCall, len(msg.ToolCalls))
			for i, tc := range msg.ToolCalls {
				llmMsg.ToolCalls[i] = llm.ToolCall{
					ID:        tc.ID.String(),
					Name:      tc.Name,
					Arguments: string(tc.Arguments),
				}
			}
		}

		// Добавляем tool call id для tool результатов
		if msg.IsToolResult() && msg.ToolCallID != nil {
			llmMsg.ToolCallID = msg.ToolCallID.String()
		}

		llmMessages = append(llmMessages, llmMsg)
	}

	return llmMessages, nil
}

// buildLLMTools строит список инструментов для LLM
func (e *Executor) buildLLMTools(agentDef *domain.AgentDefinition) ([]llm.ToolDefinition, error) {
	// Используем GetAgentTools для правильной обработки паттернов (например, "ammo-crm-*")
	agentTools, err := e.agentManager.GetAgentTools(agentDef.Key)
	if err != nil {
		return nil, err
	}

	tools := make([]llm.ToolDefinition, 0, len(agentTools))
	for _, toolDef := range agentTools {
		tools = append(tools, toolDef.ToLLMObject())
	}

	return tools, nil
}

// mapMessageRole маппит доменную роль в LLM роль
func (e *Executor) mapMessageRole(role domain.MessageRole) string {
	switch role {
	case domain.MessageRoleSystem:
		return llm.RoleSystem
	case domain.MessageRoleUser:
		return llm.RoleUser
	case domain.MessageRoleAssistant:
		return llm.RoleAssistant
	case domain.MessageRoleTool:
		return llm.RoleTool
	default:
		return llm.RoleUser
	}
}

// generateChatTitle генерирует название чата на основе первого сообщения пользователя
func (e *Executor) generateChatTitle(ctx context.Context, userMessage string) (string, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "executor.generateChatTitle")
	defer span.Finish()

	prompt := "Сгенерируй краткое и ёмкое название для этого чата на основе первого сообщения пользователя. Название должно быть коротким, не более 50 символов. В ответе укажи только одно название без лишних символов"

	// Используем специальную модель для генерации названий
	titleModel := e.cfg.GetLLMTitleGenerationModel()

	params := llm.ChatParams{
		Messages: []llm.MessageParam{
			{Role: llm.RoleSystem, Content: prompt},
			{Role: llm.RoleUser, Content: userMessage},
		},
		IncludeUsage: false,
		Model:        &titleModel,
	}

	logger.Infof(ctx, "Generating chat title with model: %s, prompt: %s", titleModel, prompt)

	title, _, err := e.llmProvider.CreateCompletion(ctx, params)
	if err != nil {
		return "", fmt.Errorf("failed to generate chat title: %w", err)
	}

	// Trim and limit length
	title = strings.TrimSpace(title)
	if len(title) > 50 {
		title = title[:50]
	}

	logger.Infof(ctx, "Generated chat title: %s", title)

	return title, nil
}
