package mysql

import (
	"database/sql"
	"errors"
	. "github.com/89hmdys/orm"
	"reflect"
)

type client struct {
	Connection *sql.DB
}

func (this *client) Begin() (Tx, error) {
	tx, errTx := this.Connection.Begin()
	if errTx != nil {
		tx.Rollback()
		return nil, errTx
	}
	return &transaction{Tx: tx}, nil
}

func (this *client) Close() error {
	return this.Connection.Close()
}

func (this *client) Execute(sql string, sqlParameter interface{}) (sql.Result, error) {

	newSql, newArgs := analysisSQL(sql, sqlParameter)

	stmt, err := this.Connection.Prepare(newSql)
	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	return execute(stmt, newArgs)
}

func (this *client) Query(v interface{}, sql string, sqlParameter interface{}) error {

	newSql, newArgs := analysisSQL(sql, sqlParameter)

	vvPtr := reflect.ValueOf(v)

	if vvPtr.Kind() != reflect.Ptr {
		return errors.New("v is not ptr")
	}

	stmt, err := this.Connection.Prepare(newSql)
	if err != nil {
		return err
	}

	defer stmt.Close()

	rows, err := stmt.Query(newArgs...)

	if err != nil {
		return err
	}

	return convert(rows, vvPtr)
}
