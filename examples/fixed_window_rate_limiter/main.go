package main

import (
	"time"

	"github.com/jpg013/ratelimiter"
)

func main() {
	r, err := ratelimiter.NewFixedWindowRateLimiter(&ratelimiter.Config{
		Limit:    5,
		Interval: 15 * time.Second,
	})

	if err != nil {
		panic(err)
	}

	ratelimiter.DoWork(r, 10)
}
