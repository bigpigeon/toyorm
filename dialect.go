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

type ForeignKey struct {
	Model *Model
	Field Field
}

type Dialect interface {
	HasTable(*Model) ExecValue
	CreateTable(*Model, map[string]ForeignKey) []ExecValue
	DropTable(*Model) ExecValue
	ConditionExec(search SearchList, limit, offset int, orderBy []Column) ExecValue
	FindExec(model *Model, columns []Column) ExecValue
	UpdateExec(*Model, []ColumnValue) ExecValue
	DeleteExec(*Model) ExecValue
	InsertExec(*Model, []ColumnValue) ExecValue
	ReplaceExec(*Model, []ColumnValue) ExecValue
	GroupByExec(*Model, []Column) ExecValue
	AddForeignKey(model, relationModel *Model, ForeignKeyField Field) ExecValue
	DropForeignKey(model *Model, ForeignKeyField Field) ExecValue
}

type DefaultDialect struct{}

func (dia DefaultDialect) DropTable(m *Model) ExecValue {
	return ExecValue{fmt.Sprintf("DROP TABLE %s", m.Name), nil}
}

func (dia DefaultDialect) ConditionExec(search SearchList, limit, offset int, orderBy []Column) (exec ExecValue) {
	if len(search) > 0 {
		searchExec := search.ToExecValue()
		exec.Query += "WHERE " + searchExec.Query
		exec.Args = append(exec.Args, searchExec.Args...)
	}
	if limit != 0 {
		exec.Query += fmt.Sprintf(" LIMIT %d", limit)
	}
	if offset != 0 {
		exec.Query += fmt.Sprintf(" OFFSET %d", offset)
	}
	if len(orderBy) > 0 {
		exec.Query += fmt.Sprintf(" ORDER BY ")
		var __list []string
		for _, column := range orderBy {
			__list = append(__list, column.Column())
		}
		exec.Query += strings.Join(__list, ",")
	}
	return
}

func (dia DefaultDialect) FindExec(model *Model, columns []Column) (exec ExecValue) {
	var _list []string
	for _, column := range columns {
		_list = append(_list, column.Column())
	}
	exec.Query = fmt.Sprintf("SELECT %s FROM %s", strings.Join(_list, ","), model.Name)
	return
}

func (dia DefaultDialect) UpdateExec(model *Model, columnValues []ColumnValue) (exec ExecValue) {
	var recordList []string
	for _, r := range columnValues {
		recordList = append(recordList, r.Column()+"=?")
		exec.Args = append(exec.Args, r.Value().Interface())
	}
	exec.Query = fmt.Sprintf("UPDATE %s SET %s", model.Name, strings.Join(recordList, ","))
	return
}

func (dia DefaultDialect) DeleteExec(model *Model) (exec ExecValue) {
	exec.Query = fmt.Sprintf("DELETE FROM %s", model.Name)
	return
}

func (dia DefaultDialect) InsertExec(model *Model, columnValues []ColumnValue) (exec ExecValue) {
	fieldStr := ""
	qStr := ""
	exec = ExecValue{}
	columnList := []string{}
	qList := []string{}

	for _, r := range columnValues {
		columnList = append(columnList, r.Column())
		qList = append(qList, "?")
		exec.Args = append(exec.Args, r.Value().Interface())
	}
	fieldStr += strings.Join(columnList, ",")
	qStr += strings.Join(qList, ",")

	exec.Query = fmt.Sprintf("INSERT INTO %s(%s) VALUES(%s)", model.Name, fieldStr, qStr)
	return
}

func (dia DefaultDialect) ReplaceExec(model *Model, columnValues []ColumnValue) (exec ExecValue) {
	fieldStr := ""
	qStr := ""
	exec = ExecValue{}
	columnList := []string{}
	qList := []string{}

	for _, r := range columnValues {
		columnList = append(columnList, r.Column())
		qList = append(qList, "?")
		exec.Args = append(exec.Args, r.Value().Interface())
	}
	fieldStr += strings.Join(columnList, ",")
	qStr += strings.Join(qList, ",")

	exec.Query = fmt.Sprintf("Replace INTO %s(%s) VALUES(%s)", model.Name, fieldStr, qStr)
	return
}

func (dia DefaultDialect) GroupByExec(model *Model, columns []Column) (exec ExecValue) {
	if len(columns) > 0 {
		exec.Query = "GROUP BY "
		var list []string
		for _, column := range columns {
			list = append(list, column.Column())
		}
		exec.Query += strings.Join(list, ",")
	}
	return
}

func (dia DefaultDialect) AddForeignKey(model, relationModel *Model, ForeignKeyField Field) (exec ExecValue) {
	exec.Query = fmt.Sprintf(
		"ALTER TABLE %s ADD FOREIGN KEY (%s) REFERENCES %s(%s)",
		model.Name, ForeignKeyField.Column(),
		relationModel.Name, relationModel.GetOnePrimary().Column(),
	)
	return
}

func (dia DefaultDialect) DropForeignKey(model *Model, ForeignKeyField Field) (exec ExecValue) {
	exec.Query = fmt.Sprintf(
		"ALTER TABLE %s DROP FOREIGN KEY (%s)", model.Name, ForeignKeyField.Column(),
	)
	return
}
