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

// CreateDocument inserts a new document
func (r *PGXRepository) CreateDocument(ctx context.Context, doc domain.Document) (domain.Document, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.CreateDocument")
	defer span.Finish()

	engine := r.engineFactory.Get(ctx)
	query := `
        INSERT INTO documents (id, organization_id, name, s3_key, file_type, file_size, status, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
        RETURNING id, organization_id, name, s3_key, file_type, file_size, status, error_message, created_at, updated_at
    `

	var created domain.Document
	err := pgxscan.Get(ctx, engine, &created, query,
		uuidToPgtype(doc.ID),
		uuidToPgtype(doc.OrganizationID),
		doc.Name,
		doc.S3Key,
		doc.FileType,
		doc.FileSize,
		doc.Status,
		doc.CreatedAt,
		doc.UpdatedAt,
	)
	if err != nil {
		logger.Errorf(ctx, "failed to create document: %v", err)
		return domain.Document{}, err
	}

	return created, nil
}

// GetDocument retrieves a document by ID
func (r *PGXRepository) GetDocument(ctx context.Context, id domain.ID) (domain.Document, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.GetDocument")
	defer span.Finish()

	engine := r.engineFactory.Get(ctx)
	query := `
        SELECT id, organization_id, name, s3_key, file_type, file_size, status, error_message, created_at, updated_at
        FROM documents
        WHERE id = $1
    `

	var doc domain.Document
	err := pgxscan.Get(ctx, engine, &doc, query, uuidToPgtype(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Document{}, domain.ErrNotFound
		}
		logger.Errorf(ctx, "failed to get document: %v", err)
		return domain.Document{}, err
	}

	return doc, nil
}

// ListDocuments retrieves documents with optional status filter
func (r *PGXRepository) ListDocuments(ctx context.Context, organizationID domain.ID, status *domain.DocumentStatus, limit, offset int) ([]domain.Document, int, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.ListDocuments")
	defer span.Finish()

	engine := r.engineFactory.Get(ctx)

	// Build query with optional status filter
	countQuery := `SELECT COUNT(*) FROM documents WHERE organization_id = $1`
	query := `
        SELECT id, organization_id, name, s3_key, file_type, file_size, status, error_message, created_at, updated_at
        FROM documents
        WHERE organization_id = $1
    `

	args := []interface{}{uuidToPgtype(organizationID)}
	if status != nil {
		countQuery += ` AND status = $2`
		query += ` AND status = $2`
		args = append(args, *status)
	}

	// Get total count
	var total int
	err := pgxscan.Get(ctx, engine, &total, countQuery, args...)
	if err != nil {
		logger.Errorf(ctx, "failed to count documents: %v", err)
		return nil, 0, err
	}

	// Get documents
	query += ` ORDER BY created_at DESC LIMIT $` + string(rune(len(args)+1)) + ` OFFSET $` + string(rune(len(args)+2))
	args = append(args, limit, offset)

	var docs []domain.Document
	err = pgxscan.Select(ctx, engine, &docs, query, args...)
	if err != nil {
		logger.Errorf(ctx, "failed to list documents: %v", err)
		return nil, 0, err
	}

	return docs, total, nil
}

// UpdateDocumentStatus updates document status
func (r *PGXRepository) UpdateDocumentStatus(ctx context.Context, id domain.ID, status domain.DocumentStatus, errorMessage string) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.UpdateDocumentStatus")
	defer span.Finish()

	engine := r.engineFactory.Get(ctx)
	query := `UPDATE documents SET status = $2, error_message = $3, updated_at = NOW() WHERE id = $1`

	tag, err := engine.Exec(ctx, query, uuidToPgtype(id), status, errorMessage)
	if err != nil {
		logger.Errorf(ctx, "failed to update document status: %v", err)
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// DeleteDocument deletes a document
func (r *PGXRepository) DeleteDocument(ctx context.Context, id domain.ID) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.DeleteDocument")
	defer span.Finish()

	engine := r.engineFactory.Get(ctx)
	query := `DELETE FROM documents WHERE id = $1`

	tag, err := engine.Exec(ctx, query, uuidToPgtype(id))
	if err != nil {
		logger.Errorf(ctx, "failed to delete document: %v", err)
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// CreateNote inserts a new note
func (r *PGXRepository) CreateNote(ctx context.Context, note domain.Note) (domain.Note, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.CreateNote")
	defer span.Finish()

	engine := r.engineFactory.Get(ctx)
	query := `
        INSERT INTO notes (id, organization_id, content, created_at)
        VALUES ($1, $2, $3, $4)
        RETURNING id, organization_id, content, created_at
    `

	var created domain.Note
	err := pgxscan.Get(ctx, engine, &created, query,
		uuidToPgtype(note.ID),
		uuidToPgtype(note.OrganizationID),
		note.Content,
		note.CreatedAt,
	)
	if err != nil {
		logger.Errorf(ctx, "failed to create note: %v", err)
		return domain.Note{}, err
	}

	return created, nil
}

// GetNote retrieves a note by ID
func (r *PGXRepository) GetNote(ctx context.Context, id domain.ID) (domain.Note, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.GetNote")
	defer span.Finish()

	engine := r.engineFactory.Get(ctx)
	query := `SELECT id, organization_id, content, created_at FROM notes WHERE id = $1`

	var note domain.Note
	err := pgxscan.Get(ctx, engine, &note, query, uuidToPgtype(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Note{}, domain.ErrNotFound
		}
		logger.Errorf(ctx, "failed to get note: %v", err)
		return domain.Note{}, err
	}

	return note, nil
}

// ListNotes retrieves notes for an organization
func (r *PGXRepository) ListNotes(ctx context.Context, organizationID domain.ID, limit int) ([]domain.Note, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.ListNotes")
	defer span.Finish()

	engine := r.engineFactory.Get(ctx)
	query := `
        SELECT id, organization_id, content, created_at
        FROM notes
        WHERE organization_id = $1
        ORDER BY created_at DESC
        LIMIT $2
    `

	var notes []domain.Note
	err := pgxscan.Select(ctx, engine, &notes, query, uuidToPgtype(organizationID), limit)
	if err != nil {
		logger.Errorf(ctx, "failed to list notes: %v", err)
		return nil, err
	}

	return notes, nil
}

// DeleteNote deletes a note
func (r *PGXRepository) DeleteNote(ctx context.Context, id domain.ID) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.DeleteNote")
	defer span.Finish()

	engine := r.engineFactory.Get(ctx)
	query := `DELETE FROM notes WHERE id = $1`

	tag, err := engine.Exec(ctx, query, uuidToPgtype(id))
	if err != nil {
		logger.Errorf(ctx, "failed to delete note: %v", err)
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}
