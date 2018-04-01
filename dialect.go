/*
 * Copyright 2018. bigpigeon. All rights reserved.
 * Use of this source code is governed by a MIT style
 * license that can be found in the LICENSE file.
 */

package toyorm

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
)

type ForeignKey struct {
	Model *Model
	Field Field
}

type Executor interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	Query(query string, args ...interface{}) (*sql.Rows, error)
}

type Dialect interface {
	// some database like postgres not support LastInsertId, need QueryRow to get the return id
	InsertExecutor(Executor, ExecValue, func(string, string, error)) (sql.Result, error)
	HasTable(*Model) ExecValue
	CreateTable(*Model, map[string]ForeignKey) []ExecValue
	DropTable(*Model) ExecValue
	ConditionExec(search SearchList, limit, offset int, orderBy []Column, groupBy []Column) ExecValue
	FindExec(model *Model, columns []Column) ExecValue
	UpdateExec(*Model, []ColumnValue) ExecValue
	DeleteExec(*Model) ExecValue
	InsertExec(*Model, []ColumnValue) ExecValue
	ReplaceExec(*Model, []ColumnValue) ExecValue
	AddForeignKey(model, relationModel *Model, ForeignKeyField Field) ExecValue
	DropForeignKey(model *Model, ForeignKeyField Field) ExecValue
	CountExec(*Model) ExecValue
	SearchExec(search SearchList) ExecValue
	TemplateExec(BasicExec, map[string]BasicExec) (ExecValue, error)
}

type DefaultDialect struct{}

func (dia DefaultDialect) InsertExecutor(db Executor, exec ExecValue, debugPrinter func(string, string, error)) (sql.Result, error) {
	query := exec.Query()
	result, err := db.Exec(query, exec.Args()...)
	debugPrinter(query, exec.JsonArgs(), err)
	return result, err
}

func (dia DefaultDialect) DropTable(m *Model) ExecValue {
	return DefaultExec{fmt.Sprintf("DROP TABLE `%s`", m.Name), nil}
}

func (dia DefaultDialect) ConditionExec(search SearchList, limit, offset int, orderBy []Column, groupBy []Column) ExecValue {
	var exec ExecValue = DefaultExec{}
	if len(search) > 0 {
		searchExec := dia.SearchExec(search)
		exec = exec.Append(" WHERE "+searchExec.Source(), searchExec.Args()...)
	}
	if limit != 0 {
		exec = exec.Append(fmt.Sprintf(" LIMIT %d", limit))
	}
	if offset != 0 {
		exec = exec.Append(fmt.Sprintf(" OFFSET %d", offset))
	}
	if len(orderBy) > 0 {
		var __list []string
		for _, column := range orderBy {
			__list = append(__list, column.Column())
		}
		exec = exec.Append(" ORDER BY " + strings.Join(__list, ","))
	}
	if len(groupBy) > 0 {
		var list []string
		for _, column := range groupBy {
			list = append(list, column.Column())
		}
		exec = exec.Append(" GROUP BY " + strings.Join(list, ","))
	}
	return exec
}

func (dia DefaultDialect) SearchExec(s SearchList) ExecValue {
	var stack []ExecValue
	for i := 0; i < len(s); i++ {

		var exec ExecValue = DefaultExec{}
		switch s[i].Type {
		case ExprAnd:
			if len(stack) < 2 {
				panic(ErrInvalidSearchTree)
			}
			last1, last2 := stack[len(stack)-1], stack[len(stack)-2]
			stack = stack[:len(stack)-2]

			exec = exec.Append(
				fmt.Sprintf(last2.Source()),
				last2.Args()...,
			)
			exec = exec.Append(" AND "+last1.Source(), last1.Args()...)
		case ExprOr:
			if len(stack) < 2 {
				panic(ErrInvalidSearchTree)
			}
			last1, last2 := stack[len(stack)-1], stack[len(stack)-2]
			stack = stack[:len(stack)-2]
			exec = exec.Append(
				fmt.Sprintf("(%s", last2.Source()),
				last2.Args()...,
			)
			exec = exec.Append(
				fmt.Sprintf(" OR %s)", last1.Source()),
				last1.Args()...,
			)
		case ExprNot:
			if len(stack) < 1 {
				panic(ErrInvalidSearchTree)
			}
			last := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			exec = exec.Append(
				fmt.Sprintf("NOT(%s)", last.Source()),
				last.Args()...,
			)
		case ExprIgnore:
			continue

		case ExprEqual:
			exec = exec.Append(
				fmt.Sprintf("%s = ?", s[i].Val.Column()),
				s[i].Val.Value().Interface(),
			)

		case ExprNotEqual:
			exec = exec.Append(
				fmt.Sprintf("%s <> ?", s[i].Val.Column()),
				s[i].Val.Value().Interface(),
			)
		case ExprGreater:
			exec = exec.Append(
				fmt.Sprintf("%s > ?", s[i].Val.Column()),
				s[i].Val.Value().Interface(),
			)
		case ExprGreaterEqual:
			exec = exec.Append(
				fmt.Sprintf("%s >= ?", s[i].Val.Column()),
				s[i].Val.Value().Interface(),
			)

		case ExprLess:
			exec = exec.Append(
				fmt.Sprintf("%s < ?", s[i].Val.Column()),
				s[i].Val.Value().Interface(),
			)

		case ExprLessEqual:
			exec = exec.Append(
				fmt.Sprintf("%s <= ?", s[i].Val.Column()),
				s[i].Val.Value().Interface(),
			)
		case ExprBetween:
			vv := reflect.ValueOf(s[i].Val.Value().Interface())
			exec = exec.Append(
				fmt.Sprintf("%s BETWEEN ? AND ?", s[i].Val.Column()),
				vv.Index(0).Interface(), vv.Index(1).Interface(),
			)

		case ExprNotBetween:
			vv := reflect.ValueOf(s[i].Val.Value().Interface())
			exec = exec.Append(
				fmt.Sprintf("%s NOT BETWEEN ? AND ?", s[i].Val.Column()),
				vv.Index(0).Interface(), vv.Index(1).Interface(),
			)

		case ExprIn:
			vv := reflect.ValueOf(s[i].Val.Value().Interface())
			questionMarks := strings.TrimSuffix(strings.Repeat("?,", vv.Len()), ",")
			var args []interface{}
			for i := 0; i < vv.Len(); i++ {
				args = append(args, vv.Index(i).Interface())
			}
			exec = exec.Append(
				fmt.Sprintf("%s IN (%s)", s[i].Val.Column(), questionMarks),
				args...,
			)

		case ExprNotIn:
			vv := reflect.ValueOf(s[i].Val.Value().Interface())
			questionMarks := strings.TrimSuffix(strings.Repeat("?,", vv.Len()), ",")
			var args []interface{}
			for i := 0; i < vv.Len(); i++ {
				args = append(args, vv.Index(i).Interface())
			}
			exec = exec.Append(
				fmt.Sprintf("%s NOT IN (%s)", s[i].Val.Column(), questionMarks),
				args...,
			)

		case ExprLike:
			exec = exec.Append(
				fmt.Sprintf("%s LIKE ?", s[i].Val.Column()),
				s[i].Val.Value().Interface(),
			)

		case ExprNotLike:
			exec = exec.Append(
				fmt.Sprintf("%s NOT LIKE ?", s[i].Val.Column()),
				s[i].Val.Value().Interface(),
			)

		case ExprNull:
			exec = exec.Append(
				fmt.Sprintf("%s IS NULL", s[i].Val.Column()),
			)

		case ExprNotNull:
			exec = exec.Append(
				fmt.Sprintf("%s IS NOT NULL", s[i].Val.Column()),
			)

		}
		stack = append(stack, exec)

	}
	return stack[0]
}

func (dia DefaultDialect) FindExec(model *Model, columns []Column) ExecValue {
	var _list []string
	for _, column := range columns {
		_list = append(_list, column.Column())
	}
	var exec ExecValue = DefaultExec{}
	exec = exec.Append(fmt.Sprintf("SELECT %s FROM `%s`", strings.Join(_list, ","), model.Name))
	return exec
}

func (dia DefaultDialect) UpdateExec(model *Model, columnValues []ColumnValue) ExecValue {
	var recordList []string
	var args []interface{}
	for _, r := range columnValues {
		recordList = append(recordList, r.Column()+"=?")
		args = append(args, r.Value().Interface())
	}
	var exec ExecValue = DefaultExec{}
	exec = exec.Append(
		fmt.Sprintf("UPDATE `%s` SET %s", model.Name, strings.Join(recordList, ",")),
		args...,
	)
	return exec
}

func (dia DefaultDialect) DeleteExec(model *Model) (exec ExecValue) {
	return DefaultExec{fmt.Sprintf("DELETE FROM `%s`", model.Name), nil}
}

func (dia DefaultDialect) InsertExec(model *Model, columnValues []ColumnValue) ExecValue {
	fieldStr := ""
	qStr := ""
	var columnList []string
	var qList []string
	var args []interface{}
	for _, r := range columnValues {
		columnList = append(columnList, r.Column())
		qList = append(qList, "?")
		args = append(args, r.Value().Interface())
	}
	fieldStr += strings.Join(columnList, ",")
	qStr += strings.Join(qList, ",")

	var exec ExecValue = DefaultExec{}
	exec = exec.Append(
		fmt.Sprintf("INSERT INTO `%s`(%s) VALUES(%s)", model.Name, fieldStr, qStr),
		args...,
	)
	return exec
}

func (dia DefaultDialect) ReplaceExec(model *Model, columnValues []ColumnValue) ExecValue {
	fieldStr := ""
	qStr := ""
	columnList := []string{}
	qList := []string{}
	var args []interface{}
	for _, r := range columnValues {
		columnList = append(columnList, r.Column())
		qList = append(qList, "?")
		args = append(args, r.Value().Interface())
	}
	fieldStr += strings.Join(columnList, ",")
	qStr += strings.Join(qList, ",")

	var exec ExecValue = DefaultExec{}
	exec = exec.Append(
		fmt.Sprintf("REPLACE INTO `%s`(%s) VALUES(%s)", model.Name, fieldStr, qStr),
		args...,
	)
	return exec
}

func (dia DefaultDialect) AddForeignKey(model, relationModel *Model, ForeignKeyField Field) ExecValue {

	return DefaultExec{fmt.Sprintf(
		"ALTER TABLE `%s` ADD FOREIGN KEY (%s) REFERENCES `%s`(%s)",
		model.Name, ForeignKeyField.Column(),
		relationModel.Name, relationModel.GetOnePrimary().Column(),
	), nil}
}

func (dia DefaultDialect) DropForeignKey(model *Model, ForeignKeyField Field) ExecValue {
	return DefaultExec{fmt.Sprintf(
		"ALTER TABLE `%s` DROP FOREIGN KEY (%s)", model.Name, ForeignKeyField.Column(),
	), nil}

}

func (dia DefaultDialect) CountExec(model *Model) ExecValue {
	return DefaultExec{fmt.Sprintf("SELECT count(*) FROM `%s`", model.Name), nil}
}

func (dia DefaultDialect) TemplateExec(tExec BasicExec, execs map[string]BasicExec) (ExecValue, error) {
	exec, err := getTemplateExec(tExec, execs)
	if err != nil {
		return nil, err
	}
	return DefaultExec{exec.query, exec.args}, nil

}
