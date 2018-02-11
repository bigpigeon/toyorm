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
	Field(field string) reflect.Value
	FieldAddress(field string) reflect.Value
	IsVariableContainer() bool
	Source() reflect.Value
	GetFieldType(string) reflect.Type
}

type ModelGroupBy map[interface{}][]ModelRecord

func (m ModelGroupBy) Keys() []interface{} {
	l := make([]interface{}, 0, len(m))
	for k, _ := range m {
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
		panic(ErrInvalidRecordType{})
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
		panic(ErrInvalidRecordType{})
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
