package chi_server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// ObservabilityConfig configures a minimal HTTP server for health and Prometheus metrics.
type ObservabilityConfig struct {
	Port                  int
	ReadinessDependencies []ReadinessDependency
}

// NewObservabilityServer returns a server exposing /health, /ready, and /metrics only.
func NewObservabilityServer(cfg ObservabilityConfig) Server {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Use(middleware.Recover())

	RegisterHealthHandlers(e, cfg.ReadinessDependencies)
	RegisterMetricsHandlers(e)

	httpServer := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Port),
		Handler:           e,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	return &serverImpl{
		echo:   e,
		server: httpServer,
	}
}

// StartObservabilityServer listens until ctx is cancelled, then shuts down gracefully.
func StartObservabilityServer(ctx context.Context, cfg ObservabilityConfig) error {
	srv := NewObservabilityServer(cfg)

	errCh := make(chan error, 1)
	go func() {
		if err := srv.Start(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	case err := <-errCh:
		return err
	}
}
