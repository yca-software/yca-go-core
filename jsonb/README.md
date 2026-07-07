# 2Chi Go JSONB helpers

Helpers for reading and writing Postgres `jsonb` columns via `database/sql`.

```go
import (
    "database/sql/driver"

    chi_jsonb "github.com/yca-software/yca-go-core/jsonb"
)
```

Define a named type for the column payload and implement `Scan` / `Value`:

```go
type Workspace struct {
    ID          string      `db:"id"`
    Permissions Permissions `db:"permissions"`
}

type Permissions []string

func (p *Permissions) Scan(value any) error {
    return chi_jsonb.JSONBScan(value, p)
}

func (p Permissions) Value() (driver.Value, error) {
    return chi_jsonb.JSONBValue(p)
}
```

`JSONBScan` accepts `[]byte` or `string` payloads and treats `nil` as a no-op.
