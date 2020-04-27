package ratelimiter

import "sync"

type Manager struct {
	errorChan    chan error
	outChan      chan *Token
	inChan       chan struct{}
	releaseChan  chan *Token
	awaiting     int64
	lock         sync.Mutex
	activeTokens map[string]*Token
}

func NewManager() (*Manager, error) {
	m := &Manager{
		errorChan:    make(chan error),
		outChan:      make(chan *Token),
		releaseChan:  make(chan *Token),
		inChan:       make(chan struct{}),
		activeTokens: make(map[string]*Token),
		awaiting:     0,
	}

	return m, nil
}

func (m *Manager) Acquire() (t *Token, err error) {
	go func() {
		m.inChan <- struct{}{}
	}()

	// Await rate limit token
	select {
	case t := <-m.outChan:
		return t, nil
	case err := <-m.errorChan:
		return nil, err
	}
}

func (m *Manager) Release(t *Token) error {
	// m.lock.Lock()
	go func() {
		m.releaseChan <- t
	}()

	// _, ok := m.tokensInUse[t]

	// if !ok {
	// 	return fmt.Errorf("unable to relase token %s - not in use", t)
	// }

	// delete(m.tokensInUse, t)

	// Call to process next needed resource
	// needNextRateLimit(m)

	return nil
}

func (m *Manager) generateToken() {
	t, err := NewToken()

	if err != nil {
		m.errorChan <- err
	} else {
		// Add token to map
		m.activeTokens[t.ID] = t
		m.outChan <- t
	}
}
