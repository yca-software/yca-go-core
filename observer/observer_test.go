package chi_observer_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/suite"

	chi_observer "github.com/yca-software/yca-go-core/observer"
)

type ObserverSuite struct {
	suite.Suite
	namespace string
	obs       *chi_observer.Observer
}

func TestObserverSuite(t *testing.T) {
	suite.Run(t, new(ObserverSuite))
}

func (s *ObserverSuite) SetupTest() {
	reg := prometheus.NewRegistry()
	prometheus.DefaultRegisterer = reg
	prometheus.DefaultGatherer = reg

	s.namespace = "obs_test"
	obs, err := chi_observer.New(chi_observer.ObserverConfig{
		Namespace: s.namespace,
		AppName:   "testapp",
	})
	s.Require().NoError(err)
	s.obs = obs
}

func (s *ObserverSuite) metricName(subsystem, name string) string {
	return strings.Join([]string{s.namespace, subsystem, name}, "_")
}

func (s *ObserverSuite) hasCounter(name string, labels map[string]string) bool {
	s.T().Helper()
	mfs, err := prometheus.DefaultGatherer.Gather()
	s.Require().NoError(err)

	for _, mf := range mfs {
		if mf.GetName() != name {
			continue
		}
		for _, m := range mf.GetMetric() {
			if s.labelsMatch(m, labels) {
				return true
			}
		}
	}
	return false
}

func (s *ObserverSuite) findCounter(name string, labels map[string]string) float64 {
	s.T().Helper()
	mfs, err := prometheus.DefaultGatherer.Gather()
	s.Require().NoError(err)

	for _, mf := range mfs {
		if mf.GetName() != name {
			continue
		}
		for _, m := range mf.GetMetric() {
			if s.labelsMatch(m, labels) {
				return m.GetCounter().GetValue()
			}
		}
	}
	s.FailNow("counter not found", "name=%s labels=%v", name, labels)
	return 0
}

func (s *ObserverSuite) findHistogramSampleCount(name string, labels map[string]string) uint64 {
	s.T().Helper()
	mfs, err := prometheus.DefaultGatherer.Gather()
	s.Require().NoError(err)

	for _, mf := range mfs {
		if mf.GetName() != name {
			continue
		}
		for _, m := range mf.GetMetric() {
			if s.labelsMatch(m, labels) {
				return m.GetHistogram().GetSampleCount()
			}
		}
	}
	s.FailNow("histogram not found", "name=%s labels=%v", name, labels)
	return 0
}

func (s *ObserverSuite) labelsMatch(m *dto.Metric, want map[string]string) bool {
	got := make(map[string]string, len(m.Label))
	for _, lp := range m.Label {
		got[lp.GetName()] = lp.GetValue()
	}
	for k, v := range want {
		if got[k] != v {
			return false
		}
	}
	return true
}

func (s *ObserverSuite) serveHTTP(e *echo.Echo, method, path string) *httptest.ResponseRecorder {
	s.T().Helper()
	req := httptest.NewRequest(method, path, nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

func (s *ObserverSuite) TestNew_registersCollectors() {
	s.obs.RecordQuery("ping", "health", "ping_health", "success", time.Millisecond)
	s.obs.RecordRateLimitHit(http.MethodGet, "/health", chi_observer.PrincipalTypeIP, "127.0.0.1", "127.0.0.1")
	s.obs.RecordJobPublished("cleanup")
	s.obs.RecordJobConsumerOutcome("cleanup", chi_observer.JobOutcomeSuccess)
	s.obs.RecordJobConsumerDuration("cleanup", time.Millisecond)

	rec := s.serveHTTP(s.echoWithHello(), http.MethodGet, "/hello")
	s.Equal(http.StatusOK, rec.Code)

	mfs, err := prometheus.DefaultGatherer.Gather()
	s.Require().NoError(err)

	names := make([]string, 0, len(mfs))
	for _, mf := range mfs {
		names = append(names, mf.GetName())
	}

	s.Contains(names, s.metricName("http", "requests_total"))
	s.Contains(names, s.metricName("http", "request_duration_seconds"))
	s.Contains(names, s.metricName("db", "query_duration_seconds"))
	s.Contains(names, s.metricName("rate_limit", "hits_total"))
	s.Contains(names, s.metricName("job_consumer", "published_total"))
	s.Contains(names, s.metricName("job_consumer", "outcome_total"))
	s.Contains(names, s.metricName("job_consumer", "duration_seconds"))
}

func (s *ObserverSuite) TestRecordRateLimitHit_observesCounter() {
	userID := "11111111-1111-4111-8111-111111111101"
	clientIP := "203.0.113.10"

	s.obs.RecordRateLimitHit(
		http.MethodPost,
		"/api/v1/auth/login",
		chi_observer.PrincipalTypeUser,
		userID,
		clientIP,
	)

	labels := map[string]string{
		"method":         "POST",
		"route":          "/api/v1/auth/login",
		"principal_type": "user",
		"principal_hash": chi_observer.HashRateLimitIdentifier(userID),
		"ip_hash":        chi_observer.HashRateLimitIdentifier(clientIP),
		"app":            "testapp",
	}
	s.Equal(1.0, s.findCounter(s.metricName("rate_limit", "hits_total"), labels))
	s.NotContains(labels["principal_hash"], userID)
	s.NotContains(labels["ip_hash"], clientIP)
}

func (s *ObserverSuite) TestRecordRateLimitHit_emptyLabelsUseFallbacks() {
	s.obs.RecordRateLimitHit("", "", "", "", "")

	labels := map[string]string{
		"method":         "unknown",
		"route":          "unmatched",
		"principal_type": "unknown",
		"principal_hash": "unknown",
		"ip_hash":        "unknown",
		"app":            "testapp",
	}
	s.Equal(1.0, s.findCounter(s.metricName("rate_limit", "hits_total"), labels))
}

func (s *ObserverSuite) TestEchoMiddleware_recordsSuccessfulRequest() {
	e := echo.New()
	e.Use(s.obs.EchoMiddleware(nil))
	e.GET("/hello", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	rec := s.serveHTTP(e, http.MethodGet, "/hello")
	s.Equal(http.StatusOK, rec.Code)

	labels := map[string]string{
		"method": "GET",
		"route":  "/hello",
		"status": "200",
		"app":    "testapp",
	}

	s.Equal(1.0, s.findCounter(s.metricName("http", "requests_total"), labels))
	s.Equal(uint64(1), s.findHistogramSampleCount(s.metricName("http", "request_duration_seconds"), labels))
}

func (s *ObserverSuite) TestEchoMiddleware_skipsMetricsEndpoint() {
	e := echo.New()
	e.Use(s.obs.EchoMiddleware(nil))
	e.GET("/metrics", func(c echo.Context) error {
		return c.String(http.StatusOK, "metrics")
	})

	rec := s.serveHTTP(e, http.MethodGet, "/metrics")
	s.Equal(http.StatusOK, rec.Code)

	labels := map[string]string{
		"method": "GET",
		"route":  "/metrics",
		"status": "200",
		"app":    "testapp",
	}
	s.False(s.hasCounter(s.metricName("http", "requests_total"), labels))
}

func (s *ObserverSuite) TestEchoMiddleware_recordsHTTPError() {
	e := echo.New()
	e.Use(s.obs.EchoMiddleware(nil))
	e.GET("/missing", func(c echo.Context) error {
		return echo.NewHTTPError(http.StatusNotFound, "not found")
	})

	rec := s.serveHTTP(e, http.MethodGet, "/missing")
	s.Equal(http.StatusNotFound, rec.Code)

	labels := map[string]string{
		"method": "GET",
		"route":  "/missing",
		"status": "404",
		"app":    "testapp",
	}
	s.Equal(1.0, s.findCounter(s.metricName("http", "requests_total"), labels))
}

func (s *ObserverSuite) TestEchoMiddleware_recordsGenericErrorAs500() {
	e := echo.New()
	e.Use(s.obs.EchoMiddleware(nil))
	e.GET("/fail", func(c echo.Context) error {
		return errors.New("boom")
	})

	rec := s.serveHTTP(e, http.MethodGet, "/fail")
	s.Equal(http.StatusInternalServerError, rec.Code)

	labels := map[string]string{
		"method": "GET",
		"route":  "/fail",
		"status": "500",
		"app":    "testapp",
	}
	s.Equal(1.0, s.findCounter(s.metricName("http", "requests_total"), labels))
}

func (s *ObserverSuite) TestEchoMiddleware_recordsUnmatchedRoute() {
	e := echo.New()
	e.Use(s.obs.EchoMiddleware(nil))

	rec := s.serveHTTP(e, http.MethodGet, "/unknown")
	s.Equal(http.StatusNotFound, rec.Code)

	labels := map[string]string{
		"method": "GET",
		"route":  "unmatched",
		"status": "404",
		"app":    "testapp",
	}
	s.Equal(1.0, s.findCounter(s.metricName("http", "requests_total"), labels))
}

func (s *ObserverSuite) TestRecordQuery_observesHistogram() {
	s.obs.RecordQuery("select", "users", "select_users", "success", 25*time.Millisecond)

	labels := map[string]string{
		"operation":  "select",
		"table":      "users",
		"query_type": "select_users",
		"status":     "success",
		"app":        "testapp",
	}
	s.Equal(uint64(1), s.findHistogramSampleCount(s.metricName("db", "query_duration_seconds"), labels))
}

func (s *ObserverSuite) TestRecordQuery_multipleObservations() {
	s.obs.RecordQuery("insert", "orders", "insert_orders", "success", time.Millisecond)
	s.obs.RecordQuery("insert", "orders", "insert_orders", "error", 2*time.Millisecond)

	success := map[string]string{"operation": "insert", "table": "orders", "query_type": "insert_orders", "status": "success", "app": "testapp"}
	errLabels := map[string]string{"operation": "insert", "table": "orders", "query_type": "insert_orders", "status": "error", "app": "testapp"}

	s.Equal(uint64(1), s.findHistogramSampleCount(s.metricName("db", "query_duration_seconds"), success))
	s.Equal(uint64(1), s.findHistogramSampleCount(s.metricName("db", "query_duration_seconds"), errLabels))
}

func (s *ObserverSuite) TestGetQueryMetricsHook() {
	hook := s.obs.GetQueryMetricsHook()
	s.NotNil(hook)
	hook.RecordQuery("select", "users", "select_users", "success", time.Millisecond)
}

func (s *ObserverSuite) TestRecordJobMetrics_perJob() {
	s.obs.RecordJobPublished("cleanup")
	s.obs.RecordJobConsumerOutcome("cleanup", chi_observer.JobOutcomeSuccess)
	s.obs.RecordJobConsumerDuration("cleanup", 150*time.Millisecond)

	published := map[string]string{"job": "cleanup", "app": "testapp"}
	outcome := map[string]string{"job": "cleanup", "outcome": chi_observer.JobOutcomeSuccess, "app": "testapp"}
	duration := map[string]string{"job": "cleanup", "app": "testapp"}

	s.Equal(1.0, s.findCounter(s.metricName("job_consumer", "published_total"), published))
	s.Equal(1.0, s.findCounter(s.metricName("job_consumer", "outcome_total"), outcome))
	s.Equal(uint64(1), s.findHistogramSampleCount(s.metricName("job_consumer", "duration_seconds"), duration))
}

func (s *ObserverSuite) TestGetJobMetricsHook() {
	hook := s.obs.GetJobMetricsHook()
	s.NotNil(hook)
	hook.RecordJobPublished("apply_scheduled_plan_changes")
}

func (s *ObserverSuite) TestObserver_implementsQueryMetricsHook() {
	var _ chi_observer.QueryMetricsHook = s.obs
}

func (s *ObserverSuite) TestNew_idempotentRegistration() {
	second, err := chi_observer.New(chi_observer.ObserverConfig{
		Namespace: s.namespace,
		AppName:   "testapp",
	})
	s.Require().NoError(err)
	s.NotNil(second)

	rec := s.serveHTTP(s.echoWithHello(), http.MethodGet, "/hello")
	s.Equal(http.StatusOK, rec.Code)
}

func (s *ObserverSuite) echoWithHello() *echo.Echo {
	e := echo.New()
	e.Use(s.obs.EchoMiddleware(nil))
	e.GET("/hello", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})
	return e
}
