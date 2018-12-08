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

type AssociationType int

const (
	JoinWith AssociationType = iota
	BelongToWith
	OneToOneWith
	OneToManyWith
	AssociationTypeEnd
)

type Field interface {
	Column                            // sql column declaration
	Name() string                     // get field name
	Offset() uintptr                  // relation position by model struct
	IsPrimary() bool                  // primary key declaration
	AutoIncrement() bool              // auto increment declaration
	Index() string                    // sql index declaration
	UniqueIndex() string              // sql unique index declaration
	IsForeign() bool                  // sql foreign declaration
	Attr(string) string               // extension attribute declaration
	Attrs() map[string]string         // get all extension attribute
	SqlType() string                  // sql type declaration
	StructField() reflect.StructField // model struct attribute
	FieldValue() reflect.Value        // model meta field value
	JoinWith() string                 // join with specified container field declaration,when call ToyBrick.Preload(<container field>) will automatic association this field
	BelongToWith() string             // BelongTo with specified container field declaration,ToyBrick.Preload(<container field>) will automatic association this field
	OneToOneWith() string             // OneToOne with specified container field declaration,ToyBrick.Preload(<container field>) will automatic association this field
	OneToManyWith() string            // OneToMany with specified container field declaration,ToyBrick.Preload(<container field>) will automatic association this field

	Source() Field
	ToColumnAlias(alias string) Field
	ToFieldValue(value reflect.Value) FieldValue
}

type FieldValue interface {
	Field
	Value() reflect.Value
}

type aliasField struct {
	Field
	columnAlias string
}

func (a *aliasField) Source() Field { return a.Field }

func (a *aliasField) Column() string {
	return a.columnAlias
}

func (a *aliasField) ToFieldValue(value reflect.Value) FieldValue {
	return &fieldValue{a, value}
}

type fieldValue struct {
	Field
	value reflect.Value
}

func (m *fieldValue) Value() reflect.Value {
	return m.value
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
	fieldValue    reflect.Value
	alias         string
	Association   map[AssociationType]string
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

func (m *modelField) FieldValue() reflect.Value {
	return m.fieldValue
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
	return m.Association[JoinWith]
}

func (m *modelField) BelongToWith() string {
	return m.Association[BelongToWith]
}

func (m *modelField) OneToOneWith() string {
	return m.Association[OneToOneWith]
}

func (m *modelField) OneToManyWith() string {
	return m.Association[OneToManyWith]
}

func (m *modelField) Source() Field { return m }

func (m *modelField) ToColumnAlias(alias string) Field {
	if alias == "" {
		return m
	}
	return &aliasField{
		Field:       m,
		columnAlias: fmt.Sprintf("%s.%s", alias, m.column),
	}
}

func (m *modelField) ToFieldValue(value reflect.Value) FieldValue {
	return &fieldValue{m, value}
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

func NewField(f *reflect.StructField, fieldVal reflect.Value, table_name string) *modelField {
	field := &modelField{
		field:       *f,
		attrs:       map[string]string{},
		column:      SqlNameConvert(f.Name),
		offset:      f.Offset,
		sqlType:     ToSqlType(f.Type),
		Association: map[AssociationType]string{},
		fieldValue:  fieldVal,
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
		case "join":
			field.Association[JoinWith] = tagKeyVal.Val
		case "belong to":
			field.Association[BelongToWith] = tagKeyVal.Val
		case "one to one":
			field.Association[OneToOneWith] = tagKeyVal.Val
		case "one to many":
			field.Association[OneToManyWith] = tagKeyVal.Val
		//case "middle model with":
		//	field.Association[MiddleModelWith] = val
		//case "left model with":
		//	field.Association[LeftModelWith] = val
		//case "right model with":
		//	field.Association[RightModelWith] = val

		default:
			field.attrs[tagKeyVal.Key] = tagKeyVal.Val
		}
	}
	if field.column == "" || field.sqlType == "" {
		field.ignore = true
	}
	return field
}

type FieldList []Field

func (l FieldList) ToColumnList() []Column {
	columns := make([]Column, len(l))
	for i := range l {
		columns[i] = l[i]
	}
	return columns
}

type FieldValueList []FieldValue

func (l FieldValueList) ToValueList() []ColumnValue {
	values := make([]ColumnValue, len(l))
	for i := range l {
		values[i] = l[i]
	}
	return values
}

func (l FieldValueList) ToNameValueList() []ColumnNameValue {
	values := make([]ColumnNameValue, len(l))
	for i := range l {
		values[i] = l[i]
	}
	return values
}
