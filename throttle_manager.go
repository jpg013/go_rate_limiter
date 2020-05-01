package ratelimiter

import (
	"errors"
	"time"
)

type ThrottleManager struct {
	state    *State
	throttle time.Duration
}

func NewThrottleRateLimiter(conf *Config) (RateLimiter, error) {
	if conf.Throttle == 0 {
		return nil, errors.New("Throttle duration must be greater than zero")
	}

	s := &State{
		closeChan:      make(chan struct{}),
		errorChan:      make(chan error),
		outChan:        make(chan *Token),
		inChan:         make(chan struct{}),
		activeTokens:   make(map[string]*Token),
		releaseChan:    make(chan *Token),
		getTokenExpire: defaultExpire,
		needToken:      0,
	}

	if conf.Limit == 0 {
		s.limit = ^int(0)
	} else {
		s.limit = conf.Limit
	}

	m := &ThrottleManager{
		state:    s,
		throttle: conf.Throttle,
	}
	m.start()

	return m, nil
}

func (m *ThrottleManager) Acquire() (t *Token, err error) {
	return acquire(m.state)
}

func (m *ThrottleManager) start() {
	state := m.state

	go func() {
		<-state.inChan
		tryGenerateToken(state)
		ticker := time.NewTicker(m.throttle)

		for {
			select {
			case <-state.inChan:
				<-ticker.C
				tryGenerateToken(state)
			case t := <-state.releaseChan:
				releaseToken(state, t)
			}
		}
	}()
}

func (m *ThrottleManager) Release(t *Token) {
	if t.IsExpired() {
		go func() {
			m.state.releaseChan <- t
		}()
	}
}
