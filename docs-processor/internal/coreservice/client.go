package coreservice

import (
	"context"
	"fmt"

	"docs-processor/internal/grpcutils"
	"docs-processor/internal/logger"
	pb "docs-processor/pkg/core"

	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client - клиент для работы с core-service
type Client struct {
	client pb.DocumentServiceClient
	conn   *grpc.ClientConn
}

// NewClient создает новый клиент для core-service
func NewClient(address string) (*Client, error) {
	conn, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(grpcutils.UnaryClientInterceptor),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to core-service: %w", err)
	}

	client := pb.NewDocumentServiceClient(conn)

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

// UpdateDocumentStatus обновляет статус документа
func (c *Client) UpdateDocumentStatus(
	ctx context.Context,
	documentID string,
	status pb.DocumentStatus,
	errorMessage string,
) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "client.coreservice.UpdateDocumentStatus")
	defer span.Finish()

	req := &pb.UpdateDocumentStatusRequest{
		Id:     documentID,
		Status: status,
	}

	if errorMessage != "" {
		req.ErrorMessage = &errorMessage
	}

	logger.Debugf(ctx, "updating document status: doc_id=%s, status=%v", documentID, status)

	_, err := c.client.UpdateDocumentStatus(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to update document status: %w", err)
	}

	logger.Infof(ctx, "document status updated: doc_id=%s, status=%v", documentID, status)
	return nil
}
