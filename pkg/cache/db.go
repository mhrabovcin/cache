package cache

import (
	"database/sql"
	"time"

	_ "github.com/lib/pq"
)

func NewDbSource(
	name string,
	driverName string,
	connStr string,
	query *DbQuery,
	frequency time.Duration,
	opts ...Option,
) StoppableSource {
	opts = append(opts, WithFetchFunc(
		dbFetchFunc(driverName, connStr, query),
		frequency,
	))
	return NewSource(
		name,
		opts...,
	)
}

// DbQuery is a way to pass SQL query into DbQuery cache source
type DbQuery struct {
	// Query is a SQL query that will be executed against the database. It must
	// be in format `SELECT key, value FROM ..`, where both `key` and `value`
	// colums must be `string` type.
	Query string

	// Query arguments
	Args []interface{}
}

func dbFetchFunc(
	driverName string,
	connStr string,
	query *DbQuery,
) func() (map[string]string, error) {
	db, err := sql.Open(driverName, connStr)
	return func() (map[string]string, error) {
		if err != nil {
			return nil, err
		}

		rows, err := db.Query(query.Query, query.Args...)
		if err != nil {
			return nil, err
		}

		data := map[string]string{}
		for rows.Next() {
			var key string
			var value string

			err := rows.Scan(&key, &value)
			if err != nil {
				return nil, err
			}

			data[key] = value
		}

		return data, nil
	}
}
