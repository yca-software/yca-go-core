# 2Chi Go Observer

Prometheus metrics for 2Chi projects: HTTP, database queries, rate limits, and background jobs.

```go
import yca_observer "github.com/yca-software/yca-go-core/observer"
```

## Setup

```go
obs, err := yca_observer.New(yca_observer.ObserverConfig{
    Namespace: "myapp",
    AppName:   "api",
})
if err != nil {
    return err
}
```

All metrics include a constant `app` label from `AppName`.

## HTTP

| Symbol                                                                | Description                       |
| --------------------------------------------------------------------- | --------------------------------- |
| `EchoMiddleware(skipper func(echo.Context) bool) echo.MiddlewareFunc` | Record request count and duration |

Default skipper ignores `/metrics`. Labels: `method`, `route`, `status`.

```go
e.Use(obs.EchoMiddleware(nil))
```

## Database

| Symbol                                                                            | Description                  |
| --------------------------------------------------------------------------------- | ---------------------------- |
| `RecordQuery(operation, table, queryType, status string, duration time.Duration)` | Observe query duration       |
| `GetQueryMetricsHook() QueryMetricsHook`                                          | Hook for repository wrappers |

Labels: `operation`, `table`, `query_type`, `status`.

## Rate limits

| Symbol                                                                   | Description                        |
| ------------------------------------------------------------------------ | ---------------------------------- |
| `RecordRateLimitHit(method, route, principalType, principal, ip string)` | Increment rate-limit hit counter   |
| `HashRateLimitIdentifier(value string) string`                           | Stable HMAC hash for metric labels |

Principal types: `ip`, `user`, `api_key`, `unknown`. Raw IDs and IPs are hashed before export.

## Jobs

Metrics are recorded **per job name** (not a single global total).

| Symbol                                                          | Description                          |
| --------------------------------------------------------------- | ------------------------------------ |
| `RecordJobPublished(job string)`                                | Job published to queue               |
| `RecordJobConsumerOutcome(job, outcome string)`                 | Consumer outcome                     |
| `RecordJobConsumerDuration(job string, duration time.Duration)` | Handler duration                     |
| `GetJobMetricsHook() JobMetricsHook`                            | Hook for job clients                 |
| `NoopJobMetricsHook`                                            | Safe no-op when metrics are disabled |

Outcomes: `success`, `dead_letter`, `retry_republished`.

## Metrics

| Metric                                      | Labels                                                           |
| ------------------------------------------- | ---------------------------------------------------------------- |
| `{namespace}_http_requests_total`           | `method`, `route`, `status`                                      |
| `{namespace}_http_request_duration_seconds` | `method`, `route`, `status`                                      |
| `{namespace}_db_query_duration_seconds`     | `operation`, `table`, `query_type`, `status`                     |
| `{namespace}_rate_limit_hits_total`         | `method`, `route`, `principal_type`, `principal_hash`, `ip_hash` |
| `{namespace}_job_consumer_published_total`  | `job`                                                            |
| `{namespace}_job_consumer_outcome_total`    | `job`, `outcome`                                                 |
| `{namespace}_job_consumer_duration_seconds` | `job`                                                            |
