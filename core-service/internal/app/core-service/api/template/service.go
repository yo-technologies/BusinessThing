package template

import (
	"context"
	pb "core-service/bin/core/api/core"
	"core-service/internal/domain"

	"github.com/opentracing/opentracing-go"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Service struct {
	pb.UnimplementedContractTemplateServiceServer
	templateService TemplateService
}

type TemplateService interface {
	CreateTemplate(ctx context.Context, organizationID domain.ID, name, description, templateType, fieldsSchema, contentTemplate string) (domain.ContractTemplate, error)
	GetTemplate(ctx context.Context, id domain.ID) (domain.ContractTemplate, error)
	UpdateTemplate(ctx context.Context, id domain.ID, name, description, fieldsSchema, contentTemplate *string) (domain.ContractTemplate, error)
	DeleteTemplate(ctx context.Context, id domain.ID) error
	ListTemplatesByOrganization(ctx context.Context, organizationID domain.ID) ([]domain.ContractTemplate, error)
}

func NewService(templateService TemplateService) *Service {
	return &Service{
		templateService: templateService,
	}
}

func templateToProto(template domain.ContractTemplate) *pb.ContractTemplate {
	return &pb.ContractTemplate{
		Id:              template.ID.String(),
		OrganizationId:  template.OrganizationID.String(),
		Name:            template.Name,
		Description:     template.Description,
		TemplateType:    template.TemplateType,
		FieldsSchema:    template.FieldsSchema,
		ContentTemplate: template.ContentTemplate,
		CreatedAt:       timestamppb.New(template.CreatedAt),
		UpdatedAt:       timestamppb.New(template.UpdatedAt),
	}
}

func (s *Service) CreateContractTemplate(ctx context.Context, req *pb.CreateTemplateRequest) (*pb.CreateTemplateResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.CreateContractTemplate")
	defer span.Finish()

	orgID, err := domain.ParseID(req.OrganizationId)
	if err != nil {
		return nil, domain.ErrInvalidArgument
	}

	template, err := s.templateService.CreateTemplate(ctx, orgID, req.Name, req.Description, req.TemplateType, req.FieldsSchema, req.ContentTemplate)
	if err != nil {
		return nil, err
	}

	return &pb.CreateTemplateResponse{
		Template: templateToProto(template),
	}, nil
}

func (s *Service) GetContractTemplate(ctx context.Context, req *pb.GetTemplateRequest) (*pb.GetTemplateResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.GetContractTemplate")
	defer span.Finish()

	id, err := domain.ParseID(req.Id)
	if err != nil {
		return nil, domain.ErrInvalidArgument
	}

	template, err := s.templateService.GetTemplate(ctx, id)
	if err != nil {
		return nil, err
	}

	return &pb.GetTemplateResponse{
		Template: templateToProto(template),
	}, nil
}

func (s *Service) UpdateContractTemplate(ctx context.Context, req *pb.UpdateTemplateRequest) (*pb.UpdateTemplateResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.UpdateContractTemplate")
	defer span.Finish()

	id, err := domain.ParseID(req.Id)
	if err != nil {
		return nil, domain.ErrInvalidArgument
	}

	template, err := s.templateService.UpdateTemplate(ctx, id, req.Name, req.Description, req.FieldsSchema, req.ContentTemplate)
	if err != nil {
		return nil, err
	}

	return &pb.UpdateTemplateResponse{
		Template: templateToProto(template),
	}, nil
}

func (s *Service) DeleteContractTemplate(ctx context.Context, req *pb.DeleteTemplateRequest) (*emptypb.Empty, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.DeleteContractTemplate")
	defer span.Finish()

	id, err := domain.ParseID(req.Id)
	if err != nil {
		return nil, domain.ErrInvalidArgument
	}

	err = s.templateService.DeleteTemplate(ctx, id)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *Service) ListContractTemplates(ctx context.Context, req *pb.ListTemplatesRequest) (*pb.ListTemplatesResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.ListContractTemplates")
	defer span.Finish()

	orgID, err := domain.ParseID(req.OrganizationId)
	if err != nil {
		return nil, domain.ErrInvalidArgument
	}

	templates, err := s.templateService.ListTemplatesByOrganization(ctx, orgID)
	if err != nil {
		return nil, err
	}

	pbTemplates := make([]*pb.ContractTemplate, 0, len(templates))
	for _, template := range templates {
		pbTemplates = append(pbTemplates, templateToProto(template))
	}

	return &pb.ListTemplatesResponse{
		Templates: pbTemplates,
	}, nil
}
