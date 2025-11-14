package contract

import (
	"context"
	"core-service/internal/domain"
	pb "core-service/pkg/core/api/core"

	"github.com/opentracing/opentracing-go"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Service struct {
	pb.UnimplementedGeneratedContractServiceServer
	contractService ContractService
}

type ContractService interface {
	RegisterGeneratedContract(ctx context.Context, organizationID, templateID domain.ID, name, filledData, s3Key, fileType string) (domain.GeneratedContract, error)
	GetGeneratedContract(ctx context.Context, id domain.ID) (domain.GeneratedContract, error)
	DeleteGeneratedContract(ctx context.Context, id domain.ID) error
	ListGeneratedContractsByOrganization(ctx context.Context, organizationID domain.ID) ([]domain.GeneratedContract, error)
	ListGeneratedContractsByTemplate(ctx context.Context, templateID domain.ID) ([]domain.GeneratedContract, error)
}

func NewService(contractService ContractService) *Service {
	return &Service{
		contractService: contractService,
	}
}

func contractToProto(contract domain.GeneratedContract) *pb.GeneratedContract {
	return &pb.GeneratedContract{
		Id:             contract.ID.String(),
		OrganizationId: contract.OrganizationID.String(),
		TemplateId:     contract.TemplateID.String(),
		Name:           contract.Name,
		FilledData:     contract.FilledData,
		S3Key:          contract.S3Key,
		FileType:       contract.FileType,
		CreatedAt:      timestamppb.New(contract.CreatedAt),
	}
}

func (s *Service) RegisterGeneratedContract(ctx context.Context, req *pb.RegisterContractRequest) (*pb.RegisterContractResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.RegisterGeneratedContract")
	defer span.Finish()

	orgID, err := domain.ParseID(req.OrganizationId)
	if err != nil {
		return nil, domain.ErrInvalidArgument
	}

	templateID, err := domain.ParseID(req.TemplateId)
	if err != nil {
		return nil, domain.ErrInvalidArgument
	}

	contract, err := s.contractService.RegisterGeneratedContract(ctx, orgID, templateID, req.Name, req.FilledData, req.S3Key, req.FileType)
	if err != nil {
		return nil, err
	}

	return &pb.RegisterContractResponse{
		Contract: contractToProto(contract),
	}, nil
}

func (s *Service) GetGeneratedContract(ctx context.Context, req *pb.GetContractRequest) (*pb.GetContractResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.GetGeneratedContract")
	defer span.Finish()

	id, err := domain.ParseID(req.Id)
	if err != nil {
		return nil, domain.ErrInvalidArgument
	}

	contract, err := s.contractService.GetGeneratedContract(ctx, id)
	if err != nil {
		return nil, err
	}

	return &pb.GetContractResponse{
		Contract: contractToProto(contract),
	}, nil
}

func (s *Service) DeleteGeneratedContract(ctx context.Context, req *pb.DeleteContractRequest) (*emptypb.Empty, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.DeleteGeneratedContract")
	defer span.Finish()

	id, err := domain.ParseID(req.Id)
	if err != nil {
		return nil, domain.ErrInvalidArgument
	}

	err = s.contractService.DeleteGeneratedContract(ctx, id)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *Service) ListGeneratedContracts(ctx context.Context, req *pb.ListContractsRequest) (*pb.ListContractsResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.ListGeneratedContracts")
	defer span.Finish()

	orgID, err := domain.ParseID(req.OrganizationId)
	if err != nil {
		return nil, domain.ErrInvalidArgument
	}

	contracts, err := s.contractService.ListGeneratedContractsByOrganization(ctx, orgID)
	if err != nil {
		return nil, err
	}

	pbContracts := make([]*pb.GeneratedContract, 0, len(contracts))
	for _, contract := range contracts {
		pbContracts = append(pbContracts, contractToProto(contract))
	}

	return &pb.ListContractsResponse{
		Contracts: pbContracts,
	}, nil
}
