package ratelimit

// import (
// 	"errors"
// 	"time"
// )

import (
	"errors"
	"time"
)

// Resource represents a rate limit resource
type Resource struct {
	ID        string
	Name      string
	CreatedAt time.Time
	Count     int
	ExpiresAt time.Time
	LockToken string
}

// func (r *Resource) release() error {
// 	stmt, err := db.MySQL.Prepare(`
// 		UPDATE
// 			` + "`rate_limit_resource`" + `
// 		SET
// 			` + "`lock_id` = NULL" + `,
// 			` + "`expires_at` = NULL" + `
// 		WHERE
// 			` + "`id` = ?" + `
// 		AND
// 			` + "`lock_id`" + ` = ?;
// 	`)

// 	if err != nil {
// 		panic(err)
// 	}

// 	result, err := stmt.Exec(r.ID, r.LockID)

// 	if err != nil {
// 		return err
// 	}

// 	updated, err := result.RowsAffected()

// 	if err != nil {
// 		return err
// 	}

// 	if updated != 1 {
// 		return errors.New("could not release resource")
// 	}

// 	return nil
// }

// // NewResource attempts to insert a new rate limit resource record
// // into the data, and returns the resource if successful.
// func NewResource(m *Manager) (*Resource, error) {
// 	// lookup pool first
// 	res, err := QueryResourcePool(m)

// 	if err != nil {
// 		return nil, err
// 	}

// 	if res != nil {
// 		return res, nil
// 	}

// 	// No luck with querying the resource pool, try try creating a new resource
// 	id, err := createResource(m)

// 	if err != nil {
// 		return nil, err
// 	}

// 	if id == "" {
// 		return nil, nil
// 	}

// 	return GetResourceByID(id)
// }

// func createResource(m *Manager) (id string, err error) {
// 	stmt, err := db.MySQL.Prepare(`
// 		INSERT INTO rate_limit_resource
// 			(
// 				id,
// 				name,
// 				count,
// 				created_at
// 			)
// 		SELECT ?, ?, ?, ? FROM dual
// 		WHERE
// 			(
// 				SELECT COUNT(*) FROM rate_limit_resource
// 			) < ?;
// 	`)

// 	if err != nil {
// 		return id, err
// 	}

// 	id = helpers.GenKsuid()

// 	results, err := stmt.Exec(
// 		id,
// 		m.name,
// 		1,
// 		time.Now().UTC(),
// 		m.limit,
// 	)

// 	if err != nil {
// 		return id, err
// 	}

// 	inserted, err := results.RowsAffected()

// 	if err != nil {
// 		return id, err
// 	}

// 	// Were we able to inserted a new resource record?
// 	if inserted == 0 {
// 		return "", nil
// 	}

// 	return id, err
// }

// func lockAvailableResource(m *Manager) error {
// 	// Attempt to get one from out of the current pool of live resources
// 	stmt, err := db.MySQL.Prepare(`
// 		UPDATE
// 			` + "`rate_limit_resource`" + `
// 		SET
// 			` + "`lock_token` = ?" + `,
// 			` + "`expires_at` = ?" + `
// 		WHERE
// 			` + "`lock_token` IS NOT NULL" + `
// 		LIMIT 1;
// 	`)

// 	if err != nil {
// 		panic(err)
// 	}

// 	duration := time.Second * time.Duration(m.resetInSeconds)
// 	result, err := stmt.Exec(
// 		lockID,
// 		time.Now().UTC().Add(duration),
// 		r.ID,
// 	)

// 	if err != nil {
// 		return err
// 	}

// 	updated, err := result.RowsAffected()

// 	if err != nil {
// 		return err
// 	}

// 	if updated != 1 {
// 		return errors.New("could not lock resource")
// 	}

// 	return nil
// }

// 	stmt, err := db.MySQL.Prepare(`
// 		UPDATE
// 			` + "`rate_limit_resource`" + `
// 		SET
// 			` + "`lock_token` = ?" + `,
// 			` + "`expires_at` = ?" + `
// 		WHERE
// 			` + "`lock_token` IS NOT NULL" + `
// 		LIMIT 1;
// 	`)

func lockAvailableResource(m *Manager) (*Resource, error) {
	// Attempt to get one from out of the current pool of free resources
	stmt, err := MySQL.Prepare(`
		UPDATE
			` + "`rate_limit_resource`" + `
		SET
			` + "`lock_token` = ?" + `,
			` + "`expires_at` = ?" + `
		WHERE
			` + "`lock_token` IS NOT NULL" + `
		LIMIT 1;
	`)

	if err != nil {
		return nil, err
	}

	// generate a lock token
	lockToken := GenKsuid()
	// duration before lock expires
	duration := time.Second * time.Duration(m.resetInSeconds)

	result, err := stmt.Exec(
		lockToken,
		time.Now().UTC().Add(duration),
	)

	if err != nil {
		return nil, err
	}

	updated, err := result.RowsAffected()

	if err != nil {
		return nil, err
	}

	if updated != 1 {
		return nil, errors.New("could not lock resource")
	}

	return nil, nil
}

// func GetResourceByID(queryID string) (*Resource, error) {
// 	stmt, err := db.MySQL.Prepare(`
// 		SELECT
// 			r.id,
// 			r.name,
// 			r.count,
// 			r.created_at,
// 			r.expires_at,
// 			r.lock_id
// 		FROM
// 			` + "`rate_limit_resource`" + ` r
// 		WHERE
// 			r.id = ?;
// 	`)

// 	if err != nil {
// 		return nil, err
// 	}

// 	var id, createdDatetime, expiresDatetime, name, lockID string
// 	var count int

// 	err = stmt.QueryRow(queryID).Scan(
// 		&id,
// 		&name,
// 		&count,
// 		&createdDatetime,
// 		&expiresDatetime,
// 		&lockID,
// 	)

// 	if err != nil {
// 		return nil, err
// 	}

// 	format := "2006-01-02 15:04:05"
// 	createdAt, err := time.Parse(format, createdDatetime)
// 	if err != nil {
// 		return nil, err
// 	}
// 	expiresAt, err := time.Parse(format, expiresDatetime)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// no error, create a rate limit resource :)
// 	return &Resource{
// 		ID:        id,
// 		Name:      name,
// 		Count:     count,
// 		CreatedAt: createdAt,
// 		ExpiresAt: expiresAt,
// 		LockID:    lockID,
// 	}, nil
// }
