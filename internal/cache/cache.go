package cache

import (
	"context"
	"encoding/json"
	"time"

	"feature-flags/internal/flags"
	"github.com/redis/go-redis/v9"
)

type Cache struct {
	client *redis.Client
	ttl    time.Duration
}

func New(url string, ttl time.Duration) *Cache {
	opt, err := redis.ParseURL(url)
	if err != nil {
		panic(err)
	}
	return &Cache{client: redis.NewClient(opt), ttl: ttl}
}

func (c *Cache) Get(ctx context.Context, name string) (*flags.Flag, error) {
	data, err := c.client.Get(ctx, key(name)).Bytes()
	if err != nil {
		return nil, err
	}
	var flag flags.Flag
	if err := json.Unmarshal(data, &flag); err != nil {
		return nil, err
	}
	return &flag, nil
}

func (c *Cache) Set(ctx context.Context, flag *flags.Flag) error {
	data, err := json.Marshal(flag)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, key(flag.Name), data, c.ttl).Err()
}

func (c *Cache) Delete(ctx context.Context, name string) error {
	return c.client.Del(ctx, key(name)).Err()
}
func (c *Cache) Close() error { return c.client.Close() }
func key(name string) string  { return "flag:" + name }
