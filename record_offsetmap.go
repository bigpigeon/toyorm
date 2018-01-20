package toyorm

import (
	"reflect"
	"time"
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

	for i := 0; i < records.source.Len(); i++ {
		// why need LoopIndirect? because m could be []*map[uintptr]interface{}
		elem := LoopIndirect(records.source.Index(i))
		c := map[string]reflect.Value{}
		for offset, field := range model.GetOffsetFieldMap() {
			if elemField := elem.MapIndex(reflect.ValueOf(offset)); elemField.IsValid() {
				elemField = elemField.Elem()
				c[field.Name()] = elemField
			}
		}
		records.FieldValuesList = append(records.FieldValuesList, c)
	}
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
	m.source.Set(reflect.Append(m.source, v))
	v = LoopIndirectAndNew(v)
	c := map[string]reflect.Value{}
	for offset, field := range m.model.OffsetFields {
		if elemField := v.MapIndex(reflect.ValueOf(offset)); elemField.IsValid() {
			elemField = elemField.Elem()
			c[field.Name()] = elemField
		}
	}

	m.FieldValuesList = append(m.FieldValuesList, c)
	return &ModelOffsetMapRecord{
		FieldValues: c,
		source:      LoopIndirect(m.source.Index(m.source.Len() - 1)),
		model:       m.model,
	}
}

func (m *ModelOffsetMapRecords) GetFieldType(name string) reflect.Type {
	fieldType := m.model.GetFieldWithName(name).StructField().Type
	t := LoopTypeIndirect(fieldType)
	if _, ok := reflect.Zero(t).Interface().(time.Time); ok {
		return fieldType
	}
	switch t.Kind() {
	case reflect.Struct:
		return reflect.TypeOf(map[uintptr]interface{}{})
	case reflect.Slice:
		return reflect.TypeOf([]map[uintptr]interface{}{})
	default:
		return fieldType
	}
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
		if fieldValue := m.source.MapIndex(offset); fieldValue.IsValid() {
			SafeMapSet(m.source, offset, value)
		} else if value.IsValid() {
			elem := reflect.New(field.StructField().Type).Elem()
			SafeSet(elem, value)
			m.source.SetMapIndex(offset, elem)
		}
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
	fieldType := m.model.GetFieldWithName(name).StructField().Type
	t := LoopTypeIndirect(fieldType)
	if _, ok := reflect.Zero(t).Interface().(time.Time); ok {
		return fieldType
	}
	switch t.Kind() {
	case reflect.Struct:
		return reflect.TypeOf(map[uintptr]interface{}{})
	case reflect.Slice:
		return reflect.TypeOf([]map[uintptr]interface{}{})
	default:
		return fieldType
	}
}
