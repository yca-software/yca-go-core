package yca_logger

import (
	"context"
	"io"
	"log/slog"
	"os"
	"regexp"
	"strings"
)

const DEFAULT_REDACTED_PLACEHOLDER = "[REDACTED]"

// ContextExtractor pulls values out of a context and returns a map.
// This allows APIs to define exactly what tracing data (like request IDs or User IDs)
// gets automatically attached to logs via WithContext().
type ContextExtractor func(ctx context.Context) map[string]any

// RedactionConfig holds the rules for sanitizing sensitive data in flat string attributes.
// Note: For redacting fields inside structs, models should implement the slog.LogValuer interface.
type RedactionConfig struct {
	// Keys lists attribute names (lowercase) that trigger redaction when the
	// log attribute key equals the entry (attribute key is lowercased for comparison).
	Keys []string
	// ValueRegexes scans string values for sensitive patterns (e.g., SSN, Bearer Tokens).
	ValueRegexes []*regexp.Regexp

	RedactedPlaceholder string // The placeholder to use for redacted values default is "[REDACTED]"
}

// DefaultRedactionConfig provides a sensible baseline for APIs.
func DefaultRedactionConfig() *RedactionConfig {
	return &RedactionConfig{
		Keys: []string{
			"password", "passwd", "pwd", "secret", "token", "apikey", "api_key",
			"authorization", "cookie", "session", "cvv", "signature",
		},
		ValueRegexes: []*regexp.Regexp{
			// Credentials in strings/queries
			regexp.MustCompile(`(?i)(password|passwd|pwd)=[^\s&]+`),
			regexp.MustCompile(`(?i)(secret|token|apikey|api_key)=[^\s&]+`),
			// Connection strings (DSNs)
			regexp.MustCompile(`(?i)(:[^:@]+)@`), // :password@ in URLs
			regexp.MustCompile(`(?i)postgres(?:ql)?://[^\s]+`),
			regexp.MustCompile(`(?i)redis(?:s)?://[^\s]+`),
			regexp.MustCompile(`(?i)amqp(?:s)?://[^\s]+`),
			regexp.MustCompile(`(?i)Bearer\s+[a-zA-Z0-9\-\._~+/]+=*`),
		},
		RedactedPlaceholder: DEFAULT_REDACTED_PLACEHOLDER,
	}
}

type LoggerConfig struct {
	// Output is where log records are written. If nil, os.Stdout is used.
	Output io.Writer
	// OutputType is "json" or "text"; default is "text".
	OutputType string
	// ThresholdLevel is "debug", "info", "warn", or "error"; default is "info".
	ThresholdLevel string
	// ContextExtractor injects fields from context via WithContext.
	ContextExtractor ContextExtractor
	// Redaction, if non-nil, applies ReplaceAttr sanitization to attributes.
	Redaction *RedactionConfig
}

type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
	With(args ...any) Logger
	WithContext(ctx context.Context) Logger
}

type slogLogger struct {
	logger    *slog.Logger
	extractor ContextExtractor
}

func New(cfg LoggerConfig) Logger {
	var slogLevel slog.Level
	switch strings.ToLower(cfg.ThresholdLevel) {
	case "debug":
		slogLevel = slog.LevelDebug
	case "warn":
		slogLevel = slog.LevelWarn
	case "error":
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level:     slogLevel,
		AddSource: slogLevel == slog.LevelDebug, // Show file/line only on debug
	}

	if cfg.Redaction != nil {
		redactedPlaceholder := cfg.Redaction.RedactedPlaceholder
		if redactedPlaceholder == "" {
			redactedPlaceholder = DEFAULT_REDACTED_PLACEHOLDER
		}
		opts.ReplaceAttr = func(groups []string, a slog.Attr) slog.Attr {
			// Skip internal slog base keys
			if a.Key == slog.TimeKey || a.Key == slog.LevelKey || a.Key == slog.MessageKey {
				return a
			}

			// Format native errors safely
			if err, ok := a.Value.Any().(error); ok {
				return slog.String(a.Key, err.Error())
			}

			lowerKey := strings.ToLower(a.Key)
			for _, k := range cfg.Redaction.Keys {
				if lowerKey == k {
					return slog.String(a.Key, redactedPlaceholder)
				}
			}

			// Regex Match on explicit Strings (e.g., value contains "Bearer ...")
			if len(cfg.Redaction.ValueRegexes) > 0 && a.Value.Kind() == slog.KindString {
				valStr := a.Value.String()
				for _, re := range cfg.Redaction.ValueRegexes {
					if re.MatchString(valStr) {
						return slog.String(a.Key, redactedPlaceholder)
					}
				}
			}

			return a
		}
	}

	out := cfg.Output
	if out == nil {
		out = os.Stdout
	}

	var handler slog.Handler
	if strings.ToLower(cfg.OutputType) == "json" {
		handler = slog.NewJSONHandler(out, opts)
	} else {
		handler = slog.NewTextHandler(out, opts)
	}

	return &slogLogger{
		logger:    slog.New(handler),
		extractor: cfg.ContextExtractor,
	}
}

func (l *slogLogger) Debug(msg string, args ...any) { l.logger.Debug(msg, args...) }
func (l *slogLogger) Info(msg string, args ...any)  { l.logger.Info(msg, args...) }
func (l *slogLogger) Warn(msg string, args ...any)  { l.logger.Warn(msg, args...) }
func (l *slogLogger) Error(msg string, args ...any) { l.logger.Error(msg, args...) }

func (l *slogLogger) With(args ...any) Logger {
	return &slogLogger{
		logger:    l.logger.With(args...),
		extractor: l.extractor,
	}
}

func (l *slogLogger) WithContext(ctx context.Context) Logger {
	if ctx == nil || l.extractor == nil {
		return l
	}

	fields := l.extractor(ctx)
	if len(fields) == 0 {
		return l
	}

	// Pre-allocate slice capacity
	args := make([]any, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}

	return &slogLogger{
		logger:    l.logger.With(args...),
		extractor: l.extractor,
	}
}
