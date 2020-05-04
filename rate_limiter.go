package ratelimiter

import (
	"errors"
	"time"
)

// RateLimiter defines two methods for acquiring
// and releasing tokens
type RateLimiter interface {
	Acquire() (*Token, error)
	Release(*Token)
}

// Config represents a rate limiter config object
type Config struct {
	Limit           int
	FixedInterval   time.Duration
	Throttle        time.Duration
	TokenResetAfter time.Duration
}

// NewMaxConcurrencyRateLimiter returns a max concurrency rate limiter
func NewMaxConcurrencyRateLimiter(conf *Config) (RateLimiter, error) {
	if conf.Limit <= 0 {
		return nil, ErrInvalidLimit
	}

	m := NewManager(conf)
	await := func() {
		go func() {
			for {
				select {
				case <-m.inChan:
					m.tryGenerateToken()
				case t := <-m.releaseChan:
					m.releaseToken(t)
				}
			}
		}()
	}

	await()
	return m, nil
}

// NewThrottleRateLimiter returns a throttle rate limiter
func NewThrottleRateLimiter(conf *Config) (RateLimiter, error) {
	if conf.Throttle == 0 {
		return nil, errors.New("Throttle duration must be greater than zero")
	}

	m := NewManager(conf)
	await := func(throttle time.Duration) {
		ticker := time.NewTicker(throttle)
		go func() {
			<-m.inChan
			m.tryGenerateToken()
			for {
				select {
				case <-m.inChan:
					<-ticker.C
					m.tryGenerateToken()
				case t := <-m.releaseChan:
					m.releaseToken(t)
				}
			}
		}()
	}
	// Call await to start
	await(conf.Throttle)
	return m, nil
}

// NewFixedWindowRateLimiter returns a fixed window rate limiter
func NewFixedWindowRateLimiter(conf *Config) (RateLimiter, error) {
	if conf.FixedInterval == 0 {
		return nil, ErrInvalidInterval
	}

	if conf.Limit == 0 {
		return nil, ErrInvalidLimit
	}

	m := NewManager(conf)
	w := &FixedWindowInterval{interval: conf.FixedInterval}
	m.makeToken = func() *Token {
		t := NewToken()
		t.ExpiresAt = w.endTime
		return t
	}
	await := func() {
		go func() {
			for {
				select {
				case <-m.inChan:
					m.tryGenerateToken()
				case token := <-m.releaseChan:
					m.releaseToken(token)
				}
			}
		}()
	}

	w.run(m.releaseExpiredTokens)
	await()
	return m, nil
}
