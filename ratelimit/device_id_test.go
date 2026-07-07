package chi_ratelimit_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/suite"

	chi_ratelimit "github.com/yca-software/yca-go-core/ratelimit"
)

type DeviceIDSuite struct {
	suite.Suite
}

func TestDeviceIDSuite(t *testing.T) {
	suite.Run(t, new(DeviceIDSuite))
}

func (s *DeviceIDSuite) newContext(req *http.Request) echo.Context {
	e := echo.New()
	return e.NewContext(req, httptest.NewRecorder())
}

func (s *DeviceIDSuite) TestEnsureDeviceID_IssuesCookie() {
	var issued string
	mw := chi_ratelimit.EnsureDeviceID("local")
	handler := mw(func(c echo.Context) error {
		issued, _ = c.Get("deviceId").(string)
		return nil
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)

	s.Require().NoError(handler(c))

	s.NotEmpty(issued)
	s.True(isUUIDv4(issued))

	var cookie *http.Cookie
	for _, ck := range rec.Result().Cookies() {
		if ck.Name == chi_ratelimit.DeviceIDCookieName {
			cookie = ck
			break
		}
	}
	s.Require().NotNil(cookie)
	s.Equal(issued, cookie.Value)
	s.True(cookie.HttpOnly)
	s.False(cookie.Secure)
}

func (s *DeviceIDSuite) TestEnsureDeviceID_ReusesCookie() {
	deviceID := uuid.NewString()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: chi_ratelimit.DeviceIDCookieName, Value: deviceID})

	var resolved string
	mw := chi_ratelimit.EnsureDeviceID("local")
	s.Require().NoError(mw(func(c echo.Context) error {
		resolved, _ = c.Get("deviceId").(string)
		return nil
	})(s.newContext(req)))

	s.Equal(deviceID, resolved)
}

func (s *DeviceIDSuite) TestEnsureDeviceID_ReusesHeader() {
	deviceID := uuid.NewString()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Device-Id", deviceID)

	var resolved string
	mw := chi_ratelimit.EnsureDeviceID("local")
	s.Require().NoError(mw(func(c echo.Context) error {
		resolved, _ = c.Get("deviceId").(string)
		return nil
	})(s.newContext(req)))

	s.Equal(deviceID, resolved)
}

func (s *DeviceIDSuite) TestEnsureDeviceID_InvalidValuesIgnored() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: chi_ratelimit.DeviceIDCookieName, Value: "not-a-uuid"})
	req.Header.Set("X-Device-Id", "also-invalid")

	var resolved string
	mw := chi_ratelimit.EnsureDeviceID("local")
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)
	s.Require().NoError(mw(func(c echo.Context) error {
		resolved, _ = c.Get("deviceId").(string)
		return nil
	})(c))

	s.True(isUUIDv4(resolved))
	s.NotEqual("not-a-uuid", resolved)
}

func (s *DeviceIDSuite) TestEnsureDeviceID_SecureOutsideLocal() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)

	s.Require().NoError(chi_ratelimit.EnsureDeviceID("production")(func(c echo.Context) error {
		return nil
	})(c))

	var cookie *http.Cookie
	for _, ck := range rec.Result().Cookies() {
		if ck.Name == chi_ratelimit.DeviceIDCookieName {
			cookie = ck
			break
		}
	}
	s.Require().NotNil(cookie)
	s.True(cookie.Secure)
}

func isUUIDv4(id string) bool {
	parsed, err := uuid.Parse(id)
	return err == nil && parsed.Version() == 4
}
