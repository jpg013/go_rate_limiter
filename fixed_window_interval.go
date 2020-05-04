package ratelimiter

import "time"

// FixedWindowInterval represents a fixed window of time with a start / end time
type FixedWindowInterval struct {
	startTime time.Time
	endTime   time.Time
	interval  time.Duration
}

func (w *FixedWindowInterval) setWindowTime() {
	w.startTime = time.Now().UTC()
	w.endTime = time.Now().UTC().Add(w.interval)
}

func (w *FixedWindowInterval) run(cb func()) {
	go func() {
		ticker := time.NewTicker(w.interval)
		w.setWindowTime()
		for range ticker.C {
			cb()
			w.setWindowTime()
		}
	}()
}
