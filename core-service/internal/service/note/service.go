package note

import (
	"context"
	"core-service/internal/domain"
	"strings"

	"github.com/opentracing/opentracing-go"
)

type repository interface {
	CreateNote(ctx context.Context, note domain.Note) (domain.Note, error)
	GetNote(ctx context.Context, id domain.ID) (domain.Note, error)
	ListNotes(ctx context.Context, organizationID domain.ID, limit int) ([]domain.Note, error)
	DeleteNote(ctx context.Context, id domain.ID) error
}

type Service struct {
	repo repository
}

func New(repo repository) *Service {
	return &Service{repo: repo}
}

const MaxNotesLimit = 100

// CreateNote creates a new note
func (s *Service) CreateNote(ctx context.Context, organizationID domain.ID, content string) (domain.Note, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.note.CreateNote")
	defer span.Finish()

	content = strings.TrimSpace(content)
	if content == "" {
		return domain.Note{}, domain.NewInvalidArgumentError("note content is required")
	}

	note := domain.NewNote(organizationID, content)

	created, err := s.repo.CreateNote(ctx, note)
	if err != nil {
		return domain.Note{}, err
	}

	return created, nil
}

// GetNote retrieves a note by ID
func (s *Service) GetNote(ctx context.Context, id domain.ID) (domain.Note, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.note.GetNote")
	defer span.Finish()

	return s.repo.GetNote(ctx, id)
}

// ListNotes retrieves notes for an organization
func (s *Service) ListNotes(ctx context.Context, organizationID domain.ID, limit int) ([]domain.Note, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.note.ListNotes")
	defer span.Finish()

	if limit < 1 || limit > MaxNotesLimit {
		limit = 50
	}

	return s.repo.ListNotes(ctx, organizationID, limit)
}

// DeleteNote deletes a note
func (s *Service) DeleteNote(ctx context.Context, id domain.ID) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.note.DeleteNote")
	defer span.Finish()

	return s.repo.DeleteNote(ctx, id)
}
