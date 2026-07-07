package chi_server

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"golang.org/x/sync/errgroup"
)

type ReadinessDependency interface {
	Check(ctx context.Context) error
}

func RegisterHealthHandlers(
	server *echo.Echo,
	readinessDependencies []ReadinessDependency,
) {
	// Health godoc
	// @Summary      Liveness check
	// @Description  Returns 200 if the application is running. Used by orchestrators to determine if the process is alive.
	// @Tags         health
	// @Produce      json
	// @Success      200  {object}  map[string]string
	// @Router       /health [get]
	server.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status": "ok",
		})
	})

	// Ready godoc
	// @Summary      Readiness check
	// @Description  Returns 200 if the application can serve traffic (e.g. database is reachable). Used by orchestrators to determine if the instance should receive traffic.
	// @Tags         health
	// @Produce      json
	// @Success      200   {object}  map[string]string
	// @Failure      503   {object}  map[string]string
	// @Router       /ready [get]
	server.GET("/ready", func(c echo.Context) error {
		if err := checkDependenciesConcurrently(
			c.Request().Context(),
			readinessDependencies,
		); err != nil {
			return c.JSON(http.StatusServiceUnavailable, map[string]string{
				"status": "unavailable",
			})
		}

		return c.JSON(http.StatusOK, map[string]string{
			"status": "ok",
		})
	})
}

// checkDependenciesConcurrently runs all dependency checks in parallel.
//
// The first dependency that fails cancels the remaining checks.
// Returns nil only if all dependencies succeed.
func checkDependenciesConcurrently(
	ctx context.Context,
	dependencies []ReadinessDependency,
) error {
	if len(dependencies) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)

	for _, dependency := range dependencies {
		dep := dependency

		g.Go(func() error {
			return dep.Check(ctx)
		})
	}

	return g.Wait()
}
