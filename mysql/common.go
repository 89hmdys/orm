package mysql

import (
	"database/sql"
	"errors"
	"fmt"
	. "orm"
	"reflect"
	"regexp"
	"strings"
	"time"
)

const (
	true           int64  = 1
	defaultMaxIdle int    = 10
	defaultMaxOpen int    = 100
	driver         string = "mysql"
	prefix         string = "#"
)

var analysisSQLRegexp *regexp.Regexp

func init() {
	analysisSQLRegexp = regexp.MustCompile("#[a-zA-Z0-9]+")
}

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
					switch field.Type().Name() {
					case "bool":
						{
							val, ok := v.Interface().(int64)
							if !ok {
								errors.New("orm error:convert to bool fail,database field must be number")
							}
							field.SetBool(val == true)
						}
					case "Time":
						{
							fieldType, _ := elemType.FieldByName(k)

							format := fieldType.Tag.Get("format")

							if format == "" {
								errors.New("orm error:no specified date format")
							}

							datetime, err := time.ParseInLocation(format, string(v.Interface().([]byte)), time.Local)
							if err != nil {
								errors.New(fmt.Sprintf("orm error:parse time fail,%s", err.Error()))
							}

							field.Set(reflect.ValueOf(datetime))
						}
					default:
						{
							if v.Kind() == reflect.Slice {
								field.SetString(string(v.Interface().([]byte)))
							} else {
								field.Set(v)
							}
						}
					}
				} else {
					errors.New(fmt.Sprintf("orm error:%s set fail", k))
				}
			}
			return elemValue, nil
		}
	default:
		{
			return reflect.Value{}, errors.New("orm error:only support struct and map as element")
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
				colsLen := len(cols)
				if colsLen != 1 {
					return errors.New(fmt.Sprintf("orm error : expect 1 columns,but got %d.", colsLen))
				}
				if err := rows.Scan(scanArgs...); err != nil {
					return err
				}
				reflect.ValueOf(v).Elem().Set(reflect.ValueOf(values[0]))
			}
		}
	}
	return nil
}

func analysisSQL(sql string, args []interface{}) (string, []interface{}) {

	keys := analysisSQLRegexp.FindAllString(sql, -1)

	sql = analysisSQLRegexp.ReplaceAllString(sql, "?")

	argsLen := len(args)

	want := len(keys)

	switch argsLen {
	case 0:
		{
			break
		}
	case 1:
		{
			var argArray []interface{}

			argValue := reflect.ValueOf(args[0])

			switch argValue.Kind() {
			case reflect.Ptr:
				{
					panic("can not be ptr")
				}
			case reflect.Struct:
				{
					if len(keys) != argValue.NumField() {
						panic(fmt.Sprintf("sql: expected %d arguments, got %d", want, len(args)))
					}

					for _, v := range keys {
						argArray = append(argArray, argValue.FieldByName(strings.TrimPrefix(v, prefix)).Interface())
					}
				}
			case reflect.Map:
				{
					if len(keys) != argValue.Len() {
						panic(fmt.Sprintf("sql: expected %d arguments, got %d", want, len(args)))
					}

					for _, v := range keys {
						key := reflect.ValueOf(strings.TrimPrefix(v, prefix))
						argArray = append(argArray, argValue.MapIndex(key).Interface())
					}
				}
			default:
				{
					if len(keys) != argsLen {
						panic(fmt.Sprintf("sql: expected %d arguments, got %d", want, len(args)))
					}
					argArray = args
				}
			}
			return sql, argArray
		}
	default:
		{
			if len(keys) != argsLen {
				panic(fmt.Sprintf("sql: expected %d arguments, got %d", want, len(args)))
			}
			return sql, args
		}
	}
	return sql, args
}

func execute(stmt *sql.Stmt, args ...interface{}) (sql.Result, error) {

	res, err := stmt.Exec(args...)
	if err != nil {
		return nil, err
	}
	return res, nil
}
