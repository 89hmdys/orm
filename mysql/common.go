package mysql

import (
	"database/sql"
	"errors"
	"fmt"
	. "orm"
	"reflect"
)

const (
	defaultMaxIdle int    = 10
	defaultMaxOpen int    = 100
	driver         string = "mysql"
)

func New(connStr string, maxIdle, maxOpen int) (Client, error) {
	connection, err := sql.Open(driver, connStr)
	if err != nil {
		return nil, err
	}
	connection.SetMaxIdleConns(maxIdle)
	connection.SetMaxOpenConns(maxOpen)
	return &client{Connection: connection}, nil
}

func NewDefault(connStr string) (Client, error) {
	connection, err := sql.Open(driver, connStr)
	if err != nil {
		return nil, err
	}
	connection.SetMaxIdleConns(defaultMaxIdle)
	connection.SetMaxOpenConns(defaultMaxOpen)
	return &client{Connection: connection}, nil
}

func buildElement(elemType reflect.Type, keys []string, values []interface{}) (reflect.Value, error) {

	//TODO 还对Struct需要支持bool,datetime,关于datetime，考虑在属性后新增tag写明转换格式

	switch elemType.Kind() {
	case reflect.Map:
		{
			elemValue := reflect.MakeMap(elemType)

			for i, k := range keys {

				v := reflect.ValueOf(values[i])

				if v.Kind() == reflect.Slice {
					elemValue.SetMapIndex(reflect.ValueOf(k), reflect.ValueOf(string(v.Interface().([]byte))))
				} else {
					elemValue.SetMapIndex(reflect.ValueOf(k), v)
				}
			}

			return elemValue, nil
		}
	case reflect.Struct:
		{
			elemValue := reflect.New(elemType)
			for i, k := range keys {

				v := reflect.ValueOf(values[i])

				field := elemValue.FieldByName(k)

				if field.CanSet() {
					if v.Kind() == reflect.Slice {
						field.SetString(string(v.Interface().([]byte)))
					} else {
						field.Set(v)
					}
				} else {
					errors.New(fmt.Sprintf("%s can not set", k))
				}
			}
			return elemValue, nil
		}
	default:
		{
			return reflect.Value{}, errors.New("struct map only")
		}
	}
}

func convert(rows *sql.Rows, v interface{}) error {

	vt := reflect.TypeOf(v)

	cols, err := rows.Columns()
	if err != nil {
		return err
	}

	values := make([]interface{}, len(cols))

	scanArgs := make([]interface{}, len(values))

	for i := range values {
		scanArgs[i] = &values[i]
	}

	switch vt.Elem().Kind() {
	case reflect.Slice:
		{
			elemType := vt.Elem().Elem()
			sliceType := reflect.SliceOf(elemType)
			newSlice := reflect.MakeSlice(sliceType, 0, 0)

			for rows.Next() {
				if err := rows.Scan(scanArgs...); err != nil {
					return err
				}
				elemValue, err := buildElement(elemType, cols, values)

				if err != nil {
					return err
				}

				newSlice = reflect.Append(newSlice, elemValue)
			}
			reflect.ValueOf(v).Elem().Set(newSlice)
		}
	case reflect.Map, reflect.Struct:
		{
			if rows.Next() {
				if err := rows.Scan(scanArgs...); err != nil {
					return err
				}
				elemValue, err := buildElement(vt, cols, values)

				if err != nil {
					return err
				}
				reflect.ValueOf(v).Elem().Set(elemValue)
			}
		}
	default:
		{
			if rows.Next() {
				if err := rows.Scan(scanArgs...); err != nil {
					return err
				}

				if len(cols) == 1 {
					reflect.ValueOf(v).Elem().Set(reflect.ValueOf(values[0]))
				} else {
					return errors.New("para have to be 1")
				}
			}
		}
	}
	return nil
}

func execute(stmt *sql.Stmt, args ...interface{}) (sql.Result, error) {
	res, err := stmt.Exec(args...)
	if err != nil {
		return nil, err
	}
	return res, nil
}
