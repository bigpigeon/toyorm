package toyorm

import (
	"database/sql"
	"io"
	"os"
	"reflect"
	"strings"
)

type Toy struct {
	db                       *sql.DB
	CacheModels              map[reflect.Type]*Model
	CacheMiddleModels        map[reflect.Type]*Model
	CacheReverseMiddleModels map[reflect.Type]*Model
	// map[model][container_field]
	oneToOnePreload          map[*Model]map[*ModelField]*OneToOnePreload
	oneToManyPreload         map[*Model]map[*ModelField]*OneToManyPreload
	manyToManyPreload        map[*Model]map[*ModelField]*ManyToManyPreload
	Dialect                  Dialect
	DefaultHandlerChain      map[string]HandlersChain
	DefaultModelHandlerChain map[string]map[*Model]HandlersChain
	Logger                   io.Writer
}

type RowsTree struct {
	Rows  *sql.Rows
	Child map[int]*sql.Rows
}

func Open(driverName, dataSourceName string) (*Toy, error) {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}
	var dialect Dialect
	switch driverName {
	case "mysql":
		dialect = MySqlDialect{}
	case "sqlite3":
		dialect = Sqlite3Dialect{}
	default:
		panic(ErrNotMatchDialect)
	}
	return &Toy{
		db:                db,
		CacheModels:       map[reflect.Type]*Model{},
		CacheMiddleModels: map[reflect.Type]*Model{},
		oneToOnePreload:   map[*Model]map[*ModelField]*OneToOnePreload{},
		oneToManyPreload:  map[*Model]map[*ModelField]*OneToManyPreload{},
		manyToManyPreload: map[*Model]map[*ModelField]*ManyToManyPreload{},
		Dialect:           dialect,
		DefaultHandlerChain: map[string]HandlersChain{
			"CreateTable":           {HandlerNotRecordPreload("CreateTable"), HandlerCreateTable},
			"CreateTableIfNotExist": {HandlerNotRecordPreload("CreateTableIfNotExist"), HandlerExistTableAbort, HandlerCreateTable},
			"DropTableIfExist":      {HandlerNotRecordPreload("DropTableIfExist"), HandlerNotExistTableAbort, HandlerDropTable},
			"DropTable":             {HandlerNotRecordPreload("DropTable"), HandlerDropTable},
			"Insert":                {HandlerPreloadInsertOrSave("Insert"), HandlerInsertTimeGenerate, HandlerInsert},
			"Find":                  {HandlerSoftDeleteCheck, HandlerFind, HandlerPreloadFind},
			"Update":                {HandlerSoftDeleteCheck, HandlerUpdateTimeGenerate, HandlerUpdate},
			"Save":                  {HandlerPreloadInsertOrSave("Save"), HandlerSaveTimeProcess, HandlerUpdateTimeGenerate, HandlerSave},
			"HardDelete":            {HandlerPreloadDelete, HandlerHardDelete},
			"SoftDelete":            {HandlerPreloadDelete, HandlerSoftDelete},
		},
		DefaultModelHandlerChain: map[string]map[*Model]HandlersChain{},
		Logger: os.Stdout,
	}, nil
}

func (t *Toy) Model(v interface{}) *ToyBrick {
	var model *Model
	vType := LoopTypeIndirect(reflect.ValueOf(v).Type())
	if vType.Kind() != reflect.Struct {
		panic("v must be struct or address with struct")
	}
	// lazy init model
	model = t.GetModel(vType)
	toyBrick := NewToyBrick(t, model)
	return toyBrick
}

func (t *Toy) MiddleModel(v, sv interface{}) *ToyBrick {
	vType := LoopTypeIndirect(reflect.ValueOf(v).Type())
	svType := LoopTypeIndirect(reflect.ValueOf(sv).Type())
	model, subModel := t.GetModel(vType), t.GetModel(svType)
	middleModel := NewMiddleModel(model, subModel, t.Dialect)
	return NewToyBrick(t, middleModel)
}

func (t *Toy) GetModel(_type reflect.Type) *Model {
	if model, ok := t.CacheModels[_type]; ok == false {
		model = NewModel(_type, t.Dialect)
		t.CacheModels[_type] = model
	}
	return t.CacheModels[_type]
}

func (t *Toy) GetMiddleModel(_type reflect.Type) *Model {
	if model, ok := t.CacheModels[_type]; ok == false {
		model = NewModel(_type, t.Dialect)
		t.CacheModels[_type] = model
	}
	return t.CacheModels[_type]
}

func (t *Toy) OneToOnePreload(model *Model, field *ModelField) *OneToOnePreload {
	// try to find cache data
	if t.oneToOnePreload[model] != nil && t.oneToOnePreload[model][field] != nil {
		return t.oneToOnePreload[model][field]
	}
	_type := LoopTypeIndirect(field.Field.Type)
	if _type.Kind() != reflect.Struct {
		return nil
	}
	if subModel := t.CacheModels[_type]; subModel != nil {
		if relationField, ok := subModel.LowerNameFields[strings.ToLower(GetRelationFieldName(model))]; ok {
			if v := t.oneToOnePreload[model]; v == nil {
				t.oneToOnePreload[model] = map[*ModelField]*OneToOnePreload{}
			}
			t.oneToOnePreload[model][field] = &OneToOnePreload{
				IsBelongTo:     false,
				Model:          model,
				SubModel:       subModel,
				RelationField:  relationField,
				ContainerField: field,
			}

		} else if relationField, ok := model.LowerNameFields[strings.ToLower(GetBelongsIDFieldName(subModel, field))]; ok {

			if v := t.oneToOnePreload[model]; v == nil {
				t.oneToOnePreload[model] = map[*ModelField]*OneToOnePreload{}
			}
			t.oneToOnePreload[model][field] = &OneToOnePreload{
				IsBelongTo:     true,
				Model:          model,
				SubModel:       subModel,
				RelationField:  relationField,
				ContainerField: field,
			}
		}
	}
	return t.oneToOnePreload[model][field]
}

func (t *Toy) OneToManyPreload(model *Model, field *ModelField) *OneToManyPreload {
	// try to find cache data
	if t.oneToManyPreload[model] != nil && t.oneToManyPreload[model][field] != nil {
		return t.oneToManyPreload[model][field]
	}

	_type := LoopTypeIndirect(field.Field.Type)
	if _type.Kind() == reflect.Slice {
		elemType := LoopTypeIndirect(_type.Elem())
		if subModel, ok := t.CacheModels[elemType]; ok {
			if relationField, ok := subModel.LowerNameFields[strings.ToLower(GetRelationFieldName(model))]; ok {
				if v := t.oneToManyPreload[model]; v == nil {
					t.oneToManyPreload[model] = map[*ModelField]*OneToManyPreload{}
				}
				t.oneToManyPreload[model][field] = &OneToManyPreload{
					Model:          model,
					SubModel:       subModel,
					RelationField:  relationField,
					ContainerField: field,
				}
			}
		}
	}
	return t.oneToManyPreload[model][field]
}

func (t *Toy) ManyToManyPreload(model *Model, field *ModelField) *ManyToManyPreload {
	// try to find cache data
	if t.manyToManyPreload[model] != nil && t.manyToManyPreload[model][field] != nil {
		return t.manyToManyPreload[model][field]
	}
	if v := t.manyToManyPreload[model]; v == nil {
		t.manyToManyPreload[model] = map[*ModelField]*ManyToManyPreload{}
	}

	_type := LoopTypeIndirect(field.Field.Type)
	if _type.Kind() == reflect.Slice {
		elemType := LoopTypeIndirect(_type.Elem())
		if subModel, ok := t.CacheModels[elemType]; ok {
			middleModel := NewMiddleModel(model, subModel, t.Dialect)
			t.CacheMiddleModels[middleModel.ReflectType] = middleModel
			t.manyToManyPreload[model][field] = &ManyToManyPreload{
				MiddleModel:    middleModel,
				Model:          model,
				SubModel:       subModel,
				ContainerField: field,
			}
		}
	}
	return t.manyToManyPreload[model][field]
}

func (t *Toy) OneToOnePreloadBind(model, subTable *Model, relationField, containerField *ModelField, isOwn bool) {
	preload := OneToOnePreload{
		Model:          model,
		SubModel:       subTable,
		RelationField:  relationField,
		ContainerField: containerField,
		IsBelongTo:     isOwn,
	}
	t.oneToOnePreload[model][containerField] = &preload
}

func (t *Toy) OneToManyPreloadBind(model, subTable *Model, relationField, containerField *ModelField) {
	preload := OneToManyPreload{
		Model:          model,
		SubModel:       subTable,
		RelationField:  relationField,
		ContainerField: containerField,
	}
	t.oneToManyPreload[model][containerField] = &preload
}

func (t *Toy) ManyToManyPreloadBind(model, subTable, middleTable *Model, containerField *ModelField) {
	preload := ManyToManyPreload{
		Model:          model,
		SubModel:       subTable,
		MiddleModel:    middleTable,
		ContainerField: containerField,
	}
	t.manyToManyPreload[model][containerField] = &preload
}

func (t *Toy) ModelHandlers(option string, model *Model) HandlersChain {
	handlers := make(HandlersChain, 0, len(t.DefaultHandlerChain[option])+len(t.DefaultModelHandlerChain[option][model]))
	handlers = append(handlers, t.DefaultModelHandlerChain[option][model]...)
	handlers = append(handlers, t.DefaultHandlerChain[option]...)
	return handlers
}
