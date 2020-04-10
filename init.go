package ratelimit

import (
	"database/sql"
	// mysql driver
	_ "github.com/go-sql-driver/mysql"
)

var (
	// MySQL client is the sql db handle
	MySQL *sql.DB
)

func init() {
	var err error

	MySQL, err = sql.Open("mysql", "justin:password@tcp(localhost:3306)/rate_limiter")

	if err != nil {
		panic(err)
	}
}
