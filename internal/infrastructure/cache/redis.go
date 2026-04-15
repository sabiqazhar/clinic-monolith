package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisCache is the connection type
type RedisCache struct{ client *redis.Client }

// RedisAddr is a tagged type to avoid string ambiguity in Wire
type RedisAddr string

func NewRedisClient(addr RedisAddr) (*RedisCache, error) {
	c := redis.NewClient(&redis.Options{Addr: string(addr)})
	if err := c.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}
	return &RedisCache{client: c}, nil
}

func (r *RedisCache) Get(ctx context.Context, key string) ([]byte, error) {
	return r.client.Get(ctx, key).Bytes()
}

func (r *RedisCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return r.client.Set(ctx, key, value, ttl).Err()
}

func (r *RedisCache) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}
