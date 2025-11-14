package document

import (
	"context"
	"core-service/internal/domain"
	"encoding/json"
	"strings"

	"github.com/opentracing/opentracing-go"
	amqp "github.com/rabbitmq/amqp091-go"
)

type repository interface {
	CreateDocument(ctx context.Context, doc domain.Document) (domain.Document, error)
	GetDocument(ctx context.Context, id domain.ID) (domain.Document, error)
	ListDocuments(ctx context.Context, organizationID domain.ID, status *domain.DocumentStatus, limit, offset int) ([]domain.Document, int, error)
	UpdateDocumentStatus(ctx context.Context, id domain.ID, status domain.DocumentStatus, errorMessage string) error
	DeleteDocument(ctx context.Context, id domain.ID) error
}

type Service struct {
	repo      repository
	queue     *amqp.Channel
	queueName string
}

func New(repo repository, queue *amqp.Channel, queueName string) *Service {
	return &Service{
		repo:      repo,
		queue:     queue,
		queueName: queueName,
	}
}

// DocumentProcessingJob represents a job for document processing
type DocumentProcessingJob struct {
	DocumentID     string `json:"document_id"`
	OrganizationID string `json:"organization_id"`
	S3Key          string `json:"s3_key"`
	FileType       string `json:"file_type"`
}

// RegisterDocument registers a new document and publishes processing job
func (s *Service) RegisterDocument(ctx context.Context, organizationID domain.ID, name, s3Key, fileType string, fileSize int64) (domain.Document, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.document.RegisterDocument")
	defer span.Finish()

	name = strings.TrimSpace(name)
	if name == "" {
		return domain.Document{}, domain.NewInvalidArgumentError("document name is required")
	}
	if s3Key == "" {
		return domain.Document{}, domain.NewInvalidArgumentError("s3 key is required")
	}

	doc := domain.NewDocument(organizationID, name, s3Key, fileType, fileSize)

	created, err := s.repo.CreateDocument(ctx, doc)
	if err != nil {
		return domain.Document{}, err
	}

	// Publish processing job to queue
	if s.queue != nil {
		err = s.publishProcessingJob(ctx, created)
		if err != nil {
			// Log error but don't fail the request
			// The document is created, processing can be retried
			span.SetTag("queue_error", err.Error())
		}
	}

	return created, nil
}

// GetDocument retrieves a document by ID
func (s *Service) GetDocument(ctx context.Context, id domain.ID) (domain.Document, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.document.GetDocument")
	defer span.Finish()

	return s.repo.GetDocument(ctx, id)
}

// ListDocuments retrieves documents for an organization
func (s *Service) ListDocuments(ctx context.Context, organizationID domain.ID, status *domain.DocumentStatus, page, pageSize int) ([]domain.Document, int, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.document.ListDocuments")
	defer span.Finish()

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	return s.repo.ListDocuments(ctx, organizationID, status, pageSize, offset)
}

// ListDocumentsByOrganization retrieves all documents for an organization (without pagination)
func (s *Service) ListDocumentsByOrganization(ctx context.Context, organizationID domain.ID, status *domain.DocumentStatus) ([]domain.Document, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.document.ListDocumentsByOrganization")
	defer span.Finish()

	docs, _, err := s.repo.ListDocuments(ctx, organizationID, status, 1000, 0)
	return docs, err
}

// UpdateDocumentStatus updates document processing status
func (s *Service) UpdateDocumentStatus(ctx context.Context, id domain.ID, status domain.DocumentStatus, errorMessage string) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.document.UpdateDocumentStatus")
	defer span.Finish()

	return s.repo.UpdateDocumentStatus(ctx, id, status, errorMessage)
}

// DeleteDocument deletes a document
func (s *Service) DeleteDocument(ctx context.Context, id domain.ID) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.document.DeleteDocument")
	defer span.Finish()

	return s.repo.DeleteDocument(ctx, id)
}

// publishProcessingJob publishes a document processing job to RabbitMQ
func (s *Service) publishProcessingJob(ctx context.Context, doc domain.Document) error {
	job := DocumentProcessingJob{
		DocumentID:     doc.ID.String(),
		OrganizationID: doc.OrganizationID.String(),
		S3Key:          doc.S3Key,
		FileType:       doc.FileType,
	}

	body, err := json.Marshal(job)
	if err != nil {
		return err
	}

	return s.queue.Publish(
		"",          // exchange
		s.queueName, // routing key
		false,       // mandatory
		false,       // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}
