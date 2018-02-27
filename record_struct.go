package toyorm

import (
	"fmt"
	"reflect"
)

type ModelStructRecords struct {
	model                  *Model
	Type                   reflect.Type
	source                 reflect.Value
	RelationFieldPos       map[string]int
	FieldTypes             map[string]reflect.Type
	FieldValuesList        []map[string]reflect.Value
	VirtualFieldValuesList []map[string]reflect.Value
}

func NewStructRecords(model *Model, value reflect.Value) *ModelStructRecords {
	vElemType := value.Type().Elem()
	records := &ModelStructRecords{
		model,
		vElemType,
		value,
		map[string]int{},
		map[string]reflect.Type{},
		[]map[string]reflect.Value{},
		[]map[string]reflect.Value{},
	}
	vIndirectElemType := LoopTypeIndirectSliceAndPtr(vElemType)
	if vIndirectElemType == model.ReflectType {
		for i := 0; i < len(model.AllFields); i++ {
			f := model.GetPosField(i)
			records.FieldTypes[f.Name()] = f.StructField().Type
			records.RelationFieldPos[f.Name()] = i
		}
	} else {
		structFieldList := GetStructFields(vIndirectElemType)
		repeatName := map[string]struct{}{}
		// RelationFieldPos[model_field]->struct_field_pos
		for si, field := range structFieldList {
			if _, ok := repeatName[field.Name]; ok {

			}
			if mField := model.GetFieldWithName(field.Name); mField != nil {
				records.FieldTypes[field.Name] = field.Type
				records.RelationFieldPos[field.Name] = si
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

	record := map[string]reflect.Value{}
	if vtype == model.ReflectType {
		for i := 0; i < len(model.AllFields); i++ {
			record[model.GetPosField(i).Name()] = fieldValueList[i]
		}
	} else {
		for i, field := range structFieldList {
			if mField := model.GetFieldWithName(field.Name); mField != nil {
				record[field.Name] = fieldValueList[i]
			}
		}
	}
	return &ModelStructRecord{
		FieldValues:        record,
		VirtualFieldValues: map[string]reflect.Value{},
		source:             value,
		model:              model,
	}
}

func (m *ModelStructRecords) GetFieldType(field string) reflect.Type {
	return m.FieldTypes[field]
}

func (m *ModelStructRecords) GetFieldAddressType(field string) reflect.Type {
	return reflect.PtrTo(m.FieldTypes[field])
}

func (m *ModelStructRecords) GroupBy(key string) ModelGroupBy {
	field := m.model.GetFieldWithName(key)
	if field.StructField().Type.Comparable() == false {
		panic(fmt.Sprintf("%s is not compareable field", field.Name()))
	}
	result := ModelGroupBy{}
	for i := 0; i < len(m.FieldValuesList); i++ {
		keyValue := m.FieldValuesList[i][key].Interface()
		result[keyValue] = append(result[keyValue], ModelIndexRecord{&ModelStructRecord{
			FieldValues:        m.FieldValuesList[i],
			VirtualFieldValues: m.VirtualFieldValuesList[i],
			source:             LoopIndirect(m.source.Index(i)),
			model:              m.model,
		}, i})
	}
	return result
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
		c := map[string]reflect.Value{}

		fieldList := GetStructValueFields(LoopIndirectAndNew(m.source.Index(i)))
		for name, si := range m.RelationFieldPos {
			c[name] = fieldList[si]
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
		c := map[string]reflect.Value{}

		fieldList := GetStructValueFields(LoopIndirectAndNew(m.source.Index(i)))
		for name, si := range m.RelationFieldPos {
			c[name] = fieldList[si]
		}
		vc := map[string]reflect.Value{}
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
	FieldValues        map[string]reflect.Value
	VirtualFieldValues map[string]reflect.Value
	source             reflect.Value
	model              *Model
}

// set field to field value map or virtual field value map
// but if value is invalid delete it on map
func (m *ModelStructRecord) SetField(name string, value reflect.Value) {
	if name == "" {
		return
	}
	fieldValue := m.FieldValues[name]
	//if value.Kind() == reflect.Ptr {
	//	panic("RecordFieldSetError: value cannot be a ptr")
	//}
	if fieldValue.IsValid() == false {
		m.VirtualFieldValues[name] = reflect.New(m.model.GetFieldWithName(name).StructField().Type).Elem()
		fieldValue = m.VirtualFieldValues[name]
	}
	//fieldValue = LoopIndirectAndNew(fieldValue)
	safeSet(fieldValue, value)
	//fmt.Printf("source :%#v\n", m.source)
}

func (m *ModelStructRecord) Field(name string) reflect.Value {
	if v := m.FieldValues[name]; v.IsValid() {
		return v
	}
	return m.VirtualFieldValues[name]
}

func (m *ModelStructRecord) FieldAddress(name string) reflect.Value {
	if v := m.FieldValues[name]; v.IsValid() {
		return v.Addr()
	}
	return m.VirtualFieldValues[name].Addr()
}

func (m *ModelStructRecord) AllField() map[string]reflect.Value {
	fieldValues := map[string]reflect.Value{}
	for name, fieldValue := range m.FieldValues {
		fieldValues[name] = fieldValue
	}
	for name, fieldValue := range m.VirtualFieldValues {
		fieldValues[name] = fieldValue
	}
	return fieldValues
}

func (m *ModelStructRecord) IsVariableContainer() bool {
	return false
}

func (m *ModelStructRecord) Source() reflect.Value {
	return m.source
}

func (m *ModelStructRecord) GetFieldType(name string) reflect.Type {
	if fieldValue := m.FieldValues[name]; fieldValue.IsValid() {
		return fieldValue.Type()
	}
	return nil
}
