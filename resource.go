package ratelimit

import (
	"errors"
	"time"

	"github.com/segmentio/ksuid"
)

// ResourceToken represents a token associated with a resource
type ResourceToken string

// Resource represents a rate limit resource
type Resource struct {
	Token       ResourceToken
	RateLimitID int
	CreatedAt   time.Time
	Count       int
	ExpiresAt   time.Time
	InUse       bool
}

func genToken() ResourceToken {
	return ResourceToken(ksuid.New().String())
}

// sets the in_use flag to true on the resource in the database or
// returns error if unable
func unlockResource(token ResourceToken) error {
	stmt, err := MySQL.Prepare(`
		UPDATE
			` + "`rate_limit_resource`" + `
		SET
			` + "`in_use`" + ` = false
		WHERE
			` + "`token`" + ` = ?
		AND
			` + "`in_use`" + ` = true;
	`)

	if err != nil {
		panic(err)
	}

	result, err := stmt.Exec(token)

	if err != nil {
		return err
	}

	updated, err := result.RowsAffected()

	if err != nil {
		return err
	}

	if updated != 1 {
		return errors.New("could not unlock resource")
	}

	return nil
}

// NewResource attempts to insert a new rate limit resource record
// into the data, and returns the resource if successful.
func NewResource(m *Manager) (r *Resource, err error) {
	stmt, err := MySQL.Prepare(`
		INSERT INTO rate_limit_resource
			(
				token,
				rate_limit_id,
				count,
				created_at,
				expires_at,
				in_use
			)
		SELECT
			?, ?, ?, ?, ?, ? FROM dual
		WHERE
			(
				SELECT COUNT(*)
				FROM
					` + `rate_limit_resource` + `
				WHERE
					in_use = true
			) <= ?;
	`)

	if err != nil {
		return r, err
	}

	// duration before lock expires
	duration := time.Second * time.Duration(m.resetInSeconds)
	// generate a new lock token
	token := genToken()
	// default count to 1
	count := 1
	// default inUse to true
	inUse := true

	results, err := stmt.Exec(
		token,
		m.id,
		count,
		time.Now().UTC(),
		time.Now().UTC().Add(duration),
		inUse,
		m.limit,
	)

	if err != nil {
		return r, err
	}

	affecedCount, err := results.RowsAffected()

	if err != nil {
		return r, err
	}

	// Were we able to insert a new resource record?
	if affecedCount == 0 {
		return r, errors.New("unable to insert rate limit resource")
	}

	return getResourceByToken(token)
}

// queries resource table by id and returns if found
func getResourceByToken(t ResourceToken) (r *Resource, err error) {
	stmt, err := MySQL.Prepare(`
		SELECT
			r.token,
			r.rate_limit_id,
			r.count,
			r.created_at,
			r.expires_at,
			r.in_use
		FROM
			` + "`rate_limit_resource`" + ` r
		WHERE
			r.token = ?;
	`)

	if err != nil {
		return r, err
	}

	sqlRow := stmt.QueryRow(string(t))

	var created, expires string
	var count, rateLimitID int
	var inUse bool
	var token ResourceToken

	err = sqlRow.Scan(
		&token,
		&rateLimitID,
		&count,
		&created,
		&expires,
		&inUse,
	)

	if err != nil {
		return r, err
	}

	return makeResourceFromSQL(
		token,
		rateLimitID,
		count,
		created,
		expires,
		inUse,
	)
}

func getResourcesInUse(rateLimitID int) (res []*Resource, err error) {
	// Load all resources in use from database
	stmt, err := MySQL.Prepare(`
		SELECT
			r.token,
			r.rate_limit_id,
			r.count,
			r.created_at,
			r.expires_at,
			r.in_use
		FROM 
			` + "`rate_limit_resource`" + ` r
		WHERE 
			r.rate_limit_id = ?
		AND
			r.in_use = true;
	`)

	if err != nil {
		panic(err)
	}

	sqlRows, err := stmt.Query(rateLimitID)

	if err != nil {
		return res, err
	}

	res = make([]*Resource, 0)

	for sqlRows.Next() {
		var created, expires string
		var count, rateLimitID int
		var inUse bool
		var token ResourceToken

		err = sqlRows.Scan(
			&token,
			&rateLimitID,
			&count,
			&created,
			&expires,
			&inUse,
		)

		if err != nil {
			return res, err
		}

		r, err := makeResourceFromSQL(
			token,
			rateLimitID,
			count,
			created,
			expires,
			inUse,
		)

		if err != nil {
			return res, err
		}

		res = append(res, r)
	}

	return res, err
}

func makeResourceFromSQL(
	token ResourceToken,
	rateLimitID int,
	count int,
	created string,
	expires string,
	inUse bool) (r *Resource, err error) {
	createdAt, err := time.Parse("2006-01-02 15:04:05", created)
	if err != nil {
		return nil, err
	}

	expiresAt, err := time.Parse("2006-01-02 15:04:05", expires)
	if err != nil {
		return nil, err
	}

	return &Resource{
		Token:       token,
		RateLimitID: rateLimitID,
		Count:       count,
		CreatedAt:   createdAt,
		ExpiresAt:   expiresAt,
		InUse:       inUse,
	}, nil
}
