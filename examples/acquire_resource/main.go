package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/jpg013/ratelimit"
)

func main() {
	m, err := ratelimit.NewManager("crawlera")

	if err != nil {
		panic(err)
	}

	var wg sync.WaitGroup
	rand.Seed(time.Now().UnixNano())

	doWork := func(id int) {
		// Acquire a rate limit token
		token := m.Acquire()
		// Simulate some work
		n := rand.Intn(5)
		fmt.Printf("Worker %d Sleeping %d seconds...\n", id, n)
		time.Sleep(time.Duration(n) * time.Second)
		fmt.Printf("Worker %d Done\n", id)
		// release the resource
		m.Release(token)
		wg.Done()
	}

	// Spin up a 1000 workers that need a rate limit resource
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go doWork(i)
	}

	// Wait for all workers to finish
	wg.Wait()

	// Allow 1 second for all rate limits to release
	time.Sleep(1 * time.Second)
}
