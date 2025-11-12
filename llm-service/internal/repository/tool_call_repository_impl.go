package repository

import (
	"context"
	"encoding/json"
	"errors"
	"llm-service/internal/domain"
	"time"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/opentracing/opentracing-go"
)

// toolCallRow - структура для маппинга из БД
type toolCallRow struct {
	ID          string     `db:"id"`
	MessageID   string     `db:"message_id"`
	Name        string     `db:"name"`
	Arguments   []byte     `db:"arguments"`
	Result      *[]byte    `db:"result"`
	Status      string     `db:"status"`
	CreatedAt   time.Time  `db:"created_at"`
	CompletedAt *time.Time `db:"completed_at"`
}

func (r *toolCallRow) toDomain() (*domain.ToolCall, error) {
	id, err := domain.ParseID(r.ID)
	if err != nil {
		return nil, err
	}

	tc := &domain.ToolCall{
		Model: domain.Model{
			ID:        id,
			CreatedAt: r.CreatedAt,
			UpdatedAt: r.CreatedAt,
		},
		Name:      r.Name,
		Arguments: json.RawMessage(r.Arguments),
		Status:    domain.ToolCallStatus(r.Status),
	}

	if r.Result != nil {
		tc.Result = json.RawMessage(*r.Result)
	}

	if r.CompletedAt != nil {
		tc.UpdatedAt = *r.CompletedAt
		tc.CompletedAt = r.CompletedAt
	}

	return tc, nil
}

// CreateToolCall создает новый tool call
func (r *PGXRepository) CreateToolCall(ctx context.Context, messageID domain.ID, toolCall *domain.ToolCall) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.CreateToolCall")
	defer span.Finish()

	engine := r.engineFactory.Get(ctx)

	query := `
		INSERT INTO tool_calls (id, message_id, name, arguments, result, status, created_at, completed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	var resultBytes *[]byte
	if toolCall.Result != nil {
		bytes := []byte(toolCall.Result)
		resultBytes = &bytes
	}

	var completedAt *time.Time
	if toolCall.Status == domain.ToolCallStatusCompleted || toolCall.Status == domain.ToolCallStatusFailed {
		completedAt = &toolCall.UpdatedAt
	}

	_, err := engine.Exec(ctx, query,
		toolCall.ID.String(),
		messageID.String(),
		toolCall.Name,
		[]byte(toolCall.Arguments),
		resultBytes,
		string(toolCall.Status),
		toolCall.CreatedAt,
		completedAt,
	)

	if err != nil {
		return domain.NewInternalError("failed to create tool call", err)
	}

	return nil
}

// GetToolCallByID получает tool call по ID
func (r *PGXRepository) GetToolCallByID(ctx context.Context, id domain.ID) (*domain.ToolCall, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.GetToolCallByID")
	defer span.Finish()

	engine := r.engineFactory.Get(ctx)

	query := `
		SELECT id, message_id, name, arguments, result, status, created_at, completed_at
		FROM tool_calls
		WHERE id = $1
	`

	var row toolCallRow
	if err := pgxscan.Get(ctx, engine, &row, query, id.String()); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NewNotFoundError("tool call not found")
		}
		return nil, domain.NewInternalError("failed to get tool call", err)
	}

	return row.toDomain()
}

// ListToolCallsByMessageID получает tool calls для сообщения
func (r *PGXRepository) ListToolCallsByMessageID(ctx context.Context, messageID domain.ID) ([]*domain.ToolCall, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.ListToolCallsByMessageID")
	defer span.Finish()

	engine := r.engineFactory.Get(ctx)

	query := `
		SELECT id, message_id, name, arguments, result, status, created_at, completed_at
		FROM tool_calls
		WHERE message_id = $1
		ORDER BY created_at ASC
	`

	var rows []toolCallRow
	if err := pgxscan.Select(ctx, engine, &rows, query, messageID.String()); err != nil {
		return nil, domain.NewInternalError("failed to list tool calls", err)
	}

	toolCalls := make([]*domain.ToolCall, 0, len(rows))
	for _, row := range rows {
		tc, err := row.toDomain()
		if err != nil {
			return nil, domain.NewInternalError("failed to convert tool call", err)
		}
		toolCalls = append(toolCalls, tc)
	}

	return toolCalls, nil
}

// UpdateToolCall обновляет tool call
func (r *PGXRepository) UpdateToolCall(ctx context.Context, toolCall *domain.ToolCall) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.UpdateToolCall")
	defer span.Finish()

	engine := r.engineFactory.Get(ctx)

	var resultBytes *[]byte
	if toolCall.Result != nil {
		bytes := []byte(toolCall.Result)
		resultBytes = &bytes
	}

	var completedAt *time.Time
	if toolCall.Status == domain.ToolCallStatusCompleted || toolCall.Status == domain.ToolCallStatusFailed {
		completedAt = &toolCall.UpdatedAt
	}

	query := `
		UPDATE tool_calls
		SET result = $2, status = $3, completed_at = $4
		WHERE id = $1
	`

	tag, err := engine.Exec(ctx, query,
		toolCall.ID.String(),
		resultBytes,
		string(toolCall.Status),
		completedAt,
	)

	if err != nil {
		return domain.NewInternalError("failed to update tool call", err)
	}

	if tag.RowsAffected() == 0 {
		return domain.NewNotFoundError("tool call not found")
	}

	return nil
}

// UpdateToolCallStatus обновляет статус tool call
func (r *PGXRepository) UpdateToolCallStatus(ctx context.Context, id domain.ID, status domain.ToolCallStatus) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.UpdateToolCallStatus")
	defer span.Finish()

	engine := r.engineFactory.Get(ctx)

	query := `
		UPDATE tool_calls
		SET status = $2
		WHERE id = $1
	`

	tag, err := engine.Exec(ctx, query, id.String(), string(status))
	if err != nil {
		return domain.NewInternalError("failed to update tool call status", err)
	}

	if tag.RowsAffected() == 0 {
		return domain.NewNotFoundError("tool call not found")
	}

	return nil
}

// CompleteToolCall завершает tool call с результатом
func (r *PGXRepository) CompleteToolCall(ctx context.Context, id domain.ID, result []byte) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.CompleteToolCall")
	defer span.Finish()

	engine := r.engineFactory.Get(ctx)

	query := `
		UPDATE tool_calls
		SET result = $2, status = $3, completed_at = NOW()
		WHERE id = $1
	`

	tag, err := engine.Exec(ctx, query,
		id.String(),
		result,
		string(domain.ToolCallStatusCompleted),
	)

	if err != nil {
		return domain.NewInternalError("failed to complete tool call", err)
	}

	if tag.RowsAffected() == 0 {
		return domain.NewNotFoundError("tool call not found")
	}

	return nil
}
