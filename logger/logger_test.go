package yca_logger_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	yca_logger "github.com/yca-software/yca-go-core/logger"
)

// LoggerSuite exercises the production [logger.New] implementation with a
// capture buffer and JSON output for stable assertions.
type LoggerSuite struct {
	suite.Suite
	buf *bytes.Buffer
}

func TestLoggerSuite(t *testing.T) {
	suite.Run(t, new(LoggerSuite))
}

func (s *LoggerSuite) SetupTest() {
	s.buf = &bytes.Buffer{}
}

func (s *LoggerSuite) newLogger(cfg yca_logger.LoggerConfig) yca_logger.Logger {
	cfg.Output = s.buf
	return yca_logger.New(cfg)
}

func (s *LoggerSuite) lastJSON() map[string]any {
	s.T().Helper()
	lines := strings.Split(strings.TrimSpace(s.buf.String()), "\n")
	s.Require().NotEmpty(lines)
	var m map[string]any
	s.Require().NoError(json.Unmarshal([]byte(lines[len(lines)-1]), &m))
	return m
}

func (s *LoggerSuite) allJSON() []map[string]any {
	lines := strings.Split(strings.TrimSpace(s.buf.String()), "\n")
	out := make([]map[string]any, 0, len(lines))
	for _, line := range lines {
		if line == "" {
			continue
		}
		var m map[string]any
		s.Require().NoError(json.Unmarshal([]byte(line), &m))
		out = append(out, m)
	}
	return out
}

func (s *LoggerSuite) TestNew_TextDefault_InfoLevel() {
	l := s.newLogger(yca_logger.LoggerConfig{OutputType: "text"})
	l.Info("hello", "k", "v")
	body := s.buf.String()
	s.Contains(body, "hello")
	s.Contains(body, "level=INFO")
	s.Contains(body, "k=v")
}

func (s *LoggerSuite) TestNew_Threshold_DebugWarnErrorDefault() {
	cases := []struct {
		name     string
		cfgLevel string
		logFn    func(yca_logger.Logger)
		want     string
	}{
		{"debug_upper", "DEBUG", func(l yca_logger.Logger) { l.Debug("x") }, "DEBUG"},
		{"warn", "warn", func(l yca_logger.Logger) { l.Warn("w") }, "WARN"},
		{"error", "error", func(l yca_logger.Logger) { l.Error("e") }, "ERROR"},
		{"default_info_empty", "", func(l yca_logger.Logger) { l.Info("i") }, "INFO"},
		{"default_info_unknown", "verbose", func(l yca_logger.Logger) { l.Info("i") }, "INFO"},
	}
	for _, tc := range cases {
		s.Run(tc.name, func() {
			s.buf.Reset()
			l := s.newLogger(yca_logger.LoggerConfig{
				OutputType:     "json",
				ThresholdLevel: tc.cfgLevel,
			})
			tc.logFn(l)
			s.Equal(tc.want, s.lastJSON()["level"])
		})
	}
}

func (s *LoggerSuite) TestNew_Threshold_FiltersDebugWhenInfo() {
	l := s.newLogger(yca_logger.LoggerConfig{
		OutputType:     "json",
		ThresholdLevel: "info",
	})
	l.Debug("hidden")
	s.Empty(strings.TrimSpace(s.buf.String()))
	l.Info("visible")
	s.Equal("visible", s.lastJSON()["msg"])
}

func (s *LoggerSuite) TestNew_JSONStructuredFields() {
	l := s.newLogger(yca_logger.LoggerConfig{OutputType: "json"})
	l.Info("m", "n", 42, "ok", true)
	rec := s.lastJSON()
	s.Equal("m", rec["msg"])
	s.EqualValues(42, rec["n"])
	s.Equal(true, rec["ok"])
}

func (s *LoggerSuite) TestNew_Debug_AddsSourceField() {
	l := s.newLogger(yca_logger.LoggerConfig{
		OutputType:     "json",
		ThresholdLevel: "debug",
	})
	l.Debug("with-source")
	rec := s.lastJSON()
	_, has := rec["source"]
	s.True(has, "debug JSON should include source")
}

func (s *LoggerSuite) TestNew_Output_NilWriterUsesStdoutDoesNotPanic() {
	l := yca_logger.New(yca_logger.LoggerConfig{OutputType: "json"})
	s.NotNil(l)
	l.Info("smoke")
}

func (s *LoggerSuite) TestNew_Output_ExplicitWriter() {
	var buf bytes.Buffer
	l := yca_logger.New(yca_logger.LoggerConfig{
		Output:     &buf,
		OutputType: "json",
	})
	l.Warn("w")
	s.Contains(buf.String(), `"level":"WARN"`)
}

func (s *LoggerSuite) TestNew_AllLevelsWhenDebug() {
	l := s.newLogger(yca_logger.LoggerConfig{
		OutputType:     "json",
		ThresholdLevel: "debug",
	})
	l.Debug("d")
	l.Info("i")
	l.Warn("w")
	l.Error("e")
	recs := s.allJSON()
	s.Require().Len(recs, 4)
	s.Equal("DEBUG", recs[0]["level"])
	s.Equal("INFO", recs[1]["level"])
	s.Equal("WARN", recs[2]["level"])
	s.Equal("ERROR", recs[3]["level"])
}

func (s *LoggerSuite) TestNew_Redaction_NilPassesThrough() {
	l := s.newLogger(yca_logger.LoggerConfig{
		OutputType: "json",
		Redaction:  nil,
	})
	l.Info("m", "password", "secret")
	rec := s.lastJSON()
	s.Equal("secret", rec["password"])
}

func (s *LoggerSuite) TestNew_Redaction_KeyMatchLowercasesAttributeKey() {
	l := s.newLogger(yca_logger.LoggerConfig{
		OutputType: "json",
		Redaction: &yca_logger.RedactionConfig{
			Keys: []string{"password"},
		},
	})
	l.Info("m", "Password", "x")
	rec := s.lastJSON()
	s.Equal(yca_logger.DEFAULT_REDACTED_PLACEHOLDER, rec["Password"])
}

func (s *LoggerSuite) TestNew_Redaction_EmptyPlaceholderUsesDefault() {
	l := s.newLogger(yca_logger.LoggerConfig{
		OutputType: "json",
		Redaction: &yca_logger.RedactionConfig{
			Keys:                []string{"token"},
			RedactedPlaceholder: "",
		},
	})
	l.Info("m", "token", "abc")
	rec := s.lastJSON()
	s.Equal(yca_logger.DEFAULT_REDACTED_PLACEHOLDER, rec["token"])
}

func (s *LoggerSuite) TestNew_Redaction_CustomPlaceholder() {
	l := s.newLogger(yca_logger.LoggerConfig{
		OutputType: "json",
		Redaction: &yca_logger.RedactionConfig{
			Keys:                []string{"api_key"},
			RedactedPlaceholder: "***",
		},
	})
	l.Info("m", "api_key", "k")
	rec := s.lastJSON()
	s.Equal("***", rec["api_key"])
}

func (s *LoggerSuite) TestNew_Redaction_ValueRegex() {
	re := regexp.MustCompile(`Bearer\s+\S+`)
	l := s.newLogger(yca_logger.LoggerConfig{
		OutputType: "json",
		Redaction: &yca_logger.RedactionConfig{
			ValueRegexes: []*regexp.Regexp{re},
		},
	})
	l.Info("m", "Authorization", "Bearer abc.def")
	rec := s.lastJSON()
	s.Equal(yca_logger.DEFAULT_REDACTED_PLACEHOLDER, rec["Authorization"])
}

func (s *LoggerSuite) TestNew_Redaction_ErrorAttrUsesErrorString() {
	l := s.newLogger(yca_logger.LoggerConfig{
		OutputType: "json",
		Redaction: &yca_logger.RedactionConfig{
			Keys: []string{"password"},
		},
	})
	err := errors.New("boom")
	l.Info("m", "err", err)
	rec := s.lastJSON()
	s.Equal("boom", rec["err"])
}

func (s *LoggerSuite) TestNew_Redaction_DoesNotRewriteSlogBaseKeys() {
	l := s.newLogger(yca_logger.LoggerConfig{
		OutputType: "json",
		Redaction: &yca_logger.RedactionConfig{
			Keys: []string{"time", "level", "msg", "source"},
		},
	})
	l.Info("hello")
	rec := s.lastJSON()
	s.Contains(rec, "time")
	s.Equal("INFO", rec["level"])
	s.Equal("hello", rec["msg"])
}

func (s *LoggerSuite) TestDefaultRedactionConfig_Shape() {
	cfg := yca_logger.DefaultRedactionConfig()
	s.NotEmpty(cfg.Keys)
	s.NotEmpty(cfg.ValueRegexes)
	s.Equal(yca_logger.DEFAULT_REDACTED_PLACEHOLDER, cfg.RedactedPlaceholder)
}

func (s *LoggerSuite) TestDefaultRedactionConfig_BearerAndQueryString() {
	l := s.newLogger(yca_logger.LoggerConfig{
		OutputType: "json",
		Redaction:  yca_logger.DefaultRedactionConfig(),
	})
	l.Info("a", "hdr", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9")
	l.Info("b", "query", "next=/x&password=secret123&ok=1")
	recs := s.allJSON()
	s.Require().Len(recs, 2)
	s.Equal(yca_logger.DEFAULT_REDACTED_PLACEHOLDER, recs[0]["hdr"])
	s.Equal(yca_logger.DEFAULT_REDACTED_PLACEHOLDER, recs[1]["query"])
}

func (s *LoggerSuite) TestWith_ChainsStaticAttributes() {
	l := s.newLogger(yca_logger.LoggerConfig{OutputType: "json"})
	l.With("svc", "api").With("region", "eu").Info("rolled")
	rec := s.lastJSON()
	s.Equal("api", rec["svc"])
	s.Equal("eu", rec["region"])
	s.Equal("rolled", rec["msg"])
}

func (s *LoggerSuite) TestWithContext_NilContext_NoFields() {
	ext := func(context.Context) map[string]any { return map[string]any{"rid": "1"} }
	l := s.newLogger(yca_logger.LoggerConfig{
		OutputType:       "json",
		ContextExtractor: ext,
	})
	// nil ctx: production returns the same logger without merging extractor fields.
	l.WithContext(nil).Info("x") //nolint:staticcheck // contract: nil ctx is a no-op
	rec := s.lastJSON()
	_, ok := rec["rid"]
	s.False(ok)
}

func (s *LoggerSuite) TestWithContext_NilExtractor_NoFields() {
	l := s.newLogger(yca_logger.LoggerConfig{OutputType: "json"})
	ctx := context.WithValue(context.Background(), struct{ k string }{"k"}, "v")
	l.WithContext(ctx).Info("x")
	rec := s.lastJSON()
	s.Equal("x", rec["msg"])
}

func (s *LoggerSuite) TestWithContext_EmptyMap_NoFields() {
	ext := func(context.Context) map[string]any { return map[string]any{} }
	l := s.newLogger(yca_logger.LoggerConfig{
		OutputType:       "json",
		ContextExtractor: ext,
	})
	l.WithContext(context.Background()).Info("x")
	rec := s.lastJSON()
	s.Equal("x", rec["msg"])
}

func (s *LoggerSuite) TestWithContext_MergesExtractorFields() {
	ext := func(context.Context) map[string]any {
		return map[string]any{"request_id": "abc"}
	}
	l := s.newLogger(yca_logger.LoggerConfig{
		OutputType:       "json",
		ContextExtractor: ext,
	})
	l.WithContext(context.Background()).Info("done")
	rec := s.lastJSON()
	s.Equal("abc", rec["request_id"])
	s.Equal("done", rec["msg"])
}

func (s *LoggerSuite) TestWith_ThenWithContext_Combines() {
	ext := func(context.Context) map[string]any {
		return map[string]any{"rid": "r1"}
	}
	l := s.newLogger(yca_logger.LoggerConfig{
		OutputType:       "json",
		ContextExtractor: ext,
	})
	l.With("svc", "x").WithContext(context.Background()).Info("m")
	rec := s.lastJSON()
	s.Equal("x", rec["svc"])
	s.Equal("r1", rec["rid"])
}

func (s *LoggerSuite) TestDiscardWriter_NoOutput() {
	l := yca_logger.New(yca_logger.LoggerConfig{
		Output:     io.Discard,
		OutputType: "json",
	})
	l.Info("silent")
	s.NotPanics(func() { l.Info("again") })
}
