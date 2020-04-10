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
		res := m.AcquireResource()
		n := rand.Intn(10)
		fmt.Printf("Worker %d Sleeping %d seconds...\n", id, n)
		time.Sleep(time.Duration(n) * time.Second)
		fmt.Printf("Worker %d Done\n", id)
		// release the resource
		m.ReleaseResource(res)
		wg.Done()
	}

	for i := 0; i < 30; i++ {
		wg.Add(1)
		go doWork(i)
	}

	wg.Wait()
}
