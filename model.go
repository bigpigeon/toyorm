/*
 * Copyright 2018. bigpigeon. All rights reserved.
 * Use of this source code is governed by a MIT style
 * license that can be found in the LICENSE file.
 */

package toyorm

import (
	"errors"
	"fmt"
	"reflect"
	"time"
)

type tabler interface {
	TableName() string
}

// table struct used to save all sql attribute with struct
// column is table name attribute
// AllFields is all struct field and it's Anonymous struct
// SqlFields is all about sql field
// OffsetFields is map field offset
// NameFields is map field name
// PrimaryFields is all about primary key field
// IndexFields is all about index fields
// UniqueIndexFields is all about unique index fields
// StructFieldFields is map unknown struct type or slice struct type
type Model struct {
	Name              string
	ReflectType       reflect.Type
	AllFields         []*modelField
	SqlFields         []*modelField
	OffsetFields      map[uintptr]*modelField
	NameFields        map[string]*modelField
	SqlFieldMap       map[string]*modelField
	PrimaryFields     []*modelField
	IndexFields       map[string][]*modelField
	UniqueIndexFields map[string][]*modelField
	StructFieldFields map[reflect.Type][]*modelField
	Association       [AssociationTypeEnd]map[string]*modelField
}

func (m *Model) GetPosFields(pos []int) []Field {
	fields := make([]Field, 0, len(pos))
	for _, i := range pos {
		fields = append(fields, m.AllFields[i])
	}
	return fields
}

func (m *Model) GetPosField(pos int) Field {
	return m.AllFields[pos]
}

func (m *Model) GetSqlFields() []Field {
	fields := make([]Field, len(m.SqlFields))
	for i, f := range m.SqlFields {
		fields[i] = f
	}
	return fields
}

func (m *Model) GetNameFieldMap() map[string]Field {
	fields := make(map[string]Field, len(m.NameFields))
	for n, f := range m.NameFields {
		fields[n] = f
	}
	return fields
}

func (m *Model) GetOffsetFieldMap() map[uintptr]Field {
	fields := make(map[uintptr]Field, len(m.OffsetFields))
	for o, f := range m.OffsetFields {
		fields[o] = f
	}
	return fields
}

func (m *Model) GetOffsetField(offset uintptr) Field {
	if field, ok := m.OffsetFields[offset]; ok {
		return field
	}
	return nil
}

func (m *Model) GetFieldWithName(name string) Field {
	if field, ok := m.NameFields[name]; ok {
		return field
	}
	return nil
}

func (m *Model) GetPrimary() []Field {
	fields := make([]Field, len(m.PrimaryFields))
	for i, f := range m.PrimaryFields {
		fields[i] = f
	}
	return fields
}

func (m *Model) GetOnePrimary() Field {
	if len(m.PrimaryFields) > 1 {
		panic(errors.New(fmt.Sprintf("%s have more than 1 primary", m.Name)))
	}
	return m.PrimaryFields[0]
}

func (m *Model) GetIndexMap() map[string][]Field {
	fieldMap := make(map[string][]Field, len(m.IndexFields))
	for s, fields := range m.IndexFields {
		fieldMap[s] = make([]Field, 0, len(fields))
		for _, f := range fields {
			fieldMap[s] = append(fieldMap[s], f)
		}
	}
	return fieldMap
}

func (m *Model) GetUniqueIndexMap() map[string][]Field {
	fieldMap := make(map[string][]Field, len(m.UniqueIndexFields))
	for s, fields := range m.UniqueIndexFields {
		fieldMap[s] = make([]Field, 0, len(fields))
		for _, f := range fields {
			fieldMap[s] = append(fieldMap[s], f)
		}
	}
	return fieldMap
}

func NewModel(val reflect.Value) *Model {
	if val.Kind() != reflect.Struct {
		panic(ErrInvalidModelType(val.Type().Name()))
	}
	if val.Type().Name() == "" {
		panic(ErrInvalidModelName{})
	}
	return newModel(val, ModelName(val))
}

func NewMiddleModel(model, subModel *Model) *Model {
	return newMiddleModel(model, subModel, `toyorm:"primary key"`)
}

func newMiddleModel(model, subModel *Model, tag reflect.StructTag) *Model {
	var fields [2]reflect.StructField
	var sortdModel [2]*Model
	if model.Name < subModel.Name {
		sortdModel = [2]*Model{model, subModel}
	} else {
		sortdModel = [2]*Model{subModel, model}
	}
	for i, model := range sortdModel {
		field := model.GetOnePrimary().StructField()
		field.Tag = tag
		field.Name = GetRelationFieldName(model)
		fields[i] = field
	}
	if fields[0].Name == fields[1].Name {
		fields[0].Name = "L_" + fields[0].Name
		fields[1].Name = "R_" + fields[1].Name
	}

	return newModel(reflect.Zero(reflect.StructOf(fields[:])), fmt.Sprintf("%s_%s", sortdModel[0].Name, sortdModel[1].Name))
}

func newModel(val reflect.Value, modelName string) *Model {
	_type := val.Type()
	model := &Model{
		Name:              modelName,
		ReflectType:       _type,
		OffsetFields:      map[uintptr]*modelField{},
		NameFields:        map[string]*modelField{},
		SqlFieldMap:       map[string]*modelField{},
		PrimaryFields:     []*modelField{},
		IndexFields:       map[string][]*modelField{},
		UniqueIndexFields: map[string][]*modelField{},
		StructFieldFields: map[reflect.Type][]*modelField{},
		Association:       [AssociationTypeEnd]map[string]*modelField{},
	}
	for i := range model.Association {
		model.Association[i] = make(map[string]*modelField)
	}
	for i := 0; i < _type.NumField(); i++ {
		field := _type.Field(i)
		fieldVal := val.Field(i)
		tField := NewField(&field, fieldVal, model.Name)
		if field.Anonymous && field.Type.Kind() == reflect.Struct && tField.Name() == "" {
			embedTable := newModel(fieldVal, model.Name)
			for _, tabField := range embedTable.AllFields {
				tabField.offset += field.Offset
				model.AllFields = append(model.AllFields, tabField)
			}
		} else {
			model.AllFields = append(model.AllFields, tField)
		}
	}
	// cache field attribute with model
	for i := 0; i < len(model.AllFields); i++ {
		field := model.AllFields[i]

		model.OffsetFields[field.offset] = field
		if _, ok := model.NameFields[field.Name()]; ok {
			panic(ErrRepeatField{model.Name, field.field.Name})
		}
		model.NameFields[field.Name()] = field

		if field.isPrimary {
			model.PrimaryFields = append(model.PrimaryFields, field)
		}
		if field.index != "" {
			model.IndexFields[field.index] = append(model.IndexFields[field.index], field)
		}
		if field.uniqueIndex != "" {
			model.UniqueIndexFields[field.uniqueIndex] = append(model.UniqueIndexFields[field.uniqueIndex], field)
		}
		for association, val := range field.Association {
			if _, ok := model.Association[association][val]; ok {
				panic(ErrModelDuplicateAssociation{model.Name, association, field.field.Name})
			}
			model.Association[association][val] = field
		}
		if field.ignore == false {
			if oldField, ok := model.SqlFieldMap[field.column]; ok {
				panic(ErrSameColumnName{model.Name, field.column, oldField.field.Name, field.field.Name})
			}
			model.SqlFieldMap[field.column] = field
			model.SqlFields = append(model.SqlFields, field)
		} else {
			// preload
			var fieldType = LoopTypeIndirectSliceAndPtr(field.field.Type)
			model.StructFieldFields[fieldType] = append(model.StructFieldFields[fieldType], field)
		}
	}
	return model
}

type FieldSelection interface{}

func (m *Model) fieldSelect(v FieldSelection) Field {
	switch v := v.(type) {
	case int:
		return m.SqlFields[v]
	case uintptr:
		return m.OffsetFields[v]
	case string:
		return m.NameFields[v]
	case Field:
		return v
	default:
		panic("invalid field value")
	}
}

type ModelDefault struct {
	ID        uint32     `toyorm:"primary key;auto_increment"`
	CreatedAt time.Time  `toyorm:"NULL"`
	UpdatedAt time.Time  `toyorm:"NULL"`
	DeletedAt *time.Time `toyorm:"index;NULL"`
}
