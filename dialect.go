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
	InsertExecutor(Executor, ExecValue, func(ExecValue, error)) (sql.Result, error)
	// sqlite3/postgresql use RowsAffected to check success or failure, but mysql can't,
	// because it's RowsAffected is zero when update value not change
	SaveExecutor(Executor, ExecValue, func(ExecValue, error)) (sql.Result, error)
	HasTable(*Model) ExecValue
	CreateTable(*Model, map[string]ForeignKey) []ExecValue
	DropTable(*Model) ExecValue
	ConditionExec(search SearchList, limit, offset int, orderBy []Column, groupBy []Column) ExecValue
	FindExec(model *Model, columns []Column, alias string) ExecValue
	UpdateExec(*Model, []ColumnValue) ExecValue
	DeleteExec(*Model) ExecValue
	InsertExec(*Model, []ColumnNameValue) ExecValue
	SaveExec(*Model, []ColumnNameValue) ExecValue
	AddForeignKey(model, relationModel *Model, ForeignKeyField Field) ExecValue
	DropForeignKey(model *Model, ForeignKeyField Field) ExecValue
	CountExec(model *Model, alias string) ExecValue
	SearchExec(search SearchList) ExecValue
	TemplateExec(BasicExec, map[string]BasicExec) (ExecValue, error)
	JoinExec(*JoinSwap) ExecValue
}

type DefaultDialect struct{}

func (dia DefaultDialect) SaveExecutor(db Executor, exec ExecValue, debugPrinter func(ExecValue, error)) (sql.Result, error) {
	query := exec.Query()
	result, err := db.Exec(query, exec.Args()...)
	// use RowsAffected to check save success or failure
	if err == nil {
		affected, e := result.RowsAffected()
		// e must be nil
		if e != nil {
			panic(e)
		}
		if affected == 0 {
			err = ErrSaveFailure{}
		}
	}
	debugPrinter(exec, err)
	return result, err
}

func (dia DefaultDialect) InsertExecutor(db Executor, exec ExecValue, debugPrinter func(ExecValue, error)) (sql.Result, error) {
	query := exec.Query()
	result, err := db.Exec(query, exec.Args()...)
	debugPrinter(exec, err)
	return result, err
}

// use to test
func (dia DefaultDialect) HasTable(model *Model) ExecValue {
	return DefaultExec{
		"SELECT count(*) FROM INFORMATION_SCHEMA.TABLES WHERE  table_schema = (SELECT DATABASE()) AND table_name = ?",
		[]interface{}{model.Name},
	}
}

// use to test
func (dia DefaultDialect) CreateTable(model *Model, foreign map[string]ForeignKey) (execlist []ExecValue) {
	// lazy init model
	strList := []string{}
	// use to create foreign definition

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

func (dia DefaultDialect) DropTable(m *Model) ExecValue {
	return DefaultExec{fmt.Sprintf("DROP TABLE `%s`", m.Name), nil}
}

func (dia DefaultDialect) ConditionExec(search SearchList, limit, offset int, orderBy []Column, groupBy []Column) ExecValue {
	var exec ExecValue = DefaultExec{}
	if len(search) > 0 {
		searchExec := dia.SearchExec(search)
		exec = exec.Append(" WHERE "+searchExec.Source(), searchExec.Args()...)
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

	if limit != 0 {
		exec = exec.Append(fmt.Sprintf(" LIMIT %d", limit))
	}
	if offset != 0 {
		exec = exec.Append(fmt.Sprintf(" OFFSET %d", offset))
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

func (dia DefaultDialect) FindExec(model *Model, columns []Column, alias string) ExecValue {
	var _list []string
	for _, column := range columns {
		_list = append(_list, column.Column())
	}
	var exec ExecValue = DefaultExec{}
	if alias != "" {
		exec = exec.Append(fmt.Sprintf("SELECT %s FROM `%s` as `%s`", strings.Join(_list, ","), model.Name, alias))
	} else {
		exec = exec.Append(fmt.Sprintf("SELECT %s FROM `%s`", strings.Join(_list, ","), model.Name))
	}
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

func (dia DefaultDialect) InsertExec(model *Model, columnValues []ColumnNameValue) ExecValue {
	// optimization column format
	fieldStr, qStr, args := insertValuesFormat(model, columnValues)

	var exec ExecValue = DefaultExec{}
	exec = exec.Append(
		fmt.Sprintf("INSERT INTO `%s`(%s) VALUES(%s)", model.Name, fieldStr, qStr),
		args...,
	)
	return exec
}

func (dia DefaultDialect) SaveExec(model *Model, columnValues []ColumnNameValue) ExecValue {
	// optimization column format
	fieldStr, qStr, args := insertValuesFormat(model, columnValues)

	var exec ExecValue = DefaultExec{}
	exec = exec.Append(
		fmt.Sprintf("INSERT INTO `%s`(%s) VALUES(%s)", model.Name, fieldStr, qStr),
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

func (dia DefaultDialect) CountExec(model *Model, alias string) ExecValue {
	if alias != "" {
		return DefaultExec{fmt.Sprintf("SELECT count(*) FROM `%s` as `%s`", model.Name, alias), nil}
	}
	return DefaultExec{fmt.Sprintf("SELECT count(*) FROM `%s`", model.Name), nil}
}

func (dia DefaultDialect) TemplateExec(tExec BasicExec, execs map[string]BasicExec) (ExecValue, error) {
	exec, err := getTemplateExec(tExec, execs)
	if err != nil {
		return nil, err
	}
	return DefaultExec{exec.query, exec.args}, nil

}

func (dia DefaultDialect) JoinExec(mainSwap *JoinSwap) ExecValue {
	var strList []string
	for name := range mainSwap.JoinMap {
		join := mainSwap.JoinMap[name]
		swap := mainSwap.SwapMap[name]
		strList = append(strList, fmt.Sprintf("JOIN `%s` AS `%s` ON %s.%s = %s.%s",
			join.SubModel.Name,
			swap.Alias,
			mainSwap.Alias, join.OnMain.Column(),
			swap.Alias, join.OnSub.Column(),
		))
	}
	var exec ExecValue = DefaultExec{strings.Join(strList, " "), nil}
	for _, subSwap := range mainSwap.SwapMap {
		subExec := dia.JoinExec(subSwap)
		exec = exec.Append(" "+subExec.Source(), subExec.Args()...)
	}
	return exec
}
