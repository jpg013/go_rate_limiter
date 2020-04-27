package ratelimiter

import "errors"

// Errors used throughout the codebase
var (
	ErrNotInitialized    = errors.New("Rate Limiter already initialized")
	ErrAleadyInitialized = errors.New("rate limiter already initialized")
)
