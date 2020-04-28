package ratelimiter

import (
	"time"
)

func withThrottle(m *Manager) error {
	throttle := m.config.Throttle

	go func() {
		ticker := time.NewTicker(throttle)

		<-m.inChan
		m.generateToken()

		for {
			select {
			case <-m.inChan:
				<-ticker.C
				m.generateToken()
			case t := <-m.releaseChan:
				m.releaseToken(t)
			}
		}
	}()

	return nil
}
