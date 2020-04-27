package ratelimiter

// import (
// 	"errors"
// 	"fmt"
// 	"sync"
// 	"time"
// )

type Manager struct {
	cancelChan chan struct{}
	errorChan  chan error
	outChan    chan *Token
	// activeTokens map[string]*Token
	// limit int
	await func() (<-chan *Token, <-chan error)
}

func NewManager() (*Manager, error) {
	m := &Manager{
		cancelChan: make(chan struct{}),
		errorChan:  make(chan error),
		outChan:    make(chan *Token),
		// activeTokens: make(map[string]*Token),
	}

	return m, nil
}

func (m *Manager) Acquire() (t *Token, err error) {
	outChan, errorChan := m.await()

	// Await rate limit token
	select {
	case t := <-outChan:
		return t, nil
	case err := <-errorChan:
		return nil, err
	}
}

func (m *Manager) Release(t *Token) error {
	// m.lock.Lock()

	// _, ok := m.tokensInUse[t]

	// if !ok {
	// 	return fmt.Errorf("unable to relase token %s - not in use", t)
	// }

	// delete(m.tokensInUse, t)

	// Call to process next needed resource
	// needNextRateLimit(m)

	return nil
}

// /*
// 	3 types of rate limiting
// 	throttle_in_seconds - connection throttle time in seconds (if set to 0 then not throttled at all)
// 	max_concurrent_connections - maximum amount of concurrent connections allows at one time. (if set to 0 then no max on connections)
// 	interval_window_in_seconds - window time in seconds before rate limit resources reset. If set to 0 then no interval. If non-zero value is present, then resources do not "un-expire" at use.
// */

// type RateLimitType uint32

// // Errors used throughout the codebase
// var (
// 	ErrRateLimiterExists         = errors.New("rate limiter type already exists")
// 	ErrRateLimiterNotInitialized = errors.New("rate limiter has not been initialized")
// )

// const (
// 	UndefinedLimiterType RateLimitType = iota
// 	ThrottleType
// 	MaxConcurrencyType
// 	IntervalWindowType
// )

// type RateLimitManager interface {
// 	Acquire() (*Type, error)
// 	Release(*Type) error
// 	WithThrottle(d time.Duration) error
// }

// type Type struct{}

// type RateLimitExecutor func(chan struct{}, chan struct{}) chan struct{}

// type Manager struct {
// 	rateLimiterType RateLimitType
// 	dataChan        chan *Type
// 	cancelChan      chan struct{}
// 	awaitLimiter    func() chan struct{}
// }

// func (m *Manager) requestRateLimit() {
// 	go func() {
// 		m.requestsChan <- struct{}{}
// 	}()
// }

// func tryGetRateLimit(m *Manager) {
// 	go func() {
// 		// Await for rate limiter
// 		<-m.awaitLimiter()

// 		// Create a new rate limit type and add it to the data chan
// 		m.dataChan <- NewType()
// 	}()
// }

// func (m *Manager) Acquire() (*Type, error) {
// 	if m.rateLimiterType == UndefinedLimiterType {
// 		return nil, ErrRateLimiterNotInitialized
// 	}

// 	tryGetRateLimit(m)

// 	// await response from the data channel
// 	rateLimit := <-m.dataChan

// 	return rateLimit, nil
// }

// func (m *Manager) Release(t *Type) error {
// 	return nil
// }

// func makeThrottler(d time.Duration) func() {
// 	ticker := time.NewTicker(d)

// 	return func() {
// 		// Wait for ticker
// 		<-ticker.C
// 	}
// }

// func (m *Manager) WithThrottle(d time.Duration) error {
// 	if m.rateLimiterType != UndefinedLimiterType {
// 		return ErrRateLimiterExists
// 	}

// 	throttler := makeThrottler(d)
// 	outChan := make(chan struct{})

// 	go func(inChan chan struct{}, cancelChan chan struct{}) {
// 		var wg sync.WaitGroup

// 		for {
// 			select {
// 			case <-inChan:
// 				wg.Add(1)
// 				// await rate limit throttler
// 				throttler()
// 				outChan <- struct{}{}
// 				wg.Done()
// 			case <-cancelChan:
// 				break
// 			}
// 		}

// 		defer func() {
// 			wg.Wait()
// 			fmt.Println("do something here??")
// 		}()
// 	}(m.requestsChan, m.cancelChan)

// 	m.dataChan = outChan

// 	return nil
// }

// func (m *Manager) WithMaxConcurrentConnections(maxConns int) error {
// 	return nil
// }

// func NewRateLimiter() RateLimitManager {
// 	m := &Manager{
// 		requestsChan: make(chan struct{}),
// 		cancelChan:   make(chan struct{}),
// 	}

// 	return m
// }

// // Manager represents a rate limit manager
// // type Manager struct {
// // 	// id                  int
// // 	// name                string
// // 	// limit               int
// // 	// resetInSeconds      int
// // 	// acquireResourceChan chan struct{}
// // 	// waitResourceChan    chan *Resource
// // 	// releaseChan         chan ResourceToken
// // 	// lock                sync.RWMutex
// // 	// resetTicker  struct {
// // 	// id                  int
// // 	// name                string
// // 	// limit               int
// // 	// resetInSeconds      int
// // 	// acquireResourceChan chan struct{}
// // 	// waitR

// func NewType() *Type {
// 	return &Type{}
// }

// // func NewManager() *Manager {
// // 	return &Manager{
// // 		dataCh:    make(chan struct{}),
// // 		cancelCh:  make(chan struct{}),
// // 		executors: make([]RateLimitExecutor, 0),
// // 	}
// // }

// // func throttleFunc(d time.Duration) RateLimitExecutor {
// // 	return func(inCh chan struct{}, cancelCh chan struct{}) chan struct{} {
// // 		outCh := make(chan struct{})
// // 		ticker := time.NewTicker(d)

// // 		go func() {
// // 			select {
// // 			case <-inCh:
// // 				// Wait for throttle ticker
// // 				<-ticker.C
// // 				// Send an empty struct to the out channel
// // 				outCh <- struct{}{}
// // 			case <-cancelCh:
// // 				break
// // 			}
// // 		}()

// // 		return outCh
// // 	}
// // }

// // func (m *Manager) WithThrottle(d time.Duration) *Manager {
// // 	fn := throttleFunc(d)
// // 	m.executors = append(m.executors, fn)

// // 	return m
// // }

// // func (m *Manager) Start() {
// // 	// Merge all the exeutors into a single channel pipeline
// // 	for _, fn := range m.executors {
// // 		m.dataCh = fn(m.dataCh, m.cancelCh)
// // 	}
// // }

// // func (m *Manager) AcquireRateLimit() *Type {
// // 	// Send a message to the in channel
// // 	go func() {
// // 		m.dataCh <- struct{}{}
// // 	}()

// // 	// wait for rate limit from the outChan
// // 	<-m.dataCh

// // 	// Create a new RateLimitType
// // 	return NewType()
// // }

// // func (m *Manager) Release(t *Type) {

// // }

// // func releaseToken(m *Manager, token ResourceToken) {
// // 	r, ok := m.resourcesInUse[token]

// // 	if !ok {
// // 		log.Printf("unable to relase token %s - not in use", token)
// // 		return
// // 	}

// // 	// Unlock the resource
// // 	err := unlockResource(r.Token)

// // 	if err != nil {
// // 		fmt.Printf("error releasing resource %s", err.Error())
// // 	}

// // 	// delete the resource from the map
// // 	delete(m.resourcesInUse, token)

// // 	// Call to process next needed resource
// // 	needNextResource(m)
// // }

// // Release takes a resource token and expires the resource
// // and removes it from the inUse map. If there are pending needed resources
// // then it will run acquire resource.
// // func (m *Manager) Release(tokenStr string) {
// // 	go func() {
// // 		m.releaseChan <- ResourceToken(tokenStr)
// // 	}()
// // }

// func needNextRateLimit(m *Manager) {
// 	val := atomic.LoadInt64(&m.needRateLimit)

// 	// Exit if no rate limits needed
// 	if val == 0 {
// 		return
// 	}

// 	// Try to decrement the needResource counter by 1 and call acquireResource
// 	// if successful
// 	if atomic.CompareAndSwapInt64(
// 		&m.needRateLimit,
// 		val,
// 		val-1,
// 	) {
// 		go requestRateLimit(m)
// 	}
// }

// // func (m *Manager) getAvailableResource() *Resource {
// // 	for _, r := range m.resources {
// // 		if !r.InUse {
// // 			return r
// // 		}
// // 	}

// // 	return nil
// // }

// // // attempts to acquire a new rate limit resource and sent it to the
// // // token channel. If no resources are available, or there is an error
// // // then we increment the needResource counter.
// // func tryAcquireResource(m *Manager) {
// // 	var r *Resource
// // 	var err error

// // 	defer func() {
// // 		if err != nil {
// // 			return
// // 		}

// // 		// If the acquired resource is not in the resource map, then add it
// // 		if _, ok := m.resources[r.Token]; !ok {
// // 			m.resources[r.Token] = r
// // 		}

// // 		// lock the resource
// // 		r.lock()

// // 		// send the resource to the "waiting" channel
// // 		m.waitResourceChan <- r
// // 	}()

// // 	// Try to get an existing available resource
// // 	r = m.getAvailableResource()

// // 	// If found, exit
// // 	if r != nil {
// // 		return
// // 	}

// // 	// Check to see if all resources are in use
// // 	if len(m.resources) > m.limit {
// // 		err = errors.New("all resources in use")
// // 		return
// // 	}

// // 	// We were unable to find an available resource, try to create a new one.
// // 	// Lock to prevent deadlock at the database level
// // 	r, err = NewResource(m)

// // 	if err != nil {
// // 		return
// // 	}

// // 	// All resources in use, increment the needResource counter
// // 	atomic.AddInt64(&m.needResource, 1)
// // }

// // // Acquire is called to receive a ResourceToken
// // func (m *Manager) Acquire() *Resource {
// // 	// request a resource to be put on the need resource chan
// // 	go func() {
// // 		m.acquireResourceChan <- struct{}{}
// // 	}()

// // 	// wait for an available resource from the channel
// // 	r := <-m.waitResourceChan
// // 	return r
// // }

// // func (m *Manager) resetExpiredResources() {
// // 	for token, r := range m.resourcesInUse {
// // 		if r.ExpiresAt.Before(time.Now().UTC()) {
// // 			go func(t ResourceToken) {
// // 				m.releaseChan <- t
// // 			}(token)
// // 		}
// // 	}
// // }

// // // periodically checks for expired resources and resets them
// // func (m *Manager) runResetExpiredTask() {
// // 	go func() {
// // 		for {
// // 			select {
// // 			case <-m.resetTicker.C:
// // 				m.resetExpiredResources()
// // 			case <-m.closeChan:
// // 				break
// // 			}
// // 		}
// // 	}()
// // }

// // // list to release channel and release tokens.
// // func (m *Manager) runReceiveReleaseChan() {
// // 	go func() {
// // 		for {
// // 			select {
// // 			case token := <-m.releaseChan:
// // 				releaseToken(m, token)
// // 			case <-m.closeChan:
// // 				break
// // 			}
// // 		}
// // 	}()
// // }

// // // NewManager is called with rate limit name to create a new manager instance
// // func NewManager(name string) (*Manager, error) {
// // 	// Init the manager
// // 	m, err := initManager(name)

// // 	if err != nil {
// // 		return m, err
// // 	}

// // 	// start task to check expired resources
// // 	m.runResetExpiredTask()

// // 	// start receiving messages on the release chan
// // 	m.runReceiveReleaseChan()

// // 	return m, nil
// // }

// // // syncResourceInUseState should be called when the manager is initialized
// // // to sync the database resource state with the manager.
// // func syncResourceInUseState(m *Manager) error {
// // 	res, err := getResourcesInUse(m.id)

// // 	if err != nil {
// // 		return err
// // 	}

// // 	// Override the existing resource map
// // 	m.resourcesInUse = make(map[ResourceToken]*Resource)

// // 	for _, r := range res {
// // 		m.resourcesInUse[r.Token] = r
// // 	}

// // 	return nil
// // }

// // func initManager(name string) (*Manager, error) {
// // 	// Load all rate limits from database
// // 	stmt, err := MySQL.Prepare(`
// // 		SELECT DISTINCT
// // 			r.id,
// // 			r.limit,
// // 			r.reset_in_seconds
// // 		FROM
// // 			` + "`rate_limit`" + ` r
// // 		WHERE
// // 			r.name = ?;
// // 	`)

// // 	if err != nil {
// // 		return nil, err
// // 	}

// // 	var limit, resetInSeconds, id int

// // 	err = stmt.QueryRow("crawlera").Scan(&id, &limit, &resetInSeconds)

// // 	if err != nil {
// // 		return nil, err
// // 	}

// // 	m := &Manager{
// // 		id:             id,
// // 		name:           name,
// // 		limit:          limit,
// // 		resetInSeconds: resetInSeconds,
// // 		resetTicker:    time.NewTicker(time.Second * 3),
// // 		closeChan:      make(chan struct{}),
// // 		releaseChan:    make(chan ResourceToken),
// // 		tokenChan:      make(chan ResourceToken),
// // 		needResource:   0,
// // 	}

// // 	// sync the database resource state
// // 	return m, syncResourceInUseState(m)
// // }

// func main() {
// 	rateLimiter := NewRateLimiter()
// 	err := rateLimiter.WithThrottle(1 * time.Second)

// 	if err != nil {
// 		panic(err)
// 	}

// 	time.Sleep(5 * time.Second)

// 	for i := 0; i < 10; i++ {
// 		go func() {
// 			_, err := rateLimiter.Acquire()

// 			if err != nil {
// 				panic(err)
// 			}
// 			fmt.Println("Have a rate limit")
// 		}()
// 	}

// 	time.Sleep(10 * time.Minute)
// }
