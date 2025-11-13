package service

import (
	"context"
	"fmt"

	"doc-processing/internal/domain"
	"doc-processing/internal/embeddings"
	"doc-processing/internal/vectordb"

	"github.com/opentracing/opentracing-go"
)

type SearchService struct {
	embeddingsCli *embeddings.Client
	vectorDB      *vectordb.OpenSearchClient
}

func NewSearchService(
	embeddingsCli *embeddings.Client,
	vectorDB *vectordb.OpenSearchClient,
) *SearchService {
	return &SearchService{
		embeddingsCli: embeddingsCli,
		vectorDB:      vectorDB,
	}
}

func (s *SearchService) SearchChunks(ctx context.Context, organizationID domain.ID, query string, limit int, minScore float32) ([]*domain.SearchResult, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.SearchService.SearchChunks")
	defer span.Finish()

	queryEmbedding, err := s.embeddingsCli.GenerateEmbedding(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	results, err := s.vectorDB.SearchChunks(ctx, organizationID, queryEmbedding, limit, minScore)
	if err != nil {
		return nil, fmt.Errorf("failed to search chunks: %w", err)
	}

	return results, nil
}
