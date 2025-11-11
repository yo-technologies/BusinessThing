package repository

import (
	"context"
	"errors"

	"llm-service/internal/domain"
	"llm-service/internal/logger"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/opentracing/opentracing-go"
)

func (r *PGXRepository) CreateChat(ctx context.Context, chat domain.Chat) (domain.Chat, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.CreateChat")
	defer span.Finish()

	query := `
		INSERT INTO llm_chats (id, user_id, title)
		VALUES ($1, $2, $3)
		RETURNING id, user_id, title, created_at, updated_at
	`

	engine := r.engineFactory.Get(ctx)

	var created domain.Chat
	if err := pgxscan.Get(ctx, engine, &created, query,
		uuidToPgtype(chat.ID),
		uuidToPgtype(chat.UserID),
		chat.Title,
	); err != nil {
		logger.Errorf(ctx, "failed to create chat: %v", err)
		return domain.Chat{}, err
	}

	return created, nil
}

func (r *PGXRepository) GetChatByID(ctx context.Context, chatID domain.ID) (domain.Chat, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.GetChatByID")
	defer span.Finish()

	query := `
		SELECT id, user_id, title, created_at, updated_at
		FROM llm_chats
		WHERE id = $1
	`

	engine := r.engineFactory.Get(ctx)

	var entity domain.Chat
	if err := pgxscan.Get(ctx, engine, &entity, query, uuidToPgtype(chatID)); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Chat{}, domain.ErrNotFound
		}
		logger.Errorf(ctx, "failed to get chat by id: %v", err)
		return domain.Chat{}, err
	}

	return entity, nil
}

func (r *PGXRepository) GetLatestChat(ctx context.Context, userID domain.ID) (domain.Chat, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.GetLatestChat")
	defer span.Finish()

	query := `
		SELECT id, user_id, title, created_at, updated_at
		FROM llm_chats
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	engine := r.engineFactory.Get(ctx)

	var entity domain.Chat
	if err := pgxscan.Get(ctx, engine, &entity, query, uuidToPgtype(userID)); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Chat{}, domain.ErrNotFound
		}
		logger.Errorf(ctx, "failed to get latest chat: %v", err)
		return domain.Chat{}, err
	}

	return entity, nil
}

func (r *PGXRepository) CreateChatMessage(ctx context.Context, message domain.ChatMessage) (domain.ChatMessage, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.CreateChatMessage")
	defer span.Finish()

	query := `
		INSERT INTO llm_chat_messages (id, chat_id, role, content, tool_name, tool_call_id, tool_arguments, token_usage, error, tool_state)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, chat_id, role, content, tool_name, tool_call_id, tool_arguments, token_usage, error, tool_state, created_at, updated_at
	`

	engine := r.engineFactory.Get(ctx)

	var created domain.ChatMessage
	if err := pgxscan.Get(ctx, engine, &created, query,
		uuidToPgtype(message.ID),
		uuidToPgtype(message.ChatID),
		message.Role,
		message.Content,
		message.ToolName,
		message.ToolCallID,
		message.ToolArguments,
		message.TokenUsage,
		message.Error,
		message.ToolState,
	); err != nil {
		logger.Errorf(ctx, "failed to create chat message: %v", err)
		return domain.ChatMessage{}, err
	}

	return created, nil
}

func (r *PGXRepository) ListChatMessages(ctx context.Context, chatID domain.ID, limit, offset int) ([]domain.ChatMessage, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.ListChatMessages")
	defer span.Finish()

	query := `
		SELECT id, chat_id, role, content, tool_name, tool_call_id, tool_arguments, token_usage, error, tool_state, created_at, updated_at
		FROM llm_chat_messages
		WHERE chat_id = $1
		ORDER BY created_at ASC
		LIMIT $2 OFFSET $3
	`

	engine := r.engineFactory.Get(ctx)

	var entities []domain.ChatMessage
	if err := pgxscan.Select(ctx, engine, &entities, query, uuidToPgtype(chatID), limit, offset); err != nil {
		logger.Errorf(ctx, "failed to list chat messages: %v", err)
		return nil, err
	}

	return entities, nil
}

func (r *PGXRepository) GetMessageByToolCallID(ctx context.Context, chatID domain.ID, toolCallID string) (domain.ChatMessage, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.GetMessageByToolCallID")
	defer span.Finish()

	query := `
        SELECT id, chat_id, role, content, tool_name, tool_call_id, tool_arguments, token_usage, error, tool_state, created_at, updated_at
        FROM llm_chat_messages
        WHERE chat_id = $1 AND tool_call_id = $2
        ORDER BY created_at DESC
        LIMIT 1
    `

	engine := r.engineFactory.Get(ctx)

	var msg domain.ChatMessage
	if err := pgxscan.Get(ctx, engine, &msg, query, uuidToPgtype(chatID), toolCallID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ChatMessage{}, domain.ErrNotFound
		}
		return domain.ChatMessage{}, err
	}

	return msg, nil
}

func (r *PGXRepository) UpdateMessageToolState(ctx context.Context, messageID domain.ID, state domain.ToolState) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.UpdateMessageToolState")
	defer span.Finish()

	query := `
        UPDATE llm_chat_messages
        SET tool_state = $2, updated_at = NOW()
        WHERE id = $1
    `

	engine := r.engineFactory.Get(ctx)
	if _, err := engine.Exec(ctx, query, uuidToPgtype(messageID), state); err != nil {
		return err
	}
	return nil
}
