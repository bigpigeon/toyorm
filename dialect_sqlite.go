package toyorm

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"time"
)

type Sqlite3Dialect struct {
}

func (dia Sqlite3Dialect) HasTable(model *Model) ExecValue {
	return ExecValue{
		`select count(*) from sqlite_master where type="table" and name= ? `,
		[]interface{}{model.Name},
	}
}

func (dia Sqlite3Dialect) CreateTable(model *Model) (execlist []ExecValue) {
	// lazy init model
	fieldStrList := []string{}

	// for strange auto_increment syntax to do strange codition
	isSinglePrimary := len(model.PrimaryFields) == 1
	for _, sqlField := range model.SqlFields {
		s := fmt.Sprintf("%s %s", sqlField.Name, sqlField.Type)
		if isSinglePrimary && sqlField == model.PrimaryFields[0] {
			s += " PRIMARY KEY"
		}
		if sqlField.AutoIncrement {
			s += " AUTOINCREMENT"
		}
		for k, v := range sqlField.CommonAttr {
			if v == "" {
				s += " " + k
			} else {
				s += " " + fmt.Sprintf("%s=%s", k, v)
			}
		}

		fieldStrList = append(fieldStrList, s)
	}

	if isSinglePrimary == false {
		primaryStrList := []string{}
		for _, p := range model.GetPrimary() {
			primaryStrList = append(primaryStrList, p.Name)
		}
		sqlStr := fmt.Sprintf("CREATE TABLE %s (%s, PRIMARY KEY(%s))",
			model.Name,
			strings.Join(fieldStrList, ","),
			strings.Join(primaryStrList, ","),
		)
		execlist = append(execlist, ExecValue{sqlStr, nil})
	} else {
		sqlStr := fmt.Sprintf("CREATE TABLE %s (%s)",
			model.Name,
			strings.Join(fieldStrList, ","),
		)
		execlist = append(execlist, ExecValue{sqlStr, nil})
	}

	indexStrList := []string{}
	for key, fieldList := range model.GetIndexMap() {
		fieldStrList := []string{}
		for _, f := range fieldList {
			fieldStrList = append(fieldStrList, f.Name)
		}
		indexStrList = append(indexStrList, fmt.Sprintf("CREATE INDEX %s ON %s(%s)", key, model.Name, strings.Join(fieldStrList, ",")))
	}
	uniqueIndexStrList := []string{}
	for key, fieldList := range model.GetUniqueIndexMap() {
		fieldStrList := []string{}
		for _, f := range fieldList {
			fieldStrList = append(fieldStrList, f.Name)
		}
		uniqueIndexStrList = append(uniqueIndexStrList, fmt.Sprintf("CREATE UNIQUE INDEX %s ON %s(%s)", key, model.Name, strings.Join(fieldStrList, ",")))
	}

	for _, indexStr := range indexStrList {
		execlist = append(execlist, ExecValue{indexStr, nil})
	}
	for _, indexStr := range uniqueIndexStrList {
		execlist = append(execlist, ExecValue{indexStr, nil})
	}
	return
}

func (dia Sqlite3Dialect) ToSqlType(_type reflect.Type) (sqlType string) {
	switch _type.Kind() {
	case reflect.Ptr:
		sqlType = dia.ToSqlType(_type.Elem())
	case reflect.Bool:
		sqlType = "BOOLEAN"
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Int, reflect.Uint:
		sqlType = "INTEGER"
	case reflect.Int64, reflect.Uint64:
		sqlType = "BIGINT"
	case reflect.Float32, reflect.Float64:
		sqlType = "FLOAT"
	case reflect.String:
		sqlType = "VARCHAR(255)"
	case reflect.Struct:
		if _, ok := reflect.New(_type).Elem().Interface().(time.Time); ok {
			sqlType = "TIMESTAMP"
		} else if _, ok := reflect.New(_type).Elem().Interface().(sql.NullBool); ok {
			sqlType = "BOOLEAN"
		} else if _, ok := reflect.New(_type).Elem().Interface().(sql.NullInt64); ok {
			sqlType = "BIGINT"
		} else if _, ok := reflect.New(_type).Elem().Interface().(sql.NullString); ok {
			sqlType = "VARCHAR(255)"
		} else if _, ok := reflect.New(_type).Elem().Interface().(sql.NullFloat64); ok {
			sqlType = "FLOAT"
		} else if _, ok := reflect.New(_type).Elem().Interface().(sql.RawBytes); ok {
			sqlType = "VARCHAR(255)"
		}
	default:
		if _, ok := reflect.New(_type).Elem().Interface().([]byte); ok {
			sqlType = "VARCHAR(255)"
		}
	}
	return
}
