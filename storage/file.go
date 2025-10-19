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
	name     string
}

func (f *File) WithName(name string) *File {
	f.name = name
	return f
}

func (f *File) Name() string {
	return f.name
}

func (f *File) GetKey(path string) string {
	var ext string
	parts := strings.Split(f.name, ".")
	if len(parts) > 1 {
		ext = parts[1]
	}
	return fmt.Sprintf("%s/%s.%s", path, f.ID, ext)
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
	bucket, err := s.Upload(uploadCtx, f.GetKey(path), contentReader, nil)
	if err != nil {
		return err
	}

	f.Bucket = bucket
	f.MimeType = mimeType
	return nil
}
