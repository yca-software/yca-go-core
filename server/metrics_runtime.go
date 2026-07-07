package chi_server

import "context"

// StartDedicatedMetricsServer runs /health, /ready, and /metrics on metricsPort.
func StartDedicatedMetricsServer(ctx context.Context, metricsPort int, readiness []ReadinessDependency) {
	if metricsPort <= 0 {
		return
	}
	go func() {
		_ = StartObservabilityServer(ctx, ObservabilityConfig{
			Port:                  metricsPort,
			ReadinessDependencies: readiness,
		})
	}()
}
