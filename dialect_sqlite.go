/*
 * Copyright 2018. bigpigeon. All rights reserved.
 * Use of this source code is governed by a MIT style
 * license that can be found in the LICENSE file.
 */

package toyorm

import (
	"fmt"
	"strings"
)

type Sqlite3Dialect struct {
	DefaultDialect
}

func (dia Sqlite3Dialect) HasTable(model *Model) ExecValue {
	return DefaultExec{
		`select count(*) from sqlite_master where type="table" and name= ? `,
		[]interface{}{model.Name},
	}
}

func (dia Sqlite3Dialect) CreateTable(model *Model, foreign map[string]ForeignKey) (execlist []ExecValue) {
	// lazy init model
	strList := []string{}

	// for strange auto_increment syntax to do strange codition
	isSinglePrimary := len(model.GetPrimary()) == 1
	// use to create foreign definition
	for _, sqlField := range model.GetSqlFields() {

		s := fmt.Sprintf("%s %s", sqlField.Column(), sqlField.SqlType())
		if isSinglePrimary && sqlField.Name() == model.GetOnePrimary().Name() {
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

		strList = append(strList, s)
	}

	if isSinglePrimary == false {
		primaryStrList := []string{}
		for _, p := range model.GetPrimary() {
			primaryStrList = append(primaryStrList, p.Column())
		}
		strList = append(strList, fmt.Sprintf("PRIMARY KEY(%s)", strings.Join(primaryStrList, ",")))

	}

	for name, key := range foreign {
		f := model.GetFieldWithName(name)
		strList = append(strList,
			fmt.Sprintf("FOREIGN KEY (%s) REFERENCES %s(%s)", f.Column(), key.Model.Name, key.Field.Column()),
		)
	}

	sqlStr := fmt.Sprintf("CREATE TABLE `%s` (%s)",
		model.Name,
		strings.Join(strList, ","),
	)
	execlist = append(execlist, DefaultExec{sqlStr, nil})

	indexStrList := []string{}
	for key, fieldList := range model.GetIndexMap() {
		fieldStrList := []string{}
		for _, f := range fieldList {
			fieldStrList = append(fieldStrList, f.Column())
		}
		indexStrList = append(indexStrList, fmt.Sprintf("CREATE INDEX %s ON `%s`(%s)", key, model.Name, strings.Join(fieldStrList, ",")))
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
		execlist = append(execlist, DefaultExec{indexStr, nil})
	}
	for _, indexStr := range uniqueIndexStrList {
		execlist = append(execlist, DefaultExec{indexStr, nil})
	}
	return
}
