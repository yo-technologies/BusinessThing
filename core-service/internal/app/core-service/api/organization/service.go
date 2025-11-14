package organization

import (
	"context"
	"core-service/internal/domain"
	pb "core-service/pkg/core/api/core"

	"github.com/opentracing/opentracing-go"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Service struct {
	pb.UnimplementedOrganizationServiceServer
	orgService OrganizationService
}

type OrganizationService interface {
	CreateOrganization(ctx context.Context, name, industry, region, description, profileData string) (domain.Organization, error)
	GetOrganization(ctx context.Context, id domain.ID) (domain.Organization, error)
	UpdateOrganization(ctx context.Context, id domain.ID, name, industry, region, description, profileData *string) (domain.Organization, error)
	DeleteOrganization(ctx context.Context, id domain.ID) error
}

func NewService(orgService OrganizationService) *Service {
	return &Service{
		orgService: orgService,
	}
}

func organizationToProto(org domain.Organization) *pb.Organization {
	pbOrg := &pb.Organization{
		Id:          org.ID.String(),
		Name:        org.Name,
		Industry:    org.Industry,
		Region:      org.Region,
		Description: org.Description,
		ProfileData: org.ProfileData,
		CreatedAt:   timestamppb.New(org.CreatedAt),
		UpdatedAt:   timestamppb.New(org.UpdatedAt),
	}
	if org.DeletedAt != nil {
		pbOrg.DeletedAt = timestamppb.New(*org.DeletedAt)
	}
	return pbOrg
}

func (s *Service) CreateOrganization(ctx context.Context, req *pb.CreateOrganizationRequest) (*pb.CreateOrganizationResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.CreateOrganization")
	defer span.Finish()

	org, err := s.orgService.CreateOrganization(ctx, req.Name, req.Industry, req.Region, req.Description, req.ProfileData)
	if err != nil {
		return nil, err
	}

	return &pb.CreateOrganizationResponse{
		Organization: organizationToProto(org),
	}, nil
}

func (s *Service) GetOrganization(ctx context.Context, req *pb.GetOrganizationRequest) (*pb.GetOrganizationResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.GetOrganization")
	defer span.Finish()

	id, err := domain.ParseID(req.Id)
	if err != nil {
		return nil, domain.ErrInvalidArgument
	}

	org, err := s.orgService.GetOrganization(ctx, id)
	if err != nil {
		return nil, err
	}

	return &pb.GetOrganizationResponse{
		Organization: organizationToProto(org),
	}, nil
}

func (s *Service) UpdateOrganization(ctx context.Context, req *pb.UpdateOrganizationRequest) (*pb.UpdateOrganizationResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.UpdateOrganization")
	defer span.Finish()

	id, err := domain.ParseID(req.Id)
	if err != nil {
		return nil, domain.ErrInvalidArgument
	}

	org, err := s.orgService.UpdateOrganization(ctx, id, req.Name, req.Industry, req.Region, req.Description, req.ProfileData)
	if err != nil {
		return nil, err
	}

	return &pb.UpdateOrganizationResponse{
		Organization: organizationToProto(org),
	}, nil
}

func (s *Service) DeleteOrganization(ctx context.Context, req *pb.DeleteOrganizationRequest) (*emptypb.Empty, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.DeleteOrganization")
	defer span.Finish()

	id, err := domain.ParseID(req.Id)
	if err != nil {
		return nil, domain.ErrInvalidArgument
	}

	err = s.orgService.DeleteOrganization(ctx, id)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}
