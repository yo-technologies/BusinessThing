package rag

import (
	"context"
	"fmt"

	"llm-service/internal/domain"
	"llm-service/internal/logger"
	desc "llm-service/pkg/document"

	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client - клиент для работы с document processing service
type Client struct {
	client desc.DocumentServiceClient
	conn   *grpc.ClientConn
}

// NewClient создает новый RAG клиент
func NewClient(address string) (*Client, error) {
	conn, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to docs processor: %w", err)
	}

	client := desc.NewDocumentServiceClient(conn)

	return &Client{
		client: client,
		conn:   conn,
	}, nil
}

// Close закрывает соединение с сервисом
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// SearchRelevantChunks ищет релевантные фрагменты документов
func (c *Client) SearchRelevantChunks(
	ctx context.Context,
	organizationID domain.ID,
	query string,
	limit int,
	minScore float32,
) ([]string, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "client.rag.SearchRelevantChunks")
	defer span.Finish()

	if limit <= 0 {
		limit = 5
	}

	if minScore <= 0 {
		minScore = 0.5
	}

	req := &desc.SearchChunksRequest{
		Query:          query,
		OrganizationId: organizationID.String(),
		Limit:          int32(limit),
		MinScore:       minScore,
	}

	logger.Debugf(ctx, "searching chunks: query=%s, org_id=%s, limit=%d, min_score=%.2f",
		query, organizationID, limit, minScore)

	resp, err := c.client.SearchChunks(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to search chunks: %w", err)
	}

	chunks := make([]string, 0, len(resp.GetChunks()))
	for _, chunk := range resp.GetChunks() {
		// Форматируем фрагмент для включения в контекст
		formatted := fmt.Sprintf(
			"[Документ: %s]\n%s",
			chunk.GetDocumentName(),
			chunk.GetContent(),
		)
		chunks = append(chunks, formatted)
	}

	logger.Debugf(ctx, "found %d relevant chunks", len(chunks))

	return chunks, nil
}
