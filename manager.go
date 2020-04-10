package ratelimit

import (
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"time"
)

type LockToken string

type Manager struct {
	id             int
	name           string
	limit          int
	resetInSeconds int
	resourceChan   chan *Resource
	mux            sync.Mutex
	resetTicker    *time.Ticker
	closeChan      chan struct{}
	needResource   int64
}

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

func (m *Manager) ReleaseResource(r *Resource) {
	err := r.unlock()

	if err != nil {
		fmt.Printf("error releasing resource %s", err.Error())
	}

	needResource := atomic.LoadInt64(&m.needResource)

	if atomic.CompareAndSwapInt64(&m.needResource, needResource, int64(math.Max(float64(needResource-1), 0))) {
		go acquireResource(m)
	}
}

func acquireResource(m *Manager) {
	m.mux.Lock()
	res, err := NewResource(m)
	m.mux.Unlock()

	if err != nil {
		atomic.AddInt64(&m.needResource, 1)
	} else {
		m.resourceChan <- res
	}
}

func (m *Manager) AcquireResource() *Resource {
	// request a resource to be put on the need resource chan
	go acquireResource(m)

	// wait for an available resource from the channel
	return <-m.resourceChan
}

func (m *Manager) unlockExpiredResourceTask() {
	// get all resource IDs that are expired
	stmt, err := MySQL.Prepare(`
		SELECT
			r.id
		FROM
			` + `rate_limit_resource` + ` r
		WHERE
			r.expires_at > ?;
	`)

	if err != nil {
		panic(err.Error())
	}

	// now := time.Now().UTC().Format("2006-01-02 15:04:05")
	rows, err := stmt.Query(time.Now().UTC())

	if err != nil {
		panic(err.Error())
	}

	ids := make([]string, 0)

	for rows.Next() {
		var resourceID string
		err := rows.Scan(&resourceID)
		if err != nil {
			panic(err.Error())
		}
		ids = append(ids, resourceID)
	}

	// Loop through each resource and check if expired
	for _, id := range ids {
		stmt, err := MySQL.Prepare(`DELETE FROM rate_limit_resource WHERE id=?;`)
		if err != nil {
			panic(err.Error())
		}
		results, err := stmt.Exec(id)
		if err != nil {
			panic(err.Error())
		}
		affectedCount, err := results.RowsAffected()
		if err != nil {
			panic(err.Error())
		}
		if affectedCount != 1 {
			panic("could not delete expired resource")
		}
	}
}

func NewManager(name string) (*Manager, error) {
	// Init the manager
	m, err := initManager(name)

	if err != nil {
		return m, err
	}

	m.startPollingTask()

	return m, nil
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

	return &Manager{
		id:             id,
		name:           name,
		limit:          limit,
		resetInSeconds: resetInSeconds,
		resetTicker:    time.NewTicker(time.Second * 3),
		closeChan:      make(chan struct{}),
		resourceChan:   make(chan *Resource),
		needResource:   0,
	}, nil
}
