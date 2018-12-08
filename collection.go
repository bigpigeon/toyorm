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
	DefaultModelHandlerChain map[reflect.Type]map[string]CollectionHandlersChain
	ToyKernel
}

func OpenCollection(driverName string, dataSourceName ...string) (*ToyCollection, error) {
	t := ToyCollection{
		ToyKernel: ToyKernel{
			CacheModels:       map[reflect.Type]map[CacheMeta]*Model{},
			CacheMiddleModels: map[reflect.Type]map[CacheMeta]*Model{},
			Logger:            os.Stdout,
		},
		DefaultHandlerChain: map[string]CollectionHandlersChain{
			"CreateTable":              {CollectionHandlerSimplePreload("CreateTable"), CollectionHandlerAssignToAllDb, CollectionHandlerCreateTable},
			"CreateTableIfNotExist":    {CollectionHandlerSimplePreload("CreateTableIfNotExist"), CollectionHandlerAssignToAllDb, CollectionHandlerExistTableAbort, CollectionHandlerCreateTable},
			"DropTableIfExist":         {CollectionHandlerDropTablePreload("DropTableIfExist"), CollectionHandlerAssignToAllDb, CollectionHandlerNotExistTableAbort, CollectionHandlerDropTable},
			"DropTable":                {CollectionHandlerDropTablePreload("DropTable"), CollectionHandlerAssignToAllDb, CollectionHandlerDropTable},
			"Insert":                   {CollectionHandlerPreloadContainerCheck, CollectionHandlerPreloadInsertOrSave("Insert"), HandlerCollectionCasVersionPushOne, CollectionHandlerInsertTimeGenerate, CollectionHandlerInsertAssignDbIndex, CollectionHandlerInsert},
			"Save":                     {CollectionHandlerPreloadContainerCheck, CollectionHandlerPreloadInsertOrSave("Save"), HandlerCollectionCasVersionPushOne, CollectionHandlerInsertAssignDbIndex, CollectionHandlerSaveTimeGenerate, CollectionHandlerSave},
			"USave":                    {CollectionHandlerPreloadContainerCheck, CollectionHandlerPreloadInsertOrSave("USave"), HandlerCollectionCasVersionPushOne, CollectionHandlerInsertAssignDbIndex, CollectionHandlerSaveTimeGenerate, CollectionHandlerUSave},
			"Find":                     {CollectionHandlerPreloadContainerCheck, CollectionHandlerSoftDeleteCheck, CollectionHandlerPreloadFind, CollectionHandlerAssignToAllDb, CollectionHandlerFind},
			"FindOne":                  {CollectionHandlerPreloadContainerCheck, CollectionHandlerSoftDeleteCheck, CollectionHandlerPreloadFind, CollectionHandlerFindOneAssignDbIndex, CollectionHandlerFindOne},
			"Update":                   {CollectionHandlerSoftDeleteCheck, CollectionHandlerUpdateTimeGenerate, CollectionHandlerAssignToAllDb, CollectionHandlerUpdate},
			"HardDelete":               {HandlerCollectionPreloadDelete, CollectionHandlerAssignToAllDb, HandlerCollectionHardDelete},
			"SoftDelete":               {HandlerCollectionPreloadDelete, CollectionHandlerAssignToAllDb, HandlerCollectionSoftDelete},
			"HardDeleteWithPrimaryKey": {HandlerCollectionPreloadDelete, HandlerCollectionSearchWithPrimaryKey, CollectionHandlerAssignToAllDb, HandlerCollectionHardDelete},
			"SoftDeleteWithPrimaryKey": {HandlerCollectionPreloadDelete, HandlerCollectionSearchWithPrimaryKey, CollectionHandlerAssignToAllDb, HandlerCollectionSoftDelete},
		},
		DefaultModelHandlerChain: map[reflect.Type]map[string]CollectionHandlersChain{},
	}
	switch driverName {
	case "mysql":
		t.Dialect = MySqlDialect{}
	case "sqlite3":
		t.Dialect = Sqlite3Dialect{}
	case "postgres":
		t.Dialect = PostgreSqlDialect{}
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
	vVal := LoopIndirect(reflect.ValueOf(v))
	// lazy init model
	model = t.GetModel(vVal)
	brick := NewCollectionBrick(t, model)
	if t.debug {
		brick = brick.Debug()
	}
	return brick
}

func (t *ToyCollection) ModelHandlers(option string, model *Model) CollectionHandlersChain {
	handlers := make(CollectionHandlersChain, 0, len(t.DefaultHandlerChain[option])+len(t.DefaultModelHandlerChain[model.ReflectType][option]))
	handlers = append(handlers, t.DefaultModelHandlerChain[model.ReflectType][option]...)
	handlers = append(handlers, t.DefaultHandlerChain[option]...)
	return handlers
}

func (t *ToyCollection) SetModelHandlers(option string, model *Model, handlers CollectionHandlersChain) {
	if t.DefaultModelHandlerChain[model.ReflectType] == nil {
		t.DefaultModelHandlerChain[model.ReflectType] = map[string]CollectionHandlersChain{}
	}
	t.DefaultModelHandlerChain[model.ReflectType][option] = handlers
}

func (t *ToyCollection) MiddleModel(v, sv interface{}) *CollectionBrick {
	vVal := LoopIndirect(reflect.ValueOf(v))
	svVal := LoopIndirect(reflect.ValueOf(sv))
	model, subModel := t.GetModel(vVal), t.GetModel(svVal)
	middleModel := NewMiddleModel(model, subModel)
	return NewCollectionBrick(t, middleModel)
}

func (t *ToyCollection) GetMiddleModel(val reflect.Value) *Model {
	name := ModelName(val)
	if name == "" {
		panic(ErrInvalidModelName{})
	}
	if model, ok := t.CacheMiddleModels[val.Type()][CacheMeta{name}]; ok == false {
		model = newModel(val, name)
		t.CacheMiddleModels[val.Type()][CacheMeta{name}] = model
	}
	return t.CacheMiddleModels[val.Type()][CacheMeta{name}]
}

func (t *ToyCollection) Close() error {
	if t == nil {
		return nil
	}
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
	val := LoopIndirect(field.FieldValue())
	if val.Kind() != reflect.Struct {
		return nil
	}
	if subModel := t.GetModel(val); subModel != nil {
		if relationField := model.GetFieldWithName(GetBelongsIDFieldName(subModel, field)); relationField != nil {
			return t.BelongToBind(model, subModel, field, relationField)
		}
	}

	return nil
}

func (t *ToyCollection) OneToOnePreload(model *Model, field Field) *OneToOnePreload {
	val := LoopIndirect(field.FieldValue())
	if val.Kind() != reflect.Struct {
		return nil
	}
	if subModel := t.GetModel(val); subModel != nil {
		if relationField := subModel.GetFieldWithName(GetRelationFieldName(model)); relationField != nil {
			return t.OneToOneBind(model, subModel, field, relationField)
		}
	}
	return nil
}

func (t *ToyCollection) OneToManyPreload(model *Model, field Field) *OneToManyPreload {
	val := LoopIndirect(field.FieldValue())
	if val.Kind() == reflect.Slice {
		elemVal := LoopGetElemAndPtr(val)
		if subModel := t.GetModel(elemVal); subModel != nil {
			if relationField := subModel.GetFieldWithName(GetRelationFieldName(model)); relationField != nil {
				return t.OneToManyBind(model, subModel, field, relationField)
			}
		}
	}
	return nil
}

func (t *ToyCollection) ManyToManyPreload(model *Model, field Field, isRight bool) *ManyToManyPreload {
	return t.manyToManyPreloadWithTag(model, field, isRight, `toyorm:"primary key"`)
}

func (t *ToyCollection) manyToManyPreloadWithTag(model *Model, field Field, isRight bool, tag reflect.StructTag) *ManyToManyPreload {
	val := LoopIndirect(field.FieldValue())
	if val.Kind() == reflect.Slice {
		elemVal := LoopGetElemAndPtr(val)
		if subModel := t.GetModel(elemVal); subModel != nil {
			middleModel := newMiddleModel(model, subModel, tag)
			relationField := GetMiddleField(model, middleModel, isRight)
			subRelationField := GetMiddleField(subModel, middleModel, !isRight)
			return t.ManyToManyPreloadBind(model, subModel, middleModel, field, relationField, subRelationField)
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

	t.CacheMiddleModels[middleModel.ReflectType][CacheMeta{middleModel.Name}] = middleModel

	return &ManyToManyPreload{
		Model:            model,
		SubModel:         subModel,
		MiddleModel:      middleModel,
		ContainerField:   containerField,
		RelationField:    relationField,
		SubRelationField: subRelationField,
	}
}
