package ratelimiter

import (
	"log"
	"sync/atomic"
	"time"
)

type RateLimiter interface {
	Acquire() (*Token, error)
	Release(*Token) error
}

type RateLimiterType uint32

const (
	ThrottleType RateLimiterType = iota + 1
	ConcurrencyType
	IntervalType
)

func New(conf *Config) (RateLimiter, error) {
	m, err := NewManager()

	if err != nil {
		return m, err
	}

	switch conf.Type {
	case ThrottleType:
		err = withThrottle(m, conf)
	case ConcurrencyType:
		err = withMaxConcurrency(m, conf)
	default:
		err = ErrInvalidLimiterType
	}

	if err != nil {
		return nil, err
	}

	return m, err
}

type Config struct {
	Type           RateLimiterType
	Throttle       time.Duration
	MaxConcurrency int
}

func withThrottle(m *Manager, conf *Config) error {
	go func() {
		ticker := time.NewTicker(conf.Throttle)

		for ; true; <-ticker.C {
			<-m.inChan
			// Generate Token
			m.generateToken()
		}
	}()

	return nil
}

func withMaxConcurrency(m *Manager, conf *Config) error {
	limit := conf.MaxConcurrency

	go func() {
		for {
			select {
			case <-m.inChan:
				if atomic.LoadInt64(&m.awaiting) > 0 {
					atomic.AddInt64(&m.awaiting, 1)
					continue
				}
				if len(m.activeTokens) >= limit {
					atomic.AddInt64(&m.awaiting, 1)
					continue
				}
				m.generateToken()
			case t := <-m.releaseChan:
				if t == nil {
					log.Print("unable to relase nil token")
					continue
				}
				if _, ok := m.activeTokens[t.ID]; !ok {
					log.Printf("unable to relase token %s - not in use", t)
					continue
				}
				// Delete from map
				delete(m.activeTokens, t.ID)

				// Is anything waiting for a rate limit?
				if atomic.LoadInt64(&m.awaiting) > 0 {
					atomic.AddInt64(&m.awaiting, -1)
					m.generateToken()
				}
			}
		}
	}()

	return nil
}

// attempts to acquire a new rate limit resource and sent it to the
// token channel. If no resources are available, or there is an error
// then we increment the needResource counter.
// func tryAcquireResource(m *Manager) {
// 	var r *Resource
// 	var err error

// 	defer func() {
// 		if err != nil {
// 			return
// 		}

// 		// If the acquired resource is not in the resource map, then add it
// 		if _, ok := m.resources[r.Token]; !ok {
// 			m.resources[r.Token] = r
// 		}

// 		// lock the resource
// 		r.lock()

// 		// send the resource to the "waiting" channel
// 		m.waitResourceChan <- r
// 	}()

// 	// Try to get an existing available resource
// 	r = m.getAvailableResource()

// 	// If found, exit
// 	if r != nil {
// 		return
// 	}

// 	// Check to see if all resources are in use
// 	if len(m.resources) > m.limit {
// 		err = errors.New("all resources in use")
// 		return
// 	}

// 	// We were unable to find an available resource, try to create a new one.
// 	// Lock to prevent deadlock at the database level
// 	r, err = NewResource(m)

// 	if err != nil {
// 		return
// 	}

// 	// All resources in use, increment the needResource counter
// 	atomic.AddInt64(&m.needResource, 1)
// }
