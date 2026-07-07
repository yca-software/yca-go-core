package chi_observer

import (
	"time"
)

const (
	JobOutcomeSuccess          = "success"
	JobOutcomeDeadLetter       = "dead_letter"
	JobOutcomeRetryRepublished = "retry_republished"
)

// JobMetricsHook records background job publish and consumer metrics.
type JobMetricsHook interface {
	RecordJobPublished(job string)
	RecordJobConsumerOutcome(job, outcome string)
	RecordJobConsumerDuration(job string, duration time.Duration)
}

// RecordJobPublished increments the publish counter for a job queue.
func (o *Observer) RecordJobPublished(job string) {
	if o == nil || o.jobConsumerPublishedTotal == nil || job == "" {
		return
	}
	o.jobConsumerPublishedTotal.WithLabelValues(job).Inc()
}

// RecordJobConsumerOutcome increments the consumer outcome counter.
func (o *Observer) RecordJobConsumerOutcome(job, outcome string) {
	if o == nil || o.jobConsumerOutcomeTotal == nil || job == "" || outcome == "" {
		return
	}
	o.jobConsumerOutcomeTotal.WithLabelValues(job, outcome).Inc()
}

// RecordJobConsumerDuration observes job handler duration in seconds.
func (o *Observer) RecordJobConsumerDuration(job string, duration time.Duration) {
	if o == nil || o.jobConsumerDuration == nil || job == "" {
		return
	}
	o.jobConsumerDuration.WithLabelValues(job).Observe(duration.Seconds())
}

// GetJobMetricsHook returns a JobMetricsHook for job clients.
func (o *Observer) GetJobMetricsHook() JobMetricsHook {
	return o
}

type noopJobMetricsHook struct{}

func (noopJobMetricsHook) RecordJobPublished(string)                       {}
func (noopJobMetricsHook) RecordJobConsumerOutcome(string, string)         {}
func (noopJobMetricsHook) RecordJobConsumerDuration(string, time.Duration) {}

// NoopJobMetricsHook is a safe default when metrics are disabled.
var NoopJobMetricsHook JobMetricsHook = noopJobMetricsHook{}
