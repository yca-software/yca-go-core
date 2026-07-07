package chi_ratelimit_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	redisclient "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"

	chi_error "github.com/yca-software/yca-go-core/error"
	chi_ratelimit "github.com/yca-software/yca-go-core/ratelimit"
	chi_types "github.com/yca-software/yca-go-core/types"
)

type RateLimitSuite struct {
	suite.Suite
	mr    *miniredis.Miniredis
	redis *redisclient.Client
}

func TestRateLimitSuite(t *testing.T) {
	suite.Run(t, new(RateLimitSuite))
}

func (s *RateLimitSuite) SetupSuite() {
	mr, err := miniredis.Run()
	s.Require().NoError(err)
	s.mr = mr
	s.redis = redisclient.NewClient(&redisclient.Options{Addr: mr.Addr()})
}

func (s *RateLimitSuite) TearDownSuite() {
	if s.redis != nil {
		s.redis.Close()
	}
	if s.mr != nil {
		s.mr.Close()
	}
}

func (s *RateLimitSuite) SetupTest() {
	s.mr.SetError("")
	s.mr.FlushAll()
}

func testHTTPErrorHandler(err error, c echo.Context) {
	if apiErr, ok := chi_error.AsError(err); ok {
		_ = c.JSON(apiErr.StatusCode, apiErr)
		return
	}
	_ = c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
}

func (s *RateLimitSuite) newEcho() *echo.Echo {
	e := echo.New()
	e.HTTPErrorHandler = testHTTPErrorHandler
	return e
}

func (s *RateLimitSuite) TestNilRedis_PassThrough() {
	rl := chi_ratelimit.NewRateLimiter(nil, nil, nil)
	e := s.newEcho()
	e.Use(rl.IPRateLimit("1-M"))
	e.GET("/", func(c echo.Context) error { return c.NoContent(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	s.Equal(http.StatusOK, rec.Code)
}

func (s *RateLimitSuite) TestIPRateLimit_EnforcesLimit() {
	rl := chi_ratelimit.NewRateLimiter(s.redis, nil, nil)
	e := s.newEcho()
	e.Use(rl.IPRateLimit("1-M"))
	e.GET("/", func(c echo.Context) error { return c.NoContent(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "203.0.113.10:12345"

	rec1 := httptest.NewRecorder()
	e.ServeHTTP(rec1, req)
	s.Equal(http.StatusOK, rec1.Code)
	s.NotEmpty(rec1.Header().Get("X-RateLimit-Limit"))

	rec2 := httptest.NewRecorder()
	e.ServeHTTP(rec2, req)
	s.Equal(http.StatusTooManyRequests, rec2.Code)
}

func (s *RateLimitSuite) serveWithMiddleware(mw echo.MiddlewareFunc, req *http.Request) *httptest.ResponseRecorder {
	e := s.newEcho()
	e.Use(mw)
	e.GET("/", func(c echo.Context) error { return c.NoContent(http.StatusOK) })
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

func (s *RateLimitSuite) TestIPDeviceRateLimit_SeparateBucketsPerDevice() {
	rl := chi_ratelimit.NewRateLimiter(s.redis, nil, nil)
	mw := rl.IPDeviceRateLimit("1-M")

	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	req1.RemoteAddr = "203.0.113.10:12345"
	req1.AddCookie(&http.Cookie{Name: chi_ratelimit.DeviceIDCookieName, Value: uuid.NewString()})

	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.RemoteAddr = "203.0.113.10:12345"
	req2.AddCookie(&http.Cookie{Name: chi_ratelimit.DeviceIDCookieName, Value: uuid.NewString()})

	s.Equal(http.StatusOK, s.serveWithMiddleware(mw, req1).Code)
	s.Equal(http.StatusOK, s.serveWithMiddleware(mw, req2).Code)
}

func (s *RateLimitSuite) TestPrincipalRateLimit_UsesAuthenticatedSubject() {
	rl := chi_ratelimit.NewRateLimiter(s.redis, nil, nil)
	mw := rl.PrincipalRateLimit("1-M")

	userID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "203.0.113.10:12345"

	e := s.newEcho()
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("accessInfo", &chi_types.AccessInfo{
				Type:      chi_types.AccessTypeUser,
				SubjectID: userID,
			})
			return next(c)
		}
	})
	e.Use(mw)
	e.GET("/", func(c echo.Context) error { return c.NoContent(http.StatusOK) })

	rec1 := httptest.NewRecorder()
	e.ServeHTTP(rec1, req)
	s.Equal(http.StatusOK, rec1.Code)

	rec2 := httptest.NewRecorder()
	e.ServeHTTP(rec2, req)
	s.Equal(http.StatusTooManyRequests, rec2.Code)
}

func (s *RateLimitSuite) TestScopedPrincipalRateLimit_ScopesKeys() {
	rl := chi_ratelimit.NewRateLimiter(s.redis, nil, nil)
	scopeA := rl.ScopedPrincipalRateLimit("1-M", "email")
	scopeB := rl.ScopedPrincipalRateLimit("1-M", "sms")

	userID := uuid.New()
	setUser := func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("accessInfo", &chi_types.AccessInfo{
				Type:      chi_types.AccessTypeUser,
				SubjectID: userID,
			})
			return next(c)
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "203.0.113.10:12345"

	e := s.newEcho()
	e.Use(setUser)
	e.Use(scopeA)
	e.GET("/", func(c echo.Context) error { return c.NoContent(http.StatusOK) })

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)

	e2 := s.newEcho()
	e2.Use(setUser)
	e2.Use(scopeB)
	e2.GET("/", func(c echo.Context) error { return c.NoContent(http.StatusOK) })

	rec2 := httptest.NewRecorder()
	e2.ServeHTTP(rec2, req)
	s.Equal(http.StatusOK, rec2.Code)
}

func (s *RateLimitSuite) TestIPRateLimit_RedisFailureReturns429() {
	rl := chi_ratelimit.NewRateLimiter(s.redis, nil, nil)
	mw := rl.IPRateLimit("1-M")
	s.mr.SetError("redis unavailable")

	rec := s.serveWithMiddleware(mw, httptest.NewRequest(http.MethodGet, "/", nil))
	s.Equal(http.StatusTooManyRequests, rec.Code)
}
