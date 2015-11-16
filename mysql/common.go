package mysql

import (
	"database/sql"
	"errors"
	"fmt"
	. "github.com/89hmdys/orm"
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

	boolean  string = "bool"
	datetime string = "Time"
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

func setValue(elem reflect.Value, value reflect.Value) error {
	switch elem.Type().Name() {
	case boolean:
		{
			val, ok := value.Interface().(int64)
			if !ok {
				errors.New("orm error:convert to bool fail,database field must be number")
			}
			elem.SetBool(val == true)
		}
	case datetime:
		{
			return errors.New("orm error:unsupport time.Time for single v")
		}
	default:
		{
			//TODO 不支持int8 int16 int32 int 预测同样情况也存在于float8 float16 float32 float,需要支持
			if value.Kind() == reflect.Slice {
				elem.SetString(string(value.Interface().([]byte)))
			} else {
				elem.Set(value)
			}
		}
	}
	return nil
}

func buildElement(elem reflect.Value, keys []string, values []interface{}) error {

	switch elem.Kind() {
	case reflect.Map:
		{
			for i, k := range keys {

				v := reflect.ValueOf(values[i])

				if v.Kind() == reflect.Slice {
					elem.SetMapIndex(reflect.ValueOf(k), reflect.ValueOf(string(v.Interface().([]byte))))
				} else {
					elem.SetMapIndex(reflect.ValueOf(k), v)
				}
			}
		}
	case reflect.Struct:
		{
			for i, k := range keys {

				v := reflect.ValueOf(values[i])

				field := elem.FieldByName(k)

				if field.CanSet() {
					switch field.Type().Name() {
					case datetime:
						{
							fieldType, _ := elem.Type().FieldByName(k)

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
							err := setValue(elem, v)
							if err != nil {
								return err
							}
						}
					}
				} else {
					errors.New(fmt.Sprintf("orm error:%s set fail", k))
				}
			}
		}
	default:
		{
			err := setValue(elem, reflect.ValueOf(values[0]))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func convert(rows *sql.Rows, vvPtr reflect.Value) error {

	cols, err := rows.Columns()
	if err != nil {
		return err
	}

	values := make([]interface{}, len(cols))

	scanArgs := make([]interface{}, len(values))

	for i := range values {
		scanArgs[i] = &values[i]
	}

	switch vvPtr.Elem().Kind() {
	case reflect.Slice:
		{
			elemType := vvPtr.Type().Elem().Elem()
			sliceType := reflect.SliceOf(elemType)
			newSlice := reflect.MakeSlice(sliceType, 0, 0)

			for rows.Next() {
				if err := rows.Scan(scanArgs...); err != nil {
					return err
				}

				var elem reflect.Value
				switch elemType.Kind() {
				case reflect.Map:
					{
						elem = reflect.MakeMap(elemType)
					}
				default:
					{
						elem = reflect.New(elemType).Elem()
					}
				}

				err := buildElement(elem, cols, values)
				if err != nil {
					return err
				}

				newSlice = reflect.Append(newSlice, elem)
			}
			vvPtr.Elem().Set(newSlice)
		}
	case reflect.Map, reflect.Struct:
		{
			if rows.Next() {
				if err := rows.Scan(scanArgs...); err != nil {
					return err
				}
				err := buildElement(vvPtr.Elem(), cols, values)

				if err != nil {
					return err
				}
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
				err := buildElement(vvPtr.Elem(), cols, values)

				if err != nil {
					return err
				}
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
					panic("orm error:args can not be ptr")
				}
			case reflect.Struct:
				{
					if len(keys) != argValue.NumField() {
						panic(fmt.Sprintf("orm error:expected %d arguments, got %d", want, len(args)))
					}

					for _, v := range keys {
						argArray = append(argArray, argValue.FieldByName(strings.TrimPrefix(v, prefix)).Interface())
					}
				}
			case reflect.Map:
				{
					if len(keys) != argValue.Len() {
						panic(fmt.Sprintf("orm error:expected %d arguments, got %d", want, len(args)))
					}

					for _, v := range keys {
						key := reflect.ValueOf(strings.TrimPrefix(v, prefix))
						argArray = append(argArray, argValue.MapIndex(key).Interface())
					}
				}
			default:
				{
					if len(keys) != argsLen {
						panic(fmt.Sprintf("orm error:expected %d arguments, got %d", want, len(args)))
					}
					argArray = args
				}
			}
			return sql, argArray
		}
	default:
		{
			if len(keys) != argsLen {
				panic(fmt.Sprintf("orm error:expected %d arguments, got %d", want, len(args)))
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
