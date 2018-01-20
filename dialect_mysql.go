package toyorm

import (
	"fmt"
	"strings"
)

type MySqlDialect struct {
	DefaultDialect
}

func (dia MySqlDialect) HasTable(model *Model) ExecValue {
	return ExecValue{
		"SELECT count(*) FROM INFORMATION_SCHEMA.TABLES WHERE  table_schema = (SELECT DATABASE()) AND table_name = ?",
		[]interface{}{model.Name},
	}
}

func (dia MySqlDialect) CreateTable(model *Model) (execlist []ExecValue) {
	// lazy init model
	fieldStrList := []string{}
	for _, sqlField := range model.GetSqlFields() {

		s := fmt.Sprintf("%s %s", sqlField.Column(), sqlField.SqlType())
		if sqlField.AutoIncrement() {
			s += " AUTO_INCREMENT"
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
