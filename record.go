/*
 * Copyright 2018. bigpigeon. All rights reserved.
 * Use of this source code is governed by a MIT style
 * license that can be found in the LICENSE file.
 */

package toyorm

import (
	"reflect"
)

type ModelRecordFieldTypes interface {
	GetFieldType(field string) reflect.Type
}

type ModelRecord interface {
	AllField() map[string]reflect.Value
	SetField(field string, v reflect.Value)
	DeleteField(string)
	Field(field string) reflect.Value
	FieldAddress(field string) reflect.Value
	IsVariableContainer() bool
	Source() reflect.Value
	GetFieldType(string) reflect.Type
}

type ModelIndexRecord struct {
	ModelRecord
	Index int
}

type ModelGroupBy map[interface{}][]ModelIndexRecord

func (m ModelGroupBy) Keys() []interface{} {
	l := make([]interface{}, 0, len(m))
	for k := range m {
		l = append(l, k)
	}
	return l
}

type ModelRecords interface {
	GetRecord(int) ModelRecord
	GetRecords() []ModelRecord
	Add(v reflect.Value) ModelRecord
	GetFieldType(string) reflect.Type
	GetFieldAddressType(string) reflect.Type
	IsVariableContainer() bool
	ElemType() reflect.Type
	Len() int
	Source() reflect.Value
	GroupBy(key string) ModelGroupBy
}

func NewRecords(model *Model, value reflect.Value) ModelRecords {
	if value.Kind() != reflect.Slice {
		panic("value must be slice")
	}
	elemType := LoopTypeIndirectSliceAndPtr(value.Type())
	if _, ok := reflect.Zero(elemType).Interface().(map[string]interface{}); ok {
		return NewNameMapRecords(model, value)
	} else if _, ok := reflect.Zero(elemType).Interface().(map[uintptr]interface{}); ok {
		return NewOffsetMapRecords(model, value)
	} else if elemType.Kind() == reflect.Struct {
		return NewStructRecords(model, value)
	} else {
		panic(ErrInvalidRecordType{elemType})
	}
}

func NewRecord(model *Model, value reflect.Value) ModelRecord {
	vType := LoopTypeIndirect(value.Type())
	if _, ok := reflect.Zero(vType).Interface().(map[string]interface{}); ok {
		return NewNameMapRecord(model, value)
	} else if _, ok := reflect.Zero(vType).Interface().(map[uintptr]interface{}); ok {
		return NewOffsetMapRecord(model, value)
	} else if vType.Kind() == reflect.Struct {
		return NewStructRecord(model, value)
	} else {
		panic(ErrInvalidRecordType{vType})
	}
}

// use element type to create ModelRecords
func MakeRecordsWithElem(model *Model, _type reflect.Type) ModelRecords {
	v := reflect.New(reflect.SliceOf(_type)).Elem()
	return NewRecords(model, v)
}

func MakeRecords(model *Model, _type reflect.Type) ModelRecords {
	v := reflect.New(_type).Elem()
	return NewRecords(model, v)
}

func MakeRecord(model *Model, _type reflect.Type) ModelRecord {
	v := reflect.New(_type).Elem()
	return NewRecord(model, v)
}
