package storage

import (
	"context"
	"core-service/internal/app/interceptors"
	"core-service/internal/domain"
	"core-service/internal/service/storage"
	pb "core-service/pkg/core"

	"github.com/opentracing/opentracing-go"
)

type Service struct {
	pb.UnimplementedStorageServiceServer
	storageService StorageService
}

type StorageService interface {
	GenerateUploadURL(ctx context.Context, organizationID domain.ID, fileName, contentType string) (string, string, error)
	GenerateDownloadURL(ctx context.Context, s3Key string) (string, error)
}

func NewService(storageService StorageService) *Service {
	return &Service{
		storageService: storageService,
	}
}

func (s *Service) GenerateUploadURL(ctx context.Context, req *pb.GenerateUploadURLRequest) (*pb.GenerateUploadURLResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.storage.GenerateUploadURL")
	defer span.Finish()

	// Проверяем авторизацию
	_, err := interceptors.UserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	organizationID, err := domain.ParseID(req.GetOrganizationId())
	if err != nil {
		return nil, domain.ErrInvalidArgument
	}

	// TODO: проверить, что пользователь является членом организации

	url, s3Key, err := s.storageService.GenerateUploadURL(ctx, organizationID, req.GetFileName(), req.GetContentType())
	if err != nil {
		return nil, err
	}

	return &pb.GenerateUploadURLResponse{
		UploadUrl:        url,
		S3Key:            s3Key,
		ExpiresInSeconds: int64(storage.PresignedURLExpiration.Seconds()),
	}, nil
}

func (s *Service) GenerateDownloadURL(ctx context.Context, req *pb.GenerateDownloadURLRequest) (*pb.GenerateDownloadURLResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.storage.GenerateDownloadURL")
	defer span.Finish()

	// Проверяем авторизацию
	_, err := interceptors.UserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// TODO: проверить права доступа к документу по s3_key

	url, err := s.storageService.GenerateDownloadURL(ctx, req.GetS3Key())
	if err != nil {
		return nil, err
	}

	return &pb.GenerateDownloadURLResponse{
		DownloadUrl:      url,
		ExpiresInSeconds: int64(storage.PresignedURLExpiration.Seconds()),
	}, nil
}
