package service

import (
	"context"
	"fmt"
	"io"

	"docs-processor/internal/chunker"
	"docs-processor/internal/coreservice"
	"docs-processor/internal/domain"
	"docs-processor/internal/embeddings"
	"docs-processor/internal/logger"
	"docs-processor/internal/parser"
	"docs-processor/internal/storage"
	"docs-processor/internal/vectordb"
	pb "docs-processor/pkg/core"

	"github.com/opentracing/opentracing-go"
)

type DocumentProcessor struct {
	s3Client       *storage.S3Client
	parserRegistry *parser.Registry
	chunker        *chunker.Chunker
	embeddingsCli  *embeddings.Client
	vectorDB       *vectordb.OpenSearchClient
	coreClient     *coreservice.Client
	batchSize      int
}

func NewDocumentProcessor(
	s3Client *storage.S3Client,
	parserRegistry *parser.Registry,
	chunker *chunker.Chunker,
	embeddingsCli *embeddings.Client,
	vectorDB *vectordb.OpenSearchClient,
	coreClient *coreservice.Client,
	batchSize int,
) *DocumentProcessor {
	return &DocumentProcessor{
		s3Client:       s3Client,
		parserRegistry: parserRegistry,
		chunker:        chunker,
		embeddingsCli:  embeddingsCli,
		vectorDB:       vectorDB,
		coreClient:     coreClient,
		batchSize:      batchSize,
	}
}

func (p *DocumentProcessor) ProcessDocument(ctx context.Context, job *domain.ProcessingJob) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.DocumentProcessor.ProcessDocument")
	defer span.Finish()

	var err error
	defer func() {
		if err != nil {
			logger.Error(ctx, "Document processing failed", "document_id", job.DocumentID, "error", err)
			updateErr := p.coreClient.UpdateDocumentStatus(ctx, job.DocumentID.String(), pb.DocumentStatus_DOCUMENT_STATUS_FAILED, err.Error())
			if updateErr != nil {
				logger.Error(ctx, "Failed to update document status to FAILED", "error", updateErr)
			}
		}
	}()

	logger.Info(ctx, "Starting document processing", "document_id", job.DocumentID)

	// Обновляем статус на PROCESSING
	if err := p.coreClient.UpdateDocumentStatus(ctx, job.DocumentID.String(), pb.DocumentStatus_DOCUMENT_STATUS_PROCESSING, ""); err != nil {
		logger.Warn(ctx, "Failed to update document status to PROCESSING", "error", err)
	}

	var reader io.ReadCloser
	reader, err = p.s3Client.GetObject(ctx, job.S3Key)
	if err != nil {
		logger.Error(ctx, "Failed to get document from S3", "error", err)
		return fmt.Errorf("failed to get document from storage: %w", err)
	}
	defer reader.Close()

	var text string
	text, err = p.parserRegistry.Parse(ctx, job.DocumentType, reader)
	if err != nil {
		logger.Error(ctx, "Failed to parse document", "error", err)
		return fmt.Errorf("failed to parse document: %w", err)
	}

	logger.Info(ctx, "Document parsed", "text_length", len(text))

	textPreview := text
	if len(text) > 400 {
		textPreview = text[:400]
	}
	logger.Info(ctx, "Document", "text", textPreview)

	var chunks []*domain.Chunk
	chunks, err = p.chunker.ChunkText(ctx, job.DocumentID, text)
	if err != nil {
		logger.Error(ctx, "Failed to chunk document", "error", err)
		return fmt.Errorf("failed to chunk document: %w", err)
	}

	logger.Info(ctx, "Document chunked", "chunk_count", len(chunks))

	if err := p.generateAndIndexEmbeddings(ctx, chunks, job); err != nil {
		return err
	}

	// Обновляем статус на INDEXED при успешной обработке
	if err := p.coreClient.UpdateDocumentStatus(ctx, job.DocumentID.String(), pb.DocumentStatus_DOCUMENT_STATUS_INDEXED, ""); err != nil {
		logger.Warn(ctx, "Failed to update document status to INDEXED", "error", err)
	}

	logger.Info(ctx, "Document processing completed", "document_id", job.DocumentID)
	return nil
}

func (p *DocumentProcessor) generateAndIndexEmbeddings(ctx context.Context, chunks []*domain.Chunk, job *domain.ProcessingJob) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.DocumentProcessor.generateAndIndexEmbeddings")
	defer span.Finish()

	for i := 0; i < len(chunks); i += p.batchSize {
		end := i + p.batchSize
		if end > len(chunks) {
			end = len(chunks)
		}

		batch := chunks[i:end]
		texts := make([]string, len(batch))
		for j, chunk := range batch {
			texts[j] = chunk.Content
		}

		embeddings, err := p.embeddingsCli.GenerateEmbeddings(ctx, texts)
		if err != nil {
			logger.Error(ctx, "Failed to generate embeddings", "error", err)
			return fmt.Errorf("failed to generate embeddings: %w", err)
		}

		for j, chunk := range batch {
			if j < len(embeddings) {
				chunk.WithEmbedding(embeddings[j])
			}

			if err := p.vectorDB.IndexChunk(ctx, chunk, job.OrganizationID, job.DocumentName); err != nil {
				logger.Error(ctx, "Failed to index chunk", "error", err, "chunk_id", chunk.ID)
				return fmt.Errorf("failed to index chunk: %w", err)
			}
		}

		logger.Info(ctx, "Batch indexed", "batch_start", i, "batch_end", end)
	}

	return nil
}
