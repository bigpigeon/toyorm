package toyorm

import (
	"fmt"
	"strings"
)

type Sqlite3Dialect struct {
	DefaultDialect
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
	isSinglePrimary := len(model.GetPrimary()) == 1
	for _, sqlField := range model.GetSqlFields() {

		s := fmt.Sprintf("%s %s", sqlField.Column(), sqlField.SqlType())
		if isSinglePrimary && sqlField == model.PrimaryFields[0] {
			s += " PRIMARY KEY"
		}
		if sqlField.AutoIncrement() {
			s += " AUTOINCREMENT"
		}
		for k, v := range sqlField.Attrs() {
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
			primaryStrList = append(primaryStrList, p.Column())
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
			fieldStrList = append(fieldStrList, f.Column())
		}
		indexStrList = append(indexStrList, fmt.Sprintf("CREATE INDEX %s ON %s(%s)", key, model.Name, strings.Join(fieldStrList, ",")))
	}
	uniqueIndexStrList := []string{}
	for key, fieldList := range model.GetUniqueIndexMap() {
		fieldStrList := []string{}
		for _, f := range fieldList {
			fieldStrList = append(fieldStrList, f.Column())
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
