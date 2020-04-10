package ratelimit

import (
	"database/sql"
	"errors"
	"time"
)

// ResourceToken represents a token associated with a resource
type ResourceToken string

// Resource represents a rate limit resource
type Resource struct {
	ID        int
	Name      string
	CreatedAt time.Time
	Count     int
	ExpiresAt time.Time
	Token     ResourceToken
	IsExpired bool
}

// sets the is_expired flag on the resource in the database or
// returns error if unable
func unlockResource(id int) error {
	stmt, err := MySQL.Prepare(`
		UPDATE
			` + "`rate_limit_resource`" + `
		SET
			` + "`is_expired`" + ` = true
		WHERE
			` + "`id`" + ` = ?
		AND
			` + "`is_expired`" + ` = false;
	`)

	if err != nil {
		panic(err)
	}

	result, err := stmt.Exec(id)

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
				rate_limit_id,
				count,
				created_at,
				expires_at,
				token,
				is_expired
			)
		SELECT
			?, ?, ?, ?, ?, ? FROM dual
		WHERE
			(
				SELECT COUNT(*)
				FROM
					` + `rate_limit_resource` + `
				WHERE
					is_expired = false
			) <= ?;
	`)

	if err != nil {
		return r, err
	}

	// duration before lock expires
	duration := time.Second * time.Duration(m.resetInSeconds)
	results, err := stmt.Exec(
		m.id,
		1,
		time.Now().UTC(),
		time.Now().UTC().Add(duration),
		GenKsuid(),
		false,
		m.limit,
	)

	if err != nil {
		return r, err
	}

	insertedID, err := results.LastInsertId()

	if err != nil {
		return r, err
	}

	// Were we able to insert a new resource record?
	if insertedID == 0 {
		return r, errors.New("unable to insert rate limit resource")
	}

	return getResourceByID(insertedID)
}

// queries resource table by id and returns if found
func getResourceByID(rowID int64) (*Resource, error) {
	stmt, err := MySQL.Prepare(`
		SELECT
			r.id,
			r.rate_limit_id,
			r.count,
			r.created_at,
			r.expires_at,
			r.token,
			r.is_expired
		FROM
			` + "`rate_limit_resource`" + ` r
		WHERE
			r.id = ?;
	`)

	if err != nil {
		return nil, err
	}

	return sqlRowToResource(stmt.QueryRow(rowID))
}

// helper function that takes a resource sql row and transforms it into a resource
func sqlRowToResource(row *sql.Row) (r *Resource, err error) {
	var createdDatetime, expiresDatetime, name string
	var count, id, rateLimitID int
	var isExpired bool
	var token ResourceToken

	err = row.Scan(
		&id,
		&rateLimitID,
		&count,
		&createdDatetime,
		&expiresDatetime,
		&token,
		&isExpired,
	)

	if err != nil {
		return r, err
	}

	if err != nil {
		return r, err
	}

	createdAt, err := time.Parse("2006-01-02 15:04:05", createdDatetime)
	if err != nil {
		return nil, err
	}

	expiresAt, err := time.Parse("2006-01-02 15:04:05", expiresDatetime)
	if err != nil {
		return nil, err
	}

	return &Resource{
		ID:        id,
		Name:      name,
		Count:     count,
		CreatedAt: createdAt,
		ExpiresAt: expiresAt,
		Token:     token,
		IsExpired: isExpired,
	}, nil
}
