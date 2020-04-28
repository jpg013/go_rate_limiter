package ratelimiter

import (
	"log"
	"sync/atomic"
)

type Manager struct {
	errorChan    chan error
	outChan      chan *Token
	inChan       chan struct{}
	releaseChan  chan *Token
	awaiting     int64
	activeTokens map[string]*Token
	config       *Config
}

func NewManager(conf *Config) (*Manager, error) {
	m := &Manager{
		errorChan:    make(chan error),
		outChan:      make(chan *Token),
		releaseChan:  make(chan *Token),
		inChan:       make(chan struct{}),
		activeTokens: make(map[string]*Token),
		awaiting:     0,
		config:       conf,
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

	return nil
}

func (m *Manager) generateToken() {
	t, err := NewToken(m.config.ResetsIn)

	if err != nil {
		m.errorChan <- err
	} else {
		// Add token to map
		m.activeTokens[t.ID] = t
		m.outChan <- t
	}
}

func (m *Manager) releaseToken(t *Token) {
	if t == nil {
		log.Print("unable to relase nil token")
		return
	}

	if _, ok := m.activeTokens[t.ID]; !ok {
		log.Printf("unable to relase token %s - not in use", t)
		return
	}

	// Delete from map
	delete(m.activeTokens, t.ID)

	// Is anything waiting for a rate limit?
	if atomic.LoadInt64(&m.awaiting) > 0 {
		atomic.AddInt64(&m.awaiting, -1)
		m.generateToken()
	}
}

func (m *Manager) hasRemaining() bool {
	if atomic.LoadInt64(&m.awaiting) > 0 {
		return false
	}

	if len(m.activeTokens) >= m.config.Limit {
		return false
	}

	return true
}
