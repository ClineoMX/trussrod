package keys

import (
	"context"
)

type SignInput struct {
	ARN     string
	Message []byte
}

type VerifyInput struct {
	ARN       string
	Message   []byte
	Signature []byte
}

type Manager interface {
	Decrypt(ctx context.Context, target []byte) ([]byte, error)
	CreateDEK(ctx context.Context) ([]byte, []byte, error)
	Sign(context.Context, *SignInput) ([]byte, error)
	Verify(context.Context, *VerifyInput) (bool, error)
}
