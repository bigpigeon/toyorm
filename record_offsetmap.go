package toyorm

import (
	"reflect"
	"time"
)

type ModelOffsetMapRecords struct {
	model           *Model
	elemType        reflect.Type
	source          reflect.Value
	FieldValuesList []map[*ModelField]reflect.Value
}

func NewOffsetMapRecords(model *Model, m reflect.Value) *ModelOffsetMapRecords {
	records := &ModelOffsetMapRecords{
		model,
		m.Type().Elem(),
		m,
		[]map[*ModelField]reflect.Value{},
	}

	for i := 0; i < records.source.Len(); i++ {
		// why need LoopIndirect? because m could be []*map[uintptr]interface{}
		elem := LoopIndirect(records.source.Index(i))
		c := map[*ModelField]reflect.Value{}
		for offset, field := range model.OffsetFields {
			if elemField := elem.MapIndex(reflect.ValueOf(offset)); elemField.IsValid() {
				elemField = elemField.Elem()
				c[field] = elemField
			}
		}
		records.FieldValuesList = append(records.FieldValuesList, c)
	}
	return records
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
	c := map[*ModelField]reflect.Value{}
	for name, field := range m.model.OffsetFields {
		if elemField := v.MapIndex(reflect.ValueOf(name)); elemField.IsValid() {
			elemField = elemField.Elem()
			c[field] = elemField
		}
	}

	m.FieldValuesList = append(m.FieldValuesList, c)
	return &ModelNameMapRecord{
		FieldValues: c,
		source:      LoopIndirect(m.source.Index(m.source.Len() - 1)),
		model:       m.model,
	}
}

func (m *ModelOffsetMapRecords) GetFieldType(field *ModelField) reflect.Type {
	t := LoopTypeIndirect(field.Field.Type)
	if _, ok := reflect.Zero(t).Interface().(time.Time); ok {
		return field.Field.Type
	}
	switch t.Kind() {
	case reflect.Struct:
		return reflect.TypeOf(map[uintptr]interface{}{})
	case reflect.Slice:
		return reflect.TypeOf([]map[uintptr]interface{}{})
	default:
		return field.Field.Type
	}
}

func (m *ModelOffsetMapRecords) GetFieldAddressType(field *ModelField) reflect.Type {
	return m.GetFieldType(field)
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
	FieldValues map[*ModelField]reflect.Value
	source      reflect.Value
	model       *Model
}

func (m *ModelOffsetMapRecord) SetField(field *ModelField, value reflect.Value) {
	if field == nil {
		return
	}
	offset := reflect.ValueOf(field.Offset)
	if fieldValue := m.source.MapIndex(offset); fieldValue.IsValid() {
		SafeMapSet(m.source, offset, value)
	} else if value.IsValid() {
		elem := reflect.New(field.Field.Type).Elem()
		SafeSet(elem, value)
		m.source.SetMapIndex(offset, elem)
		m.FieldValues[field] = m.source.MapIndex(offset).Elem()
	}
}

func (m *ModelOffsetMapRecord) Field(field *ModelField) reflect.Value {
	return m.FieldValues[field]
}

func (m *ModelOffsetMapRecord) FieldAddress(field *ModelField) reflect.Value {
	return m.FieldValues[field]
}

func (m *ModelOffsetMapRecord) AllField() map[*ModelField]reflect.Value {
	return m.FieldValues
}

func (m *ModelOffsetMapRecord) IsVariableContainer() bool {
	return true
}

func (m *ModelOffsetMapRecord) Source() reflect.Value {
	return m.source
}

func (m *ModelOffsetMapRecord) GetFieldType(field *ModelField) reflect.Type {
	t := LoopTypeIndirect(field.Field.Type)
	if _, ok := reflect.Zero(t).Interface().(time.Time); ok {
		return field.Field.Type
	}
	switch t.Kind() {
	case reflect.Struct:
		return reflect.TypeOf(map[string]interface{}{})
	case reflect.Slice:
		return reflect.TypeOf([]map[string]interface{}{})
	default:
		return field.Field.Type
	}
}
