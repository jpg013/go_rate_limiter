package ratelimiter

import "errors"

// Errors used throughout the codebase
var (
	ErrInvalidLimiterType     = errors.New("Invalid Limiter Type")
	ErrNotInitialized         = errors.New("Rate Limiter is not  initialized")
	ErrRateLimiterInitialized = errors.New("rate limiter already initialized")
)
