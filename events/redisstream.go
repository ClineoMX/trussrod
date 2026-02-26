package events

import (
	"context"
	"fmt"
	"maps"

	"github.com/clineomx/trussrod/cache"
	"github.com/redis/go-redis/v9"
)

type RedisStream struct {
	client     *cache.RedisClient
	streamName string
}

func NewRedisStream(client *cache.RedisClient, streamName string) (*RedisStream, error) {
	if client == nil {
		return nil, fmt.Errorf("redis client cannot be nil")
	}
	if streamName == "" {
		return nil, fmt.Errorf("stream name cannot be empty")
	}

	if err := client.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	return &RedisStream{
		client:     client,
		streamName: streamName,
	}, nil
}

func (rs *RedisStream) Publish(ctx context.Context, message map[string]any) error {
	values := make(map[string]any)
	maps.Copy(values, message)

	_, err := rs.client.XAdd(ctx, &redis.XAddArgs{
		Stream: rs.streamName,
		Values: values,
	})

	if err != nil {
		return fmt.Errorf("failed to publish to redis stream: %w", err)
	}

	return nil
}

func (rs *RedisStream) Close() error {
	return nil
}
