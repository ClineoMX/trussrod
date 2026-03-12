package storage

import (
	"context"
	"io"
	"mime/multipart"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3 implements the Storage interface using AWS S3.
type S3 struct {
	client *s3.Client
	bucket string
}

// NewS3 returns an S3 storage implementation.
func NewS3(cfg *aws.Config, bucket string) *S3 {
	return &S3{
		client: s3.NewFromConfig(*cfg),
		bucket: bucket,
	}
}

// Upload uploads a file to S3 and returns the object key as []byte.
func (s *S3) Upload(ctx context.Context, key string, file multipart.File, fileSize int64, metadata map[string]string) ([]byte, error) {
	putInput := &s3.PutObjectInput{
		Bucket:        aws.String(s.bucket),
		Key:           aws.String(key),
		Body:          file,
		ContentLength: aws.Int64(fileSize),
	}
	if len(metadata) > 0 {
		putInput.Metadata = metadata
	}

	_, err := s.client.PutObject(ctx, putInput)
	if err != nil {
		return nil, err
	}
	return []byte(key), nil
}

// Get downloads an object from S3 and returns its body as io.ReadCloser.
func (s *S3) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	getInput := &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}
	out, err := s.client.GetObject(ctx, getInput)
	if err != nil {
		return nil, err
	}
	return out.Body, nil
}

// Delete removes an object from S3.
func (s *S3) Delete(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	return err
}
