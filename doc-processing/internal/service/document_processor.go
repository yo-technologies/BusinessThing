package service

import (
	"context"
	"fmt"

	"doc-processing/internal/chunker"
	"doc-processing/internal/domain"
	"doc-processing/internal/embeddings"
	"doc-processing/internal/logger"
	"doc-processing/internal/parser"
	"doc-processing/internal/storage"
	"doc-processing/internal/vectordb"

	"github.com/opentracing/opentracing-go"
)

type DocumentProcessor struct {
	s3Client       *storage.S3Client
	parserRegistry *parser.Registry
	chunker        *chunker.Chunker
	embeddingsCli  *embeddings.Client
	vectorDB       *vectordb.OpenSearchClient
	batchSize      int
}

func NewDocumentProcessor(
	s3Client *storage.S3Client,
	parserRegistry *parser.Registry,
	chunker *chunker.Chunker,
	embeddingsCli *embeddings.Client,
	vectorDB *vectordb.OpenSearchClient,
	batchSize int,
) *DocumentProcessor {
	return &DocumentProcessor{
		s3Client:       s3Client,
		parserRegistry: parserRegistry,
		chunker:        chunker,
		embeddingsCli:  embeddingsCli,
		vectorDB:       vectorDB,
		batchSize:      batchSize,
	}
}

func (p *DocumentProcessor) ProcessDocument(ctx context.Context, job *domain.ProcessingJob) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.DocumentProcessor.ProcessDocument")
	defer span.Finish()

	logger.Info(ctx, "Starting document processing", "document_id", job.DocumentID)

	reader, err := p.s3Client.GetObject(ctx, job.S3Key)
	if err != nil {
		logger.Error(ctx, "Failed to get document from S3", "error", err)
		return fmt.Errorf("failed to get document from storage: %w", err)
	}
	defer reader.Close()

	text, err := p.parserRegistry.Parse(ctx, job.DocumentType, reader)
	if err != nil {
		logger.Error(ctx, "Failed to parse document", "error", err)
		return fmt.Errorf("failed to parse document: %w", err)
	}

	logger.Info(ctx, "Document parsed", "text_length", len(text))
	logger.Info(ctx, "Document", "text", text[:400])

	chunks, err := p.chunker.ChunkText(ctx, job.DocumentID, text)
	if err != nil {
		logger.Error(ctx, "Failed to chunk document", "error", err)
		return fmt.Errorf("failed to chunk document: %w", err)
	}

	logger.Info(ctx, "Document chunked", "chunk_count", len(chunks))

	if err := p.generateAndIndexEmbeddings(ctx, chunks, job); err != nil {
		return err
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
