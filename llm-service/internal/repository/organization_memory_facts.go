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

// CreateOrganizationMemoryFact inserts a new short fact for organization.
func (r *PGXRepository) CreateOrganizationMemoryFact(ctx context.Context, fact domain.OrganizationMemoryFact) (domain.OrganizationMemoryFact, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.CreateOrganizationMemoryFact")
	defer span.Finish()

	engine := r.engineFactory.Get(ctx)
	query := `
        INSERT INTO llm_organization_memory_facts (id, organization_id, content)
        VALUES ($1, $2, $3)
        ON CONFLICT (organization_id, content) DO NOTHING
        RETURNING id, organization_id, content, created_at, updated_at
    `

	created := fact
	if err := pgxscan.Get(ctx, engine, &created, query,
		created.ID,
		created.OrganizationID,
		created.Content,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// duplicate (organization_id, content) — fetch existing
			var existing domain.OrganizationMemoryFact
			sel := `SELECT id, organization_id, content, created_at, updated_at FROM llm_organization_memory_facts WHERE organization_id = $1 AND content = $2`
			if gerr := pgxscan.Get(ctx, engine, &existing, sel, uuidToPgtype(fact.OrganizationID), fact.Content); gerr == nil {
				return existing, nil
			}
			return domain.OrganizationMemoryFact{}, nil
		}
		logger.Errorf(ctx, "failed to insert organization memory fact: %v", err)
		return domain.OrganizationMemoryFact{}, err
	}

	return created, nil
}

// ListOrganizationMemoryFacts returns facts for organization ordered by created_at asc, limited.
func (r *PGXRepository) ListOrganizationMemoryFacts(ctx context.Context, organizationID domain.ID, limit int) ([]domain.OrganizationMemoryFact, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.ListOrganizationMemoryFacts")
	defer span.Finish()

	engine := r.engineFactory.Get(ctx)
	query := `
        SELECT id, organization_id, content, created_at, updated_at
        FROM llm_organization_memory_facts
        WHERE organization_id = $1
        ORDER BY created_at ASC
        LIMIT $2
    `
	var rows []domain.OrganizationMemoryFact
	if err := pgxscan.Select(ctx, engine, &rows, query, uuidToPgtype(organizationID), limit); err != nil {
		logger.Errorf(ctx, "failed to list organization memory facts: %v", err)
		return nil, err
	}
	return rows, nil
}

// CountOrganizationMemoryFacts returns current number of facts for the organization.
func (r *PGXRepository) CountOrganizationMemoryFacts(ctx context.Context, organizationID domain.ID) (int, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.CountOrganizationMemoryFacts")
	defer span.Finish()

	engine := r.engineFactory.Get(ctx)
	query := `SELECT COUNT(*) AS cnt FROM llm_organization_memory_facts WHERE organization_id = $1`
	var row struct {
		Cnt int `db:"cnt"`
	}
	if err := pgxscan.Get(ctx, engine, &row, query, uuidToPgtype(organizationID)); err != nil {
		return 0, err
	}
	return row.Cnt, nil
}

// DeleteOrganizationMemoryFact deletes a fact by id for organization (ensures ownership).
func (r *PGXRepository) DeleteOrganizationMemoryFact(ctx context.Context, organizationID domain.ID, factID domain.ID) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.DeleteOrganizationMemoryFact")
	defer span.Finish()

	engine := r.engineFactory.Get(ctx)
	query := `DELETE FROM llm_organization_memory_facts WHERE id = $1 AND organization_id = $2`
	tag, err := engine.Exec(ctx, query, uuidToPgtype(factID), uuidToPgtype(organizationID))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
