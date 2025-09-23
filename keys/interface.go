package keys

import (
	"context"
)

type Manager interface {
	Decrypt(ctx context.Context, target []byte) ([]byte, error)
	CreateDEK(ctx context.Context) ([]byte, []byte, error)
	Sign(ctx context.Context, arn, message string) ([]byte, error)
}
