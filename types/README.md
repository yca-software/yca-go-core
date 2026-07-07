# 2Chi Go Types

Shared Go types for 2Chi projects.

```go
import chi_types "github.com/yca-software/yca-go-core/types"
```

## Models

Embedded base structs for domain models and database rows. Use them to keep `id`, timestamps, and soft-delete fields consistent across modules.

| Type | Description |
| --- | --- |
| `ModelBase` | `ID`, `CreatedAt`, `UpdatedAt` with `json` and `db` tags |
| `ModelBaseWithArchive` | `ModelBase` plus optional `DeletedAt` for soft deletes |

### Example

```go
type User struct {
    chi_types.ModelBaseWithArchive
    Email string `json:"email" db:"email"`
}
```

## Access

Caller identity DTOs for JWT-authenticated requests. AuthZ logic stays in the app; these types describe who is calling.

| Type | Description |
| --- | --- |
| `AccessType` | `user` or `api_key` |
| `AccessInfo` | Request context plus authenticated subject |
| `JWTAccessTokenPermissionData` | Organization-scoped permissions from the JWT |

`AccessInfo` fields:

| Field | Description |
| --- | --- |
| `RequestID`, `IPAddress`, `UserAgent` | Request metadata |
| `Type`, `SubjectID` | Caller type and user or API key ID |
| `Email` | User email; empty for API keys |
| `Roles` | Organization permissions (users and API keys) |
| `IsAdmin` | Admin flag; always `false` for API keys |
| `ImpersonatedBy`, `ImpersonatedByEmail` | Impersonation context (users only) |

Use `organizationId` in JWT payloads — never `workspaceId`.

### Example

```go
info := chi_types.AccessInfo{
    Type:      chi_types.AccessTypeUser,
    SubjectID: userID,
    Email:     "ada@example.com",
    Roles: []chi_types.JWTAccessTokenPermissionData{
        {
            OrganizationID: orgID,
            Permissions:    []string{"members:read", "org:write"},
        },
    },
}
```

## Geo

WGS84 (SRID 4326) types for PostGIS `geography` / `geometry` columns. Both implement `sql.Scanner` and `driver.Valuer`.

| Type | Description |
| --- | --- |
| `Point` | A single coordinate (`Lng`, `Lat`) |
| `Polygon` | A slice of `Point` values forming a closed ring |

**Read:** accepts PostGIS EWKB as a hex string or raw bytes.

**Write:** emits EWKT (`SRID=4326;POINT(lng lat)` or `SRID=4326;POLYGON((...))`).

### Example

```go
p := chi_types.Point{Lng: 2.3522, Lat: 48.8566}

val, err := p.Value() // "SRID=4326;POINT(2.3522 48.8566)"

var scanned chi_types.Point
err = scanned.Scan(hexEWKBFromDB)
```

## Pagination

Generic JSON shape for paginated list API responses.

| Field | Description |
| --- | --- |
| `Items` | Page of results |
| `HasNext` | Whether another page exists |

### Example

```go
type User struct {
    ID   string `json:"id"`
    Name string `json:"name"`
}

resp := chi_types.PaginatedListResponse[User]{
    Items:   users,
    HasNext: len(users) == pageSize,
}
```
