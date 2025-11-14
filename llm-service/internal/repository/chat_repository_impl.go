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

// chatRow - структура для маппинга из БД
type chatRow struct {
	ID               string    `db:"id"`
	OrganizationID   string    `db:"organization_id"`
	UserID           string    `db:"user_id"`
	AgentKey         string    `db:"agent_key"`
	Title            string    `db:"title"`
	Status           string    `db:"status"`
	ParentChatID     *string   `db:"parent_chat_id"`
	ParentToolCallID *string   `db:"parent_tool_call_id"`
	CreatedAt        time.Time `db:"created_at"`
	UpdatedAt        time.Time `db:"updated_at"`
}

func (r *chatRow) toDomain() (*domain.Chat, error) {
	id, err := domain.ParseID(r.ID)
	if err != nil {
		return nil, err
	}

	orgID, err := domain.ParseID(r.OrganizationID)
	if err != nil {
		return nil, err
	}

	userID, err := domain.ParseID(r.UserID)
	if err != nil {
		return nil, err
	}

	chat := &domain.Chat{
		Model: domain.Model{
			ID:        id,
			CreatedAt: r.CreatedAt,
			UpdatedAt: r.UpdatedAt,
		},
		OrganizationID: orgID,
		UserID:         userID,
		AgentKey:       r.AgentKey,
		Title:          r.Title,
		Status:         domain.ChatStatus(r.Status),
	}

	if r.ParentChatID != nil {
		parentID, err := domain.ParseID(*r.ParentChatID)
		if err != nil {
			return nil, err
		}
		chat.ParentChatID = &parentID
	}

	if r.ParentToolCallID != nil {
		toolCallID, err := domain.ParseID(*r.ParentToolCallID)
		if err != nil {
			return nil, err
		}
		chat.ParentToolCallID = &toolCallID
	}

	return chat, nil
}

// Create создает новый чат
func (r *PGXRepository) CreateChat(ctx context.Context, chat *domain.Chat) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.CreateChat")
	defer span.Finish()

	engine := r.engineFactory.Get(ctx)

	query := `
		INSERT INTO chats (
			id, organization_id, user_id, agent_key, title, status,
			parent_chat_id, parent_tool_call_id, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := engine.Exec(ctx, query,
		chat.ID.String(),
		chat.OrganizationID.String(),
		chat.UserID.String(),
		chat.AgentKey,
		chat.Title,
		string(chat.Status),
		nullableIDToString(chat.ParentChatID),
		nullableIDToString(chat.ParentToolCallID),
		chat.CreatedAt,
		chat.UpdatedAt,
	)

	if err != nil {
		return domain.NewInternalError("failed to create chat", err)
	}

	return nil
}

// GetChatByID получает чат по ID
func (r *PGXRepository) GetChatByID(ctx context.Context, id domain.ID) (*domain.Chat, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.GetChatByID")
	defer span.Finish()

	engine := r.engineFactory.Get(ctx)

	query := `
		SELECT id, organization_id, user_id, agent_key, title, status,
		       parent_chat_id, parent_tool_call_id, created_at, updated_at
		FROM chats
		WHERE id = $1
	`

	var row chatRow
	if err := pgxscan.Get(ctx, engine, &row, query, id.String()); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NewNotFoundError("chat not found")
		}
		return nil, domain.NewInternalError("failed to get chat", err)
	}

	return row.toDomain()
}

// ListChats получает список чатов с фильтрацией и пагинацией
func (r *PGXRepository) ListChats(ctx context.Context, filter ChatFilter, page, pageSize int) ([]*domain.Chat, int, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.ListChats")
	defer span.Finish()

	engine := r.engineFactory.Get(ctx)

	// Построение WHERE условий с помощью squirrel
	qb := sq.Select("COUNT(*)").From("chats").PlaceholderFormat(sq.Dollar)

	if filter.OrganizationID != nil {
		qb = qb.Where(sq.Eq{"organization_id": filter.OrganizationID.String()})
	}

	if filter.UserID != nil {
		qb = qb.Where(sq.Eq{"user_id": filter.UserID.String()})
	}

	if filter.Status != nil {
		qb = qb.Where(sq.Eq{"status": string(*filter.Status)})
	}

	if filter.ParentChatID != nil {
		qb = qb.Where(sq.Eq{"parent_chat_id": filter.ParentChatID.String()})
	} else {
		// По умолчанию исключаем субчаты (где parent_chat_id IS NOT NULL)
		qb = qb.Where(sq.Eq{"parent_chat_id": nil})
	}

	// Подсчет общего количества
	countQuery, args, err := qb.ToSql()
	if err != nil {
		return nil, 0, domain.NewInternalError("failed to build count query", err)
	}

	var countResult []struct {
		Count int `db:"count"`
	}
	if err := pgxscan.Select(ctx, engine, &countResult, countQuery, args...); err != nil {
		return nil, 0, domain.NewInternalError("failed to count chats", err)
	}

	total := 0
	if len(countResult) > 0 {
		total = countResult[0].Count
	}

	// Получение списка с пагинацией
	offset := (page - 1) * pageSize
	qb = sq.Select(
		"id", "organization_id", "user_id", "agent_key", "title", "status",
		"parent_chat_id", "parent_tool_call_id", "created_at", "updated_at",
	).From("chats").PlaceholderFormat(sq.Dollar)

	if filter.OrganizationID != nil {
		qb = qb.Where(sq.Eq{"organization_id": filter.OrganizationID.String()})
	}

	if filter.UserID != nil {
		qb = qb.Where(sq.Eq{"user_id": filter.UserID.String()})
	}

	if filter.Status != nil {
		qb = qb.Where(sq.Eq{"status": string(*filter.Status)})
	}

	if filter.ParentChatID != nil {
		qb = qb.Where(sq.Eq{"parent_chat_id": filter.ParentChatID.String()})
	} else {
		qb = qb.Where(sq.Eq{"parent_chat_id": nil})
	}

	qb = qb.OrderBy("created_at DESC").Limit(uint64(pageSize)).Offset(uint64(offset))

	query, args, err := qb.ToSql()
	if err != nil {
		return nil, 0, domain.NewInternalError("failed to build list query", err)
	}

	var chatRows []chatRow
	if err := pgxscan.Select(ctx, engine, &chatRows, query, args...); err != nil {
		return nil, 0, domain.NewInternalError("failed to list chats", err)
	}

	chats := make([]*domain.Chat, 0, len(chatRows))
	for _, row := range chatRows {
		chat, err := row.toDomain()
		if err != nil {
			return nil, 0, domain.NewInternalError("failed to convert chat", err)
		}
		chats = append(chats, chat)
	}

	return chats, total, nil
}

// UpdateChat обновляет чат
func (r *PGXRepository) UpdateChat(ctx context.Context, chat *domain.Chat) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.UpdateChat")
	defer span.Finish()

	engine := r.engineFactory.Get(ctx)

	query := `
		UPDATE chats
		SET title = $2, status = $3, updated_at = $4
		WHERE id = $1
	`

	tag, err := engine.Exec(ctx, query,
		chat.ID.String(),
		chat.Title,
		string(chat.Status),
		chat.UpdatedAt,
	)

	if err != nil {
		return domain.NewInternalError("failed to update chat", err)
	}

	if tag.RowsAffected() == 0 {
		return domain.NewNotFoundError("chat not found")
	}

	return nil
}

// DeleteChat удаляет чат
func (r *PGXRepository) DeleteChat(ctx context.Context, id, userID, orgID domain.ID) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.DeleteChat")
	defer span.Finish()

	engine := r.engineFactory.Get(ctx)

	query := `DELETE FROM chats WHERE id = $1 AND user_id = $2 AND organization_id = $3`

	tag, err := engine.Exec(ctx, query, id.String(), userID.String(), orgID.String())
	if err != nil {
		return domain.NewInternalError("failed to delete chat", err)
	}

	if tag.RowsAffected() == 0 {
		return domain.NewNotFoundError("chat not found")
	}

	return nil
}

// GetChatByParentToolCallID получает дочерний чат по ID tool call родителя
func (r *PGXRepository) GetChatByParentToolCallID(ctx context.Context, parentToolCallID domain.ID) (*domain.Chat, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.GetChatByParentToolCallID")
	defer span.Finish()

	engine := r.engineFactory.Get(ctx)

	query := `
		SELECT id, organization_id, user_id, agent_key, title, status,
		       parent_chat_id, parent_tool_call_id, created_at, updated_at
		FROM chats
		WHERE parent_tool_call_id = $1
	`

	var row chatRow
	if err := pgxscan.Get(ctx, engine, &row, query, parentToolCallID.String()); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NewNotFoundError("chat not found")
		}
		return nil, domain.NewInternalError("failed to get chat", err)
	}

	return row.toDomain()
}

// GetActiveChildChat получает активный дочерний чат для родительского чата
func (r *PGXRepository) GetActiveChildChat(ctx context.Context, parentChatID domain.ID) (*domain.Chat, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.GetActiveChildChat")
	defer span.Finish()

	engine := r.engineFactory.Get(ctx)

	query := `
		SELECT id, organization_id, user_id, agent_key, title, status,
		       parent_chat_id, parent_tool_call_id, created_at, updated_at
		FROM chats
		WHERE parent_chat_id = $1 AND status = $2
		ORDER BY created_at DESC
		LIMIT 1
	`

	var row chatRow
	if err := pgxscan.Get(ctx, engine, &row, query, parentChatID.String(), string(domain.ChatStatusActive)); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NewNotFoundError("active child chat not found")
		}
		return nil, domain.NewInternalError("failed to get active child chat", err)
	}

	return row.toDomain()
}
