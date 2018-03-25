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

type ModelOffsetMapRecords struct {
	model           *Model
	elemType        reflect.Type
	source          reflect.Value
	FieldValuesList []map[string]reflect.Value
}

func NewOffsetMapRecords(model *Model, m reflect.Value) *ModelOffsetMapRecords {
	records := &ModelOffsetMapRecords{
		model,
		m.Type().Elem(),
		m,
		[]map[string]reflect.Value{},
	}
	records.sync()
	return records
}

func NewOffsetMapRecord(model *Model, v reflect.Value) *ModelOffsetMapRecord {
	record := &ModelOffsetMapRecord{
		map[string]reflect.Value{},
		v,
		model,
	}

	for offset, mField := range model.GetOffsetFieldMap() {
		if fieldValue := v.MapIndex(reflect.ValueOf(offset)); fieldValue.IsValid() {
			fieldValue = fieldValue.Elem()
			record.FieldValues[mField.Name()] = fieldValue
		}
	}
	return record
}

func (m *ModelOffsetMapRecords) sync() {
	for i := len(m.FieldValuesList); i < m.source.Len(); i++ {
		// why need LoopIndirect? because m could be []*map[uintptr]interface{}
		elem := LoopIndirect(m.source.Index(i))
		c := map[string]reflect.Value{}
		for offset, field := range m.model.GetOffsetFieldMap() {
			if elemField := elem.MapIndex(reflect.ValueOf(offset)); elemField.IsValid() {
				elemField = elemField.Elem()
				c[field.Name()] = elemField
			}
		}
		m.FieldValuesList = append(m.FieldValuesList, c)
	}
}

func (m *ModelOffsetMapRecords) GetRecords() []ModelRecord {
	var recordList []ModelRecord
	for i := 0; i < len(m.FieldValuesList); i++ {
		recordList = append(recordList, &ModelOffsetMapRecord{
			FieldValues: m.FieldValuesList[i],
			source:      LoopIndirect(m.source.Index(i)),
			model:       m.model,
		})
	}
	return recordList
}

func (m *ModelOffsetMapRecords) GroupBy(key string) ModelGroupBy {
	field := m.model.GetFieldWithName(key)
	if field.StructField().Type.Comparable() == false {
		panic(fmt.Sprintf("%s is not compareable field", field.Name()))
	}
	result := ModelGroupBy{}
	for i := 0; i < len(m.FieldValuesList); i++ {
		keyValue := m.FieldValuesList[i][key].Interface()
		result[keyValue] = append(result[keyValue], ModelIndexRecord{&ModelOffsetMapRecord{
			FieldValues: m.FieldValuesList[i],
			source:      LoopIndirect(m.source.Index(i)),
			model:       m.model,
		}, i})
	}
	return result
}

func (m *ModelOffsetMapRecords) GetRecord(i int) ModelRecord {
	return &ModelOffsetMapRecord{
		FieldValues: m.FieldValuesList[i],
		source:      LoopIndirect(m.source.Index(i)),
		model:       m.model,
	}
}

func (m *ModelOffsetMapRecords) Add(v reflect.Value) ModelRecord {
	if m.source.CanSet() == false {
		panic("Add need can set permission")
	}
	last := len(m.FieldValuesList)
	m.source.Set(reflect.Append(m.source, v))
	m.sync()
	return &ModelOffsetMapRecord{
		FieldValues: m.FieldValuesList[last],
		source:      LoopIndirect(m.source.Index(last)),
		model:       m.model,
	}
}

func (m *ModelOffsetMapRecords) GetFieldType(name string) reflect.Type {
	field := m.model.GetFieldWithName(name)
	fieldType := field.StructField().Type

	// TODO check the field is or not container field ?
	if field.Column() == "" || field.SqlType() == "" {
		t := LoopTypeIndirect(fieldType)
		switch t.Kind() {
		case reflect.Struct:
			return reflect.TypeOf(map[uintptr]interface{}{})
		case reflect.Slice:
			return reflect.TypeOf([]map[uintptr]interface{}{})
		}
	}
	return fieldType
}

func (m *ModelOffsetMapRecords) GetFieldAddressType(name string) reflect.Type {
	return m.GetFieldType(name)
}

func (m *ModelOffsetMapRecords) IsVariableContainer() bool {
	return true
}

func (m *ModelOffsetMapRecords) ElemType() reflect.Type {
	return reflect.TypeOf(map[uintptr]interface{}{})
}

func (m *ModelOffsetMapRecords) Len() int {
	return len(m.FieldValuesList)
}

func (m *ModelOffsetMapRecords) Source() reflect.Value {
	return m.source
}

type ModelOffsetMapRecord struct {
	FieldValues map[string]reflect.Value
	source      reflect.Value
	model       *Model
}

func (m *ModelOffsetMapRecord) SetField(name string, value reflect.Value) {
	if name == "" {
		return
	}
	if field := m.model.GetFieldWithName(name); field != nil {
		offset := reflect.ValueOf(field.Offset())
		elem := reflect.New(field.StructField().Type).Elem()
		safeSet(elem, value)
		m.source.SetMapIndex(offset, elem)
		m.FieldValues[name] = m.source.MapIndex(offset).Elem()
	}
}

func (m *ModelOffsetMapRecord) Field(name string) reflect.Value {
	return m.FieldValues[name]
}

func (m *ModelOffsetMapRecord) FieldAddress(name string) reflect.Value {
	return m.FieldValues[name]
}

func (m *ModelOffsetMapRecord) AllField() map[string]reflect.Value {
	return m.FieldValues
}

func (m *ModelOffsetMapRecord) IsVariableContainer() bool {
	return true
}

func (m *ModelOffsetMapRecord) Source() reflect.Value {
	return m.source
}

func (m *ModelOffsetMapRecord) GetFieldType(name string) reflect.Type {
	field := m.model.GetFieldWithName(name)
	fieldType := field.StructField().Type

	// TODO check the field is or not container field ?
	if field.Column() == "" || field.SqlType() == "" {
		t := LoopTypeIndirect(fieldType)
		switch t.Kind() {
		case reflect.Struct:
			return reflect.TypeOf(map[uintptr]interface{}{})
		case reflect.Slice:
			return reflect.TypeOf([]map[uintptr]interface{}{})
		}
	}
	return fieldType
}
