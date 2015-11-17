package orm

import (
	"database/sql"
)

type Tx interface {
	SqlCommand
	Fail()
	Close()
}

type SqlCommand interface {
	Execute(query string, sqlParameter interface{}) (sql.Result, error)
	Query(ptr interface{}, sql string, sqlParameter interface{}) error
}

type SqlParameter interface {
	Set(key string, value interface{}) SqlParameter
}

type Client interface {
	SqlCommand
	Begin() (Tx, error)
	Close() error
}
