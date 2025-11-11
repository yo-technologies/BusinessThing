package chat

import (
	"context"
	"fmt"
	"strings"
	"time"

	"llm-service/internal/llm"

	"llm-service/internal/domain"
	"llm-service/internal/domain/dto"

	"github.com/guregu/null/v6"
	"github.com/opentracing/opentracing-go"
)

const (
	chatHistoryLimit       = 200
	maxChatCompletionLoops = 25
	assistantErrorMessage  = "Не удалось ответить. Попробуйте ещё раз чуть позже."
)

const defaultChatSystemPrompt = `Ты — виртуальный банковский помощник и агент. Отвечай конкретно и понятно, без выдумок.

СТИЛЬ РЕЧИ:
Говори живо и просто, по‑пацански. КРАЙНЕ ЧАСТО используй слово "брат". Используй искренние, прямые фразы, не усложняй. Подчёркивай дружбу, братство, уличный слог. Темы — про жизнь на районе, машины, мечты, боль и любовь. Стиль эмоциональный, с долей самоиронии и ноткой гордости. Общайся как с другом, не как с аудиторией.»

Правила:
- Если нужны данные из систем (счета, транзакции, переводы) — используй инструменты.
- Никогда не придумывай реквизиты, суммы или статусы операций.
- Объясняй коротко, шаг за шагом, и предлагай следующий шаг.
- Если инструмент не нужен (приветствие, уточнение, справка) — отвечай сразу.
`

type chatSession struct {
	chat     domain.Chat
	messages []llm.MessageParam
	toolDefs []llm.ToolDefinition
}

func (s *Service) SendChatMessageStream(ctx context.Context, userID domain.ID, req dto.SendChatMessageRequest, callbacks dto.ChatStreamCallbacks) (dto.ChatCompletionDTO, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.chat.SendChatMessageStream")
	defer span.Finish()

	session, err := s.startChatSession(ctx, userID, req)
	if err != nil {
		return dto.ChatCompletionDTO{}, err
	}

	messages := make([]llm.MessageParam, len(session.messages))
	copy(messages, session.messages)
	toolDefs := session.toolDefs
	chat := session.chat

	// Reserve 1 token upfront for admission control
	allowed, err := s.quotaService.Reserve(ctx, userID, 1)
	if err != nil {
		return dto.ChatCompletionDTO{}, err
	}
	if !allowed {
		// Return a typed quota-exceeded error and try to emit a structured error event
		te := domain.QuotaExceededError(domain.ErrTooManyRequests)
		if callbacks.OnError != nil {
			_ = callbacks.OnError(te)
		}

		_, _ = s.handleAssistantFailure(ctx, chat.ID, te)
		return dto.ChatCompletionDTO{}, te
	}
	return s.runAgentCycle(ctx, userID, chat, messages, toolDefs, callbacks)
}

func (s *Service) startChatSession(ctx context.Context, userID domain.ID, req dto.SendChatMessageRequest) (chatSession, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.chat.startChatSession")
	defer span.Finish()

	content := strings.TrimSpace(req.Content)
	if content == "" {
		return chatSession{}, fmt.Errorf("message content cannot be empty: %w", domain.ErrInvalidArgument)
	}

	chat, err := s.ensureChatForRequest(ctx, userID, req)
	if err != nil {
		return chatSession{}, err
	}

	userMessage, err := domain.NewChatMessage(
		chat.ID,
		domain.ChatMessageRoleUser,
		content,
		null.String{},
		null.String{},
		nil,
	)
	if err != nil {
		return chatSession{}, err
	}

	if _, err := s.chatRepository.CreateChatMessage(ctx, userMessage); err != nil {
		return chatSession{}, fmt.Errorf("failed to save user chat message: %w", err)
	}

	system := s.buildSystemPrompt(ctx, chat.UserID)
	systemMessages := []llm.MessageParam{{Role: llm.RoleSystem, Content: system}}

	history, err := s.chatRepository.ListChatMessages(ctx, chat.ID, chatHistoryLimit, 0)
	if err != nil {
		return chatSession{}, fmt.Errorf("failed to load chat history: %w", err)
	}

	messages := make([]llm.MessageParam, 0, len(systemMessages)+len(history)+4)
	messages = append(messages, systemMessages...)

	for _, msg := range history {
		param, err := s.chatMessageToLLMParam(msg)
		if err != nil {
			return chatSession{}, fmt.Errorf("failed to convert chat message to OpenAI param: %w", err)
		}
		messages = append(messages, param)
	}

	return chatSession{
		chat:     chat,
		messages: messages,
		toolDefs: s.toolsService.ChatAgentToolDefinitions(),
	}, nil
}

func (s *Service) ensureChatForRequest(ctx context.Context, userID domain.ID, req dto.SendChatMessageRequest) (domain.Chat, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.chat.ensureChatForRequest")
	defer span.Finish()

	if req.ChatID.Valid {
		cid, err := domain.ParseID(req.ChatID.String)
		if err != nil {
			return domain.Chat{}, err
		}
		chat, err := s.chatRepository.GetChatByID(ctx, cid)
		if err != nil {
			return domain.Chat{}, err
		}
		if chat.UserID != userID {
			return domain.Chat{}, domain.ErrForbidden
		}
		return chat, nil
	}

	title := fmt.Sprintf("Чат %s", time.Now().Format("02.01.2006 15:04"))
	newChat := domain.NewChat(userID, title)

	createdChat, err := s.chatRepository.CreateChat(ctx, newChat)
	if err != nil {
		return domain.Chat{}, fmt.Errorf("failed to create chat: %w", err)
	}

	return createdChat, nil
}
