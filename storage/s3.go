package storage

import (
	"context"
	"io"
	"time"

	"github.com/Domedik/trussrod/identity"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3 struct {
	client    *s3.Client
	bucket    string
	presigner *s3.PresignClient
}

type UploaderOptions struct {
	Size        int64
	Concurrency int
}

func (s *S3) Upload(ctx context.Context, key string, object io.Reader, options *UploaderOptions) (string, error) {
	var uploader *manager.Uploader
	if options != nil {
		uploader = manager.NewUploader(s.client, func(u *manager.Uploader) {
			u.PartSize = options.Size
			u.Concurrency = options.Concurrency
		})
	} else {
		uploader = manager.NewUploader(s.client)
	}
	_, err := uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		Body:   object,
	})
	if err != nil {
		return "", err
	}

	return s.bucket, nil
}

func (s *S3) GetURL(ctx context.Context, key string, ttl time.Duration) (string, error) {
	request, err := s.presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = ttl
	})
	if err != nil {
		return "", err
	}

	return request.URL, nil
}

func NewS3(ctx context.Context, bucket, region string, grants *identity.Credentials) (*S3, error) {
	cfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			grants.AccessKey,
			grants.SecretKey,
			grants.SessionToken,
		)),
	)
	if err != nil {
		return nil, err
	}
	client := s3.NewFromConfig(cfg)
	return &S3{
		client:    client,
		bucket:    bucket,
		presigner: s3.NewPresignClient(client),
	}, nil
}
