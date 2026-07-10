# yca-go-core

Single Go module for shared 2chi backend primitives. Published as `github.com/yca-software/yca-go-core`.

## Layout

Each top-level directory is a subpackage:

| Package | Import |
| ------- | ------ |
| `error` | `github.com/yca-software/yca-go-core/error` |
| `server` | `github.com/yca-software/yca-go-core/server` |
| `validator` | `github.com/yca-software/yca-go-core/validator` |
| `token` | `github.com/yca-software/yca-go-core/token` |
| `password` | `github.com/yca-software/yca-go-core/password` |
| `logger` | `github.com/yca-software/yca-go-core/logger` |
| `localizer` | `github.com/yca-software/yca-go-core/localizer` |
| `repository` | `github.com/yca-software/yca-go-core/repository` |
| `postgresql` | `github.com/yca-software/yca-go-core/postgresql` |
| `redis` | `github.com/yca-software/yca-go-core/redis` |
| `jsonb` | `github.com/yca-software/yca-go-core/jsonb` |
| `types` | `github.com/yca-software/yca-go-core/types` |
| `observer` | `github.com/yca-software/yca-go-core/observer` |
| `ratelimit` | `github.com/yca-software/yca-go-core/ratelimit` |
| `paddle` | `github.com/yca-software/yca-go-core/paddle` |
| `aws` | `github.com/yca-software/yca-go-core/aws` |
| `google` | `github.com/yca-software/yca-go-core/google` |
| `template` | `github.com/yca-software/yca-go-core/template` |
| `archive` | `github.com/yca-software/yca-go-core/archive` |
| `test` | `github.com/yca-software/yca-go-core/test` |

Prefer `yca_<pkg>` import aliases:

```go
yca_error "github.com/yca-software/yca-go-core/error"
yca_server "github.com/yca-software/yca-go-core/server"
```

## Verify

```bash
go test ./... -count=1
```

Integration tests (`postgresql`, `repository`) require Docker.

## Migration from `2chi-go-*`

Replace separate modules with subpackage imports:

```diff
-yca_error "github.com/yca-software/2chi-go-error"
+yca_error "github.com/yca-software/yca-go-core/error"
```

```diff
-require github.com/yca-software/2chi-go-error v1.0.0
-require github.com/yca-software/2chi-go-server v1.0.1
+require github.com/yca-software/yca-go-core v1.0.0
```

Per-package READMEs under each subdirectory document package-specific usage.
