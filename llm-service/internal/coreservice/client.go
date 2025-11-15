package coreservice

import (
	"context"
	"fmt"
	"time"

	pb "llm-service/pkg/core"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client interface {
	GetTemplate(ctx context.Context, templateID string) (*Template, error)
	RegisterContract(ctx context.Context, organizationID, templateID, name, filledData, s3Key, fileType string) (*Contract, error)
	ListContracts(ctx context.Context, organizationID string, limit, offset int) ([]*Contract, int, error)
}

type Template struct {
	ID              string
	OrganizationID  string
	Name            string
	Description     string
	TemplateType    string
	FieldsSchema    string
	S3TemplateKey   string
	ContentTemplate string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type Contract struct {
	ID             string
	OrganizationID string
	TemplateID     string
	Name           string
	FilledData     string
	S3Key          string
	FileType       string
	CreatedAt      time.Time
}

type grpcClient struct {
	conn                  *grpc.ClientConn
	templateServiceClient pb.ContractTemplateServiceClient
	contractServiceClient pb.GeneratedContractServiceClient
}

func NewClient(address string) (Client, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to core-service: %w", err)
	}

	return &grpcClient{
		conn:                  conn,
		templateServiceClient: pb.NewContractTemplateServiceClient(conn),
		contractServiceClient: pb.NewGeneratedContractServiceClient(conn),
	}, nil
}

func (c *grpcClient) Close() error {
	return c.conn.Close()
}

func (c *grpcClient) GetTemplate(ctx context.Context, templateID string) (*Template, error) {
	resp, err := c.templateServiceClient.GetTemplate(ctx, &pb.GetTemplateRequest{
		Id: templateID,
	})
	if err != nil {
		return nil, err
	}

	return &Template{
		ID:              resp.Template.Id,
		OrganizationID:  resp.Template.OrganizationId,
		Name:            resp.Template.Name,
		Description:     resp.Template.Description,
		TemplateType:    resp.Template.TemplateType,
		FieldsSchema:    resp.Template.FieldsSchema,
		S3TemplateKey:   resp.Template.S3TemplateKey,
		ContentTemplate: resp.Template.ContentTemplate,
		CreatedAt:       resp.Template.CreatedAt.AsTime(),
		UpdatedAt:       resp.Template.UpdatedAt.AsTime(),
	}, nil
}

func (c *grpcClient) RegisterContract(ctx context.Context, organizationID, templateID, name, filledData, s3Key, fileType string) (*Contract, error) {
	resp, err := c.contractServiceClient.RegisterContract(ctx, &pb.RegisterContractRequest{
		OrganizationId: organizationID,
		TemplateId:     templateID,
		Name:           name,
		FilledData:     filledData,
		S3Key:          s3Key,
		FileType:       fileType,
	})
	if err != nil {
		return nil, err
	}

	return &Contract{
		ID:             resp.Contract.Id,
		OrganizationID: resp.Contract.OrganizationId,
		TemplateID:     resp.Contract.TemplateId,
		Name:           resp.Contract.Name,
		FilledData:     resp.Contract.FilledData,
		S3Key:          resp.Contract.S3Key,
		FileType:       resp.Contract.FileType,
		CreatedAt:      resp.Contract.CreatedAt.AsTime(),
	}, nil
}

func (c *grpcClient) ListContracts(ctx context.Context, organizationID string, limit, offset int) ([]*Contract, int, error) {
	page := max((offset/limit)+1, 1)

	resp, err := c.contractServiceClient.ListContracts(ctx, &pb.ListContractsRequest{
		OrganizationId: organizationID,
		Page:           int32(page),
		PageSize:       int32(limit),
	})
	if err != nil {
		return nil, 0, err
	}

	contracts := make([]*Contract, 0, len(resp.Contracts))
	for _, c := range resp.Contracts {
		contracts = append(contracts, &Contract{
			ID:             c.Id,
			OrganizationID: c.OrganizationId,
			TemplateID:     c.TemplateId,
			Name:           c.Name,
			FilledData:     c.FilledData,
			S3Key:          c.S3Key,
			FileType:       c.FileType,
			CreatedAt:      c.CreatedAt.AsTime(),
		})
	}

	return contracts, int(resp.Total), nil
}
