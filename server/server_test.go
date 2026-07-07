package chi_server_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/suite"

	chi_server "github.com/yca-software/yca-go-core/server"
)

type stubDep struct {
	err error
}

func (d stubDep) Check(_ context.Context) error { return d.err }

type ServerSuite struct {
	suite.Suite
	echo *echo.Echo
}

func TestServerSuite(t *testing.T) {
	suite.Run(t, new(ServerSuite))
}

func (s *ServerSuite) SetupTest() {
	s.echo = echo.New()
}

func testServerConfig(registerRoutes func(e *echo.Echo)) chi_server.ServerConfig {
	return chi_server.ServerConfig{
		Port:               0,
		BodyLimit:          "32M",
		ServerReadTimeout:  30,
		ServerWriteTimeout: 30,
		ServerIdleTimeout:  120,
		RegisterRoutes:     registerRoutes,
	}
}

func (s *ServerSuite) get(path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	return rec
}

func (s *ServerSuite) registerHealth(deps ...chi_server.ReadinessDependency) {
	chi_server.RegisterHealthHandlers(s.echo, deps)
}

func (s *ServerSuite) TestNew_echoHTTPError_returnsErrorCode() {
	srv := chi_server.New(testServerConfig(func(e *echo.Echo) {
		e.GET("/missing", func(c echo.Context) error {
			return echo.ErrNotFound
		})
	}))
	req := httptest.NewRequest(http.MethodGet, "/missing", nil)
	rec := httptest.NewRecorder()
	srv.GetEcho().ServeHTTP(rec, req)

	s.Equal(http.StatusNotFound, rec.Code)
	var body map[string]string
	s.Require().NoError(json.Unmarshal(rec.Body.Bytes(), &body))
	s.Equal("NotFound", body["errorCode"])
	s.NotContains(body, "message")
}

func (s *ServerSuite) TestNew_invokesRegisterRoutes() {
	var registered bool
	srv := chi_server.New(testServerConfig(func(e *echo.Echo) {
		registered = true
		e.GET("/ping", func(c echo.Context) error {
			return c.String(http.StatusOK, "pong")
		})
	}))
	s.True(registered)
	s.NotNil(srv)
}

func (s *ServerSuite) TestShutdown_onNonStartedServer() {
	srv := chi_server.New(testServerConfig(nil))
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	s.NoError(srv.Shutdown(ctx))
}

func (s *ServerSuite) TestHealth_returnsOK() {
	s.registerHealth()
	rec := s.get("/health")
	s.Equal(http.StatusOK, rec.Code)

	var body map[string]string
	s.Require().NoError(json.Unmarshal(rec.Body.Bytes(), &body))
	s.Equal("ok", body["status"])
}

func (s *ServerSuite) TestReady_noDependencies_returnsOK() {
	s.registerHealth()
	rec := s.get("/ready")
	s.Equal(http.StatusOK, rec.Code)
}

func (s *ServerSuite) TestReady_dependencyFails_returns503() {
	s.registerHealth(stubDep{err: errors.New("database connection failed")})
	rec := s.get("/ready")
	s.Equal(http.StatusServiceUnavailable, rec.Code)

	var body map[string]string
	s.Require().NoError(json.Unmarshal(rec.Body.Bytes(), &body))
	s.Equal("unavailable", body["status"])
	s.NotContains(body, "reason")
}

func (s *ServerSuite) TestReady_allDependenciesPass_returnsOK() {
	s.registerHealth(stubDep{}, stubDep{})
	rec := s.get("/ready")
	s.Equal(http.StatusOK, rec.Code)
}

func (s *ServerSuite) TestRegisterMetricsHandlers() {
	chi_server.RegisterMetricsHandlers(s.echo)
	rec := s.get("/metrics")

	s.Equal(http.StatusOK, rec.Code)
	s.Contains(rec.Header().Get("Content-Type"), "text/plain")
	s.True(strings.Contains(rec.Body.String(), "# HELP") || rec.Body.Len() > 0)
}
