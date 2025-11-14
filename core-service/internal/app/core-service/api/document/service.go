package document

import (
	"context"
	"core-service/internal/domain"
	pb "core-service/pkg/core/api/core"

	"github.com/opentracing/opentracing-go"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Service struct {
	pb.UnimplementedDocumentServiceServer
	docService DocumentService
}

type DocumentService interface {
	RegisterDocument(ctx context.Context, organizationID domain.ID, name, s3Key, fileType string, fileSize int64) (domain.Document, error)
	GetDocument(ctx context.Context, id domain.ID) (domain.Document, error)
	ListDocumentsByOrganization(ctx context.Context, organizationID domain.ID, status *domain.DocumentStatus) ([]domain.Document, error)
	UpdateDocumentStatus(ctx context.Context, id domain.ID, status domain.DocumentStatus, errorMessage string) error
	DeleteDocument(ctx context.Context, id domain.ID) error
}

func NewService(docService DocumentService) *Service {
	return &Service{
		docService: docService,
	}
}

func documentToProto(doc domain.Document) *pb.Document {
	return &pb.Document{
		Id:             doc.ID.String(),
		OrganizationId: doc.OrganizationID.String(),
		Name:           doc.Name,
		S3Key:          doc.S3Key,
		FileType:       doc.FileType,
		FileSize:       doc.FileSize,
		Status:         documentStatusToProto(doc.Status),
		ErrorMessage:   doc.ErrorMessage,
		CreatedAt:      timestamppb.New(doc.CreatedAt),
		UpdatedAt:      timestamppb.New(doc.UpdatedAt),
	}
}

func documentStatusToProto(status domain.DocumentStatus) pb.DocumentStatus {
	switch status {
	case domain.DocumentStatusPending:
		return pb.DocumentStatus_DOCUMENT_STATUS_PENDING
	case domain.DocumentStatusProcessing:
		return pb.DocumentStatus_DOCUMENT_STATUS_PROCESSING
	case domain.DocumentStatusIndexed:
		return pb.DocumentStatus_DOCUMENT_STATUS_INDEXED
	case domain.DocumentStatusFailed:
		return pb.DocumentStatus_DOCUMENT_STATUS_FAILED
	default:
		return pb.DocumentStatus_DOCUMENT_STATUS_UNSPECIFIED
	}
}

func documentStatusFromProto(status pb.DocumentStatus) domain.DocumentStatus {
	switch status {
	case pb.DocumentStatus_DOCUMENT_STATUS_PENDING:
		return domain.DocumentStatusPending
	case pb.DocumentStatus_DOCUMENT_STATUS_PROCESSING:
		return domain.DocumentStatusProcessing
	case pb.DocumentStatus_DOCUMENT_STATUS_INDEXED:
		return domain.DocumentStatusIndexed
	case pb.DocumentStatus_DOCUMENT_STATUS_FAILED:
		return domain.DocumentStatusFailed
	default:
		return domain.DocumentStatusPending
	}
}

func (s *Service) RegisterDocument(ctx context.Context, req *pb.RegisterDocumentRequest) (*pb.RegisterDocumentResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.RegisterDocument")
	defer span.Finish()

	orgID, err := domain.ParseID(req.OrganizationId)
	if err != nil {
		return nil, domain.ErrInvalidArgument
	}

	doc, err := s.docService.RegisterDocument(ctx, orgID, req.Name, req.S3Key, req.FileType, req.FileSize)
	if err != nil {
		return nil, err
	}

	return &pb.RegisterDocumentResponse{
		Document: documentToProto(doc),
	}, nil
}

func (s *Service) GetDocument(ctx context.Context, req *pb.GetDocumentRequest) (*pb.GetDocumentResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.GetDocument")
	defer span.Finish()

	id, err := domain.ParseID(req.Id)
	if err != nil {
		return nil, domain.ErrInvalidArgument
	}

	doc, err := s.docService.GetDocument(ctx, id)
	if err != nil {
		return nil, err
	}

	return &pb.GetDocumentResponse{
		Document: documentToProto(doc),
	}, nil
}

func (s *Service) ListDocuments(ctx context.Context, req *pb.ListDocumentsRequest) (*pb.ListDocumentsResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.ListDocuments")
	defer span.Finish()

	orgID, err := domain.ParseID(req.OrganizationId)
	if err != nil {
		return nil, domain.ErrInvalidArgument
	}

	var statusFilter *domain.DocumentStatus
	if req.Status != nil && *req.Status != pb.DocumentStatus_DOCUMENT_STATUS_UNSPECIFIED {
		status := documentStatusFromProto(*req.Status)
		statusFilter = &status
	}

	docs, err := s.docService.ListDocumentsByOrganization(ctx, orgID, statusFilter)
	if err != nil {
		return nil, err
	}

	pbDocs := make([]*pb.Document, 0, len(docs))
	for _, doc := range docs {
		pbDocs = append(pbDocs, documentToProto(doc))
	}

	return &pb.ListDocumentsResponse{
		Documents: pbDocs,
	}, nil
}

func (s *Service) UpdateDocumentStatus(ctx context.Context, req *pb.UpdateDocumentStatusRequest) (*emptypb.Empty, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.UpdateDocumentStatus")
	defer span.Finish()

	id, err := domain.ParseID(req.Id)
	if err != nil {
		return nil, domain.ErrInvalidArgument
	}

	status := documentStatusFromProto(req.Status)

	errorMsg := ""
	if req.ErrorMessage != nil {
		errorMsg = *req.ErrorMessage
	}

	err = s.docService.UpdateDocumentStatus(ctx, id, status, errorMsg)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *Service) DeleteDocument(ctx context.Context, req *pb.DeleteDocumentRequest) (*emptypb.Empty, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.DeleteDocument")
	defer span.Finish()

	id, err := domain.ParseID(req.Id)
	if err != nil {
		return nil, domain.ErrInvalidArgument
	}

	err = s.docService.DeleteDocument(ctx, id)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}
