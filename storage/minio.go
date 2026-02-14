package storage

import (
	"context"
	"crypto/sha256"
	"io"
	"mime/multipart"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Minio struct {
	client *minio.Client
	bucket string
}

func NewMinio(ctx context.Context, endpoint, accessKey, secretKey, bucket string, useSSL bool) (*Minio, error) {
	options := &minio.Options{
		Creds:           credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure:          useSSL,
		TrailingHeaders: true,
	}
	client, err := minio.New(endpoint, options)
	if err != nil {
		return nil, err
	}

	exists, err := client.BucketExists(ctx, bucket)
	if err != nil {
		return nil, err
	}
	if !exists {
		err = client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
		if err != nil {
			return nil, err
		}
	}
	return &Minio{client: client, bucket: bucket}, nil
}

func (m *Minio) Upload(ctx context.Context, key string, file multipart.File, fileSize int64, metadata map[string]string) ([]byte, error) {
	checksum := sha256.New()
	tee := io.TeeReader(file, checksum)
	_, err := m.client.PutObject(ctx, m.bucket, key, tee, fileSize, minio.PutObjectOptions{
		UserMetadata: metadata,
	})
	if err != nil {
		return nil, err
	}

	return checksum.Sum(nil), nil
}

func (m *Minio) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	object, err := m.client.GetObject(ctx, m.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	return object, nil
}

func (m *Minio) Delete(ctx context.Context, key string) error {
	return m.client.RemoveObject(ctx, m.bucket, key, minio.RemoveObjectOptions{})
}
