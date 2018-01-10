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

func NewStructRecord(model *Model, value reflect.Value) ModelRecord {
	vtype := value.Type()
	fieldValueList := GetStructValueFields(value)
	structFieldList := GetStructFields(vtype)

	record := map[*ModelField]reflect.Value{}
	if vtype == model.ReflectType {
		for i := 0; i < len(model.AllFields); i++ {
			record[model.GetPosField(i)] = fieldValueList[i]
		}
	} else {
		for i, field := range structFieldList {
			if mField, ok := model.NameFields[field.Name]; ok {
				record[mField] = fieldValueList[i]
			}
		}
	}
	return &ModelStructRecord{
		FieldValues:        record,
		VirtualFieldValues: map[*ModelField]reflect.Value{},
		source:             value,
		model:              model,
	}
}

func NewRecords(model *Model, value reflect.Value) ModelRecords {
	if value.Kind() != reflect.Slice {
		panic("value must be slice")
	}
	elemType := LoopTypeIndirectSliceAndPtr(value.Type())
	if _, ok := reflect.New(elemType).Elem().Interface().(map[string]interface{}); ok {
		return NewNameMapRecords(model, value)
	} else if _, ok := reflect.New(elemType).Elem().Interface().(map[uintptr]interface{}); ok {
		return NewOffsetMapRecords(model, value)
	} else {
		return NewStructRecords(model, value)
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
