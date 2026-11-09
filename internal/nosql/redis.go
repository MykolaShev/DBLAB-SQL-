package nosql

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisStore struct {
	client *redis.Client
}

func NewRedisStore(ctx context.Context, addr string) (*RedisStore, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &RedisStore{client: rdb}, nil
}

func (r *RedisStore) CacheTopProjects(ctx context.Context, key string, projects interface{}, expiration time.Duration) error {
	data, err := json.Marshal(projects)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, key, data, expiration).Err()
}

func (r *RedisStore) GetCachedProjects(ctx context.Context, key string, dest interface{}) error {
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(val), dest)
}

func (r *RedisStore) Close() error {
	return r.client.Close()
}