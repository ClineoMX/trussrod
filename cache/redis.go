package cache

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	conn *redis.Client
}

func NewRedisClient(host, port, password, db string) (*RedisClient, error) {
	uri := fmt.Sprintf("%s:%s", host, port)
	dbInt, err := strconv.Atoi(db)
	if err != nil {
		return nil, err
	}
	client := &RedisClient{
		conn: redis.NewClient(&redis.Options{
			Addr:     uri,
			Password: password,
			DB:       dbInt,
		}),
	}
	if err := client.conn.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}
	return client, nil
}

func (c *RedisClient) Get(ctx context.Context, key string) ([]byte, error) {
	result, err := c.conn.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, ErrCacheMiss
		}
		return nil, fmt.Errorf("%w: %v", ErrConnection, err)
	}
	return result, nil
}

func (c *RedisClient) Set(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	return c.conn.Set(ctx, key, value, expiration).Err()
}

func (c *RedisClient) Close() error {
	return c.conn.Close()
}

func (c *RedisClient) Del(ctx context.Context, key string) error {
	return c.conn.Del(ctx, key).Err()
}

func (c *RedisClient) Ping(ctx context.Context) error {
	return c.conn.Ping(ctx).Err()
}
