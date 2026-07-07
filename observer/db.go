package chi_observer

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// QueryMetricsHook records database query metrics without tying the repository to Prometheus.
type QueryMetricsHook interface {
	RecordQuery(operation, table, queryType, status string, duration time.Duration)
}

// GetQueryMetricsHook returns a QueryMetricsHook for repository metrics.
func (o *Observer) GetQueryMetricsHook() QueryMetricsHook {
	return &queryMetricsHook{dbQueryDuration: o.dbQueryDuration}
}

// RecordQuery records a query observation on this observer.
func (o *Observer) RecordQuery(operation, table, queryType, status string, duration time.Duration) {
	o.GetQueryMetricsHook().RecordQuery(operation, table, queryType, status, duration)
}

type queryMetricsHook struct {
	dbQueryDuration *prometheus.HistogramVec
}

func (h *queryMetricsHook) RecordQuery(operation, table, queryType, status string, duration time.Duration) {
	if h.dbQueryDuration == nil {
		return
	}
	h.dbQueryDuration.WithLabelValues(operation, table, queryType, status).Observe(duration.Seconds())
}
