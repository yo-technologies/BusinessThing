package storage

import (
	"context"
	"core-service/internal/domain"
	"fmt"
	"path/filepath"
	"time"

	"github.com/opentracing/opentracing-go"
)

const (
	// PresignedURLExpiration время жизни presigned URL (1 час)
	PresignedURLExpiration = 1 * time.Hour
)

type s3Client interface {
	GeneratePresignedUploadURL(ctx context.Context, key string, contentType string, expiration time.Duration) (string, error)
	GeneratePresignedDownloadURL(ctx context.Context, key string, expiration time.Duration) (string, error)
}

type Service struct {
	s3Client s3Client
}

func New(s3Client s3Client) *Service {
	return &Service{
		s3Client: s3Client,
	}
}

// GenerateUploadURL генерирует presigned URL для загрузки документа
func (s *Service) GenerateUploadURL(ctx context.Context, organizationID domain.ID, fileName, contentType string) (string, string, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.storage.GenerateUploadURL")
	defer span.Finish()

	// Валидация
	if fileName == "" {
		return "", "", domain.NewInvalidArgumentError("file name is required")
	}

	// Генерируем безопасное имя файла
	fileExt := filepath.Ext(fileName)
	documentID := domain.NewID()

	// Формируем S3 ключ: documents/{organization_id}/{document_id}{extension}
	s3Key := fmt.Sprintf("documents/%s/%s%s", organizationID.String(), documentID.String(), fileExt)

	// Генерируем presigned URL
	url, err := s.s3Client.GeneratePresignedUploadURL(ctx, s3Key, contentType, PresignedURLExpiration)
	if err != nil {
		return "", "", domain.NewInternalError("failed to generate upload URL", err)
	}

	return url, s3Key, nil
}

// GenerateDownloadURL генерирует presigned URL для скачивания документа
func (s *Service) GenerateDownloadURL(ctx context.Context, s3Key string) (string, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.storage.GenerateDownloadURL")
	defer span.Finish()

	if s3Key == "" {
		return "", domain.NewInvalidArgumentError("s3 key is required")
	}

	url, err := s.s3Client.GeneratePresignedDownloadURL(ctx, s3Key, PresignedURLExpiration)
	if err != nil {
		return "", domain.NewInternalError("failed to generate download URL", err)
	}

	return url, nil
}
