package toyorm

import (
	"database/sql"
	"io"
	"os"
	"reflect"
)

type ToyKernel struct {
	CacheModels              map[reflect.Type]*Model
	CacheMiddleModels        map[reflect.Type]*Model
	CacheReverseMiddleModels map[reflect.Type]*Model
	// map[model][container_field_name]

	belongToPreload   map[*Model]map[string]*BelongToPreload
	oneToOnePreload   map[*Model]map[string]*OneToOnePreload
	oneToManyPreload  map[*Model]map[string]*OneToManyPreload
	manyToManyPreload map[*Model]map[string]map[bool]*ManyToManyPreload
	Dialect           Dialect
	Logger            io.Writer
}

type Toy struct {
	db                       *sql.DB
	DefaultHandlerChain      map[string]HandlersChain
	DefaultModelHandlerChain map[*Model]map[string]HandlersChain
	ToyKernel
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
		db: db,
		DefaultHandlerChain: map[string]HandlersChain{
			"CreateTable":              {HandlerSimplePreload("CreateTable"), HandlerCreateTable},
			"CreateTableIfNotExist":    {HandlerSimplePreload("CreateTableIfNotExist"), HandlerExistTableAbort, HandlerCreateTable},
			"DropTableIfExist":         {HandlerDropTablePreload("DropTableIfExist"), HandlerNotExistTableAbort, HandlerDropTable},
			"DropTable":                {HandlerDropTablePreload("DropTable"), HandlerDropTable},
			"Insert":                   {HandlerPreloadContainerCheck, HandlerPreloadInsertOrSave("Insert"), HandlerInsertTimeGenerate, HandlerInsert},
			"Find":                     {HandlerPreloadContainerCheck, HandlerSoftDeleteCheck, HandlerFind, HandlerPreloadFind},
			"Update":                   {HandlerSoftDeleteCheck, HandlerUpdateTimeGenerate, HandlerUpdate},
			"Save":                     {HandlerPreloadContainerCheck, HandlerPreloadInsertOrSave("Save"), HandlerSaveTimeGenerate, HandlerSave},
			"HardDelete":               {HandlerPreloadDelete, HandlerHardDelete},
			"SoftDelete":               {HandlerPreloadDelete, HandlerSoftDelete},
			"HardDeleteWithPrimaryKey": {HandlerPreloadDelete, HandlerSearchWithPrimaryKey, HandlerHardDelete},
			"SoftDeleteWithPrimaryKey": {HandlerPreloadDelete, HandlerSearchWithPrimaryKey, HandlerSoftDelete},
		},
		DefaultModelHandlerChain: map[*Model]map[string]HandlersChain{},
		ToyKernel: ToyKernel{
			CacheModels:       map[reflect.Type]*Model{},
			CacheMiddleModels: map[reflect.Type]*Model{},
			belongToPreload:   map[*Model]map[string]*BelongToPreload{},
			oneToOnePreload:   map[*Model]map[string]*OneToOnePreload{},
			oneToManyPreload:  map[*Model]map[string]*OneToManyPreload{},
			// because have isRight feature need two point to save
			manyToManyPreload: map[*Model]map[string]map[bool]*ManyToManyPreload{},
			Dialect:           dialect,
			Logger:            os.Stdout,
		},
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
func (t *ToyKernel) GetModel(_type reflect.Type) *Model {
	if model, ok := t.CacheModels[_type]; ok == false {
		model = NewModel(_type)
		t.CacheModels[_type] = model
	}
	return t.CacheModels[_type]
}

func (t *ToyKernel) GetMiddleModel(_type reflect.Type) *Model {
	if model, ok := t.CacheModels[_type]; ok == false {
		model = NewModel(_type)
		t.CacheModels[_type] = model
	}
	return t.CacheModels[_type]
}

func (t *ToyKernel) BelongToPreload(model *Model, field Field) *BelongToPreload {
	// try to find cache data
	if t.belongToPreload[model] != nil && t.belongToPreload[model][field.Name()] != nil {
		return t.belongToPreload[model][field.Name()]
	}
	_type := LoopTypeIndirect(field.StructField().Type)
	if _type.Kind() != reflect.Struct {
		return nil
	}
	if subModel := t.CacheModels[_type]; subModel != nil {
		if relationField := model.GetFieldWithName(GetBelongsIDFieldName(subModel, field)); relationField != nil {
			if v := t.belongToPreload[model]; v == nil {
				t.belongToPreload[model] = map[string]*BelongToPreload{}
			}
			t.belongToPreload[model][field.Name()] = t.BelongToBind(model, subModel, field, relationField)
			return t.belongToPreload[model][field.Name()]
		}
	}

	return nil
}

func (t *ToyKernel) OneToOnePreload(model *Model, field Field) *OneToOnePreload {
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
			if v := t.oneToOnePreload[model]; v == nil {
				t.oneToOnePreload[model] = map[string]*OneToOnePreload{}
			}
			t.oneToOnePreload[model][field.Name()] = t.OneToOneBind(model, subModel, field, relationField)
			return t.oneToOnePreload[model][field.Name()]
		}
	}
	return nil
}

func (t *ToyKernel) OneToManyPreload(model *Model, field Field) *OneToManyPreload {
	// try to find cache data
	if t.oneToManyPreload[model] != nil && t.oneToManyPreload[model][field.Name()] != nil {
		return t.oneToManyPreload[model][field.Name()]
	}

	_type := LoopTypeIndirect(field.StructField().Type)
	if _type.Kind() == reflect.Slice {
		elemType := LoopTypeIndirect(_type.Elem())
		if subModel, ok := t.CacheModels[elemType]; ok {
			if relationField := subModel.GetFieldWithName(GetRelationFieldName(model)); relationField != nil {
				if v := t.oneToManyPreload[model]; v == nil {
					t.oneToManyPreload[model] = map[string]*OneToManyPreload{}
				}
				// save cache
				t.oneToManyPreload[model][field.Name()] = t.OneToManyBind(model, subModel, field, relationField)
				return t.oneToManyPreload[model][field.Name()]
			}
		}
	}
	return nil
}

func (t *ToyKernel) ManyToManyPreload(model *Model, field Field, isRight bool) *ManyToManyPreload {
	return t.manyToManyPreloadWithTag(model, field, isRight, `toyorm:"primary key"`)
}

func (t *ToyKernel) manyToManyPreloadWithTag(model *Model, field Field, isRight bool, tag reflect.StructTag) *ManyToManyPreload {
	// try to find cache data
	if t.manyToManyPreload[model] != nil && t.manyToManyPreload[model][field.Name()][isRight] != nil {
		return t.manyToManyPreload[model][field.Name()][isRight]
	}
	if v := t.manyToManyPreload[model]; v == nil {
		t.manyToManyPreload[model] = map[string]map[bool]*ManyToManyPreload{}
	}
	if t.manyToManyPreload[model][field.Name()] == nil {
		t.manyToManyPreload[model][field.Name()] = map[bool]*ManyToManyPreload{}
	}

	_type := LoopTypeIndirect(field.StructField().Type)
	if _type.Kind() == reflect.Slice {
		elemType := LoopTypeIndirect(_type.Elem())
		if subModel, ok := t.CacheModels[elemType]; ok {
			middleModel := newMiddleModel(model, subModel, tag)
			relationField := GetMiddleField(model, middleModel, isRight)
			subRelationField := GetMiddleField(subModel, middleModel, !isRight)
			t.manyToManyPreload[model][field.Name()][isRight] = t.ManyToManyPreloadBind(model, subModel, middleModel, field, relationField, subRelationField)
			return t.manyToManyPreload[model][field.Name()][isRight]
		}
	}
	return nil
}

func (t *ToyKernel) BelongToBind(model, subModel *Model, containerField, relationField Field) *BelongToPreload {
	if LoopTypeIndirect(relationField.StructField().Type) != subModel.GetOnePrimary().StructField().Type {
		panic("relation key must have same type with sub model primary key")
	}
	if realField := model.NameFields[relationField.Name()]; realField.isForeign {
		realField.foreignModel = subModel
	}
	return &BelongToPreload{
		Model:          model,
		SubModel:       subModel,
		RelationField:  relationField,
		ContainerField: containerField,
	}
}

func (t *ToyKernel) OneToOneBind(model, subModel *Model, containerField, relationField Field) *OneToOnePreload {
	if LoopTypeIndirect(relationField.StructField().Type) != model.GetOnePrimary().StructField().Type {
		panic("relation key must have same type with model primary key")
	}
	if realField := subModel.NameFields[relationField.Name()]; realField.isForeign {
		realField.foreignModel = model
	}

	return &OneToOnePreload{
		Model:          model,
		SubModel:       subModel,
		RelationField:  relationField,
		ContainerField: containerField,
	}
}
func (t *ToyKernel) OneToManyBind(model, subModel *Model, containerField, relationField Field) *OneToManyPreload {
	if LoopTypeIndirect(relationField.StructField().Type) != model.GetOnePrimary().StructField().Type {
		panic("relation key must have same type with model primary key")
	}
	if realField := subModel.NameFields[relationField.Name()]; realField.isForeign {
		realField.foreignModel = model
	}

	return &OneToManyPreload{
		Model:          model,
		SubModel:       subModel,
		RelationField:  relationField,
		ContainerField: containerField,
	}
}

func (t *ToyKernel) ManyToManyPreloadBind(model, subModel, middleModel *Model, containerField, relationField, subRelationField Field) *ManyToManyPreload {
	if LoopTypeIndirect(relationField.StructField().Type) != model.GetOnePrimary().StructField().Type {
		panic("relation key must have same type with model primary key")
	}
	if LoopTypeIndirect(subRelationField.StructField().Type) != subModel.GetOnePrimary().StructField().Type {
		panic("sub relation key must have same type with sub model primary key")
	}
	if realField := middleModel.NameFields[relationField.Name()]; realField.isForeign {
		realField.foreignModel = model
	}
	if realField := middleModel.NameFields[subRelationField.Name()]; realField.isForeign {
		realField.foreignModel = subModel
	}

	t.CacheMiddleModels[middleModel.ReflectType] = middleModel

	return &ManyToManyPreload{
		Model:            model,
		SubModel:         subModel,
		MiddleModel:      middleModel,
		ContainerField:   containerField,
		RelationField:    relationField,
		SubRelationField: subRelationField,
	}
}

func (t *Toy) ModelHandlers(option string, model *Model) HandlersChain {
	handlers := make(HandlersChain, 0, len(t.DefaultHandlerChain[option])+len(t.DefaultModelHandlerChain[model][option]))
	handlers = append(handlers, t.DefaultModelHandlerChain[model][option]...)
	handlers = append(handlers, t.DefaultHandlerChain[option]...)
	return handlers
}

func (t *Toy) Close() error {
	return t.db.Close()
}
