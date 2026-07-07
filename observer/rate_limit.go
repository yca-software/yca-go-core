package chi_observer

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// Principal types recorded on rate limit hits (for blocking / investigation).
const (
	PrincipalTypeIP      = "ip"
	PrincipalTypeUser    = "user"
	PrincipalTypeAPIKey  = "api_key"
	PrincipalTypeUnknown = "unknown"
)

// rateLimitMetricsKey salts hashes so identifiers are not recoverable from /metrics alone.
var rateLimitMetricsKey = []byte("yca-2chi-rate-limit-metrics-v1")

// HashRateLimitIdentifier returns a stable, non-reversible label for Prometheus.
// Raw user IDs, API key IDs, and IPs must not appear on the metrics endpoint.
func HashRateLimitIdentifier(value string) string {
	if value == "" || value == "unknown" {
		return "unknown"
	}
	mac := hmac.New(sha256.New, rateLimitMetricsKey)
	_, _ = mac.Write([]byte(value))
	return hex.EncodeToString(mac.Sum(nil)[:8])
}

// RecordRateLimitHit increments the rate limit hit counter.
// principal and ip are hashed before export; pass raw values from the request context.
func (o *Observer) RecordRateLimitHit(method, route, principalType, principal, ip string) {
	if o == nil || o.rateLimitHitsTotal == nil {
		return
	}
	if method == "" {
		method = "unknown"
	}
	if route == "" {
		route = "unmatched"
	}
	if principalType == "" {
		principalType = PrincipalTypeUnknown
	}

	o.rateLimitHitsTotal.WithLabelValues(
		method,
		route,
		principalType,
		HashRateLimitIdentifier(principal),
		HashRateLimitIdentifier(ip),
	).Inc()
}
