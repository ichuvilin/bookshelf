package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type Cache struct {
	client *redis.Client
}

func New(redisURL string) (*Cache, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(opts)

	if err := client.Ping(context.Background()).Err(); err != nil {
		_ = client.Close()
		return nil, err
	}

	return &Cache{client: client}, nil
}

func (c *Cache) Get(ctx context.Context, key string) ([]byte, bool) {
	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		return nil, false
	}

	return data, true
}

func (c *Cache) Set(ctx context.Context, key, value string, ttl time.Duration) {
	c.client.Set(ctx, key, value, ttl)
}

func (c *Cache) Delete(ctx context.Context, key string) {
	c.client.Del(ctx, key)
}

func (c *Cache) GetClient() *redis.Client {
	return c.client
}

func (c *Cache) Close() error {
	return c.client.Close()
}
