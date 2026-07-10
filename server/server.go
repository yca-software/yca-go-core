package yca_server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	yca_error "github.com/yca-software/yca-go-core/error"
	yca_logger "github.com/yca-software/yca-go-core/logger"
	yca_observer "github.com/yca-software/yca-go-core/observer"
)

type ServerConfig struct {
	Port               int
	CORSAllowOrigins   []string
	BodyLimit          string
	ServerReadTimeout  int
	ServerWriteTimeout int
	ServerIdleTimeout  int
	Logger             yca_logger.Logger
	Observer           yca_observer.HTTPMiddlewareProvider
	RegisterRoutes     func(e *echo.Echo)
}

type Server interface {
	Start() error
	Shutdown(ctx context.Context) error
	GetEcho() *echo.Echo
	GetServer() *http.Server
}

type serverImpl struct {
	echo   *echo.Echo
	server *http.Server
}

func New(cfg ServerConfig) Server {
	e := echo.New()

	e.HTTPErrorHandler = func(err error, c echo.Context) {
		if apiErr, ok := yca_error.AsError(err); ok {
			c.JSON(apiErr.StatusCode, apiErr)
			return
		}

		if he, ok := err.(*echo.HTTPError); ok {
			c.JSON(he.Code, map[string]any{
				"errorCode": errorCodeFromHTTPStatus(he.Code),
			})
			return
		}

		apiErr := yca_error.NewInternalServerError(err, "", nil)
		c.JSON(apiErr.StatusCode, apiErr)
	}

	if cfg.Observer != nil {
		e.Use(cfg.Observer.EchoMiddleware(nil))
	}

	requestLogger := cfg.Logger
	e.Use(
		middleware.Recover(),
		middleware.RequestID(),
		middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
			LogLatency:   true,
			LogRemoteIP:  true,
			LogMethod:    true,
			LogURIPath:   true,
			LogRequestID: true,
			LogStatus:    true,
			LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
				if requestLogger != nil {
					requestLogger.WithContext(c.Request().Context()).Info("request",
						"time", v.StartTime.Format(time.RFC3339),
						"request_id", v.RequestID,
						"status", v.Status,
						"latency", v.Latency.String(),
						"method", v.Method,
						"path", v.URIPath,
						"remote_ip", v.RemoteIP,
					)
					return nil
				}
				_, err := fmt.Fprintf(os.Stderr, "[%s] %s %d %s %s %s %s\n",
					v.StartTime.Format(time.RFC3339),
					v.RequestID,
					v.Status,
					v.Latency,
					v.Method,
					v.URIPath,
					v.RemoteIP,
				)
				return err
			},
		}),
		middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins:     cfg.CORSAllowOrigins,
			AllowCredentials: true,
			AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions, http.MethodHead},
			AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization, "X-API-Key", "X-Device-Id"},
		}),
		middleware.BodyLimit(cfg.BodyLimit),
		middleware.SecureWithConfig(middleware.SecureConfig{
			XSSProtection:         "1; mode=block",
			ContentTypeNosniff:    "nosniff",
			XFrameOptions:         "DENY",
			HSTSMaxAge:            31536000,
			ContentSecurityPolicy: "default-src 'self'",
		}),
	)

	if cfg.RegisterRoutes != nil {
		cfg.RegisterRoutes(e)
	}

	httpServer := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Port),
		Handler:           e,
		ReadHeaderTimeout: time.Duration(cfg.ServerReadTimeout) * time.Second,
		ReadTimeout:       time.Duration(cfg.ServerReadTimeout) * time.Second,
		WriteTimeout:      time.Duration(cfg.ServerWriteTimeout) * time.Second,
		IdleTimeout:       time.Duration(cfg.ServerIdleTimeout) * time.Second,
	}

	return &serverImpl{
		echo:   e,
		server: httpServer,
	}
}

func errorCodeFromHTTPStatus(status int) string {
	switch status {
	case 400, 422:
		return "InvalidRequestBody"
	case 401:
		return "Unauthorized"
	case 403:
		return "Forbidden"
	case 404:
		return "NotFound"
	case 409:
		return "Conflict"
	case 429:
		return "TooManyRequests"
	default:
		if status >= 500 {
			return "InternalServerError"
		}
		return "Unknown"
	}
}

func (s *serverImpl) Start() error {
	return s.server.ListenAndServe()
}

func (s *serverImpl) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func (s *serverImpl) GetEcho() *echo.Echo {
	return s.echo
}

func (s *serverImpl) GetServer() *http.Server {
	return s.server
}
