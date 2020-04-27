package ratelimiter

import "github.com/segmentio/ksuid"

type Token struct {
	ID string
}

func NewToken() (*Token, error) {
	return &Token{
		ID: ksuid.New().String(),
	}, nil
}
