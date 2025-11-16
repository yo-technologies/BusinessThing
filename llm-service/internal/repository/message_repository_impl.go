package repository

import (
	"context"
	"errors"
	"llm-service/internal/domain"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/opentracing/opentracing-go"
)

// messageRow - структура для маппинга из БД
type messageRow struct {
	ID         string    `db:"id"`
	ChatID     string    `db:"chat_id"`
	Role       string    `db:"role"`
	Content    string    `db:"content"`
	Sender     *string   `db:"sender"`
	ToolCallID *string   `db:"tool_call_id"`
	CreatedAt  time.Time `db:"created_at"`
}

func (r *messageRow) toDomain() (*domain.Message, error) {
	id, err := domain.ParseID(r.ID)
	if err != nil {
		return nil, err
	}

	chatID, err := domain.ParseID(r.ChatID)
	if err != nil {
		return nil, err
	}

	msg := &domain.Message{
		Model: domain.Model{
			ID:        id,
			CreatedAt: r.CreatedAt,
			UpdatedAt: r.CreatedAt, // для messages нет updated_at
		},
		ChatID:  chatID,
		Role:    domain.MessageRole(r.Role),
		Content: r.Content,
		Sender:  r.Sender,
	}

	if r.ToolCallID != nil {
		toolCallID, err := domain.ParseID(*r.ToolCallID)
		if err != nil {
			return nil, err
		}
		msg.ToolCallID = &toolCallID
	}

	return msg, nil
}

// CreateMessage создает новое сообщение
func (r *PGXRepository) CreateMessage(ctx context.Context, message *domain.Message) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.CreateMessage")
	defer span.Finish()

	engine := r.engineFactory.Get(ctx)

	query := `
		INSERT INTO messages (id, chat_id, role, content, sender, tool_call_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := engine.Exec(ctx, query,
		message.ID.String(),
		message.ChatID.String(),
		string(message.Role),
		message.Content,
		message.Sender,
		nullableIDToString(message.ToolCallID),
		message.CreatedAt,
	)

	if err != nil {
		return domain.NewInternalError("failed to create message", err)
	}

	return nil
}

// GetMessageByID получает сообщение по ID
func (r *PGXRepository) GetMessageByID(ctx context.Context, id domain.ID) (*domain.Message, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.GetMessageByID")
	defer span.Finish()

	engine := r.engineFactory.Get(ctx)

	query := `
		SELECT id, chat_id, role, content, sender, tool_call_id, created_at
		FROM messages
		WHERE id = $1
	`

	var row messageRow
	if err := pgxscan.Get(ctx, engine, &row, query, id.String()); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NewNotFoundError("message not found")
		}
		return nil, domain.NewInternalError("failed to get message", err)
	}

	return row.toDomain()
}

// ListMessagesByChatID получает список сообщений чата
func (r *PGXRepository) ListMessagesByChatID(ctx context.Context, chatID domain.ID, limit, offset int) ([]*domain.Message, int, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.ListMessagesByChatID")
	defer span.Finish()

	engine := r.engineFactory.Get(ctx)

	// Подсчет общего количества
	countQuery := `SELECT COUNT(*) AS count FROM messages WHERE chat_id = $1`
	var countResult []struct {
		Count int `db:"count"`
	}
	if err := pgxscan.Select(ctx, engine, &countResult, countQuery, chatID.String()); err != nil {
		return nil, 0, domain.NewInternalError("failed to count messages", err)
	}

	total := 0
	if len(countResult) > 0 {
		total = countResult[0].Count
	}

	// Получение списка
	query := `
		SELECT id, chat_id, role, content, sender, tool_call_id, created_at
		FROM messages
		WHERE chat_id = $1
		ORDER BY created_at ASC
		LIMIT $2 OFFSET $3
	`

	var rows []messageRow
	if err := pgxscan.Select(ctx, engine, &rows, query, chatID.String(), limit, offset); err != nil {
		return nil, 0, domain.NewInternalError("failed to list messages", err)
	}

	messages := make([]*domain.Message, 0, len(rows))
	for _, row := range rows {
		msg, err := row.toDomain()
		if err != nil {
			return nil, 0, domain.NewInternalError("failed to convert message", err)
		}
		messages = append(messages, msg)
	}

	return messages, total, nil
}

// ListMessagesByChatIDWithToolCalls получает сообщения вместе с tool calls
func (r *PGXRepository) ListMessagesByChatIDWithToolCalls(ctx context.Context, chatID domain.ID, limit, offset int) ([]*domain.Message, int, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.ListMessagesByChatIDWithToolCalls")
	defer span.Finish()

	// Сначала получаем сообщения
	messages, total, err := r.ListMessagesByChatID(ctx, chatID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	// Загружаем tool calls для каждого сообщения
	for _, msg := range messages {
		toolCalls, err := r.ListToolCallsByMessageID(ctx, msg.ID)
		if err != nil {
			return nil, 0, err
		}
		msg.ToolCalls = toolCalls
	}

	return messages, total, nil
}

// ListMessagesWithSubchatsWithToolCalls получает сообщения родительского чата и всех его субчатов с tool calls
func (r *PGXRepository) ListMessagesWithSubchatsWithToolCalls(ctx context.Context, parentChatID domain.ID, limit, offset int) ([]*domain.Message, int, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.ListMessagesWithSubchatsWithToolCalls")
	defer span.Finish()

	engine := r.engineFactory.Get(ctx)

	// Создаем подзапрос для получения ID субчатов (не используем SelectBuilder напрямую в Where,
	// чтобы не передавать его как аргумент в запрос и не вызывать проблему кодирования).

	// Подсчет общего количества сообщений из родительского чата и всех субчатов
	countQb := sq.Select("COUNT(*) AS count").From("messages").PlaceholderFormat(sq.Dollar)
	// Используем SQL-выражение с подзапросом (вставляем placeholder для parentChatID),
	// чтобы не передавать SelectBuilder как аргумент в запрос (это вызывает ошибку кодирования).
	countQb = countQb.Where(sq.Expr("(chat_id = ? OR chat_id IN (SELECT id FROM chats WHERE parent_chat_id = ?))", parentChatID.String(), parentChatID.String()))

	countQuery, countArgs, err := countQb.ToSql()
	if err != nil {
		return nil, 0, domain.NewInternalError("failed to build count query", err)
	}

	var countResult []struct {
		Count int `db:"count"`
	}
	if err := pgxscan.Select(ctx, engine, &countResult, countQuery, countArgs...); err != nil {
		return nil, 0, domain.NewInternalError("failed to count messages with subchats", err)
	}

	total := 0
	if len(countResult) > 0 {
		total = countResult[0].Count
	}

	// Получение сообщений из родительского чата и всех субчатов
	qb := sq.Select(
		"id", "chat_id", "role", "content", "sender", "tool_call_id", "created_at",
	).From("messages").PlaceholderFormat(sq.Dollar)

	qb = qb.Where(sq.Expr("(chat_id = ? OR chat_id IN (SELECT id FROM chats WHERE parent_chat_id = ?))", parentChatID.String(), parentChatID.String()))

	qb = qb.OrderBy("created_at DESC").Limit(uint64(limit)).Offset(uint64(offset))

	query, args, err := qb.ToSql()
	if err != nil {
		return nil, 0, domain.NewInternalError("failed to build list query", err)
	}

	var rows []messageRow
	if err := pgxscan.Select(ctx, engine, &rows, query, args...); err != nil {
		return nil, 0, domain.NewInternalError("failed to list messages with subchats", err)
	}

	messages := make([]*domain.Message, 0, len(rows))
	for _, row := range rows {
		msg, err := row.toDomain()
		if err != nil {
			return nil, 0, domain.NewInternalError("failed to convert message", err)
		}
		messages = append(messages, msg)
	}

	// Загружаем tool calls для каждого сообщения
	for _, msg := range messages {
		toolCalls, err := r.ListToolCallsByMessageID(ctx, msg.ID)
		if err != nil {
			return nil, 0, err
		}
		msg.ToolCalls = toolCalls
	}

	return messages, total, nil
}

// DeleteMessage удаляет сообщение
func (r *PGXRepository) DeleteMessage(ctx context.Context, id domain.ID) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.DeleteMessage")
	defer span.Finish()

	engine := r.engineFactory.Get(ctx)

	query := `DELETE FROM messages WHERE id = $1`

	tag, err := engine.Exec(ctx, query, id.String())
	if err != nil {
		return domain.NewInternalError("failed to delete message", err)
	}

	if tag.RowsAffected() == 0 {
		return domain.NewNotFoundError("message not found")
	}

	return nil
}
