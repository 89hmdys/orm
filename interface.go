package orm

import "database/sql"

type Tx interface {
	SqlCommand
	Fail()
	Close()
}

type SqlCommand interface {
	Execute(query string, args ...interface{}) (sql.Result, error)
	Query(ptr interface{}, sql string, args ...interface{}) error
}

type Client interface {
	SqlCommand
	Begin() (Tx, error)
	Close() error
}
