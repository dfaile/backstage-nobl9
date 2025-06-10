package nobl9

import (
	"context"
	"time"

	"golang.org/x/time/rate"
)

// SimpleRateLimiter implements the RateLimiter interface using golang.org/x/time/rate
type SimpleRateLimiter struct {
	limiter *rate.Limiter
}

// NewSimpleRateLimiter creates a new SimpleRateLimiter
func NewSimpleRateLimiter(rps int, period time.Duration) *SimpleRateLimiter {
	return &SimpleRateLimiter{
		limiter: rate.NewLimiter(rate.Limit(float64(rps)/period.Seconds()), rps),
	}
}

// Wait waits for rate limiting before proceeding
func (r *SimpleRateLimiter) Wait(ctx context.Context) error {
	return r.limiter.Wait(ctx)
}

// Success records a successful API call
func (r *SimpleRateLimiter) Success() {
	// No action needed for simple rate limiter
}

// Failure records a failed API call
func (r *SimpleRateLimiter) Failure() {
	// No action needed for simple rate limiter
} 