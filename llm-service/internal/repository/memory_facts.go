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

// CreateUserMemoryFact inserts a new short fact for user.
func (r *PGXRepository) CreateUserMemoryFact(ctx context.Context, fact domain.UserMemoryFact) (domain.UserMemoryFact, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.CreateUserMemoryFact")
	defer span.Finish()

	engine := r.engineFactory.Get(ctx)
	query := `
        INSERT INTO llm_user_memory_facts (id, user_id, content)
        VALUES ($1, $2, $3)
        ON CONFLICT (user_id, content) DO NOTHING
        RETURNING id, user_id, content, created_at, updated_at
    `

	created := fact
	if err := pgxscan.Get(ctx, engine, &created, query,
		created.ID,
		created.UserID,
		created.Content,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// duplicate (user_id, content) — fetch existing
			var existing domain.UserMemoryFact
			sel := `SELECT id, user_id, content, created_at, updated_at FROM llm_user_memory_facts WHERE user_id = $1 AND content = $2`
			if gerr := pgxscan.Get(ctx, engine, &existing, sel, uuidToPgtype(fact.UserID), fact.Content); gerr == nil {
				return existing, nil
			}
			return domain.UserMemoryFact{}, nil
		}
		logger.Errorf(ctx, "failed to insert memory fact: %v", err)
		return domain.UserMemoryFact{}, err
	}

	return created, nil
}

// ListUserMemoryFacts returns facts for user ordered by created_at asc, limited.
func (r *PGXRepository) ListUserMemoryFacts(ctx context.Context, userID domain.ID, limit int) ([]domain.UserMemoryFact, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.ListUserMemoryFacts")
	defer span.Finish()

	engine := r.engineFactory.Get(ctx)
	query := `
        SELECT id, user_id, content, created_at, updated_at
        FROM llm_user_memory_facts
        WHERE user_id = $1
        ORDER BY created_at ASC
        LIMIT $2
    `
	var rows []domain.UserMemoryFact
	if err := pgxscan.Select(ctx, engine, &rows, query, uuidToPgtype(userID), limit); err != nil {
		logger.Errorf(ctx, "failed to list memory facts: %v", err)
		return nil, err
	}
	return rows, nil
}

// CountUserMemoryFacts returns current number of facts for the user.
func (r *PGXRepository) CountUserMemoryFacts(ctx context.Context, userID domain.ID) (int, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.CountUserMemoryFacts")
	defer span.Finish()

	engine := r.engineFactory.Get(ctx)
	query := `SELECT COUNT(*) AS cnt FROM llm_user_memory_facts WHERE user_id = $1`
	var row struct {
		Cnt int `db:"cnt"`
	}
	if err := pgxscan.Get(ctx, engine, &row, query, uuidToPgtype(userID)); err != nil {
		return 0, err
	}
	return row.Cnt, nil
}

// DeleteUserMemoryFact deletes a fact by id for user (ensures ownership).
func (r *PGXRepository) DeleteUserMemoryFact(ctx context.Context, userID domain.ID, factID domain.ID) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.DeleteUserMemoryFact")
	defer span.Finish()

	engine := r.engineFactory.Get(ctx)
	query := `DELETE FROM llm_user_memory_facts WHERE id = $1 AND user_id = $2`
	tag, err := engine.Exec(ctx, query, uuidToPgtype(factID), uuidToPgtype(userID))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
