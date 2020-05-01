package ratelimiter

import (
	"time"

	"github.com/segmentio/ksuid"
)

type Token struct {
	ID        string
	CreatedAt time.Time
	// Minimum amount of time the rate limit is alive before being released
	ExpiresAt time.Time
}

func NewToken(expiresAt time.Time) (*Token, error) {
	return &Token{
		ID:        ksuid.New().String(),
		CreatedAt: time.Now().UTC(),
		ExpiresAt: expiresAt,
	}, nil
}

func (t *Token) IsExpired() bool {
	now := time.Now().UTC()

	return t.ExpiresAt.Before(now)
}
