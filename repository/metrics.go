package chi_repository

import (
	"strings"
	"time"

	chi_observer "github.com/yca-software/yca-go-core/observer"
)

// recordMetrics records query metrics if a hook is provided.
func recordMetrics(hook chi_observer.QueryMetricsHook, tableName, operation string, start time.Time, err error) {
	if hook == nil {
		return
	}

	duration := time.Since(start)
	status := "success"
	if err != nil {
		status = "error"
	}

	table := strings.ToLower(tableName)
	queryType := strings.ToLower(operation) + "_" + table
	hook.RecordQuery(operation, table, queryType, status, duration)
}
