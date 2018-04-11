/*
 * Copyright 2018. bigpigeon. All rights reserved.
 * Use of this source code is governed by a MIT style
 * license that can be found in the LICENSE file.
 */

package toyorm

import (
	"fmt"
	"reflect"
	"strings"
)

type Field interface {
	Column
	Name() string
	Offset() uintptr
	IsPrimary() bool
	AutoIncrement() bool
	Index() string
	UniqueIndex() string
	IsForeign() bool
	Attr(string) string
	Attrs() map[string]string
	SqlType() string
	StructField() reflect.StructField
	JoinWith() string
}

type modelField struct {
	column        string
	sqlType       string
	offset        uintptr
	isPrimary     bool
	index         string
	uniqueIndex   string
	ignore        bool
	attrs         map[string]string
	autoIncrement bool
	isForeign     bool
	field         reflect.StructField
	alias         string
	joinWith      string
}

func (m *modelField) Column() string {
	return m.column
}

func (m *modelField) Name() string {
	if m.alias != "" {
		return m.alias
	}
	return m.field.Name
}

func (m *modelField) Offset() uintptr {
	return m.offset
}

func (m *modelField) IsPrimary() bool {
	return m.isPrimary
}

func (m *modelField) AutoIncrement() bool {
	return m.autoIncrement
}

func (m *modelField) Index() string {
	return m.index
}

func (m *modelField) UniqueIndex() string {
	return m.uniqueIndex
}

func (m *modelField) IsForeign() bool {
	return m.isForeign
}

func (m *modelField) StructField() reflect.StructField {
	return m.field
}

func (m *modelField) Attr(s string) string {
	return m.attrs[s]
}

func (m *modelField) Attrs() map[string]string {
	return m.attrs
}

func (m *modelField) SqlType() string {
	return m.sqlType
}

func (m *modelField) JoinWith() string {
	return m.joinWith
}

func NewField(f *reflect.StructField, table_name string) *modelField {
	field := &modelField{
		field:   *f,
		attrs:   map[string]string{},
		column:  SqlNameConvert(f.Name),
		offset:  f.Offset,
		sqlType: ToSqlType(f.Type),
	}

	// set attribute by tag
	tags := strings.Split(f.Tag.Get("toyorm"), ";")
	for _, t := range tags {
		keyval := strings.Split(t, ":")
		var key, val string
		switch len(keyval) {
		case 0:
			continue
		case 1:
			key = keyval[0]
		case 2:
			key, val = keyval[0], keyval[1]
		default:
			panic(ErrInvalidTag)
		}
		switch strings.ToLower(key) {
		case "auto_increment", "autoincrement":
			field.autoIncrement = true
		case "primary key":
			field.isPrimary = true
		case "type":
			field.sqlType = val
		case "index":
			if val == "" {
				field.index = fmt.Sprintf("idx_%s_%s", table_name, field.column)
			} else {
				field.index = val
			}
		case "unique index":
			if val == "" {
				field.uniqueIndex = fmt.Sprintf("udx_%s_%s", table_name, field.column)
			} else {
				field.uniqueIndex = val
			}
		case "foreign key":
			field.isForeign = true
		case "column":
			field.column = val
		case "-":
			field.ignore = true
		case "alias":
			field.alias = val
			// container field must be ignore in sql
			field.ignore = true
		case "join":
			field.joinWith = val
		default:
			field.attrs[key] = val
		}
	}
	if field.column == "" || field.sqlType == "" {
		field.ignore = true
	}
	return field
}
