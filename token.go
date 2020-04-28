package ratelimiter

import (
	"time"

	"github.com/segmentio/ksuid"
)

type Token struct {
	ID        string
	ExpiresAt time.Time
	CreatedAt time.Time
}

func NewToken(resetsIn time.Duration) (*Token, error) {
	return &Token{
		ID:        ksuid.New().String(),
		CreatedAt: time.Now().UTC(),
		ExpiresAt: time.Now().UTC().Add(resetsIn),
	}, nil
}
