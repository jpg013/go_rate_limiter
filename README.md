### Go Rate Limiter
A rate limiting module powered by Golang and MySQL.

`$ go get github.com/jpg013/ratelimit`

```golang
func main() {
	m, err := ratelimit.NewManager("crawlera")

	if err != nil {
		panic(err)
	}

	var wg sync.WaitGroup
	rand.Seed(time.Now().UnixNano())

	doWork := func(id int) {
    		// Acquire a resource token
    		token := m.Acquire()
    		// Simulate some work for some random time 
    		n := rand.Intn(5)
		fmt.Printf("Worker %d Sleeping %d seconds...\n", id, n)
		time.Sleep(time.Duration(n) * time.Second)
		fmt.Printf("Worker %d Done\n", id)
		// always release the resource token
		m.Release(token)
		wg.Done()
	}

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go doWork(i)
	}

	wg.Wait()
}
```
