# 2Chi Go Server

Echo HTTP server bootstrap for 2Chi APIs: middleware, error handling, health checks, metrics, and graceful shutdown helpers.

```go
import chi_server "github.com/yca-software/yca-go-core/server"
```

## API server

```go
srv := chi_server.New(chi_server.ServerConfig{
    Port:             cfg.Port,
    CORSAllowOrigins: cfg.CORSAllowOrigins,
    BodyLimit:        "32M",
    ServerReadTimeout:  30,
    ServerWriteTimeout: 30,
    ServerIdleTimeout:  120,
    Logger:           logger,
    Observer:         observer,
    RegisterRoutes:   routes.Register,
})

go srv.Start()
defer srv.Shutdown(ctx)
```

`New` wires:

- `chi_error` HTTP error handler (typed API errors + generic Echo errors)
- Optional `2chi-go-observer` HTTP metrics middleware
- Recover, request ID, structured request logging
- CORS, body limit, security headers

## Health and metrics

| Symbol | Description |
| --- | --- |
| `RegisterHealthHandlers(e, deps)` | `/health` (liveness) and `/ready` (readiness) |
| `RegisterMetricsHandlers(e)` | `/metrics` Prometheus scrape endpoint |
| `NewObservabilityServer(cfg)` | Standalone server with health + metrics only |
| `StartObservabilityServer(ctx, cfg)` | Run observability server until context cancel |
| `StartDedicatedMetricsServer(ctx, port, deps)` | Background metrics server on `metricsPort` |

Readiness checks run concurrently with a 5s timeout. Failed checks return 503 without exposing dependency details.

## Context helpers

| Symbol | Description |
| --- | --- |
| `GetAccessInfo(c)` | Read `*chi_types.AccessInfo` from Echo context (`accessInfo` key) |

Auth middleware in the app sets `accessInfo`; handlers use `GetAccessInfo` for caller identity.

## Shutdown

| Symbol | Description |
| --- | --- |
| `CleanupDependency` | Resource with `Cleanup()` (DB pool, Redis, etc.) |
| `CleanupDependenciesConcurrently(deps)` | Parallel cleanup on shutdown |

## Example

```go
chi_server.RegisterHealthHandlers(e, []chi_server.ReadinessDependency{postgres, redis})
chi_server.RegisterMetricsHandlers(e)

chi_server.StartDedicatedMetricsServer(ctx, cfg.MetricsPort, []chi_server.ReadinessDependency{postgres})

chi_server.CleanupDependenciesConcurrently([]chi_server.CleanupDependency{postgres, redis})
```
