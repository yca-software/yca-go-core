# 2Chi Go PostgreSQL

PostgreSQL client wrapper for 2Chi projects, built on [sqlx](https://github.com/jmoiron/sqlx) and [pgx](https://github.com/jackc/pgx).

```go
import chi_postgresql "github.com/yca-software/yca-go-core/postgresql"
```

## Configuration

| Field             | Description                                     |
| ----------------- | ----------------------------------------------- |
| `DSN`             | PostgreSQL connection URL                       |
| `MaxOpenConns`    | Maximum open connections in the pool            |
| `MaxIdleConns`    | Maximum idle connections in the pool            |
| `ConnMaxLifetime` | Maximum lifetime of a connection                |
| `ConnMaxIdleTime` | Maximum idle time before a connection is closed |

## API

| Symbol                                                           | Description                                     |
| ---------------------------------------------------------------- | ----------------------------------------------- |
| `NewPostgreSQL(cfg PostgreSQLClientConfig) (*PostgreSQL, error)` | Connect and configure the pool                  |
| `Check(ctx context.Context) error`                               | Ping the database for health checks             |
| `Cleanup()`                                                      | Close the connection pool                       |
| `GetClient() any`                                                | Return the underlying `*sqlx.DB` for direct use |

## Example

```go
pg, err := chi_postgresql.NewPostgreSQL(chi_postgresql.PostgreSQLClientConfig{
    DSN:             cfg.PostgresDSN,
    MaxOpenConns:    10,
    MaxIdleConns:    5,
    ConnMaxLifetime: time.Hour,
    ConnMaxIdleTime: 15 * time.Minute,
})
if err != nil {
    return err
}
defer pg.Cleanup()

if err := pg.Check(ctx); err != nil {
    return err
}

db := pg.GetClient().(*sqlx.DB)
```

## Tests

Requires Docker. Tests spin up PostgreSQL via [testcontainers-go](https://golang.testcontainers.org/).

```bash
go test -race -count=1 ./...
```
