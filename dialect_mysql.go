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
	strList := []string{}
	// use to create foreign definition
	var foreignKeyList []Field
	for _, sqlField := range model.GetSqlFields() {
		if sqlField.ForeignModel() != nil {
			foreignKeyList = append(foreignKeyList, sqlField)
		}
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
		strList = append(strList, s)
	}
	var primaryStrList []string
	for _, p := range model.GetPrimary() {
		primaryStrList = append(primaryStrList, p.Column())
	}
	strList = append(strList, fmt.Sprintf("PRIMARY KEY(%s)", strings.Join(primaryStrList, ",")))

	for _, f := range foreignKeyList {
		strList = append(strList,
			fmt.Sprintf("FOREIGN KEY (%s) REFERENCES %s(%s)", f.Column(), f.ForeignModel().Name, f.ForeignModel().GetOnePrimary().Column()),
		)
	}

	sqlStr := fmt.Sprintf("CREATE TABLE %s (%s)",
		model.Name,
		strings.Join(strList, ","),
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
