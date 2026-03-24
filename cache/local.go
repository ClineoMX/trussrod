package cache

import (
	"context"
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
)

type CacheRecord struct {
	Value     []byte
	TTL       time.Duration
	CreatedAt time.Time
}

func (r *CacheRecord) IsExpired() bool {
	return time.Now().After(r.CreatedAt.Add(r.TTL))
}

type LocalCache struct {
	mu     sync.Mutex
	ticker *time.Ticker
	done   chan struct{}
	lru    *lru.Cache[string, *CacheRecord]
}

func NewLocalCache(size int) (*LocalCache, error) {
	core, err := lru.New[string, *CacheRecord](size)
	if err != nil {
		return nil, err
	}

	local := &LocalCache{
		lru:    core,
		ticker: time.NewTicker(5 * time.Second),
		done:   make(chan struct{}),
	}

	go local.cleanupLoop()
	return local, nil
}

func (c *LocalCache) Get(ctx context.Context, key string) ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	cached, ok := c.lru.Get(key)
	if !ok {
		return nil, ErrCacheMiss
	}

	if cached.IsExpired() {
		c.lru.Remove(key)
		return nil, ErrCacheMiss
	}

	return cached.Value, nil
}

func (c *LocalCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	record := &CacheRecord{Value: value, TTL: ttl, CreatedAt: time.Now()}
	c.lru.Add(key, record)
	return nil
}

func (c *LocalCache) Del(ctx context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.lru.Remove(key)
	return nil
}

func (c *LocalCache) Close() error {
	close(c.done)
	return nil
}

func (c *LocalCache) Ping(ctx context.Context) error {
	return nil
}

func (c *LocalCache) cleanupLoop() {
	defer c.ticker.Stop()

	for {
		select {
		case <-c.ticker.C:
			c.removeExpired()
		case <-c.done:
			return
		}
	}
}

func (c *LocalCache) removeExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, key := range c.lru.Keys() {
		val, ok := c.lru.Peek(key)
		if ok && val.IsExpired() {
			c.lru.Remove(key)
		}
	}
}
