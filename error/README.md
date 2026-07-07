# 2Chi Go Error

Typed HTTP/API errors for 2Chi projects. Handlers return `*Error` values with a status code and a JSON body shaped for client i18n via `errorCode`.

```go
import chi_error "github.com/yca-software/yca-go-core/error"
```

## Error shape

| Field        | JSON        | Description                                                 |
| ------------ | ----------- | ----------------------------------------------------------- |
| `StatusCode` | omitted     | HTTP status for the handler                                 |
| `Err`        | omitted     | Internal cause for logs; never sent to clients              |
| `ErrorCode`  | `errorCode` | Stable machine-readable code (maps to i18n keys in the SPA) |
| `Message`    | `message`   | Optional client-facing text                                 |
| `Extra`      | `extra`     | Optional structured context (validation fields, etc.)       |

`ErrorResponse` documents the same JSON fields for Swagger/OpenAPI.

## Constructors

Each constructor takes `(err error, errorCode string, extra any)` and sets the matching HTTP status:

| Function                      | Status |
| ----------------------------- | ------ |
| `NewBadRequestError`          | 400    |
| `NewUnauthorizedError`        | 401    |
| `NewPaymentRequiredError`     | 402    |
| `NewForbiddenError`           | 403    |
| `NewNotFoundError`            | 404    |
| `NewConflictError`            | 409    |
| `NewUnprocessableEntityError` | 422    |
| `NewTooManyRequestsError`     | 429    |
| `NewInternalServerError`      | 500    |
| `NewServiceUnavailableError`  | 503    |

## Message rules

`message` in the JSON response is derived automatically:

- **5xx** — always empty (internal `Err` is for logging only)
- **`errorCode` set** — empty (clients localize via `errorCode`)
- **Otherwise** — uses `err.Error()` when `err` is non-nil

## Handler helpers

`AsError(err error) (*Error, bool)` extracts `*Error` from a returned error, including wrapped errors. Use it in HTTP error handlers to read `StatusCode` and serialize the JSON body.

`Error` implements `Unwrap()` so `errors.Is` and `errors.As` work with the underlying cause.

## Example

```go
// Domain error with i18n key — preferred for API responses
return chi_error.NewNotFoundError(nil, "USER_NOT_FOUND", nil)

// Validation with structured extra
return chi_error.NewBadRequestError(nil, "INVALID_EMAIL", map[string]any{
    "field": "email",
})

// Internal failure — log Err, return safe JSON
return chi_error.NewInternalServerError(dbErr, "InternalServerError", nil)

// In an HTTP error handler:
if apiErr, ok := chi_error.AsError(err); ok {
    return c.JSON(apiErr.StatusCode, apiErr)
}
```
