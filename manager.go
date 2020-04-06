package ratelimit

import (
	"sync"
	"time"
)

type LockToken string

type Manager struct {
	id               int
	name             string
	limit            int
	remaining        int
	resetInSeconds   int
	needResourceChan chan Resource
	mux              sync.Mutex
	resetTicker      *time.Ticker
	lockedResources  map[LockToken]*Resource
	closeChan        chan struct{}
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

func (m *Manager) ReleaseResource(r *Resource) error {
	m.mux.Lock()
	defer m.mux.Unlock()

	// if _, ok := m.pool[r.ID]; !ok {
	// 	return errors.New("cannot release resource, not in use.")
	// }

	// err := r.release()

	// if err != nil {
	// 	return err
	// }

	// // remove resource from pool
	// delete(m.pool, r.ID)

	// if atomic.LoadInt32(&m.needResource) > 0 {
	// 	res := m.acquireResource()

	// 	if res == nil {
	// 		panic("could not retrieve resource from pool.")
	// 	}

	// 	atomic.AddInt32(&m.needResource, -1)

	// 	if atomic.LoadInt32(&m.needResource) < 0 {
	// 		panic("need resource cannot be less than zero")
	// 	}

	// 	// run a goroutine and add the resource to the resourceChan
	// 	go func() {
	// 		m.resourceChan <- res
	// 	}()
	// }

	return nil
}

// func (m *Manager) QueryFreeResource() (*Resource, error) {
// 	// Attempt to get one from out of the current pool of existing rate limit resources
// 	stmt, err := MySQL.Prepare(`
// 		SELECT
// 			r.id
// 		FROM
// 			` + "`rate_limit_resource`" + ` r
// 		WHERE
// 			r.lock_id IS NOT NULL
// 		AND
// 			r.name = ?
// 		LIMIT 1;
// 	`)

// 	if err != nil {
// 		return nil, err
// 	}

// 	var id string
// 	err = stmt.QueryRow(m.name).Scan(&id)

// 	if err == sql.ErrNoRows {
// 		return nil, nil
// 	}

// 	if err != nil {
// 		return nil, err
// 	}

// 	return GetResourceByID(id)
// }

// tries to acquire a resource by first looking for any free / available
// resources in the resource table, and if that fails then by creating
// a new resource if the limit has not been exceeded.
func (m *Manager) acquire() error {

	// Try to query the resource pool first
	// resource, err := NewResource(m)

	// if err != nil {
	// 	panic(err) // handle this
	// }

	// // exit if we can't get resource
	// if resource == nil {
	// 	return nil
	// }

	// // Lock the resource
	// err = resource.lock()

	// if err != nil {
	// 	panic(err)
	// }

	// // Update the map of resource
	// _, ok := m.resourceMap[resource.ID]

	// if ok == true {
	// 	panic("resource exists in pool")
	// }

	// m.resourceMap[resource.ID] = resource

	// return resource
	return nil
}

func (m *Manager) AcquireResource() Resource {
	// request a resource to be put on the need resource chan

	// wait for an available resource from the channel
	res := <-m.needResourceChan

	return res
}

func (m *Manager) unlockExpiredResourceTask() {
	// get all resource IDs that are expired
	stmt, err := MySQL.Prepare(`SELECT r.id FROM rate_limit_resource r WHERE r.expires_at > ?;`)

	if err != nil {
		panic(err.Error())
	}

	now := time.Now().UTC().Format("2006-01-02 15:04:05")
	rows, err := stmt.Query(now)

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
			r.name=?;
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
		id:               id,
		name:             name,
		limit:            limit,
		resetInSeconds:   resetInSeconds,
		resetTicker:      time.NewTicker(time.Second * 3),
		closeChan:        make(chan struct{}),
		needResourceChan: make(chan Resource),
		lockedResources:  make(map[LockToken]*Resource),
	}, nil
}
