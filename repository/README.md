# 2Chi Go Repository

Generic PostgreSQL repository for 2Chi projects. Builds queries with [Squirrel](https://github.com/Masterminds/squirrel), runs them via [sqlx](https://github.com/jmoiron/sqlx), maps SQL errors to `2chi-go-error`, and records optional Prometheus metrics via `2chi-go-observer`.

```go
import chi_repository "github.com/yca-software/yca-go-core/repository"
```

## Create a repository

```go
columns := []string{"id", "name", "email", "created_at"}
repo := chi_repository.NewRepository[User](db, "users", columns, obs.GetQueryMetricsHook())
```

`NewRepository` panics if `db` is nil, `tableName` is empty, or `columns` is empty.

## Repository API

| Method                                                          | Description                            |
| --------------------------------------------------------------- | -------------------------------------- |
| `Count(ctx, condition)`                                         | Count rows                             |
| `Get(ctx, condition, columns)`                                  | Fetch one row (`condition` required)   |
| `Select(ctx, condition, columns, sort)`                         | Fetch many rows                        |
| `PaginatedSelect(ctx, condition, columns, sort, limit, offset)` | Paginated list (`sort` required)       |
| `Create(ctx, data)`                                             | Insert one row                         |
| `CreateMany(ctx, columns, data, ignoreConflict)`                | Bulk insert                            |
| `Update(ctx, condition, data)`                                  | Update rows (`condition` required)     |
| `Delete(ctx, condition)`                                        | Delete rows (`condition` required)     |
| `WithTx(tx)`                                                    | Clone bound to a transaction           |
| `GetQueryBuilder()`                                             | Squirrel builder with `$` placeholders |
| `DB()`                                                          | Underlying `Executor`                  |

## Transactions

```go
err := chi_repository.RunInTx(ctx, db, hook, func(tx chi_repository.Tx) error {
    txRepo := repo.WithTx(tx)
    return txRepo.Create(ctx, data)
})
```

`RunInTx` commits on success and rolls back on error or panic.

## Sort safety

`sort` is validated against repository `columns` only. Invalid columns, directions, or injection attempts are rejected.

## SQL errors

`WrapSQLError` maps:

| Case               | API error                 |
| ------------------ | ------------------------- |
| `sql.ErrNoRows`    | 404 `NotFound`            |
| PostgreSQL `23505` | 409 `Conflict`            |
| PostgreSQL `23503` | 422 `UnprocessableEntity` |

`Delete` and `Update` return 404 when zero rows are affected.

## Testing

`MockRepository[T]` and `MockTx` are testify mocks for unit tests in services.

Integration tests require Docker:

```bash
go test -race -count=1 ./...
```
