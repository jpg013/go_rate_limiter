package ratelimiter

import (
	"time"
)

type FixedWindowManager struct {
	state     *State
	startTime time.Time
	endTime   time.Time
	interval  time.Duration
}

func NewFixedWindowRateLimiter(conf *Config) (RateLimiter, error) {
	if conf.Interval == 0 {
		return nil, ErrIntervalGreaterThanZero
	}

	if conf.Limit == 0 {
		return nil, ErrLimitGreaterThanZero
	}

	s := &State{
		errorChan:    make(chan error),
		outChan:      make(chan *Token),
		inChan:       make(chan struct{}),
		activeTokens: make(map[string]*Token),
		releaseChan:  make(chan *Token),
		needToken:    0,
		limit:        conf.Limit,
	}

	m := &FixedWindowManager{
		state:    s,
		interval: conf.Interval,
	}

	// define token expiration func
	s.getTokenExpire = func() time.Time {
		return m.endTime
	}
	m.start()

	return m, nil
}

func (m *FixedWindowManager) start() {
	runFixedWindowIntervalTask(m)

	go func() {
		state := m.state
		for {
			select {
			case <-state.inChan:
				tryGenerateToken(state)
			case token := <-state.releaseChan:
				releaseToken(m.state, token)
			}
		}
	}()
}

func (m *FixedWindowManager) FixedWindowManager() (t *Token, err error) {
	go func() {
		m.state.inChan <- struct{}{}
	}()

	// Await rate limit token
	select {
	case t := <-m.state.outChan:
		return t, nil
	case err := <-m.state.errorChan:
		return nil, err
	}
}

func setWindowTime(m *FixedWindowManager) {
	m.startTime = time.Now().UTC()
	m.endTime = time.Now().UTC().Add(m.interval)
}

func runFixedWindowIntervalTask(m *FixedWindowManager) {
	ticker := time.NewTicker(m.interval)
	setWindowTime(m)
	go func() {
		for range ticker.C {
			for _, token := range m.state.activeTokens {
				go func(t *Token) {
					m.state.releaseChan <- t
				}(token)
			}
			setWindowTime(m)
		}
	}()
}

func (m *FixedWindowManager) Release(t *Token) {
	if t.IsExpired() {
		go func() {
			m.state.releaseChan <- t
		}()
	}
}

func (m *FixedWindowManager) Acquire() (t *Token, err error) {
	return acquire(m.state)
}
