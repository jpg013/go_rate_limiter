package ratelimiter

import (
	"time"
)

type RateLimiter interface {
	Acquire() (*Token, error)
	Release(*Token) error
}

type RateLimiterType uint32

const (
	ThrottleType RateLimiterType = iota + 1
	MaxConcurrencyType
	IntervalWindowType
)

func New(conf *Config) (RateLimiter, error) {
	m, err := NewManager()

	if err != nil {
		return m, err
	}

	switch conf.Type {
	case ThrottleType:
		err = withThrottle(m, conf)
	default:
		err = ErrInvalidLimiterType
	}

	if err != nil {
		return nil, err
	}

	return m, err
}

type Config struct {
	Type            RateLimiterType
	Throttle        time.Duration
	ConnectionLimit int
}

// func NewMaxConcurrencyRateLimiter(conf *Config) (RateLimiter, error) {
// 	m, err := NewManager()

// 	if err != nil {
// 		return m, err
// 	}

// 	return withMaxConcurrency(m, conf)
// }

// func withMaxConcurrency(m *Manager, conf *Config) (*Manager, error) {
// 	if m.rateLimiterType != UndefinedRateLimiterType {
// 		return m, ErrAleadyInitialized
// 	}

// 	m.rateLimiterType = MaxConcurrencyType
// 	m.limit = conf.limit

// 	go func() {
// 		var wg sync.WaitGroup

// 		// Listen on the in channel
// 		for {
// 			select {
// 			case <-inChan:
// 				wg.Add(1)
// 				// Try to get a remaining rateLimit
// 				err := tryGetRemaining(m)

// 				if err == nil {
// 					t, err := NewType()

// 					if err != nil {
// 						m.errorChan <- err
// 					} else {
// 						outChan <- t
// 					}
// 				} else {

// 				}
// 			}
// 		}

// 		defer func() {
// 			wg.Wait()
// 			fmt.Println("close out channel")
// 		}()
// 	}()

// 	return m, nil
// }

// func requestToken(m *Manager) {
// 	m.lock.Lock()
// 	defer m.lock.Unlock()

// 	// If the number of rate limits in use are less than limit, create a new one
// 	if len(m.tokensInUse) >= m.limit {
// 		m.needToken++
// 		return
// 	}

// 	t, err := NewType()

// 	if err != nil {
// 		m.errorChan <- err
// 	} else {
// 		m.tokensInUse[t.Token] = t
// 		m.outChan <- t.Token
// 	}
// }

func withThrottle(m *Manager, conf *Config) error {
	if m.await != nil {
		return ErrRateLimiterInitialized
	}

	throttler := make(chan struct{})

	go func() {
		ticker := time.NewTicker(conf.Throttle)

		for ; true; <-ticker.C {
			throttler <- struct{}{}
		}
	}()

	// define the await function
	m.await = func() (<-chan *Token, <-chan error) {
		go func() {
			<-throttler
			t, err := NewToken()

			if err != nil {
				m.errorChan <- err
			} else {
				m.outChan <- t
			}
		}()

		return m.outChan, m.errorChan
	}

	return nil
}
