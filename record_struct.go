package toyorm

import (
	"fmt"
	"reflect"
)

type ModelStructRecords struct {
	model                  *Model
	Type                   reflect.Type
	source                 reflect.Value
	RelationFieldPos       map[*ModelField]int
	FieldTypes             map[*ModelField]reflect.Type
	FieldValuesList        []map[*ModelField]reflect.Value
	VirtualFieldValuesList []map[*ModelField]reflect.Value
}

func NewStructRecords(model *Model, value reflect.Value) *ModelStructRecords {
	vElemType := value.Type().Elem()
	records := &ModelStructRecords{
		model,
		vElemType,
		value,
		map[*ModelField]int{},
		map[*ModelField]reflect.Type{},
		[]map[*ModelField]reflect.Value{},
		[]map[*ModelField]reflect.Value{},
	}
	vIndirectElemType := LoopTypeIndirectSliceAndPtr(vElemType)
	if vIndirectElemType == model.ReflectType {
		for i := 0; i < len(model.AllFields); i++ {
			f := model.GetPosField(i)
			records.FieldTypes[f] = f.Field.Type
			records.RelationFieldPos[f] = i
		}
	} else {
		structFieldList := GetStructFields(vIndirectElemType)
		// RelationFieldPos[model_field]->struct_field_pos
		for si, field := range structFieldList {
			if mField, ok := model.NameFields[field.Name]; ok {
				records.FieldTypes[mField] = field.Type
				records.RelationFieldPos[mField] = si
			}
		}
	}
	records.sync()
	return records
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

func (m *ModelStructRecords) GetFieldType(field *ModelField) reflect.Type {
	return m.FieldTypes[field]
}

func (m *ModelStructRecords) GetFieldAddressType(field *ModelField) reflect.Type {
	return reflect.PtrTo(m.FieldTypes[field])
}

func (m *ModelStructRecords) GetRecords() []ModelRecord {
	var records []ModelRecord
	for i := 0; i < len(m.FieldValuesList); i++ {
		records = append(records, &ModelStructRecord{
			FieldValues:        m.FieldValuesList[i],
			VirtualFieldValues: m.VirtualFieldValuesList[i],
			source:             LoopIndirect(m.source.Index(i)),
			model:              m.model,
		})
	}
	return records
}

func (m *ModelStructRecords) GetRecord(i int) ModelRecord {
	return &ModelStructRecord{
		FieldValues:        m.FieldValuesList[i],
		VirtualFieldValues: m.VirtualFieldValuesList[i],
		source:             LoopIndirect(m.source.Index(i)),
		model:              m.model,
	}
}

func (m *ModelStructRecords) reallocate() {
	m.FieldValuesList = nil
	for i := 0; i < m.source.Len(); i++ {
		// why need loop indirect? because data type can be []*ModelData{} or []**ModelData{}
		c := map[*ModelField]reflect.Value{}

		fieldList := GetStructValueFields(LoopIndirectAndNew(m.source.Index(i)))
		for mField, si := range m.RelationFieldPos {
			c[mField] = fieldList[si]
		}
		m.FieldValuesList = append(m.FieldValuesList, c)
	}
}

func (m *ModelStructRecords) Add(v reflect.Value) ModelRecord {
	if m.source.CanSet() == false {
		panic("Add need can set permission")
	}
	if v.Type() != m.Type {
		panic(fmt.Sprintf("add data type(%s) must be %s", v.Type().Name(), m.Type.Name()))
	}
	last := m.source.Len()
	oldSource := m.source.Pointer()
	m.source.Set(reflect.Append(m.source, v))
	m.sync()
	// if list trigger the reallocate , reallocate
	if m.source.Pointer() != oldSource {
		m.reallocate()
	}
	return &ModelStructRecord{
		FieldValues:        m.FieldValuesList[last],
		VirtualFieldValues: m.VirtualFieldValuesList[last],
		source:             m.source.Index(last),
		model:              m.model,
	}
}

func (m *ModelStructRecords) sync() {
	for i := len(m.FieldValuesList); i < m.source.Len(); i++ {
		c := map[*ModelField]reflect.Value{}

		fieldList := GetStructValueFields(LoopIndirectAndNew(m.source.Index(i)))
		for mField, si := range m.RelationFieldPos {
			c[mField] = fieldList[si]
		}
		vc := map[*ModelField]reflect.Value{}
		m.FieldValuesList = append(m.FieldValuesList, c)
		m.VirtualFieldValuesList = append(m.VirtualFieldValuesList, vc)
	}
}

func (m *ModelStructRecords) IsVariableContainer() bool {
	return false
}

func (m *ModelStructRecords) ElemType() reflect.Type {
	return m.Type
}

func (m *ModelStructRecords) Len() int {
	return len(m.FieldValuesList)
}

func (m *ModelStructRecords) Source() reflect.Value {
	return m.source
}

type ModelStructRecord struct {
	FieldValues        map[*ModelField]reflect.Value
	VirtualFieldValues map[*ModelField]reflect.Value
	source             reflect.Value
	model              *Model
}

// set field to field value map or virtual field value map
// but if value is invalid delete it on map
func (m *ModelStructRecord) SetField(field *ModelField, value reflect.Value) {
	if field == nil {
		return
	}
	fieldValue := m.FieldValues[field]
	if value.Kind() == reflect.Ptr {
		panic("RecordFieldSetError: value cannot be a ptr")
	}
	if fieldValue.IsValid() == false {
		m.VirtualFieldValues[field] = reflect.New(field.Field.Type).Elem()
		fieldValue = m.VirtualFieldValues[field]
	}
	fieldValue = LoopIndirectAndNew(fieldValue)
	SafeSet(fieldValue, value)
	//fmt.Printf("source :%#v\n", m.source)
}

func (m *ModelStructRecord) Field(field *ModelField) reflect.Value {
	if v := m.FieldValues[field]; v.IsValid() {
		return v
	}
	return m.VirtualFieldValues[field]
}

func (m *ModelStructRecord) FieldAddress(field *ModelField) reflect.Value {
	if v := m.FieldValues[field]; v.IsValid() {
		return v.Addr()
	}
	return m.VirtualFieldValues[field].Addr()
}

func (m *ModelStructRecord) AllField() map[*ModelField]reflect.Value {
	fieldValues := map[*ModelField]reflect.Value{}
	for field, fieldValue := range m.FieldValues {
		fieldValues[field] = fieldValue
	}
	for field, fieldValue := range m.VirtualFieldValues {
		fieldValues[field] = fieldValue
	}
	return fieldValues
}

func (m *ModelStructRecord) IsVariableContainer() bool {
	return false
}

func (m *ModelStructRecord) Source() reflect.Value {
	return m.source
}

func (m *ModelStructRecord) GetFieldType(field *ModelField) reflect.Type {
	if fieldValue := m.FieldValues[field]; fieldValue.IsValid() {
		return fieldValue.Type()
	}
	return nil
}
