package main

import (
	"time"

	"github.com/jpg013/ratelimiter"
)

func main() {
	r, err := ratelimiter.NewMaxConcurrencyRateLimiter(&ratelimiter.Config{
		Limit:            4,
		TokenResetsAfter: 10 * time.Second,
	})

	if err != nil {
		panic(err)
	}
	ratelimiter.DoWork(r, 10)
}
