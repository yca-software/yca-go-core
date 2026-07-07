# 2Chi Go Redis

Redis client wrapper for 2Chi projects, built on [go-redis](https://github.com/redis/go-redis).

```go
import chi_redis "github.com/yca-software/yca-go-core/redis"
```

## Configuration

| Field | Description                                            |
| ----- | ------------------------------------------------------ |
| `DSN` | Redis connection URL (e.g. `redis://localhost:6379/0`) |

## API

| Symbol                                            | Description                                          |
| ------------------------------------------------- | ---------------------------------------------------- |
| `NewRedis(cfg RedisClientConfig) (*Redis, error)` | Parse the DSN and create a client                    |
| `Check(ctx context.Context) error`                | Ping Redis for health checks                         |
| `Cleanup()`                                       | Close the underlying client                          |
| `GetClient() any`                                 | Return the underlying `*redis.Client` for direct use |

## Example

```go
r, err := chi_redis.NewRedis(chi_redis.RedisClientConfig{
    DSN: cfg.RedisDSN,
})
if err != nil {
    return err
}
defer r.Cleanup()

if err := r.Check(ctx); err != nil {
    return err
}

client := r.GetClient().(*redis.Client)
```
