# 2Chi Go Test

Postgres integration test helpers for 2Chi Go modules and services. Spins up a shared PostGIS container (or uses an external DSN), runs goose migrations, and exposes `*sql.DB` / `*sqlx.DB` handles for repository tests.

```go
import yca_test "github.com/yca-software/yca-go-core/test"
```

## Quick start

```go
//go:build integration

package myrepo_test

import (
    "os"
    "path/filepath"
    "runtime"
    "testing"

    yca_test "github.com/yca-software/yca-go-core/test"
)

func migrationsDir() string {
    _, file, _, _ := runtime.Caller(0)
    return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", "migrations"))
}

func TestMain(m *testing.M) {
    code := m.Run()
    yca_test.Cleanup()
    os.Exit(code)
}

func (s *Suite) SetupSuite() {
    testDB, err := yca_test.Get(migrationsDir())
    s.Require().NoError(err)
    // ...
}
```

Run integration tests from the module root (Docker required unless `GO_API_INTEGRATION_TEST_DSN` is set):

```bash
go test -tags=integration -race -count=1 -p 1 ./...
```

Use `-p 1` so parallel packages share one reused container safely.

## API

| Symbol                                   | Description                                                                                                             |
| ---------------------------------------- | ----------------------------------------------------------------------------------------------------------------------- |
| `Get(migrationsDir string) (*DB, error)` | Returns a process-wide migrated database. `migrationsDir` must point at goose SQL migrations for the module under test. |
| `(*DB) SQLDB() *sql.DB`                  | Underlying `*sql.DB` with migrations applied.                                                                           |
| `(*DB) SQLx() (*sqlx.DB, error)`         | New `sqlx` handle on the same DSN (preferred for repositories).                                                         |
| `(*DB) ConnectionString() string`        | Postgres DSN.                                                                                                           |
| `Cleanup()`                              | Closes connections and terminates the container when this process owns it. Call from `TestMain`.                        |

## Environment

| Variable                                     | Description                                                                                             |
| -------------------------------------------- | ------------------------------------------------------------------------------------------------------- |
| `GO_API_INTEGRATION_TEST_DSN`                | Skip testcontainers and use this Postgres DSN (migrations still run).                                   |
| `GO_API_INTEGRATION_TEST_NO_CONTAINER_REUSE` | Set to `1` to start a fresh container per process instead of reusing `2chi-go-api-integration-postgis`. |

## Container defaults

| Constant                   | Value                                                                 |
| -------------------------- | --------------------------------------------------------------------- |
| `PostgresImage`            | `postgis/postgis:16-3.4` (PostGIS extensions used by core migrations) |
| `ContainerStartTimeout`    | 5 minutes                                                             |
| `integrationContainerName` | `2chi-go-api-integration-postgis`                                     |

## Tests

```bash
go test -race -count=1 ./...
```
