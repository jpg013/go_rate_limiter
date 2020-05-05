package ratelimiter

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

	// override the manager makeToken function
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
