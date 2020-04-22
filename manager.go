package ratelimit

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

// Manager represents a rate limit manager
type Manager struct {
	id             int
	name           string
	limit          int
	resetInSeconds int
	tokenChan      chan ResourceToken
	releaseChan    chan ResourceToken
	mux            sync.Mutex
	resetTicker    *time.Ticker
	closeChan      chan struct{}
	resourcesInUse map[ResourceToken]*Resource
	needResource   int64
}

func releaseToken(m *Manager, token ResourceToken) {
	r, ok := m.resourcesInUse[token]

	if !ok {
		log.Printf("unable to relase token %s - not in use", token)
	}

	// Unlock the resource
	err := unlockResource(r.Token)

	if err != nil {
		fmt.Printf("error releasing resource %s", err.Error())
	}

	// delete the resource from the map
	delete(m.resourcesInUse, token)

	// Call to process next needed resource
	needNextResource(m)
}

// Release takes a resource token and expires the resource
// and removes it from the inUse map. If there are pending needed resources
// then it will run acquire resource.
func (m *Manager) Release(tokenStr string) {
	go func() {
		m.releaseChan <- ResourceToken(tokenStr)
	}()
}

func needNextResource(m *Manager) {
	needResource := atomic.LoadInt64(&m.needResource)

	// Exit if no resources needed
	if needResource == 0 {
		return
	}

	// Try to decrement the needResource counter by 1 and call acquireResource
	// if successful
	if atomic.CompareAndSwapInt64(
		&m.needResource,
		needResource,
		needResource-1,
	) {
		go tryAcquireResource(m)
	}
}

// attempts to acquire a new rate limit resource and sent it to the
// token channel. If no resources are available, or there is an error
// then we increment the needResource counter.
func tryAcquireResource(m *Manager) (err error) {
	// Lock when trying to create a new resource
	m.mux.Lock()

	defer func() {
		// if unable to acquire a new resource, then increment the need resource counter.
		if err != nil {
			atomic.AddInt64(&m.needResource, 1)
		}

		// Always remember to unlock!
		m.mux.Unlock()
	}()

	// Number of resources in use is passed the allowed limit.
	if len(m.resourcesInUse) >= m.limit {
		err = errors.New("all available resources in use")
		return err
	}

	// try to create a new resource
	res, err := NewResource(m)

	if err != nil {
		return err
	}

	m.resourcesInUse[res.Token] = res

	// Allow the function to return to caller before sending token to channel.
	go func() {
		m.tokenChan <- res.Token
	}()

	return err
}

// Acquire is called to receive a ResourceToken
func (m *Manager) Acquire() string {
	// request a resource to be put on the need resource chan
	tryAcquireResource(m)

	// wait for an available resource from the channel
	t := <-m.tokenChan

	return string(t)
}

func (m *Manager) unlockExpiredResourceTask() {
	for token, r := range m.resourcesInUse {
		if r.ExpiresAt.Before(time.Now().UTC()) {
			go func(t ResourceToken) {
				m.releaseChan <- t
			}(token)
		}
	}
}

// polling that that checks for expired resources and
// forcefully releases them from use.
func (m *Manager) startPollingExpiredTask() {
	go func() {
		for {
			select {
			case <-m.resetTicker.C:
				m.unlockExpiredResourceTask()
			case <-m.closeChan:
				break
			}
		}
	}()
}

// list to release channel and release tokens.
func (m *Manager) startReceivingReleaseChan() {
	go func() {
		for {
			select {
			case token := <-m.releaseChan:
				releaseToken(m, token)
			case <-m.closeChan:
				break
			}
		}
	}()
}

// NewManager is called with rate limit name to create a new manager instance
func NewManager(name string) (*Manager, error) {
	// Init the manager
	m, err := initManager(name)

	if err != nil {
		return m, err
	}

	// start task to check expired resources
	m.startPollingExpiredTask()

	// start receiving messages on the release chan
	m.startReceivingReleaseChan()

	return m, nil
}

// syncResourceInUseState should be called when the manager is initialized
// to sync the database resource state with the manager.
func syncResourceInUseState(m *Manager) error {
	res, err := getResourcesInUse(m.id)

	if err != nil {
		return err
	}

	// Override the existing resource map
	m.resourcesInUse = make(map[ResourceToken]*Resource)

	for _, r := range res {
		m.resourcesInUse[r.Token] = r
	}

	return nil
}

func initManager(name string) (*Manager, error) {
	// Load all rate limits from database
	stmt, err := MySQL.Prepare(`
		SELECT DISTINCT
			r.id,
			r.limit,
			r.reset_in_seconds
		FROM 
			` + "`rate_limit`" + ` r
		WHERE 
			r.name = ?;
	`)

	if err != nil {
		return nil, err
	}

	var limit, resetInSeconds, id int

	err = stmt.QueryRow("crawlera").Scan(&id, &limit, &resetInSeconds)

	if err != nil {
		return nil, err
	}

	m := &Manager{
		id:             id,
		name:           name,
		limit:          limit,
		resetInSeconds: resetInSeconds,
		resetTicker:    time.NewTicker(time.Second * 3),
		closeChan:      make(chan struct{}),
		releaseChan:    make(chan ResourceToken),
		tokenChan:      make(chan ResourceToken),
		needResource:   0,
	}

	// sync the database resource state
	return m, syncResourceInUseState(m)
}
