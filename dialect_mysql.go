/*
 * Copyright 2018. bigpigeon. All rights reserved.
 * Use of this source code is governed by a MIT style
 * license that can be found in the LICENSE file.
 */

package toyorm

import (
	"database/sql"
	"fmt"
	"strings"
)

type MySqlDialect struct {
	DefaultDialect
}

func (dia MySqlDialect) SaveExecutor(db Executor, exec ExecValue, debugPrinter func(ExecValue, error)) (sql.Result, error) {
	query := exec.Query()
	result, err := db.Exec(query, exec.Args()...)
	debugPrinter(exec, err)
	return result, err
}

func (dia MySqlDialect) HasTable(model *Model) ExecValue {
	return DefaultExec{
		"SELECT count(*) FROM INFORMATION_SCHEMA.TABLES WHERE  table_schema = (SELECT DATABASE()) AND table_name = ?",
		[]interface{}{model.Name},
	}
}

func (dia MySqlDialect) CreateTable(model *Model, foreign map[string]ForeignKey) (execlist []ExecValue) {
	// lazy init model
	strList := []string{}
	// use to create foreign definition

	for _, sqlField := range model.GetSqlFields() {

		s := fmt.Sprintf("%s %s", sqlField.Column(), sqlField.SqlType())
		if sqlField.AutoIncrement() {
			s += " AUTO_INCREMENT"
		}
		if _default := sqlField.Default(); _default != "" {
			s += " DEFAULT " + _default
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

	for name, key := range foreign {
		f := model.GetFieldWithName(name)
		strList = append(strList,
			fmt.Sprintf("FOREIGN KEY (%s) REFERENCES `%s`(%s)", f.Column(), key.Model.Name, key.Field.Column()),
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
		uniqueIndexStrList = append(uniqueIndexStrList, fmt.Sprintf("CREATE UNIQUE INDEX %s ON `%s`(%s)", key, model.Name, strings.Join(fieldStrList, ",")))
	}
	for _, indexStr := range indexStrList {
		execlist = append(execlist, DefaultExec{indexStr, nil})
	}
	for _, indexStr := range uniqueIndexStrList {
		execlist = append(execlist, DefaultExec{indexStr, nil})
	}
	return
}

func (dia MySqlDialect) InsertExec(temp *BasicExec, model *Model, columnValues FieldValueList, condition *BasicExec, alias string) (ExecValue, error) {
	if temp == nil {
		temp = &BasicExec{"INSERT INTO $ModelDef($Columns) VALUES($Values)", nil}
	}
	execMap := dia.saveTemplate(temp, model, columnValues, condition, alias)
	basicExec, err := execMap.Render()
	if err != nil {
		return nil, err
	}
	return &DefaultExec{basicExec.query, basicExec.args}, nil
}

// replace will failure when have foreign key
func (dia MySqlDialect) SaveExec(temp *BasicExec, model *Model, columnValues FieldValueList, condition *BasicExec, alias string) (ExecValue, error) {
	if temp == nil {
		temp = &BasicExec{"INSERT INTO $ModelDef($Columns) VALUES($Values) ON DUPLICATE KEY UPDATE $Cas $UpdateValues", nil}
	}
	execMap := dia.saveTemplate(temp, model, columnValues, condition, alias)

	basicExec, err := execMap.Render()
	if err != nil {
		return nil, err
	}
	return &DefaultExec{basicExec.query, basicExec.args}, nil
}

func (dia MySqlDialect) saveTemplate(temp *BasicExec, model *Model, columnValues FieldValueList, condition *BasicExec, alias string) *SaveTemplate {
	var recordList []string
	var casField string
	saveColumnValues, casValue := GetSaveValues(model, columnValues)
	for _, r := range saveColumnValues {
		recordList = append(recordList, fmt.Sprintf("%[1]s = VALUES(%[1]s)", r.Column()))
	}
	if casValue != nil {
		casField = fmt.Sprintf("%[1]s = IF(%[1]s = VALUES(%[1]s) - 1, VALUES(%[1]s) , \"update failure\"),", casValue.Column())
	}

	columns, values := getInsertColumnExecAndValue(model, columnValues)
	execMap := SaveTemplate{
		TemplateBasic: TemplateBasic{
			Temp:  *temp,
			Model: model,
			Alias: alias,
			Quote: "`",
		},
		Columns:      columns,
		Values:       values,
		UpdateValues: BasicExec{strings.Join(recordList, ","), nil},
		Cas:          BasicExec{casField, nil},
		Condition:    *condition,
	}
	return &execMap
}
