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
	Column() string
}

type StrColumn string

func (d StrColumn) Column() string {
	return string(d)
}

type ColumnName interface {
	Column() string
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

type modelFieldValue struct {
	Field
	value reflect.Value
}

func (s *modelFieldValue) Value() reflect.Value {
	return s.value
}
