package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/opentracing/opentracing-go"
)

type S3Client struct {
	client *s3.S3
	bucket string
}

func NewS3Client(endpoint, accessKey, secretKey, region, bucket string, useSSL bool) (*S3Client, error) {
	sess, err := session.NewSession(&aws.Config{
		Endpoint:         aws.String(endpoint),
		Region:           aws.String(region),
		Credentials:      credentials.NewStaticCredentials(accessKey, secretKey, ""),
		S3ForcePathStyle: aws.Bool(true),
		DisableSSL:       aws.Bool(!useSSL),
	})
	if err != nil {
		return nil, err
	}

	return &S3Client{
		client: s3.New(sess),
		bucket: bucket,
	}, nil
}

// GeneratePresignedUploadURL генерирует presigned URL для загрузки файла в S3
func (c *S3Client) GeneratePresignedUploadURL(ctx context.Context, key string, contentType string, expiration time.Duration) (string, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "storage.S3Client.GeneratePresignedUploadURL")
	defer span.Finish()

	req, _ := c.client.PutObjectRequest(&s3.PutObjectInput{
		Bucket:      aws.String(c.bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
	})

	url, err := req.Presign(expiration)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return url, nil
}

// GeneratePresignedDownloadURL генерирует presigned URL для скачивания файла из S3
func (c *S3Client) GeneratePresignedDownloadURL(ctx context.Context, key string, expiration time.Duration) (string, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "storage.S3Client.GeneratePresignedDownloadURL")
	defer span.Finish()

	req, _ := c.client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})

	url, err := req.Presign(expiration)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return url, nil
}
