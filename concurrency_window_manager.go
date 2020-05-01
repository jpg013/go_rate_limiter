package ratelimiter

type ConcurrencyManager struct {
	state *State
}

func NewConcurrencyRateLimiter(conf *Config) (RateLimiter, error) {
	if conf.Limit <= 0 {
		return nil, ErrLimitGreaterThanZero
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
		limit:          conf.Limit,
	}
	m := &ConcurrencyManager{
		state: s,
	}
	m.start()
	return m, nil
}

func (m *ConcurrencyManager) Acquire() (t *Token, err error) {
	return acquire(m.state)
}

func (m *ConcurrencyManager) start() {
	state := m.state

	go func() {
		for {
			select {
			case <-state.inChan:
				tryGenerateToken(state)
			case t := <-state.releaseChan:
				releaseToken(state, t)
			}
		}
	}()
}

func (m *ConcurrencyManager) Release(t *Token) {
	if t.IsExpired() {
		go func() {
			m.state.releaseChan <- t
		}()
	}
}
