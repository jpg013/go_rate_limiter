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
		token := m.Acquire()
		n := rand.Intn(5)
		fmt.Printf("Worker %d Sleeping %d seconds...\n", id, n)
		time.Sleep(time.Duration(n) * time.Second)
		fmt.Printf("Worker %d Done\n", id)
		// release the resource
		m.Release(token)
		wg.Done()
	}

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go doWork(i)
	}

	wg.Wait()

	time.Sleep(10 * time.Second)
}
