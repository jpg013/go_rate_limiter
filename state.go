package ratelimiter

import (
	"log"
	"sync/atomic"
	"time"
)

type State struct {
	closeChan      chan struct{}
	errorChan      chan error
	releaseChan    chan *Token
	outChan        chan *Token
	inChan         chan struct{}
	needToken      int64
	activeTokens   map[string]*Token
	limit          int
	getTokenExpire func() time.Time
}

func defaultExpire() time.Time {
	return time.Time{}
}

func incNeedToken(s *State) {
	atomic.AddInt64(&s.needToken, 1)
}

func decNeedToken(s *State) {
	atomic.AddInt64(&s.needToken, -1)
}

func awaitingToken(s *State) bool {
	return atomic.LoadInt64(&s.needToken) > 0
}

func tryGenerateToken(s *State) {
	if isLimitExceeded(s) {
		incNeedToken(s)
		return
	}

	token, err := NewToken(s.getTokenExpire())

	if err != nil {
		s.errorChan <- err
		return
	}

	// Add token to map
	s.activeTokens[token.ID] = token

	// send token to outChan
	go func() {
		s.outChan <- token
	}()
}

func isLimitExceeded(s *State) bool {
	if len(s.activeTokens) >= s.limit {
		return true
	}

	return false
}

func releaseToken(s *State, token *Token) {
	if token == nil {
		log.Print("unable to relase nil token")
		return
	}

	if _, ok := s.activeTokens[token.ID]; !ok {
		log.Printf("unable to relase token %s - not in use", s)
		return
	}

	if !token.IsExpired() {
		log.Printf("unable to relase token %s - has not expired", s)
		return
	}

	// Delete from map
	delete(s.activeTokens, token.ID)

	// Is anything waiting for a rate limit?
	if awaitingToken(s) {
		decNeedToken(s)
		go tryGenerateToken(s)
	}
}
