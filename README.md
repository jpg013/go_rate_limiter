### Go Rate Limiter

Rate limiting examples in Golang.

`$ go get github.com/jpg013/ratelimiter`

Throttle Rate Limiter
```golang
func main() {
	r, err := ratelimiter.NewThrottleRateLimiter(&ratelimiter.Config{
		Throttle: 1 * time.Second,
	})

	if err != nil {
		panic(err)
	}

	doWork(r, 10)
}

func doWork(r RateLimiter, workerCount int) {
	var wg sync.WaitGroup
	rand.Seed(time.Now().UnixNano())

	doWork := func(id int) {
		// Acquire a rate limit token
		token, err := r.Acquire()
		fmt.Printf("Rate Limit Token acquired %s...\n", token.ID)
		if err != nil {
			panic(err)
		}
		// Simulate some work
		n := rand.Intn(5)
		fmt.Printf("Worker %d Sleeping %d seconds...\n", id, n)
		time.Sleep(time.Duration(n) * time.Second)
		fmt.Printf("Worker %d Done\n", id)
		r.Release(token)
		wg.Done()
	}

	// Spin up a 10 workers that need a rate limit resource
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go doWork(i)
	}

	wg.Wait()
}
```

Max Concurrency Rate Limiter
```golang
func main() {
	r, err := ratelimiter.NewMaxConcurrencyRateLimiter(&ratelimiter.Config{
		Limit:            4,
		TokenResetsAfter: 10 * time.Second,
	})

	if err != nil {
		panic(err)
	}

	doWork(r, 10)
}
```

Fixed Window Rate Limiter
```golang
func main() {
	r, err := ratelimiter.NewFixedWindowRateLimiter(&ratelimiter.Config{
		Limit:         5,
		FixedInterval: 15 * time.Second,
	})

	if err != nil {
		panic(err)
	}

	doWork(r, 10)
}
```