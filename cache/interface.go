package cache

import (
	"context"
	"errors"
	"time"
)

var (
	ErrCacheMiss  = errors.New("cache: miss")
	ErrConnection = errors.New("cache: connection error")
)

type Client interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, expiration time.Duration) error
	Ping(ctx context.Context) error
	Del(ctx context.Context, key string) error
	XAdd(ctx context.Context, args any) (any, error)
	Close() error
}
