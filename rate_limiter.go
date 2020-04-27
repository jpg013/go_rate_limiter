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
	UndefinedRateLimiterType RateLimiterType = iota
	ThrottleType
	MaxConcurrencyType
	IntervalWindowType
)

func New(conf *Config) (RateLimiter, error) {
	m, err := NewManager()

	if err != nil {
		return m, err
	}

	err = withThrottle(m, conf)

	if err != nil {
		return nil, err
	}

	return m, err
}

type Config struct {
	Throttle time.Duration
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
	outChan := make(chan *Token)
	errorChan := make(chan error)
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
				errorChan <- err
			} else {
				outChan <- t
			}
		}()

		return outChan, errorChan
	}

	return nil
}

// func main() {
// 	rateLimiter, err := NewRateLimiter(&Config{
// 		throttle: 3 * time.Second,
// 	})

// 	if err != nil {
// 		panic(err)
// 	}

// 	for i := 0; i < 10; i++ {
// 		go func() {
// 			token, err := rateLimiter.Acquire()

// 			if err != nil {
// 				panic(err)
// 			}

// 			fmt.Println(token.id)
// 		}()
// 	}

// 	time.Sleep(1 * time.Minute)
// }
