package chi_server

import (
	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func RegisterMetricsHandlers(server *echo.Echo) {
	server.GET("/metrics", echo.WrapHandler(promhttp.Handler()))
}
