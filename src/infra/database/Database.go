//go:generate mockgen -source Database.go -destination mock/Database_mock.go -package mock
package database

import (
	"context"
	"database/sql"
)

type Database interface {
	Many(ctx context.Context, query string, params ...interface{}) (*sql.Rows, error)
	One(ctx context.Context, query string, params ...interface{}) *sql.Row
	Exec(ctx context.Context, query string, params ...interface{}) (sql.Result, error)
	Close()
}
