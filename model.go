package toyorm

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"
)

type tabler interface {
	TableName() string
}

// table struct used to save all sql attribute with struct
// Name is table name attribute
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
	AllFields         []*ModelField
	SqlFields         []*ModelField
	FieldsPos         map[*ModelField]int
	OffsetFields      map[uintptr]*ModelField
	NameFields        map[string]*ModelField
	LowerNameFields   map[string]*ModelField
	PrimaryFields     []*ModelField
	IndexFields       map[string][]*ModelField
	UniqueIndexFields map[string][]*ModelField
	StructFieldFields map[reflect.Type][]*ModelField
}

func (t *Model) GetPosFields(pos []int) []*ModelField {
	fields := make([]*ModelField, 0, len(pos))
	for _, i := range pos {
		fields = append(fields, t.AllFields[i])
	}
	return fields
}

func (t *Model) GetPosField(pos int) *ModelField {
	return t.AllFields[pos]
}

func (t *Model) GetSqlFields() []*ModelField {
	return t.SqlFields
}

func (t *Model) GetOffsetField(offset uintptr) *ModelField {
	return t.OffsetFields[offset]
}

func (t *Model) GetFieldWithName(name string) *ModelField {
	return t.NameFields[name]
}

func (t *Model) GetPrimary() []*ModelField {
	return t.PrimaryFields
}

func (t *Model) GetOnePrimary() *ModelField {
	if len(t.PrimaryFields) > 1 {
		panic(errors.New(fmt.Sprintf("%s have more than 1 primary", t.Name)))
	}
	return t.PrimaryFields[0]
}

func (t *Model) GetIndexMap() map[string][]*ModelField {
	return t.IndexFields
}

func (t *Model) GetUniqueIndexMap() map[string][]*ModelField {
	return t.UniqueIndexFields
}

func NewModel(_type reflect.Type, dia Dialect) *Model {
	if _type.Kind() != reflect.Struct {
		panic(ErrInvalidModelType(_type.Name()))
	}
	if _type.Name() == "" {
		panic(ErrInvalidModelName{})
	}
	var modelName string
	if v, ok := reflect.New(_type).Interface().(tabler); ok {
		modelName = v.TableName()
	} else {
		modelName = SqlNameConvert(_type.Name())
	}
	return newModel(_type, modelName)
}

func NewMiddleModel(model, subModel *Model) *Model {
	var fields [2]reflect.StructField
	var sortdModel [2]*Model
	if model.Name < subModel.Name {
		sortdModel = [2]*Model{model, subModel}
	} else {
		sortdModel = [2]*Model{subModel, model}
	}
	for i, model := range sortdModel {
		field := model.GetOnePrimary().Field
		field.Tag = `toyorm:"primary key"`
		field.Name = GetRelationFieldName(model)
		fields[i] = field
	}
	if fields[0].Name == fields[1].Name {
		fields[0].Name = "L_" + fields[0].Name
		fields[1].Name = "R_" + fields[1].Name
	}
	return newModel(reflect.StructOf(fields[:]), fmt.Sprintf("%s_%s", sortdModel[0].Name, sortdModel[1].Name))
}

func newModel(_type reflect.Type, modelName string) *Model {
	model := &Model{
		Name:              modelName,
		ReflectType:       _type,
		FieldsPos:         map[*ModelField]int{},
		OffsetFields:      map[uintptr]*ModelField{},
		NameFields:        map[string]*ModelField{},
		LowerNameFields:   map[string]*ModelField{},
		PrimaryFields:     []*ModelField{},
		IndexFields:       map[string][]*ModelField{},
		UniqueIndexFields: map[string][]*ModelField{},
		StructFieldFields: map[reflect.Type][]*ModelField{},
	}

	for i := 0; i < _type.NumField(); i++ {
		field := _type.Field(i)
		if field.Anonymous && field.Type.Kind() == reflect.Struct {
			embedTable := newModel(field.Type, model.Name)
			for _, tabField := range embedTable.AllFields {
				tabField.Offset += field.Offset
				model.AllFields = append(model.AllFields, tabField)
			}
		} else {
			tField := NewField(&field, model.Name)
			model.AllFields = append(model.AllFields, tField)
		}
	}
	// cache field attribute with model
	for i := 0; i < len(model.AllFields); i++ {
		field := model.GetPosField(i)
		model.FieldsPos[field] = i
		model.OffsetFields[field.Offset] = field
		if _, ok := model.NameFields[field.Field.Name]; ok {
			panic(ErrRepeatFieldName)
		}
		if _, ok := model.LowerNameFields[strings.ToLower(field.Field.Name)]; ok {
			panic(ErrRepeatFieldName)
		}
		model.NameFields[field.Field.Name] = field
		model.LowerNameFields[strings.ToLower(field.Field.Name)] = field
		if field.PrimaryKey {
			model.PrimaryFields = append(model.PrimaryFields, field)
		}
		if field.Index != "" {
			if _, ok := model.IndexFields[field.Index]; ok == false {
				model.IndexFields[field.Index] = []*ModelField{}
			}
			model.IndexFields[field.Index] = append(model.IndexFields[field.Index], field)
		}
		if field.UniqueIndex != "" {
			if _, ok := model.UniqueIndexFields[field.UniqueIndex]; ok == false {
				model.UniqueIndexFields[field.UniqueIndex] = []*ModelField{}
			}
			model.UniqueIndexFields[field.UniqueIndex] = append(model.UniqueIndexFields[field.UniqueIndex], field)
		}
		if field.Ignore == false {
			model.SqlFields = append(model.SqlFields, field)
		} else {
			// preload
			var fieldType = LoopTypeIndirectSliceAndPtr(field.Field.Type)
			model.StructFieldFields[fieldType] = append(model.StructFieldFields[fieldType], field)
		}
	}
	return model
}

func (m *Model) fieldSelect(v interface{}) *ModelField {
	switch v := v.(type) {
	case int:
		return m.SqlFields[v]
	case uintptr:
		return m.OffsetFields[v]
	case string:
		return m.NameFields[v]
	case *ModelField:
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

var StarField = &ModelField{
	Name:          "*",
	Offset:        0,
	Ignore:        false,
	CommonAttr:    map[string]string{},
	AutoIncrement: false,
	Field: reflect.StructField{
		Name: "*",
		Type: reflect.TypeOf(struct{}{}),
	},
}
