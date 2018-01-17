package toyorm

import (
	"reflect"
)

type ModelRecordFieldTypes interface {
	GetFieldType(*ModelField) reflect.Type
}

type ModelRecord interface {
	AllField() map[*ModelField]reflect.Value
	SetField(field *ModelField, v reflect.Value)
	Field(field *ModelField) reflect.Value
	FieldAddress(field *ModelField) reflect.Value
	IsVariableContainer() bool
	Source() reflect.Value
	GetFieldType(*ModelField) reflect.Type
}

type ModelRecords interface {
	GetRecord(int) ModelRecord
	GetRecords() []ModelRecord
	Add(v reflect.Value) ModelRecord
	GetFieldType(*ModelField) reflect.Type
	GetFieldAddressType(*ModelField) reflect.Type
	IsVariableContainer() bool
	ElemType() reflect.Type
	Len() int
	Source() reflect.Value
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
