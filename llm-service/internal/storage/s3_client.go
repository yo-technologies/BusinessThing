package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"

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

func (c *S3Client) GetObject(ctx context.Context, key string) (io.ReadCloser, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "storage.S3Client.GetObject")
	defer span.Finish()

	result, err := c.client.GetObjectWithContext(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get object from S3: %w", err)
	}

	return result.Body, nil
}

func (c *S3Client) PutObject(ctx context.Context, key string, data []byte, contentType string) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "storage.S3Client.PutObject")
	defer span.Finish()

	_, err := c.client.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(c.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return fmt.Errorf("failed to put object to S3: %w", err)
	}

	return nil
}

func (c *S3Client) DeleteObject(ctx context.Context, key string) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "storage.S3Client.DeleteObject")
	defer span.Finish()

	_, err := c.client.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete object from S3: %w", err)
	}

	return nil
}
