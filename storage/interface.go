// Package storage holds all the logic to save
// static files in any file server provider.
package storage

import (
	"context"
	"io"
)

type Storage interface {
	Upload(context.Context, string, io.Reader, *UploaderOptions) (string, error)
	GetURL(context.Context, string) (string, error)
}
