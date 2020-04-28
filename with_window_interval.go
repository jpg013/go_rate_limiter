package ratelimiter

import (
	"sync/atomic"
	"time"
)

func withWindowInterval(m *Manager) error {
	go func() {
		ticker := time.NewTicker(m.config.Throttle)

		for {
			<-ticker.C
			for _, token := range m.activeTokens {
				m.releaseToken(token)
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
			case <-m.releaseChan:
				continue
			}
		}
	}()

	return nil
}
