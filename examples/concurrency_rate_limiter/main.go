package main

import (
	"github.com/jpg013/ratelimiter"
)

func main() {
	r, err := ratelimiter.NewConcurrencyRateLimiter(&ratelimiter.Config{
		Limit: 1,
	})

	if err != nil {
		panic(err)
	}
	ratelimiter.DoWork(r, 10)
}
