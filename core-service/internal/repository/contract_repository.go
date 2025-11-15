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

// CreateTemplate inserts a new contract template
func (r *PGXRepository) CreateTemplate(ctx context.Context, template domain.ContractTemplate) (domain.ContractTemplate, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.CreateTemplate")
	defer span.Finish()

	engine := r.engineFactory.Get(ctx)
	query := `
        INSERT INTO contract_templates (id, organization_id, name, description, template_type, fields_schema, content_template, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
        RETURNING id, organization_id, name, description, template_type, fields_schema, content_template, created_at, updated_at
    `

	var created domain.ContractTemplate
	err := pgxscan.Get(ctx, engine, &created, query,
		uuidToPgtype(template.ID),
		uuidToPgtype(template.OrganizationID),
		template.Name,
		template.Description,
		template.TemplateType,
		template.FieldsSchema,
		template.ContentTemplate,
		template.CreatedAt,
		template.UpdatedAt,
	)
	if err != nil {
		logger.Errorf(ctx, "failed to create template: %v", err)
		return domain.ContractTemplate{}, err
	}

	return created, nil
}

// GetTemplate retrieves a contract template by ID
func (r *PGXRepository) GetTemplate(ctx context.Context, id domain.ID) (domain.ContractTemplate, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.GetTemplate")
	defer span.Finish()

	engine := r.engineFactory.Get(ctx)
	query := `
        SELECT id, organization_id, name, description, template_type, fields_schema, content_template, created_at, updated_at
        FROM contract_templates
        WHERE id = $1
    `

	var template domain.ContractTemplate
	err := pgxscan.Get(ctx, engine, &template, query, uuidToPgtype(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ContractTemplate{}, domain.ErrNotFound
		}
		logger.Errorf(ctx, "failed to get template: %v", err)
		return domain.ContractTemplate{}, err
	}

	return template, nil
}

// ListTemplates retrieves templates with pagination
func (r *PGXRepository) ListTemplates(ctx context.Context, organizationID domain.ID, limit, offset int) ([]domain.ContractTemplate, int, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.ListTemplates")
	defer span.Finish()

	engine := r.engineFactory.Get(ctx)

	// Get total count
	countQuery := `SELECT COUNT(*) FROM contract_templates WHERE organization_id = $1`
	var total int
	err := pgxscan.Get(ctx, engine, &total, countQuery, uuidToPgtype(organizationID))
	if err != nil {
		logger.Errorf(ctx, "failed to count templates: %v", err)
		return nil, 0, err
	}

	// Get templates
	query := `
        SELECT id, organization_id, name, description, template_type, fields_schema, content_template, created_at, updated_at
        FROM contract_templates
        WHERE organization_id = $1
        ORDER BY created_at DESC
        LIMIT $2 OFFSET $3
    `

	var templates []domain.ContractTemplate
	err = pgxscan.Select(ctx, engine, &templates, query, uuidToPgtype(organizationID), limit, offset)
	if err != nil {
		logger.Errorf(ctx, "failed to list templates: %v", err)
		return nil, 0, err
	}

	return templates, total, nil
}

// UpdateTemplate updates an existing template
func (r *PGXRepository) UpdateTemplate(ctx context.Context, template domain.ContractTemplate) (domain.ContractTemplate, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.UpdateTemplate")
	defer span.Finish()

	engine := r.engineFactory.Get(ctx)
	query := `
        UPDATE contract_templates
        SET name = $2, description = $3, fields_schema = $4, content_template = $5, updated_at = $6
        WHERE id = $1
        RETURNING id, organization_id, name, description, template_type, fields_schema, content_template, created_at, updated_at
    `

	var updated domain.ContractTemplate
	err := pgxscan.Get(ctx, engine, &updated, query,
		uuidToPgtype(template.ID),
		template.Name,
		template.Description,
		template.FieldsSchema,
		template.ContentTemplate,
		template.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ContractTemplate{}, domain.ErrNotFound
		}
		logger.Errorf(ctx, "failed to update template: %v", err)
		return domain.ContractTemplate{}, err
	}

	return updated, nil
}

// DeleteTemplate deletes a template
func (r *PGXRepository) DeleteTemplate(ctx context.Context, id domain.ID) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.DeleteTemplate")
	defer span.Finish()

	engine := r.engineFactory.Get(ctx)
	query := `DELETE FROM contract_templates WHERE id = $1`

	tag, err := engine.Exec(ctx, query, uuidToPgtype(id))
	if err != nil {
		logger.Errorf(ctx, "failed to delete template: %v", err)
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// CreateContract inserts a new generated contract
func (r *PGXRepository) CreateContract(ctx context.Context, contract domain.GeneratedContract) (domain.GeneratedContract, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.CreateContract")
	defer span.Finish()

	engine := r.engineFactory.Get(ctx)
	query := `
        INSERT INTO generated_contracts (id, organization_id, template_id, name, filled_data, s3_key, file_type, created_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        RETURNING id, organization_id, template_id, name, filled_data, s3_key, file_type, created_at
    `

	var created domain.GeneratedContract
	err := pgxscan.Get(ctx, engine, &created, query,
		uuidToPgtype(contract.ID),
		uuidToPgtype(contract.OrganizationID),
		uuidToPgtype(contract.TemplateID),
		contract.Name,
		contract.FilledData,
		contract.S3Key,
		contract.FileType,
		contract.CreatedAt,
	)
	if err != nil {
		logger.Errorf(ctx, "failed to create contract: %v", err)
		return domain.GeneratedContract{}, err
	}

	return created, nil
}

// GetContract retrieves a generated contract by ID
func (r *PGXRepository) GetContract(ctx context.Context, id domain.ID) (domain.GeneratedContract, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.GetContract")
	defer span.Finish()

	engine := r.engineFactory.Get(ctx)
	query := `
        SELECT id, organization_id, template_id, name, filled_data, s3_key, file_type, created_at
        FROM generated_contracts
        WHERE id = $1
    `

	var contract domain.GeneratedContract
	err := pgxscan.Get(ctx, engine, &contract, query, uuidToPgtype(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.GeneratedContract{}, domain.ErrNotFound
		}
		logger.Errorf(ctx, "failed to get contract: %v", err)
		return domain.GeneratedContract{}, err
	}

	return contract, nil
}

// ListContracts retrieves generated contracts with pagination
func (r *PGXRepository) ListContracts(ctx context.Context, organizationID domain.ID, limit, offset int) ([]domain.GeneratedContract, int, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.ListContracts")
	defer span.Finish()

	engine := r.engineFactory.Get(ctx)

	// Get total count
	countQuery := `SELECT COUNT(*) FROM generated_contracts WHERE organization_id = $1`
	var total int
	err := pgxscan.Get(ctx, engine, &total, countQuery, uuidToPgtype(organizationID))
	if err != nil {
		logger.Errorf(ctx, "failed to count contracts: %v", err)
		return nil, 0, err
	}

	// Get contracts
	query := `
        SELECT id, organization_id, template_id, name, filled_data, s3_key, file_type, created_at
        FROM generated_contracts
        WHERE organization_id = $1
        ORDER BY created_at DESC
        LIMIT $2 OFFSET $3
    `

	var contracts []domain.GeneratedContract
	err = pgxscan.Select(ctx, engine, &contracts, query, uuidToPgtype(organizationID), limit, offset)
	if err != nil {
		logger.Errorf(ctx, "failed to list contracts: %v", err)
		return nil, 0, err
	}

	return contracts, total, nil
}

// DeleteContract deletes a generated contract
func (r *PGXRepository) DeleteContract(ctx context.Context, id domain.ID) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.DeleteContract")
	defer span.Finish()

	engine := r.engineFactory.Get(ctx)
	query := `DELETE FROM generated_contracts WHERE id = $1`

	tag, err := engine.Exec(ctx, query, uuidToPgtype(id))
	if err != nil {
		logger.Errorf(ctx, "failed to delete contract: %v", err)
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// ListContractsByTemplate retrieves contracts for a specific template
func (r *PGXRepository) ListContractsByTemplate(ctx context.Context, templateID domain.ID) ([]domain.GeneratedContract, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.ListContractsByTemplate")
	defer span.Finish()

	engine := r.engineFactory.Get(ctx)
	query := `
        SELECT id, organization_id, template_id, name, filled_data, s3_key, file_type, created_at
        FROM generated_contracts
        WHERE template_id = $1
        ORDER BY created_at DESC
    `

	var contracts []domain.GeneratedContract
	err := pgxscan.Select(ctx, engine, &contracts, query, uuidToPgtype(templateID))
	if err != nil {
		logger.Errorf(ctx, "failed to list contracts by template: %v", err)
		return nil, err
	}

	return contracts, nil
}
