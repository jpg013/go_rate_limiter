package ratelimiter

import (
	"errors"
	"time"
)

// NewThrottleRateLimiter returns a throttle rate limiter
func NewThrottleRateLimiter(conf *Config) (RateLimiter, error) {
	if conf.Throttle == 0 {
		return nil, errors.New("Throttle duration must be greater than zero")
	}

	m := NewManager(conf)

	// Throttle Await Function
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
