package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/jpg013/ratelimiter"
)

func main() {
	rateLimiter, err := ratelimiter.New(&ratelimiter.Config{
		Throttle: 3 * time.Second,
		Type:     ratelimiter.ThrottleType,
	})

	if err != nil {
		panic(err)
	}

	var wg sync.WaitGroup
	rand.Seed(time.Now().UnixNano())

	doWork := func(id int) {
		// Acquire a rate limit token
		token, err := rateLimiter.Acquire()
		fmt.Printf("Rate Limit Token acquired %s...\n", token.ID)
		if err != nil {
			panic(err)
		}
		// Simulate some work
		n := rand.Intn(5)
		fmt.Printf("Worker %d Sleeping %d seconds...\n", id, n)
		time.Sleep(time.Duration(n) * time.Second)
		fmt.Printf("Worker %d Done\n", id)
		wg.Done()
	}

	// Spin up a 10 workers that need a rate limit resource
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go doWork(i)
	}

	wg.Wait()
}
