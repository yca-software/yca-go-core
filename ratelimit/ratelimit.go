package chi_ratelimit

import (
	"fmt"

	"github.com/labstack/echo/v4"
	redisclient "github.com/redis/go-redis/v9"
	"github.com/ulule/limiter/v3"
	redisstore "github.com/ulule/limiter/v3/drivers/store/redis"
	chi_error "github.com/yca-software/yca-go-core/error"
	chi_logger "github.com/yca-software/yca-go-core/logger"
	chi_observer "github.com/yca-software/yca-go-core/observer"
	chi_types "github.com/yca-software/yca-go-core/types"
)

// RateLimiter is a factory that generates Echo rate-limiting middlewares.
type RateLimiter struct {
	redisClient *redisclient.Client
	observer    *chi_observer.Observer
	logger      chi_logger.Logger
}

// NewRateLimiter initializes the factory.
func NewRateLimiter(rdb *redisclient.Client, obs *chi_observer.Observer, logger chi_logger.Logger) *RateLimiter {
	return &RateLimiter{
		redisClient: rdb,
		observer:    obs,
		logger:      logger,
	}
}

func (rl *RateLimiter) build(rate string, keyExtractor func(echo.Context) string) echo.MiddlewareFunc {
	if rl.redisClient == nil {
		return func(next echo.HandlerFunc) echo.HandlerFunc {
			return next
		}
	}

	store, err := redisstore.NewStoreWithOptions(rl.redisClient, limiter.StoreOptions{
		Prefix: "limiter",
	})
	if err != nil {
		panic(fmt.Sprintf("failed to create Redis store for rate limiting: %v", err))
	}

	parsedRate, err := limiter.NewRateFromFormatted(rate)
	if err != nil {
		panic(fmt.Sprintf("failed to parse rate limit format '%s': %v", rate, err))
	}

	instance := limiter.New(store, parsedRate)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			key := keyExtractor(c)
			ctx := c.Request().Context()

			limiterCtx, err := instance.Get(ctx, key)
			if err != nil {
				return chi_error.NewTooManyRequestsError(err, "TooManyRequests", nil)
			}

			c.Response().Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", limiterCtx.Limit))
			c.Response().Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", limiterCtx.Remaining))
			c.Response().Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", limiterCtx.Reset))

			if limiterCtx.Reached {
				principalType, principal := rateLimitPrincipal(c)
				ip := clientIP(c)

				if rl.observer != nil {
					rl.observer.RecordRateLimitHit(c.Request().Method, c.Path(), principalType, principal, ip)
				}
				if rl.logger != nil {
					rl.logRateLimitHit(c, principalType, principal, ip)
				}
				return chi_error.NewTooManyRequestsError(nil, "TooManyRequests", nil)
			}

			return next(c)
		}
	}
}

func (rl *RateLimiter) logRateLimitHit(c echo.Context, principalType, principal, ip string) {
	args := []any{
		"method", c.Request().Method,
		"route", c.Path(),
		"principal_type", principalType,
		"client_ip", ip,
	}
	switch principalType {
	case chi_observer.PrincipalTypeUser:
		args = append(args, "user_id", principal)
	case chi_observer.PrincipalTypeAPIKey:
		args = append(args, "api_key_id", principal)
	}
	if v := c.Get("accessInfo"); v != nil {
		if ai, ok := v.(*chi_types.AccessInfo); ok && ai != nil && ai.RequestID != "" {
			args = append(args, "request_id", ai.RequestID)
		}
	}
	rl.logger.WithContext(c.Request().Context()).Warn("rate limit exceeded", args...)
}

// -----------------------------------------------------------------------------
// Public Middleware Generators
// -----------------------------------------------------------------------------

// IPRateLimit limits based strictly on the requester's IP address.
func (rl *RateLimiter) IPRateLimit(rate string) echo.MiddlewareFunc {
	return rl.build(rate, realIPKey)
}

// IPDeviceRateLimit limits by IP and device ID (cookie or X-Device-Id header).
func (rl *RateLimiter) IPDeviceRateLimit(rate string) echo.MiddlewareFunc {
	return rl.build(rate, ipDeviceKey)
}

// PrincipalRateLimit limits based on the authenticated User/API Key, falling back to IP.
func (rl *RateLimiter) PrincipalRateLimit(rate string) echo.MiddlewareFunc {
	return rl.build(rate, authenticatedPrincipalKey)
}

// ScopedPrincipalRateLimit is used for specific sensitive actions (like sending emails).
func (rl *RateLimiter) ScopedPrincipalRateLimit(rate string, scope string) echo.MiddlewareFunc {
	return rl.build(rate, func(c echo.Context) string {
		return scope + ":" + authenticatedPrincipalKey(c)
	})
}
