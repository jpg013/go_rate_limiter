package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/jpg013/ratelimiter"
)

func main() {
	conf := &ratelimiter.Config{
		Type:     ratelimiter.ConcurrencyType,
		Limit:    1,
		ResetsIn: time.Second * 10,
	}

	rateLimiter, err := ratelimiter.New(conf)

	if err != nil {
		panic(err)
	}

	var wg sync.WaitGroup
	rand.Seed(time.Now().UnixNano())

	doWork := func(id int) {
		// Acquire a rate limit token
		token, err := rateLimiter.Acquire()
		fmt.Printf("Rate Limit Token %s acquired at %s...\n", token.ID, time.Now())
		if err != nil {
			panic(err)
		}
		// Simulate some work
		n := rand.Intn(5)
		fmt.Printf("Worker %d Sleeping %d seconds...\n", id, n)
		time.Sleep(time.Duration(n) * time.Second)
		fmt.Printf("Worker %d Done\n", id)
		rateLimiter.Release(token)
		wg.Done()
	}

	// Spin up a 10 workers that need a rate limit resource
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go doWork(i)
	}

	wg.Wait()

	time.Sleep(100 * time.Second)
}
