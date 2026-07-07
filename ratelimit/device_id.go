package chi_ratelimit

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

const (
	// DeviceIDCookieName is the first-party device identifier cookie.
	DeviceIDCookieName = "2chi_device_id"
	deviceIDHeaderName = "X-Device-Id"
	deviceIDContextKey = "deviceId"
	deviceIDMaxAge     = 365 * 24 * 60 * 60 // ~1 year
	deviceIDUnknown    = "unknown"
)

// EnsureDeviceID issues a first-party device cookie and stores the ID on the request context.
func EnsureDeviceID(env string) echo.MiddlewareFunc {
	secure := env != "local"

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			id := resolveDeviceID(c)
			if id == deviceIDUnknown {
				id = uuid.NewString()
				setDeviceCookie(c, id, secure)
			}
			c.Set(deviceIDContextKey, id)
			return next(c)
		}
	}
}

func resolveDeviceID(c echo.Context) string {
	if id, ok := c.Get(deviceIDContextKey).(string); ok && isValidDeviceUUID(id) {
		return id
	}
	if cookie, err := c.Cookie(DeviceIDCookieName); err == nil && isValidDeviceUUID(cookie.Value) {
		return cookie.Value
	}
	if header := c.Request().Header.Get(deviceIDHeaderName); isValidDeviceUUID(header) {
		return header
	}
	return deviceIDUnknown
}

func isValidDeviceUUID(id string) bool {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return false
	}
	return parsed.Version() == 4
}

func setDeviceCookie(c echo.Context, id string, secure bool) {
	c.SetCookie(&http.Cookie{
		Name:     DeviceIDCookieName,
		Value:    id,
		Path:     "/",
		MaxAge:   deviceIDMaxAge,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}
