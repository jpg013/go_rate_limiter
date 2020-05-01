package ratelimiter

import (
	"time"
)

type RateLimiter interface {
	Acquire() (*Token, error)
	Release(*Token)
}

type Config struct {
	Limit    int
	Interval time.Duration
	Throttle time.Duration
}

func acquire(s *State) (*Token, error) {
	go func() {
		s.inChan <- struct{}{}
	}()

	// Await rate limit token
	select {
	case t := <-s.outChan:
		return t, nil
	case err := <-s.errorChan:
		return nil, err
	}
}
