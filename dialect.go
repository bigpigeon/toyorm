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

type DialectConditionArgs struct {
	search        SearchList
	limit, offset int
	orderBy       []Column
	groupBy       []Column
}

type DialectSaveArgs struct {
	CreatedAtField  FieldValue
	UpdatedAtField  FieldValue
	CasField        FieldValue
	InsertFieldList FieldValueList
	SaveFieldList   FieldValueList
	PrimaryFields   FieldValueList
}

type DialectUpdateArgs struct {
	UpdateFieldList FieldValueList
}

type DialectFindArgs struct {
	Columns []Column
	Swap    *JoinSwap
}

type DialectHardDeleteArgs struct {
	PrimaryFields FieldValueList
}

type DialectSoftDeleteArgs struct {
	PrimaryFields FieldValueList
	UpdatedValues FieldValueList
}

type Dialect interface {
	// some database like postgres not support LastInsertId, need QueryRow to get the return id
	InsertExecutor(Executor, ExecValue, func(ExecValue, error)) (sql.Result, error)
	// FIXME sqlite3/postgresql use RowsAffected to check success or failure, but mysql can't,
	// because it's RowsAffected is zero when update value not change
	SaveExecutor(Executor, ExecValue, func(ExecValue, error)) (sql.Result, error)
	HasTable(*Model) ExecValue
	CreateTable(*Model, map[string]ForeignKey) []ExecValue
	DropTable(*Model) ExecValue
	ConditionExec(search SearchList, limit, offset int, orderBy []Column, groupBy []Column) ExecValue
	ConditionBasicExec(args DialectConditionArgs) *BasicExec
	FindExec(temp *BasicExec, model *Model, find DialectFindArgs, condition DialectConditionArgs) (ExecValue, error)
	UpdateExec(temp *BasicExec, model *Model, update DialectUpdateArgs, condition DialectConditionArgs) (ExecValue, error)
	HardDeleteExec(temp *BasicExec, model *Model, delete DialectSoftDeleteArgs, condition DialectConditionArgs) (ExecValue, error)
	SoftDeleteExec(temp *BasicExec, model *Model, delete DialectSoftDeleteArgs, condition DialectConditionArgs) (ExecValue, error)
	InsertExec(temp *BasicExec, model *Model, save DialectSaveArgs, condition DialectConditionArgs) (ExecValue, error)
	SaveExec(temp *BasicExec, model *Model, save DialectSaveArgs, condition DialectConditionArgs) (ExecValue, error)
	USaveExec(temp *BasicExec, model *Model, alias string, save DialectSaveArgs, condition DialectConditionArgs) (ExecValue, error)
	//AddForeignKey(model, relationModel *Model, ForeignKeyField Field) ExecValue
	//DropForeignKey(model *Model, ForeignKeyField Field) ExecValue
	CountExec(model *Model, alias string) ExecValue
	SearchExec(search SearchList) ExecValue
	TemplateExec(BasicExec, map[string]BasicExec) (ExecValue, error)
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
	if len(groupBy) > 0 {
		var list []string
		for _, column := range groupBy {
			list = append(list, column.Column())
		}
		exec = exec.Append(" GROUP BY " + strings.Join(list, ","))
	}
	if len(orderBy) > 0 {
		var __list []string
		for _, column := range orderBy {
			__list = append(__list, column.Column())
		}
		exec = exec.Append(" ORDER BY " + strings.Join(__list, ","))
	}
	if limit != 0 {
		exec = exec.Append(fmt.Sprintf(" LIMIT %d", limit))
	}
	if offset != 0 {
		exec = exec.Append(fmt.Sprintf(" OFFSET %d", offset))
	}
	return exec
}

// TODO do not convert the ConditionExec result
func (dia DefaultDialect) ConditionBasicExec(args DialectConditionArgs) *BasicExec {
	exec := dia.ConditionExec(args.search, args.limit, args.offset, args.orderBy, args.groupBy)
	return &BasicExec{exec.Source(), exec.Args()}
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

func (dia DefaultDialect) FindExec(temp *BasicExec, model *Model, find DialectFindArgs, condition DialectConditionArgs) (ExecValue, error) {
	var _list []string
	if temp == nil {
		temp = &BasicExec{"SELECT $Columns FROM $ModelDef $JoinDef $Conditions", nil}
	}
	for _, column := range find.Columns {
		_list = append(_list, column.Column())
	}

	execMap := FindTemplate{
		TemplateBasic: TemplateBasic{
			Temp:  *temp,
			Model: model,
			Alias: find.Swap.Alias,
			Quote: "`",
		},
		Columns:    BasicExec{strings.Join(_list, ","), nil},
		JoinDef:    dia.JoinExec(find.Swap),
		Conditions: *dia.ConditionBasicExec(condition),
	}
	basicExec, err := execMap.Render()
	if err != nil {
		return nil, err
	}
	return &DefaultExec{basicExec.query, basicExec.args}, nil
}

func (dia DefaultDialect) UpdateExec(temp *BasicExec, model *Model, update DialectUpdateArgs, condition DialectConditionArgs) (ExecValue, error) {
	var recordList []string
	var args []interface{}
	if temp == nil {
		temp = &BasicExec{"UPDATE $ModelDef SET $UpdateValues $Conditions", nil}
	}
	for _, r := range update.UpdateFieldList {
		recordList = append(recordList, r.Column()+" = ?")
		args = append(args, r.Value().Interface())
	}

	execMap := UpdateTemplate{
		TemplateBasic: TemplateBasic{
			Temp:  *temp,
			Model: model,
			Alias: "",
			Quote: "`",
		},
		UpdateValues: BasicExec{
			query: strings.Join(recordList, ","),
			args:  args,
		},
		Conditions: *dia.ConditionBasicExec(condition),
	}
	basicExec, err := execMap.Render()
	if err != nil {
		return nil, err
	}
	return &DefaultExec{basicExec.query, basicExec.args}, nil
}

func (dia DefaultDialect) HardDeleteExec(temp *BasicExec, model *Model, delete DialectSoftDeleteArgs, condition DialectConditionArgs) (ExecValue, error) {
	if temp == nil {
		temp = &BasicExec{"DELETE FROM $ModelDef $Conditions", nil}
	}

	var primaryExec BasicExec
	if delete.PrimaryFields != nil {
		var primaryCondition SearchList
		for _, p := range delete.PrimaryFields {
			primaryCondition = primaryCondition.Condition(p, ExprIn, ExprAnd)
			condition.search = condition.search.Condition(p, ExprIn, ExprAnd)
		}
		execVal := dia.SearchExec(primaryCondition)
		primaryExec = BasicExec{execVal.Source(), execVal.Args()}
	}

	var updateValList []string
	var updateValArgs []interface{}
	for _, u := range delete.UpdatedValues {
		updateValList = append(updateValList, u.Column()+" = ?")
		updateValArgs = append(updateValArgs, u.Value().Interface())
	}

	execMap := HardDeleteTemplate{
		TemplateBasic: TemplateBasic{
			Temp:  *temp,
			Model: model,
			Alias: "",
			Quote: "`",
		},
		UpdateValues:  BasicExec{strings.Join(updateValList, ","), updateValArgs},
		PrimaryValues: primaryExec,
		Conditions:    *dia.ConditionBasicExec(condition),
	}
	basicExec, err := execMap.Render()
	if err != nil {
		return nil, err
	}
	return &DefaultExec{basicExec.query, basicExec.args}, nil
}

func (dia DefaultDialect) SoftDeleteExec(temp *BasicExec, model *Model, delete DialectSoftDeleteArgs, condition DialectConditionArgs) (ExecValue, error) {
	if temp == nil {
		temp = &BasicExec{"UPDATE $ModelDef SET $UpdateValues $Conditions", nil}
	}
	var primaryExec BasicExec
	if delete.PrimaryFields != nil {
		var primaryCondition SearchList
		for _, p := range delete.PrimaryFields {
			primaryCondition = primaryCondition.Condition(p, ExprIn, ExprAnd)
			condition.search = condition.search.Condition(p, ExprIn, ExprAnd)
		}
		execVal := dia.SearchExec(primaryCondition)
		primaryExec = BasicExec{execVal.Source(), execVal.Args()}
	}

	var updateValList []string
	var updateValArgs []interface{}
	for _, u := range delete.UpdatedValues {
		updateValList = append(updateValList, u.Column()+" = ?")
		updateValArgs = append(updateValArgs, u.Value().Interface())
	}

	execMap := HardDeleteTemplate{
		TemplateBasic: TemplateBasic{
			Temp:  *temp,
			Model: model,
			Alias: "",
			Quote: "`",
		},
		UpdateValues:  BasicExec{strings.Join(updateValList, ","), updateValArgs},
		PrimaryValues: primaryExec,
		Conditions:    *dia.ConditionBasicExec(condition),
	}
	basicExec, err := execMap.Render()
	if err != nil {
		return nil, err
	}
	return &DefaultExec{basicExec.query, basicExec.args}, nil
}

func (dia DefaultDialect) InsertExec(temp *BasicExec, model *Model, save DialectSaveArgs, condition DialectConditionArgs) (ExecValue, error) {
	// optimization column format
	if temp == nil {
		temp = &BasicExec{"INSERT INTO $ModelDef($Columns) VALUES($Values)", nil}
	}

	columns, values := getInsertColumnExecAndValue(model, save.InsertFieldList)

	execMap := SaveTemplate{
		TemplateBasic: TemplateBasic{
			Temp:  *temp,
			Model: model,
			Alias: "",
			Quote: "`",
		},
		Columns:    columns,
		Values:     values,
		Conditions: *dia.ConditionBasicExec(condition),
	}
	basicExec, err := execMap.Render()
	if err != nil {
		return nil, err
	}
	return &DefaultExec{basicExec.query, basicExec.args}, nil
}

func (dia DefaultDialect) SaveExec(temp *BasicExec, model *Model, save DialectSaveArgs, condition DialectConditionArgs) (ExecValue, error) {
	// optimization column format
	if temp == nil {
		temp = &BasicExec{"INSERT INTO $ModelDef($Columns) VALUES($Values)", nil}
	}
	columns, values := getInsertColumnExecAndValue(model, save.InsertFieldList)

	execMap := SaveTemplate{
		TemplateBasic: TemplateBasic{
			Temp:  *temp,
			Model: model,
			Alias: "",
			Quote: "`",
		},
		Columns:      columns,
		Values:       values,
		UpdateValues: BasicExec{},
		Conditions:   *dia.ConditionBasicExec(condition),
	}
	basicExec, err := execMap.Render()
	if err != nil {
		return nil, err
	}

	return &DefaultExec{basicExec.query, basicExec.args}, nil
}

func (dia DefaultDialect) USaveExec(temp *BasicExec, model *Model, alias string, save DialectSaveArgs, condition DialectConditionArgs) (ExecValue, error) {
	// optimization column format
	if temp == nil {
		temp = &BasicExec{"UPDATE $ModelDef SET $UpdateValues $Conditions", nil}
	}
	columns, values := getInsertColumnExecAndValue(model, save.InsertFieldList)

	for _, p := range save.PrimaryFields {
		condition.search = condition.search.Condition(p, ExprEqual, ExprAnd)
	}

	if save.CasField != nil {
		condition.search = condition.search.Condition(save.CasField, ExprEqual, ExprAnd)
	}

	updateValExec := BasicExec{}
	var updateQueryList []string
	for _, val := range save.SaveFieldList {
		updateQueryList = append(updateQueryList, val.Column()+" = ? ")
		updateValExec.args = append(updateValExec.args, val.Value().Interface())
	}
	updateValExec.query = strings.Join(updateQueryList, ",")

	execMap := SaveTemplate{
		TemplateBasic: TemplateBasic{
			Temp:  *temp,
			Model: model,
			Alias: alias,
			Quote: "`",
		},
		Columns:      columns,
		Values:       values,
		UpdateValues: updateValExec,
		Conditions:   *dia.ConditionBasicExec(condition),
	}
	basicExec, err := execMap.Render()
	if err != nil {
		return nil, err
	}

	return &DefaultExec{basicExec.query, basicExec.args}, nil
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
	exec, err := GetTemplateExec(tExec, execs)
	if err != nil {
		return nil, err
	}
	return DefaultExec{exec.query, exec.args}, nil

}

func (dia DefaultDialect) JoinExec(mainSwap *JoinSwap) BasicExec {
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
	var exec = BasicExec{strings.Join(strList, " "), nil}
	for _, subSwap := range mainSwap.SwapMap {
		subExec := dia.JoinExec(subSwap)
		exec.query += " " + subExec.query
		exec.args = append(exec.args, subExec.args...)
	}
	return exec
}
