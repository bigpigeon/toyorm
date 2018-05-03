/*
 * Copyright 2018. bigpigeon. All rights reserved.
 * Use of this source code is governed by a MIT style
 * license that can be found in the LICENSE file.
 */

package toyorm

import (
	"fmt"
	"reflect"
)

type Column interface {
	Column() string // sql column declaration
}

type BrickColumn struct {
	alias  string
	column string
}

func (c BrickColumn) Column() string {
	if c.alias != "" {
		return fmt.Sprintf("%s.%s", c.alias, c.column)
	}
	return string(c.column)
}

type BrickColumnList []*BrickColumn

func (l BrickColumnList) ToColumnList() []Column {
	columns := make([]Column, len(l))
	for i := range l {
		columns[i] = *l[i]
	}
	return columns
}

type ColumnName interface {
	Column
	Name() string
}

type ScanField struct {
	name   string
	column string
}

type ColumnValue interface {
	Column
	Value() reflect.Value
}

type BrickColumnValue struct {
	BrickColumn
	value reflect.Value
}

func (v BrickColumnValue) Value() reflect.Value {
	return v.value
}

type BrickColumnValueList []*BrickColumnValue

func (l BrickColumnValueList) ToValueList() []ColumnValue {
	values := make([]ColumnValue, len(l))
	for i := range l {
		values[i] = l[i]
	}
	return values
}

type modelFieldValue struct {
	Field
	value reflect.Value
}

func (s *modelFieldValue) Value() reflect.Value {
	return s.value
}
