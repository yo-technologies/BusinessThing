package repository

import (
	"context"
	"core-service/internal/domain"
	"core-service/internal/logger"
	"errors"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/opentracing/opentracing-go"
)

// normalizeProfileData ensures profile_data is valid JSON for JSONB column
func normalizeProfileData(profileData string) string {
	if profileData == "" {
		return "{}"
	}
	return profileData
}

// CreateOrganization inserts a new organization
func (r *PGXRepository) CreateOrganization(ctx context.Context, org domain.Organization) (domain.Organization, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.CreateOrganization")
	defer span.Finish()

	engine := r.engineFactory.Get(ctx)
	query := `
        INSERT INTO organizations (id, name, industry, region, description, profile_data, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        RETURNING id, name, industry, region, description, profile_data, created_at, updated_at, deleted_at
    `

	var created domain.Organization
	err := pgxscan.Get(ctx, engine, &created, query,
		uuidToPgtype(org.ID),
		org.Name,
		org.Industry,
		org.Region,
		org.Description,
		normalizeProfileData(org.ProfileData),
		org.CreatedAt,
		org.UpdatedAt,
	)
	if err != nil {
		logger.Errorf(ctx, "failed to create organization: %v", err)
		return domain.Organization{}, err
	}

	return created, nil
}

// GetOrganization retrieves an organization by ID
func (r *PGXRepository) GetOrganization(ctx context.Context, id domain.ID) (domain.Organization, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.GetOrganization")
	defer span.Finish()

	engine := r.engineFactory.Get(ctx)
	query := `
        SELECT id, name, industry, region, description, profile_data, created_at, updated_at, deleted_at
        FROM organizations
        WHERE id = $1 AND deleted_at IS NULL
    `

	var org domain.Organization
	err := pgxscan.Get(ctx, engine, &org, query, uuidToPgtype(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Organization{}, domain.ErrNotFound
		}
		logger.Errorf(ctx, "failed to get organization: %v", err)
		return domain.Organization{}, err
	}

	return org, nil
}

// UpdateOrganization updates an existing organization
func (r *PGXRepository) UpdateOrganization(ctx context.Context, org domain.Organization) (domain.Organization, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.UpdateOrganization")
	defer span.Finish()

	engine := r.engineFactory.Get(ctx)
	query := `
        UPDATE organizations
        SET name = $2, industry = $3, region = $4, description = $5, profile_data = $6, updated_at = $7
        WHERE id = $1 AND deleted_at IS NULL
        RETURNING id, name, industry, region, description, profile_data, created_at, updated_at, deleted_at
    `

	var updated domain.Organization
	err := pgxscan.Get(ctx, engine, &updated, query,
		uuidToPgtype(org.ID),
		org.Name,
		org.Industry,
		org.Region,
		org.Description,
		normalizeProfileData(org.ProfileData),
		org.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Organization{}, domain.ErrNotFound
		}
		logger.Errorf(ctx, "failed to update organization: %v", err)
		return domain.Organization{}, err
	}

	return updated, nil
}

// DeleteOrganization soft deletes an organization
func (r *PGXRepository) DeleteOrganization(ctx context.Context, id domain.ID) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.DeleteOrganization")
	defer span.Finish()

	engine := r.engineFactory.Get(ctx)
	query := `UPDATE organizations SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`

	tag, err := engine.Exec(ctx, query, uuidToPgtype(id))
	if err != nil {
		logger.Errorf(ctx, "failed to delete organization: %v", err)
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// GetOrganizationsByUserID retrieves all organizations where the user is a member
func (r *PGXRepository) GetOrganizationsByUserID(ctx context.Context, userID domain.ID) ([]domain.Organization, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.GetOrganizationsByUserID")
	defer span.Finish()

	engine := r.engineFactory.Get(ctx)
	query := `
        SELECT DISTINCT o.id, o.name, o.industry, o.region, o.description, o.profile_data, o.created_at, o.updated_at, o.deleted_at
        FROM organizations o
        INNER JOIN organization_members om ON o.id = om.organization_id
        WHERE om.user_id = $1
          AND om.status = 'active'
          AND o.deleted_at IS NULL
        ORDER BY o.created_at DESC
    `

	var orgs []domain.Organization
	err := pgxscan.Select(ctx, engine, &orgs, query, uuidToPgtype(userID))
	if err != nil {
		logger.Errorf(ctx, "failed to get organizations by user id: %v", err)
		return nil, err
	}

	return orgs, nil
}
