# 2Chi Go Logger

Structured logging for 2Chi projects, built on Go's `log/slog` with optional context field injection and sensitive-data redaction.

```go
import yca_logger "github.com/yca-software/yca-go-core/logger"
```

## Logger

| Method | Description |
| --- | --- |
| `Debug(msg string, args ...any)` | Log at debug level |
| `Info(msg string, args ...any)` | Log at info level |
| `Warn(msg string, args ...any)` | Log at warn level |
| `Error(msg string, args ...any)` | Log at error level |
| `With(args ...any) Logger` | Return a child logger with static attributes |
| `WithContext(ctx context.Context) Logger` | Return a child logger with fields from `ContextExtractor` |

Attributes use slog's key-value pairs: `l.Info("request handled", "status", 200, "ms", 12)`.

## Configuration

Create a logger with `New(LoggerConfig)`:

| Field | Description |
| --- | --- |
| `Output` | Writer for log records; defaults to `os.Stdout` |
| `OutputType` | `"json"` or `"text"`; default is `"text"` |
| `ThresholdLevel` | `"debug"`, `"info"`, `"warn"`, or `"error"`; default is `"info"` |
| `ContextExtractor` | Pulls tracing fields from `context.Context` for `WithContext` |
| `Redaction` | Optional rules to sanitize sensitive attribute values |

When the threshold is `debug`, log records include a `source` field (file and line).

### Example

```go
log := yca_logger.New(yca_logger.LoggerConfig{
    OutputType:     "json",
    ThresholdLevel: "info",
    Redaction:      yca_logger.DefaultRedactionConfig(),
})

log.Info("server started", "port", 8080)
```

## Context extraction

Define a `ContextExtractor` to attach request-scoped fields automatically:

```go
extractor := func(ctx context.Context) map[string]any {
    if id, ok := ctx.Value(requestIDKey).(string); ok {
        return map[string]any{"request_id": id}
    }
    return nil
}

log := yca_logger.New(yca_logger.LoggerConfig{
    ContextExtractor: extractor,
})

log.WithContext(ctx).Info("handled")
```

`WithContext` is a no-op when `ctx` is nil, the extractor is nil, or the extractor returns an empty map.

## Redaction

`RedactionConfig` sanitizes flat string attributes before they are written:

| Field | Description |
| --- | --- |
| `Keys` | Attribute names (compared case-insensitively) whose values are replaced |
| `ValueRegexes` | Patterns matched against string values (e.g. Bearer tokens, DSNs) |
| `RedactedPlaceholder` | Replacement text; default is `[REDACTED]` |

`DefaultRedactionConfig()` covers common sensitive keys (`password`, `token`, `authorization`, etc.) and patterns such as Bearer tokens, query-string credentials, and database connection URLs.

Redaction applies to top-level slog attributes. To redact fields inside structs, implement `slog.LogValuer` on your types.

```go
log := yca_logger.New(yca_logger.LoggerConfig{
    OutputType: "json",
    Redaction:  yca_logger.DefaultRedactionConfig(),
})

log.Info("login", "password", "secret") // password value becomes [REDACTED]
```

## Testing

`MockLogger` is a testify mock for the `Logger` interface:

```go
m := new(yca_logger.MockLogger)
m.On("Info", "hello", "k", 1).Return()
m.On("With", "svc", "api").Return(m).Maybe()
m.On("WithContext", mock.Anything).Return(m).Maybe()

m.Info("hello", "k", 1)
m.AssertExpectations(t)
```

For integration-style tests, pass a `bytes.Buffer` as `Output` and assert on written JSON or text.
