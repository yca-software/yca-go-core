package yca_ratelimit

import (
	"github.com/labstack/echo/v4"
	yca_observer "github.com/yca-software/yca-go-core/observer"
	yca_types "github.com/yca-software/yca-go-core/types"
)

// -----------------------------------------------------------------------------
// Key Extractors
// -----------------------------------------------------------------------------

func realIPKey(c echo.Context) string {
	return "ip:" + clientIP(c)
}

func ipDeviceKey(c echo.Context) string {
	return "ip:" + clientIP(c) + ":dev:" + resolveDeviceID(c)
}

func clientIP(c echo.Context) string {
	ip := c.RealIP()
	if ip == "" {
		ip = c.Request().RemoteAddr
	}
	return ip
}

func rateLimitPrincipal(c echo.Context) (principalType, principal string) {
	v := c.Get("accessInfo")
	if v == nil {
		return yca_observer.PrincipalTypeIP, clientIP(c)
	}

	ai, ok := v.(*yca_types.AccessInfo)
	if !ok || ai == nil {
		return yca_observer.PrincipalTypeIP, clientIP(c)
	}

	switch ai.Type {
	case yca_types.AccessTypeUser:
		return yca_observer.PrincipalTypeUser, ai.SubjectID.String()
	case yca_types.AccessTypeAPIKey:
		return yca_observer.PrincipalTypeAPIKey, ai.SubjectID.String()
	default:
		return yca_observer.PrincipalTypeIP, clientIP(c)
	}
}

func authenticatedPrincipalKey(c echo.Context) string {
	v := c.Get("accessInfo")
	if v == nil {
		return realIPKey(c)
	}

	ai, ok := v.(*yca_types.AccessInfo)
	if !ok || ai == nil {
		return realIPKey(c)
	}

	if ai.Type == yca_types.AccessTypeUser {
		return "u:" + ai.SubjectID.String()
	}
	if ai.Type == yca_types.AccessTypeAPIKey {
		return "k:" + ai.SubjectID.String()
	}

	return realIPKey(c)
}
