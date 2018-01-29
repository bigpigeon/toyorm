package toyorm

import (
	"database/sql"
	"io"
	"os"
	"reflect"
)

type Toy struct {
	db                       *sql.DB
	CacheModels              map[reflect.Type]*Model
	CacheMiddleModels        map[reflect.Type]*Model
	CacheReverseMiddleModels map[reflect.Type]*Model
	// map[model][container_field_name]
	oneToOnePreload          map[*Model]map[string]*OneToOnePreload
	oneToManyPreload         map[*Model]map[string]*OneToManyPreload
	manyToManyPreload        map[*Model]map[string]*ManyToManyPreload
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
		oneToOnePreload:   map[*Model]map[string]*OneToOnePreload{},
		oneToManyPreload:  map[*Model]map[string]*OneToManyPreload{},
		manyToManyPreload: map[*Model]map[string]*ManyToManyPreload{},
		Dialect:           dialect,
		DefaultHandlerChain: map[string]HandlersChain{
			"CreateTable":           {HandlerSimplePreload("CreateTable"), HandlerCreateTable},
			"CreateTableIfNotExist": {HandlerSimplePreload("CreateTableIfNotExist"), HandlerExistTableAbort, HandlerCreateTable},
			"DropTableIfExist":      {HandlerReversePreload("DropTableIfExist"), HandlerNotExistTableAbort, HandlerDropTable},
			"DropTable":             {HandlerReversePreload("DropTable"), HandlerDropTable},
			"Insert":                {HandlerPreloadInsertOrSave("Insert"), HandlerInsertTimeGenerate, HandlerInsert},
			"Find":                  {HandlerSoftDeleteCheck, HandlerFind, HandlerPreloadFind},
			"Update":                {HandlerSoftDeleteCheck, HandlerUpdateTimeGenerate, HandlerUpdate},
			"Save":                  {HandlerPreloadInsertOrSave("Save"), HandlerSaveTimeGenerate, HandlerSave},
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
	// lazy init model
	model = t.GetModel(vType)
	toyBrick := NewToyBrick(t, model)
	return toyBrick
}

func (t *Toy) MiddleModel(v, sv interface{}) *ToyBrick {
	vType := LoopTypeIndirect(reflect.ValueOf(v).Type())
	svType := LoopTypeIndirect(reflect.ValueOf(sv).Type())
	model, subModel := t.GetModel(vType), t.GetModel(svType)
	middleModel := NewMiddleModel(model, subModel)
	return NewToyBrick(t, middleModel)
}

// TODO testing thread safe? if not add lock
func (t *Toy) GetModel(_type reflect.Type) *Model {
	if model, ok := t.CacheModels[_type]; ok == false {
		model = NewModel(_type)
		t.CacheModels[_type] = model
	}
	return t.CacheModels[_type]
}

func (t *Toy) GetMiddleModel(_type reflect.Type) *Model {
	if model, ok := t.CacheModels[_type]; ok == false {
		model = NewModel(_type)
		t.CacheModels[_type] = model
	}
	return t.CacheModels[_type]
}

func (t *Toy) OneToOnePreload(model *Model, field Field) *OneToOnePreload {
	// try to find cache data
	if t.oneToOnePreload[model] != nil && t.oneToOnePreload[model][field.Name()] != nil {
		return t.oneToOnePreload[model][field.Name()]
	}
	_type := LoopTypeIndirect(field.StructField().Type)
	if _type.Kind() != reflect.Struct {
		return nil
	}
	if subModel := t.CacheModels[_type]; subModel != nil {
		if relationField := subModel.GetFieldWithName(GetRelationFieldName(model)); relationField != nil {

			return t.OneToOneBind(model, subModel, field, relationField, false)
		} else if relationField := model.GetFieldWithName(GetBelongsIDFieldName(subModel, field)); relationField != nil {

			return t.OneToOneBind(model, subModel, field, relationField, true)
		}
	}

	return nil
}

func (t *Toy) OneToManyPreload(model *Model, field Field) *OneToManyPreload {
	// try to find cache data
	if t.oneToManyPreload[model] != nil && t.oneToManyPreload[model][field.Name()] != nil {
		return t.oneToManyPreload[model][field.Name()]
	}

	_type := LoopTypeIndirect(field.StructField().Type)
	if _type.Kind() == reflect.Slice {
		elemType := LoopTypeIndirect(_type.Elem())
		if subModel, ok := t.CacheModels[elemType]; ok {
			if relationField := subModel.GetFieldWithName(GetRelationFieldName(model)); relationField != nil {

				return t.OneToManyBind(model, subModel, field, relationField)
			}
		}
	}
	return nil
}

func (t *Toy) ManyToManyPreload(model *Model, field Field, isRight bool) *ManyToManyPreload {
	// try to find cache data
	if t.manyToManyPreload[model] != nil && t.manyToManyPreload[model][field.Name()] != nil {
		return t.manyToManyPreload[model][field.Name()]
	}
	if v := t.manyToManyPreload[model]; v == nil {
		t.manyToManyPreload[model] = map[string]*ManyToManyPreload{}
	}

	_type := LoopTypeIndirect(field.StructField().Type)
	if _type.Kind() == reflect.Slice {
		elemType := LoopTypeIndirect(_type.Elem())
		if subModel, ok := t.CacheModels[elemType]; ok {
			middleModel := NewMiddleModel(model, subModel)
			relationField := GetMiddleField(model, middleModel, isRight)
			subRelationField := GetMiddleField(subModel, middleModel, !isRight)
			return t.ManyToManyPreloadBind(model, subModel, middleModel, field, relationField, subRelationField)
		}
	}
	return nil
}

func (t *Toy) OneToOneBind(model, subModel *Model, containerField, relationField Field, isBelongTo bool) *OneToOnePreload {
	if isBelongTo {
		if realField := model.NameFields[relationField.Name()]; realField.isForeign {
			realField.foreignModel = subModel
		}
	} else {
		if realField := subModel.NameFields[relationField.Name()]; realField.isForeign {
			realField.foreignModel = model
		}
	}

	if v := t.oneToOnePreload[model]; v == nil {
		t.oneToOnePreload[model] = map[string]*OneToOnePreload{}
	}
	t.oneToOnePreload[model][containerField.Name()] = &OneToOnePreload{
		IsBelongTo:     isBelongTo,
		Model:          model,
		SubModel:       subModel,
		RelationField:  relationField,
		ContainerField: containerField,
	}
	return t.oneToOnePreload[model][containerField.Name()]
}
func (t *Toy) OneToManyBind(model, subModel *Model, containerField, relationField Field) *OneToManyPreload {
	if realField := subModel.NameFields[relationField.Name()]; realField.isForeign {
		realField.foreignModel = model
	}
	if v := t.oneToManyPreload[model]; v == nil {
		t.oneToManyPreload[model] = map[string]*OneToManyPreload{}
	}
	t.oneToManyPreload[model][containerField.Name()] = &OneToManyPreload{
		Model:          model,
		SubModel:       subModel,
		RelationField:  relationField,
		ContainerField: containerField,
	}
	return t.oneToManyPreload[model][containerField.Name()]
}

func (t *Toy) ManyToManyPreloadBind(model, subModel, middleModel *Model, containerField, relationField, subRelationField Field) *ManyToManyPreload {
	if realField := middleModel.NameFields[relationField.Name()]; realField.isForeign {
		realField.foreignModel = model
	}
	if realField := middleModel.NameFields[subRelationField.Name()]; realField.isForeign {
		realField.foreignModel = subModel
	}

	if v := t.manyToManyPreload[model]; v == nil {
		t.manyToManyPreload[model] = map[string]*ManyToManyPreload{}
	}
	t.CacheMiddleModels[middleModel.ReflectType] = middleModel
	t.manyToManyPreload[model][containerField.Name()] = &ManyToManyPreload{
		Model:            model,
		SubModel:         subModel,
		MiddleModel:      middleModel,
		ContainerField:   containerField,
		RelationField:    relationField,
		SubRelationField: subRelationField,
	}
	return t.manyToManyPreload[model][containerField.Name()]
}

func (t *Toy) ModelHandlers(option string, model *Model) HandlersChain {
	handlers := make(HandlersChain, 0, len(t.DefaultHandlerChain[option])+len(t.DefaultModelHandlerChain[option][model]))
	handlers = append(handlers, t.DefaultModelHandlerChain[option][model]...)
	handlers = append(handlers, t.DefaultHandlerChain[option]...)
	return handlers
}
