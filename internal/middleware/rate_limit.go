package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"multipass/pkg/apperror"
	"multipass/pkg/logging"
	"multipass/pkg/response"

	"github.com/go-redis/redis_rate/v10"
	"github.com/redis/go-redis/v9"
)

func RateLimitMiddleware(redisClient *redis.Client, limit int, window time.Duration, logger logging.Logger, responder response.Writer) func(http.Handler) http.Handler {
	// Create a new Redis rate limiter instance using redis_rate
	rateLimiter := redis_rate.NewLimiter(redisClient)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get the IP or user ID to rate limit
			ip := r.RemoteAddr // You can adjust this to use user-specific info, like a user ID.

			// Define the key for this IP address (you can adjust this to be user-specific too)
			key := fmt.Sprintf("rate_limit:%s", ip)

			// Create a redis_rate.Limit instance with the requested limit and window
			limitObj := redis_rate.PerSecond(limit) // Example: 5 requests per second

			// Attempt to acquire the token from the rate limiter
			res, err := rateLimiter.Allow(context.Background(), key, limitObj)
			if err != nil {
				logger.Error("Rate limit check failed", err)
				apperror.NewAppError(http.StatusInternalServerError, "server error", "rate_limit", err, logger, nil).WriteJSONError(w, r, responder)
				return
			}

			// If there is no remaining allowance, the rate limit has been exceeded
			if res.Allowed == 0 {
				apperror.NewAppError(http.StatusTooManyRequests, "Rate limit exceeded", "rate_limit", nil, logger, nil).WriteJSONError(w, r, responder)
				return
			}

			// Continue to the next handler if rate limit is within bounds
			next.ServeHTTP(w, r)
		})
	}
}
