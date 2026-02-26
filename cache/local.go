package cache

import (
	"context"
	"errors"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
)

type LocalCache struct {
	LRU *lru.Cache[string, any]
}

func NewLocalCache(size int) (*LocalCache, error) {
	core, err := lru.New[string, any](size)
	if err != nil {
		return nil, err
	}
	return &LocalCache{LRU: core}, nil
}

func (m *LocalCache) Get(ctx context.Context, key string) (string, error) {
	value, ok := m.LRU.Get(key)
	if !ok {
		return "", nil
	}
	return value.(string), nil
}

func (m *LocalCache) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	m.LRU.Add(key, value)
	return nil
}

func (m *LocalCache) Del(ctx context.Context, key string) error {
	m.LRU.Remove(key)
	return nil
}

func (m *LocalCache) Close() error {
	return nil
}

func (m *LocalCache) Ping(ctx context.Context) error {
	return nil
}

func (m *LocalCache) XAdd(ctx context.Context, args any) (any, error) {
	return nil, errors.New("not implemented")
}
