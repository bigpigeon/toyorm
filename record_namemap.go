package toyorm

import (
	"reflect"
	"time"
)

type ModelNameMapRecords struct {
	model           *Model
	elemType        reflect.Type
	source          reflect.Value
	FieldValuesList []map[*ModelField]reflect.Value
}

func NewNameMapRecords(model *Model, v reflect.Value) *ModelNameMapRecords {
	records := &ModelNameMapRecords{
		model,
		v.Type().Elem(),
		v,
		[]map[*ModelField]reflect.Value{},
	}
	for i := 0; i < records.source.Len(); i++ {
		// why need LoopIndirect? because m could be []*map[string]interface{}
		elem := LoopIndirect(records.source.Index(i))
		c := map[*ModelField]reflect.Value{}
		for name, field := range model.NameFields {
			if elemField := elem.MapIndex(reflect.ValueOf(name)); elemField.IsValid() {
				elemField = elemField.Elem()
				c[field] = elemField
			}
		}
		records.FieldValuesList = append(records.FieldValuesList, c)
	}
	return records
}

func NewNameMapRecord(model *Model, v reflect.Value) *ModelNameMapRecord {
	record := &ModelNameMapRecord{
		map[*ModelField]reflect.Value{},
		v,
		model,
	}

	for name, mField := range model.NameFields {
		if fieldValue := v.MapIndex(reflect.ValueOf(name)); fieldValue.IsValid() {
			fieldValue = fieldValue.Elem()
			record.FieldValues[mField] = fieldValue
		}
	}
	return record
}

func (m *ModelNameMapRecords) GetRecords() []ModelRecord {
	var recordList []ModelRecord
	for i := 0; i < len(m.FieldValuesList); i++ {
		recordList = append(recordList, &ModelNameMapRecord{
			FieldValues: m.FieldValuesList[i],
			source:      LoopIndirect(m.source.Index(i)),
			model:       m.model,
		})
	}
	return recordList
}

func (m *ModelNameMapRecords) GetRecord(i int) ModelRecord {
	return &ModelNameMapRecord{
		FieldValues: m.FieldValuesList[i],
		source:      LoopIndirect(m.source.Index(i)),
		model:       m.model,
	}
}

func (m *ModelNameMapRecords) Add(v reflect.Value) ModelRecord {
	if m.source.CanSet() == false {
		panic("Add need can set permission")
	}
	m.source.Set(reflect.Append(m.source, v))
	v = LoopIndirectAndNew(v)
	c := map[*ModelField]reflect.Value{}

	for name, field := range m.model.NameFields {
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

func (m *ModelNameMapRecords) GetFieldType(field *ModelField) reflect.Type {
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

func (m *ModelNameMapRecords) GetFieldAddressType(field *ModelField) reflect.Type {
	return m.GetFieldType(field)
}

func (m *ModelNameMapRecords) IsVariableContainer() bool {
	return true
}

func (m *ModelNameMapRecords) ElemType() reflect.Type {
	return m.elemType
}

func (m *ModelNameMapRecords) Len() int {
	return len(m.FieldValuesList)
}

func (m *ModelNameMapRecords) Source() reflect.Value {
	return m.source
}

type ModelNameMapRecord struct {
	FieldValues map[*ModelField]reflect.Value
	source      reflect.Value
	model       *Model
}

func (m *ModelNameMapRecord) SetField(field *ModelField, value reflect.Value) {
	if field == nil {
		return
	}
	name := reflect.ValueOf(field.Field.Name)
	if fieldValue := m.source.MapIndex(name); fieldValue.IsValid() {
		SafeMapSet(m.source, name, value)
	} else if value.IsValid() {
		elem := reflect.New(field.Field.Type).Elem()
		SafeSet(elem, value)
		m.source.SetMapIndex(name, elem)
		m.FieldValues[field] = m.source.MapIndex(name).Elem()
	}
}

func (m *ModelNameMapRecord) Field(field *ModelField) reflect.Value {
	return m.FieldValues[field]
}

func (m *ModelNameMapRecord) FieldAddress(field *ModelField) reflect.Value {
	return m.FieldValues[field]
}

func (m *ModelNameMapRecord) AllField() map[*ModelField]reflect.Value {
	return m.FieldValues
}

func (m *ModelNameMapRecord) IsVariableContainer() bool {
	return true
}

func (m *ModelNameMapRecord) Source() reflect.Value {
	return m.source
}

func (m *ModelNameMapRecord) GetFieldType(field *ModelField) reflect.Type {
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
