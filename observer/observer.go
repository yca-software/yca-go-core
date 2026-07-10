package yca_observer

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

type ObserverConfig struct {
	Namespace string
	AppName   string
}

type Observer struct {
	httpRequestsTotal         *prometheus.CounterVec
	httpRequestDuration       *prometheus.HistogramVec
	dbQueryDuration           *prometheus.HistogramVec
	rateLimitHitsTotal        *prometheus.CounterVec
	jobConsumerPublishedTotal *prometheus.CounterVec
	jobConsumerOutcomeTotal   *prometheus.CounterVec
	jobConsumerDuration       *prometheus.HistogramVec
}

func New(cfg ObserverConfig) (*Observer, error) {
	var o Observer

	o.httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace:   cfg.Namespace,
			Subsystem:   "http",
			Name:        "requests_total",
			Help:        "Total HTTP requests processed.",
			ConstLabels: prometheus.Labels{"app": cfg.AppName},
		},
		[]string{"method", "route", "status"},
	)

	o.httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace:   cfg.Namespace,
			Subsystem:   "http",
			Name:        "request_duration_seconds",
			Help:        "Duration of HTTP requests.",
			ConstLabels: prometheus.Labels{"app": cfg.AppName},
			Buckets:     prometheus.DefBuckets,
		},
		[]string{"method", "route", "status"},
	)

	o.dbQueryDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace:   cfg.Namespace,
			Subsystem:   "db",
			Name:        "query_duration_seconds",
			Help:        "Duration of database queries in seconds.",
			ConstLabels: prometheus.Labels{"app": cfg.AppName},
			Buckets:     []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
		[]string{"operation", "table", "query_type", "status"},
	)

	o.rateLimitHitsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace:   cfg.Namespace,
			Subsystem:   "rate_limit",
			Name:        "hits_total",
			Help:        "Total number of rate limit hits.",
			ConstLabels: prometheus.Labels{"app": cfg.AppName},
		},
		[]string{"method", "route", "principal_type", "principal_hash", "ip_hash"},
	)

	o.jobConsumerPublishedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace:   cfg.Namespace,
			Subsystem:   "job_consumer",
			Name:        "published_total",
			Help:        "Total job triggers published to SQS.",
			ConstLabels: prometheus.Labels{"app": cfg.AppName},
		},
		[]string{"job"},
	)

	o.jobConsumerOutcomeTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace:   cfg.Namespace,
			Subsystem:   "job_consumer",
			Name:        "outcome_total",
			Help:        "Total job consumer outcomes.",
			ConstLabels: prometheus.Labels{"app": cfg.AppName},
		},
		[]string{"job", "outcome"},
	)

	o.jobConsumerDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace:   cfg.Namespace,
			Subsystem:   "job_consumer",
			Name:        "duration_seconds",
			Help:        "Duration of job consumer handler execution.",
			ConstLabels: prometheus.Labels{"app": cfg.AppName},
			Buckets:     []float64{.01, .05, .1, .25, .5, 1, 2.5, 5, 10, 30, 60, 120, 300},
		},
		[]string{"job"},
	)

	if err := RegisterCollectors(
		o.httpRequestsTotal,
		o.httpRequestDuration,
		o.dbQueryDuration,
		o.rateLimitHitsTotal,
		o.jobConsumerPublishedTotal,
		o.jobConsumerOutcomeTotal,
		o.jobConsumerDuration,
	); err != nil {
		return nil, err
	}

	return &o, nil
}

// RegisterCollectors registers Prometheus collectors, replacing any already-registered instance.
func RegisterCollectors(cs ...prometheus.Collector) error {
	for _, c := range cs {
		if c == nil {
			continue
		}
		if err := prometheus.Register(c); err != nil {
			if already, ok := err.(prometheus.AlreadyRegisteredError); ok {
				if !prometheus.Unregister(already.ExistingCollector) {
					return fmt.Errorf("failed to unregister existing collector")
				}
				if err := prometheus.Register(c); err != nil {
					return fmt.Errorf("re-registering collector after unregister: %w", err)
				}
				continue
			}
			return fmt.Errorf("registering collector: %w", err)
		}
	}
	return nil
}
