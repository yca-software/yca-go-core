# 2Chi Go Validator

Struct validation for 2Chi projects, built on [go-playground/validator](https://github.com/go-playground/validator).

```go
import chi_validator "github.com/yca-software/yca-go-core/validator"
```

## API

| Symbol                                              | Description                                 |
| --------------------------------------------------- | ------------------------------------------- |
| `New() Validator`                                   | Create a validator instance                 |
| `ValidateStruct(s any) *map[string]ValidationError` | Validate a struct; returns `nil` when valid |

Tag structs with `validate` rules (for example `required`, `email`, `min`, `max`). Errors are keyed by struct field name.

## ValidationError

| Field   | JSON    | Description                             |
| ------- | ------- | --------------------------------------- |
| `Tag`   | `tag`   | Failed validation rule                  |
| `Param` | `param` | Rule parameter (e.g. `18` for `min=18`) |
| `Value` | `value` | Invalid field value                     |
| `Error` | `error` | Human-readable message from validator   |

## Behavior

- **Valid struct** — returns `nil`
- **`nil` input** — returns `nil` (no panic)
- **Non-struct input** — returns a single entry under key `""` with the underlying error

## Example

```go
type CreateUserRequest struct {
    Email    string `validate:"required,email"`
    Age      int    `validate:"min=18,max=100"`
    Username string `validate:"required,min=3,max=20"`
}

v := chi_validator.New()

req := CreateUserRequest{Email: "bad", Age: 15, Username: "ab"}
errs := v.ValidateStruct(req)
if errs != nil {
    for field, ve := range *errs {
        // field: "Email", ve.Tag: "email", ve.Error: ...
    }
}
```
