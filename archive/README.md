# 2Chi Go Archive

Shared archive filter types and helpers for soft-delete aware list/search APIs in 2Chi projects.

```go
import chi_archive "github.com/yca-software/yca-go-core/archive"
```

## Constants

| Symbol | Value | Description |
| --- | --- | --- |
| `ArchiveFilterActive` | `active` | Non-deleted rows only |
| `ArchiveFilterArchived` | `archived` | Soft-deleted rows only |
| `DefaultArchiveFilter` | `active` | Used when no filter is specified |
| `ArchivedDataRetentionPeriod` | 30 days | How long archived rows are kept before hard delete |

## API

| Function | Description |
| --- | --- |
| `NormalizeArchiveFilter(filter ArchiveFilter) (ArchiveFilter, error)` | Validate a typed filter; empty input defaults to active |
| `ParseArchiveFilterQuery(value string) (ArchiveFilter, error)` | Parse an HTTP query value; invalid input returns a 400 `chi_error` |

Use `NormalizeArchiveFilter` in repositories and services where the filter is already typed. Use `ParseArchiveFilterQuery` at HTTP boundaries (for example `?archive=archived`).

## Example

```go
// HTTP handler
filter, err := chi_archive.ParseArchiveFilterQuery(c.QueryParam("archive"))
if err != nil {
    return err
}

// Repository
filter, err := chi_archive.NormalizeArchiveFilter(filter)
if err != nil {
    return nil, err
}

switch filter {
case chi_archive.ArchiveFilterArchived:
    // deleted_at IS NOT NULL
default:
    // deleted_at IS NULL
}
```

## Tests

```bash
go test -race -count=1 ./...
```
