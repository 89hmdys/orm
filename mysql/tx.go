package mysql

import (
	"database/sql"
	"errors"
	"reflect"
)

type transaction struct {
	Tx      *sql.Tx
	success bool
}

func (this *transaction) Close() {
	if this.success {
		this.Tx.Commit()
	} else {
		this.Tx.Rollback()
	}
}

func (this *transaction) Fail() {
	this.success = false
}

func (this *transaction) Execute(query string, args ...interface{}) (sql.Result, error) {
	stmt, err := this.Tx.Prepare(query)
	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	return execute(stmt, args)
}

func (this *transaction) Query(v interface{}, sql string, args ...interface{}) error {

	vt := reflect.TypeOf(v)

	if vt.Kind() != reflect.Ptr {
		return errors.New("v is not ptr")
	}

	stmt, err := this.Tx.Prepare(sql)
	if err != nil {
		return err
	}

	defer stmt.Close()

	rows, err := stmt.Query(args...)

	if err != nil {
		return err
	}

	return convert(rows, v)
}
