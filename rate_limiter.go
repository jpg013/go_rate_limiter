package ratelimiter

import (
	"time"
)

// RateLimiter defines two methods for acquiring and releasing tokens
type RateLimiter interface {
	Acquire() (*Token, error)
	Release(*Token)
}

// Config represents a rate limiter config object
type Config struct {
	// Limit determines how many rate limit tokens can be active at a time
	Limit int

	// FixedInterval sets the fixed time window for a Fixed Window Rate Limiter
	FixedInterval time.Duration

	// Throttle is the min time between requests for a Throttle Rate Limiter
	Throttle time.Duration

	// TokenResetsAfter is the maximum amount of time a token can live before being
	// forcefully released - if set to zero time then the token may live forever
	TokenResetsAfter time.Duration
}
