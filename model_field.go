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

type tagKeyValue struct {
	Key string
	Val string
}

func newTagKeyVal(s string) *tagKeyValue {
	sList := strings.Split(s, ":")
	keyVal := new(tagKeyValue)
	switch len(sList) {
	case 0:
		return nil
	case 1:
		keyVal.Key = strings.TrimSpace(sList[0])
	case 2:
		keyVal.Key, keyVal.Val = strings.TrimSpace(sList[0]), strings.TrimSpace(sList[1])
	default:
		panic(ErrInvalidTag)
	}
	keyVal.Key = strings.ToLower(keyVal.Key)
	return keyVal
}

func GetTagKeyVal(tag string) []*tagKeyValue {
	var keyValList []*tagKeyValue
	// set attribute by tag
	tags := strings.Split(tag, ";")
	for _, t := range tags {
		if tagKeyVal := newTagKeyVal(t); tagKeyVal != nil {
			keyValList = append(keyValList, tagKeyVal)
		}
	}
	return keyValList
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
	for _, tagKeyVal := range GetTagKeyVal(f.Tag.Get("toyorm")) {

		switch tagKeyVal.Key {
		case "auto_increment", "autoincrement":
			field.autoIncrement = true
		case "primary key":
			field.isPrimary = true
		case "type":
			field.sqlType = tagKeyVal.Val
		case "index":
			if tagKeyVal.Val == "" {
				field.index = fmt.Sprintf("idx_%s_%s", table_name, field.column)
			} else {
				field.index = tagKeyVal.Val
			}
		case "unique index":
			if tagKeyVal.Val == "" {
				field.uniqueIndex = fmt.Sprintf("udx_%s_%s", table_name, field.column)
			} else {
				field.uniqueIndex = tagKeyVal.Val
			}
		case "foreign key":
			field.isForeign = true
		case "column":
			field.column = tagKeyVal.Val
		case "-":
			field.ignore = true
		case "alias":
			field.alias = tagKeyVal.Val
			// container field must be ignore in sql
			field.ignore = true
		case "join":
			field.joinWith = tagKeyVal.Val
		default:
			field.attrs[tagKeyVal.Val] = tagKeyVal.Val
		}
	}
	if field.column == "" || field.sqlType == "" {
		field.ignore = true
	}
	return field
}
