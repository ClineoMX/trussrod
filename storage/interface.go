// Package storage holds all the logic to save
// static files in any file server provider.
package storage

import (
	"context"
	"io"
	"mime/multipart"
)

type Storage interface {
	Upload(ctx context.Context, key string, file multipart.File, fileSize int64, metadata map[string]string) ([]byte, error)
	Get(ctx context.Context, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, key string) error
}
