/*
 * Copyright 2018. bigpigeon. All rights reserved.
 * Use of this source code is governed by a MIT style
 * license that can be found in the LICENSE file.
 */

package toyorm

import (
	"reflect"
)

type Column interface {
	Column() string // sql column declaration
}

type ColumnName interface {
	Column
	Name() string
}

type ColumnValue interface {
	Column
	Value() reflect.Value
}

type ColumnNameValue interface {
	ColumnName
	Value() reflect.Value
}

type CountColumn struct{}

func (CountColumn) Column() string {
	return "count(*)"
}

//type modelFieldValue struct {
//	Field
//	value reflect.Value
//}
//
//func (s *modelFieldValue) Value() reflect.Value {
//	return s.value
//}
