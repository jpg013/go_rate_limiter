package ratelimiter

import "errors"

// Errors used throughout the codebase
var (
	ErrLimitGreaterThanZero    = errors.New("Limit must be greater than zero")
	ErrIntervalGreaterThanZero = errors.New("Interval must be greater than zero")
)
