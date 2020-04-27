package ratelimiter

// import (
// 	"errors"
// 	"time"

// 	"github.com/segmentio/ksuid"
// )

// // RequestToken represents a unqiue token associated with a Request
// type RequestToken string

// // Request represents a rate limit request
// type Request struct {
// 	Token       RequestToken
// 	RateLimitID int
// 	CreatedAt   time.Time
// 	Count       int
// 	ExpiresAt   time.Time
// 	InUse       bool
// }

// func genToken() RequestToken {
// 	return RequestToken(ksuid.New().String())
// }

// // sets the in_use flag to false on the request record or returns error if unable
// func releaseRequest(token RequestToken) error {
// 	stmt, err := MySQL.Prepare(`
// 		UPDATE
// 			` + "`rate_limit_request`" + `
// 		SET
// 			` + "`in_use`" + ` = false
// 		WHERE
// 			` + "`token`" + ` = ?
// 		AND
// 			` + "`in_use`" + ` = true;
// 	`)

// 	if err != nil {
// 		panic(err)
// 	}

// 	result, err := stmt.Exec(token)

// 	if err != nil {
// 		return err
// 	}

// 	updated, err := result.RowsAffected()

// 	if err != nil {
// 		return err
// 	}

// 	if updated != 1 {
// 		return errors.New("could not unlock request")
// 	}

// 	return nil
// }

// // 	`token` VARCHAR(255) NOT NULL,
// //  `rate_limit_id` INT NOT NULL,
// // 	`created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
// // 	`expires_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
// // 	`last_used_at` DATETIME DEFAULT NULL,
// // 	`in_use` BOOLEAN DEFAULT false,
// // 	`total_usage` INT,

// // NewRequest attempts to create a new rate limit request in the database.
// func NewRequest(m *Manager) (r *Request, err error) {
// 	stmt, err := MySQL.Prepare(`
// 		INSERT INTO rate_limit_request
// 			(
// 				token,
// 				rate_limit_id
// 			)
// 		SELECT
// 			?, ? FROM dual
// 		WHERE
// 			(
// 				SELECT COUNT(*)
// 				FROM
// 					` + `rate_limit_requeset` + `
// 				WHERE
// 					in_use = true
// 			) <= ?;
// 	`)

// 	if err != nil {
// 		return r, err
// 	}

// 	// duration before lock expires
// 	duration := time.Second * time.Duration(m.resetInSeconds)
// 	// generate a new lock token
// 	token := genToken()
// 	// default count to 1
// 	count := 1
// 	// default inUse to true
// 	inUse := true

// 	results, err := stmt.Exec(
// 		token,
// 		m.id,
// 		count,
// 		time.Now().UTC(),
// 		time.Now().UTC().Add(duration),
// 		inUse,
// 		m.limit,
// 	)

// 	if err != nil {
// 		return r, err
// 	}

// 	affecedCount, err := results.RowsAffected()

// 	if err != nil {
// 		return r, err
// 	}

// 	// Were we able to insert a new resource record?
// 	if affecedCount == 0 {
// 		return r, errors.New("unable to insert rate limit resource")
// 	}

// 	return getResourceByToken(token)
// }

// // queries resource table by id and returns if found
// func getResourceByToken(t ResourceToken) (r *Resource, err error) {
// 	stmt, err := MySQL.Prepare(`
// 		SELECT
// 			r.token,
// 			r.rate_limit_id,
// 			r.count,
// 			r.created_at,
// 			r.expires_at,
// 			r.in_use
// 		FROM
// 			` + "`rate_limit_resource`" + ` r
// 		WHERE
// 			r.token = ?;
// 	`)

// 	if err != nil {
// 		return r, err
// 	}

// 	sqlRow := stmt.QueryRow(string(t))

// 	var created, expires string
// 	var count, rateLimitID int
// 	var inUse bool
// 	var token ResourceToken

// 	err = sqlRow.Scan(
// 		&token,
// 		&rateLimitID,
// 		&count,
// 		&created,
// 		&expires,
// 		&inUse,
// 	)

// 	if err != nil {
// 		return r, err
// 	}

// 	return makeResourceFromSQL(
// 		token,
// 		rateLimitID,
// 		count,
// 		created,
// 		expires,
// 		inUse,
// 	)
// }

// func getResourcesInUse(rateLimitID int) (res []*Resource, err error) {
// 	// Load all resources in use from database
// 	stmt, err := MySQL.Prepare(`
// 		SELECT
// 			r.token,
// 			r.rate_limit_id,
// 			r.count,
// 			r.created_at,
// 			r.expires_at,
// 			r.in_use
// 		FROM
// 			` + "`rate_limit_resource`" + ` r
// 		WHERE
// 			r.rate_limit_id = ?
// 		AND
// 			r.in_use = true;
// 	`)

// 	if err != nil {
// 		panic(err)
// 	}

// 	sqlRows, err := stmt.Query(rateLimitID)

// 	if err != nil {
// 		return res, err
// 	}

// 	res = make([]*Resource, 0)

// 	for sqlRows.Next() {
// 		var created, expires string
// 		var count, rateLimitID int
// 		var inUse bool
// 		var token ResourceToken

// 		err = sqlRows.Scan(
// 			&token,
// 			&rateLimitID,
// 			&count,
// 			&created,
// 			&expires,
// 			&inUse,
// 		)

// 		if err != nil {
// 			return res, err
// 		}

// 		r, err := makeResourceFromSQL(
// 			token,
// 			rateLimitID,
// 			count,
// 			created,
// 			expires,
// 			inUse,
// 		)

// 		if err != nil {
// 			return res, err
// 		}

// 		res = append(res, r)
// 	}

// 	return res, err
// }

// func makeResourceFromSQL(
// 	token ResourceToken,
// 	rateLimitID int,
// 	count int,
// 	created string,
// 	expires string,
// 	inUse bool) (r *Resource, err error) {
// 	createdAt, err := time.Parse("2006-01-02 15:04:05", created)
// 	if err != nil {
// 		return nil, err
// 	}

// 	expiresAt, err := time.Parse("2006-01-02 15:04:05", expires)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return &Resource{
// 		Token:       token,
// 		RateLimitID: rateLimitID,
// 		Count:       count,
// 		CreatedAt:   createdAt,
// 		ExpiresAt:   expiresAt,
// 		InUse:       inUse,
// 	}, nil
// }
