/*
 * Copyright 2018. bigpigeon. All rights reserved.
 * Use of this source code is governed by a MIT style
 * license that can be found in the LICENSE file.
 */

package toyorm

import (
	"database/sql"
	"os"
	"reflect"
)

type DBValSelector interface {
	Select(int) int
}

type DBPrimarySelector func(n int, key ...interface{}) int

func dbPrimaryKeySelector(n int, keys ...interface{}) int {
	sum := 0
	for _, k := range keys {
		switch val := k.(type) {
		case int:
			sum += val
		case int32:
			sum += int(val)
		case uint:
			sum += int(val)
		case uint32:
			sum += int(val)
		case string:
			for _, c := range val {
				sum += int(c)
			}
		default:
			panic("primary key type not match")
		}
	}
	return sum % n
}

type ToyCollection struct {
	dbs                      []*sql.DB
	DefaultHandlerChain      map[string]CollectionHandlersChain
	DefaultModelHandlerChain map[*Model]map[string]CollectionHandlersChain
	ToyKernel
}

func OpenCollection(driverName string, dataSourceName ...string) (*ToyCollection, error) {
	t := ToyCollection{
		ToyKernel: ToyKernel{
			CacheModels:       map[reflect.Type]*Model{},
			CacheMiddleModels: map[reflect.Type]*Model{},
			belongToPreload:   map[*Model]map[string]*BelongToPreload{},
			oneToOnePreload:   map[*Model]map[string]*OneToOnePreload{},
			oneToManyPreload:  map[*Model]map[string]*OneToManyPreload{},
			// because have isRight feature need two point to save
			manyToManyPreload: map[*Model]map[string]map[bool]*ManyToManyPreload{},
			Logger:            os.Stdout,
		},
		DefaultHandlerChain: map[string]CollectionHandlersChain{
			"CreateTable":              {CollectionHandlerSimplePreload("CreateTable"), CollectionHandlerAssignToAllDb, CollectionHandlerCreateTable},
			"CreateTableIfNotExist":    {CollectionHandlerSimplePreload("CreateTableIfNotExist"), CollectionHandlerAssignToAllDb, CollectionHandlerExistTableAbort, CollectionHandlerCreateTable},
			"DropTableIfExist":         {CollectionHandlerDropTablePreload("DropTableIfExist"), CollectionHandlerAssignToAllDb, CollectionHandlerNotExistTableAbort, CollectionHandlerDropTable},
			"DropTable":                {CollectionHandlerDropTablePreload("DropTable"), CollectionHandlerAssignToAllDb, CollectionHandlerDropTable},
			"Insert":                   {CollectionHandlerPreloadContainerCheck, CollectionHandlerPreloadInsertOrSave("Insert"), CollectionHandlerInsertTimeGenerate, CollectionHandlerInsertAssignDbIndex, CollectionHandlerInsert},
			"Save":                     {CollectionHandlerPreloadContainerCheck, CollectionHandlerPreloadInsertOrSave("Save"), CollectionHandlerInsertAssignDbIndex, CollectionHandlerSaveTimeGenerate, CollectionHandlerSave},
			"Find":                     {CollectionHandlerPreloadContainerCheck, CollectionHandlerSoftDeleteCheck, CollectionHandlerPreloadFind, CollectionHandlerAssignToAllDb, CollectionHandlerFind},
			"FindOne":                  {CollectionHandlerPreloadContainerCheck, CollectionHandlerSoftDeleteCheck, CollectionHandlerPreloadFind, CollectionHandlerFindOneAssignDbIndex, CollectionHandlerFindOne},
			"Update":                   {CollectionHandlerSoftDeleteCheck, CollectionHandlerUpdateTimeGenerate, CollectionHandlerAssignToAllDb, CollectionHandlerUpdate},
			"HardDelete":               {HandlerCollectionPreloadDelete, CollectionHandlerAssignToAllDb, HandlerCollectionHardDelete},
			"SoftDelete":               {HandlerCollectionPreloadDelete, CollectionHandlerAssignToAllDb, HandlerCollectionSoftDelete},
			"HardDeleteWithPrimaryKey": {HandlerCollectionPreloadDelete, HandlerCollectionSearchWithPrimaryKey, CollectionHandlerAssignToAllDb, HandlerCollectionHardDelete},
			"SoftDeleteWithPrimaryKey": {HandlerCollectionPreloadDelete, HandlerCollectionSearchWithPrimaryKey, CollectionHandlerAssignToAllDb, HandlerCollectionSoftDelete},
		},
		DefaultModelHandlerChain: map[*Model]map[string]CollectionHandlersChain{},
	}
	switch driverName {
	case "mysql":
		t.Dialect = MySqlDialect{}
	case "sqlite3":
		t.Dialect = Sqlite3Dialect{}
	default:
		panic(ErrNotMatchDialect)
	}
	for _, source := range dataSourceName {
		db, err := sql.Open(driverName, source)
		if err != nil {
			t.Close()
			return nil, err
		}
		t.dbs = append(t.dbs, db)
	}
	return &t, nil
}

func (t *ToyCollection) Model(v interface{}) *CollectionBrick {
	var model *Model
	vType := LoopTypeIndirect(reflect.ValueOf(v).Type())
	// lazy init model
	model = t.GetModel(vType)
	brick := NewCollectionBrick(t, model)
	return brick
}

func (t *ToyCollection) ModelHandlers(option string, model *Model) CollectionHandlersChain {
	handlers := make(CollectionHandlersChain, 0, len(t.DefaultHandlerChain[option])+len(t.DefaultModelHandlerChain[model][option]))
	handlers = append(handlers, t.DefaultModelHandlerChain[model][option]...)
	handlers = append(handlers, t.DefaultHandlerChain[option]...)
	return handlers
}

func (t *ToyCollection) SetModelHandlers(option string, model *Model, handlers CollectionHandlersChain) {
	if t.DefaultModelHandlerChain[model] == nil {
		t.DefaultModelHandlerChain[model] = map[string]CollectionHandlersChain{}
	}
	t.DefaultModelHandlerChain[model][option] = handlers
}

func (t *ToyCollection) MiddleModel(v, sv interface{}) *CollectionBrick {
	vType := LoopTypeIndirect(reflect.ValueOf(v).Type())
	svType := LoopTypeIndirect(reflect.ValueOf(sv).Type())
	model, subModel := t.GetModel(vType), t.GetModel(svType)
	middleModel := NewMiddleModel(model, subModel)
	return NewCollectionBrick(t, middleModel)
}

func (t *ToyCollection) GetMiddleModel(_type reflect.Type) *Model {
	if model, ok := t.CacheModels[_type]; ok == false {
		model = NewModel(_type)
		t.CacheModels[_type] = model
	}
	return t.CacheModels[_type]
}

func (t *ToyCollection) Close() error {
	errs := ErrCollectionQueryRow{}
	for i, db := range t.dbs {
		err := db.Close()
		if err != nil {
			errs[i] = err
		}
	}
	return errs
}

func (t *ToyCollection) BelongToPreload(model *Model, field Field) *BelongToPreload {
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

func (t *ToyCollection) OneToOnePreload(model *Model, field Field) *OneToOnePreload {
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

func (t *ToyCollection) OneToManyPreload(model *Model, field Field) *OneToManyPreload {
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

func (t *ToyCollection) ManyToManyPreload(model *Model, field Field, isRight bool) *ManyToManyPreload {
	return t.manyToManyPreloadWithTag(model, field, isRight, `toyorm:"primary key"`)
}

func (t *ToyCollection) manyToManyPreloadWithTag(model *Model, field Field, isRight bool, tag reflect.StructTag) *ManyToManyPreload {
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

func (t *ToyCollection) BelongToBind(model, subModel *Model, containerField, relationField Field) *BelongToPreload {
	if LoopTypeIndirect(relationField.StructField().Type) != subModel.GetOnePrimary().StructField().Type {
		panic("relation key must have same type with sub model primary key")
	}

	return &BelongToPreload{
		Model:          model,
		SubModel:       subModel,
		RelationField:  relationField,
		ContainerField: containerField,
	}
}

func (t *ToyCollection) OneToOneBind(model, subModel *Model, containerField, relationField Field) *OneToOnePreload {
	if LoopTypeIndirect(relationField.StructField().Type) != model.GetOnePrimary().StructField().Type {
		panic("relation key must have same type with model primary key")
	}

	return &OneToOnePreload{
		Model:          model,
		SubModel:       subModel,
		RelationField:  relationField,
		ContainerField: containerField,
	}
}
func (t *ToyCollection) OneToManyBind(model, subModel *Model, containerField, relationField Field) *OneToManyPreload {
	if LoopTypeIndirect(relationField.StructField().Type) != model.GetOnePrimary().StructField().Type {
		panic("relation key must have same type with model primary key")
	}

	return &OneToManyPreload{
		Model:          model,
		SubModel:       subModel,
		RelationField:  relationField,
		ContainerField: containerField,
	}
}

func (t *ToyCollection) ManyToManyPreloadBind(model, subModel, middleModel *Model, containerField, relationField, subRelationField Field) *ManyToManyPreload {
	if LoopTypeIndirect(relationField.StructField().Type) != model.GetOnePrimary().StructField().Type {
		panic("relation key must have same type with model primary key")
	}
	if LoopTypeIndirect(subRelationField.StructField().Type) != subModel.GetOnePrimary().StructField().Type {
		panic("sub relation key must have same type with sub model primary key")
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
