package yca_observer

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	yca_error "github.com/yca-software/yca-go-core/error"
)

type HTTPMiddlewareProvider interface {
	EchoMiddleware(skipper func(echo.Context) bool) echo.MiddlewareFunc
}

// EchoMiddleware records HTTP metrics.
func (o *Observer) EchoMiddleware(skipper func(echo.Context) bool) echo.MiddlewareFunc {
	if skipper == nil {
		skipper = func(c echo.Context) bool {
			return c.Path() == "/metrics"
		}
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if skipper(c) {
				return next(c)
			}

			start := time.Now()
			err := next(c) // Process request

			status := c.Response().Status
			// If an error occurred and response wasn't committed,
			// the status might still be 200. We should check that.
			if err != nil && !c.Response().Committed {
				status = http.StatusInternalServerError
				if apiErr, ok := yca_error.AsError(err); ok {
					status = apiErr.StatusCode
				} else if he, ok := err.(*echo.HTTPError); ok {
					status = he.Code
				}
			}

			route := c.Path()
			if route == "" {
				route = "unmatched"
			}

			method := c.Request().Method
			statusStr := strconv.Itoa(status)

			o.httpRequestsTotal.WithLabelValues(method, route, statusStr).Inc()
			o.httpRequestDuration.WithLabelValues(method, route, statusStr).Observe(time.Since(start).Seconds())

			return err
		}
	}
}
