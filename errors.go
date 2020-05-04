package ratelimiter

import "errors"

// Errors used throughout the codebase
var (
	ErrInvalidLimit           = errors.New("Limit must be greater than zero")
	ErrInvalidInterval        = errors.New("Interval must be greater than zero")
	ErrTokenFactoryNotDefined = errors.New("Token factory must be defined")
)
