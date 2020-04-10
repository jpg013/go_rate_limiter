package ratelimit

import (
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
	mux            sync.Mutex
	resetTicker    *time.Ticker
	closeChan      chan struct{}
	resourcesInUse map[ResourceToken]*Resource
	needResource   int64
}

// polling task that listens for reset ticker channel
// and runs the unlockExpiredResourceTask
func (m *Manager) startPollingTask() {
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

// Release takes a resource token and expires the resource
// and removes it from the inUse map. If there are pending needed resources
// then it will run acquire resource.
func (m *Manager) Release(token ResourceToken) {
	m.mux.Lock()
	r, ok := m.resourcesInUse[token]

	if !ok {
		log.Printf("unable to relase token %s - not in use", token)
	}

	// Unlock the resource
	err := unlockResource(r.ID)

	if err != nil {
		fmt.Printf("error releasing resource %s", err.Error())
	}

	// delete the resource from the map
	delete(m.resourcesInUse, token)

	needResource := atomic.LoadInt64(&m.needResource)

	// If we needResource, then call acquire resource
	if needResource > 0 {
		if atomic.CompareAndSwapInt64(
			&m.needResource,
			needResource,
			needResource-1,
		) {
			go acquireResource(m)
		}
	}
	m.mux.Unlock()
}

func acquireResource(m *Manager) {
	m.mux.Lock()
	res, err := NewResource(m)
	m.mux.Unlock()

	if err != nil {
		atomic.AddInt64(&m.needResource, 1)
	} else {
		m.resourcesInUse[res.Token] = res
		m.tokenChan <- res.Token
	}
}

// Acquire is called to receive a ResourceToken
func (m *Manager) Acquire() ResourceToken {
	// request a resource to be put on the need resource chan
	go acquireResource(m)

	// wait for an available resource from the channel
	return <-m.tokenChan
}

func (m *Manager) unlockExpiredResourceTask() {
	m.mux.Lock()

	for token, r := range m.resourcesInUse {
		if r.ExpiresAt.Before(time.Now().UTC()) {
			m.Release(token)
		}
	}

	m.mux.Unlock()
}

// NewManager is called with rate limit name to create a new manager instance
func NewManager(name string) (*Manager, error) {
	// Init the manager
	m, err := initManager(name)
	fmt.Println(m.resourcesInUse)
	if err != nil {
		return m, err
	}

	m.startPollingTask()

	return m, nil
}

func syncResourceInUseState(m *Manager) error {
	m.mux.Lock()

	// Load all resources in use from database
	stmt, err := MySQL.Prepare(`
		SELECT
			r.id
		FROM 
			` + "`rate_limit_resource`" + ` r
		WHERE 
			r.rate_limit_id = ?
		AND
			r.is_expired = false;
	`)

	if err != nil {
		return err
	}

	sqlRows, err := stmt.Query(m.id)

	if err != nil {
		return err
	}

	ids := make([]int64, 0)

	for sqlRows.Next() {
		var id int64
		err = sqlRows.Scan(&id)
		if err != nil {
			return err
		}
		ids = append(ids, id)
	}

	// Reset the resource in Use map
	m.resourcesInUse = make(map[ResourceToken]*Resource)

	for _, id := range ids {
		res, err := getResourceByID(id)

		if err != nil {
			return nil
		}

		m.resourcesInUse[res.Token] = res
	}

	m.mux.Unlock()

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
		tokenChan:      make(chan ResourceToken),
		resourcesInUse: make(map[ResourceToken]*Resource),
		needResource:   0,
	}

	return m, syncResourceInUseState(m)
}
