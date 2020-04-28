package ratelimiter

import (
	"sync/atomic"
	"time"
)

func withMaxConcurrency(m *Manager) error {
	go func() {
		// Check every 10 seconds
		ticker := time.NewTicker(time.Second * 10)

		for {
			<-ticker.C
			now := time.Now().UTC()
			for _, token := range m.activeTokens {
				if token.ExpiresAt.Before(now) {
					m.Release(token)
				}
			}
		}
	}()

	go func() {
		for {
			select {
			case <-m.inChan:
				if !m.hasRemaining() {
					atomic.AddInt64(&m.awaiting, 1)
					continue
				}
				m.generateToken()
			case t := <-m.releaseChan:
				m.releaseToken(t)
			}
		}
	}()

	return nil
}
