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

func (this *transaction) Execute(sql string, sqlParameter interface{}) (sql.Result, error) {

	newSql, newArgs := analysisSQL(sql, sqlParameter)

	stmt, err := this.Tx.Prepare(newSql)
	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	return execute(stmt, newArgs)
}

func (this *transaction) Query(v interface{}, sql string, sqlParameter interface{}) error {

	newSql, newArgs := analysisSQL(sql, sqlParameter)

	vv := reflect.ValueOf(v)

	if vv.Kind() != reflect.Ptr {
		return errors.New("v is not ptr")
	}

	stmt, err := this.Tx.Prepare(newSql)
	if err != nil {
		return err
	}

	defer stmt.Close()

	rows, err := stmt.Query(newArgs...)

	if err != nil {
		return err
	}

	return convert(rows, vv)
}
