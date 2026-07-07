package chi_redis

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type RedisClientConfig struct {
	DSN string
}

type Redis struct {
	client *redis.Client
}

func NewRedis(cfg RedisClientConfig) (*Redis, error) {
	opt, err := redis.ParseURL(cfg.DSN)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(opt)

	return &Redis{client: client}, nil
}

func (r *Redis) Check(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

func (r *Redis) Cleanup() {
	r.client.Close()
}

func (r *Redis) GetClient() any { return r.client }
