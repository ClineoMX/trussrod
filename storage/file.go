package storage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type File struct {
	ID       string
	Content  []byte
	Hash     []byte
	Bucket   string
	MimeType string
	Name     string
}

// Key returns storage key of the current file, using root path
// as parameter along a uuid generated to keep sensitive original
// filenames encrypted and secure.
func (f *File) Key(path string) string {
	var ext string
	parts := strings.Split(f.Name, ".")
	if len(parts) > 1 {
		ext = parts[1]
	}
	return fmt.Sprintf("%s/%s.%s", path, f.ID, ext)
}

// Size returns file total size in bytes.
func (f *File) Size() uint64 {
	return uint64(len(f.Content))
}

func (f *File) SaveTo(ctx context.Context, s Storage, path string) error {
	uploadCtx, cancel := context.WithTimeout(ctx, time.Minute*5)
	defer cancel()

	if len(f.Content) == 0 {
		return errors.New("file content is empty")
	}

	mimeType := http.DetectContentType(f.Content)
	allowedTypes := map[string]bool{
		"image/jpeg":        true,
		"image/png":         true,
		"image/gif":         true,
		"application/pdf":   true,
		"application/dicom": true,
	}

	if !allowedTypes[mimeType] {
		return fmt.Errorf("invalid file type: %s", mimeType)
	}

	contentReader := bytes.NewReader(f.Content)
	bucket, err := s.Upload(uploadCtx, f.Key(path), contentReader, nil)
	if err != nil {
		return err
	}

	f.Bucket = bucket
	f.MimeType = mimeType
	return nil
}
