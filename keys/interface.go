package keys

import (
	"context"
)

type Manager interface {
	Decrypt(ctx context.Context, target []byte) ([]byte, error)
	CreateDEK(ctx context.Context) ([]byte, []byte, error)
	CreateSigner(key string) Signer
}

type Signer interface {
	Sign(ctx context.Context, input []byte) ([]byte, error)
	Verify(ctx context.Context, message, signature []byte) (bool, error)
}
