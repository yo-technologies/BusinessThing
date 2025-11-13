package document

import (
	"context"

	"doc-processing/internal/domain"
	"doc-processing/internal/service"
	desc "doc-processing/pkg/document"

	"github.com/opentracing/opentracing-go"
)

type Service struct {
	desc.UnimplementedDocumentServiceServer
	searchService *service.SearchService
}

func NewService(searchService *service.SearchService) *Service {
	return &Service{
		searchService: searchService,
	}
}

func (s *Service) SearchChunks(ctx context.Context, req *desc.SearchChunksRequest) (*desc.SearchChunksResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.document.Service.SearchChunks")
	defer span.Finish()

	organizationID := domain.ID(req.OrganizationId)
	limit := int(req.Limit)
	if limit == 0 {
		limit = 10
	}
	minScore := req.MinScore
	if minScore == 0 {
		minScore = 0.5
	}

	results, err := s.searchService.SearchChunks(ctx, organizationID, req.Query, limit, minScore)
	if err != nil {
		return nil, err
	}

	chunks := make([]*desc.DocumentChunk, len(results))
	for i, result := range results {
		chunks[i] = &desc.DocumentChunk{
			ChunkId:      result.ChunkID.String(),
			DocumentId:   result.DocumentID.String(),
			DocumentName: result.DocumentName,
			Content:      result.Content,
			Position:     int32(result.Position),
			Score:        result.Score,
			Metadata:     result.Metadata,
		}
	}

	return &desc.SearchChunksResponse{
		Chunks: chunks,
	}, nil
}
