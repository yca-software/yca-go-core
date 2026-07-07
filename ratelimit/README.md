# 2Chi Go Ratelimit

Redis-backed Echo rate limiting for 2Chi projects, built on [ulule/limiter](https://github.com/ulule/limiter).

```go
import chi_ratelimit "github.com/yca-software/yca-go-core/ratelimit"
```

## Setup

```go
rl := chi_ratelimit.NewRateLimiter(redisClient, observer, logger)
```

When `redisClient` is nil, middleware is a no-op pass-through.

## Middleware

| Method | Key | Use case |
| --- | --- | --- |
| `IPRateLimit(rate)` | Client IP | Anonymous/public endpoints |
| `IPDeviceRateLimit(rate)` | IP + device ID | Login, signup, password reset |
| `PrincipalRateLimit(rate)` | User or API key (falls back to IP) | Authenticated routes |
| `ScopedPrincipalRateLimit(rate, scope)` | Scoped principal key | Sensitive actions (e.g. send email) |

Rates use ulule format (e.g. `"100-M"`, `"5-H"`). Responses include `X-RateLimit-Limit`, `X-RateLimit-Remaining`, and `X-RateLimit-Reset`. Limit breaches return 429 `TooManyRequests`.

## Device ID

`EnsureDeviceID(env)` issues a first-party `2chi_device_id` cookie and stores the ID on the Echo context. Used with `IPDeviceRateLimit`.

| Source | Priority |
| --- | --- |
| Echo context (`deviceId`) | 1 |
| Cookie `2chi_device_id` | 2 |
| Header `X-Device-Id` | 3 |

Values must be UUID v4. Cookie `Secure` is enabled when `env != "local"`.

```go
e.Use(chi_ratelimit.EnsureDeviceID(cfg.Env))
e.POST("/auth/login", handler, rl.IPDeviceRateLimit("10-M"))
```

## Observability

On limit breach, the package records rate-limit metrics via `2chi-go-observer` and logs a warning via `2chi-go-logger` (when provided). Principal identifiers are hashed in metrics.

## Example

```go
rl := chi_ratelimit.NewRateLimiter(redis, obs, log)

e.Use(chi_ratelimit.EnsureDeviceID(cfg.Env))
e.POST("/auth/login", login, rl.IPDeviceRateLimit("10-M"))
e.GET("/api/v1/members", listMembers, rl.PrincipalRateLimit("200-M"))
e.POST("/api/v1/emails", sendEmail, rl.ScopedPrincipalRateLimit("5-H", "email"))
```

## Tests

```bash
go test -race -count=1 ./...
```

Uses miniredis for rate-limit middleware tests; no Docker required.
