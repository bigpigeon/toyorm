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

type Toy struct {
	db                       *sql.DB
	DefaultHandlerChain      map[string]HandlersChain
	DefaultModelHandlerChain map[reflect.Type]map[string]HandlersChain
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
	case "postgres":
		dialect = PostgreSqlDialect{}
	default:
		dialect = DefaultDialect{}
	}
	return &Toy{
		db: db,
		DefaultHandlerChain: map[string]HandlersChain{
			"CreateTable":              {HandlerCreateTablePreload("CreateTable"), HandlerCreateTable},
			"CreateTableIfNotExist":    {HandlerCreateTablePreload("CreateTableIfNotExist"), HandlerExistTableAbort, HandlerCreateTable},
			"DropTableIfExist":         {HandlerDropTablePreload("DropTableIfExist"), HandlerNotExistTableAbort, HandlerDropTable},
			"DropTable":                {HandlerDropTablePreload("DropTable"), HandlerDropTable},
			"Insert":                   {HandlerPreloadContainerCheck, HandlerPreloadInsertOrSave("Insert"), HandlerInsertTimeGenerate, HandlerInsert},
			"Find":                     {HandlerPreloadContainerCheck, HandlerSoftDeleteCheck, HandlerFind, HandlerPreloadOnJoinFind, HandlerPreloadFind},
			"Update":                   {HandlerSoftDeleteCheck, HandlerUpdateTimeGenerate, HandlerUpdate},
			"Save":                     {HandlerPreloadContainerCheck, HandlerPreloadInsertOrSave("Save"), HandlerSaveTimeGenerate, HandlerSave},
			"HardDelete":               {HandlerPreloadDelete, HandlerHardDelete},
			"SoftDelete":               {HandlerPreloadDelete, HandlerSoftDelete},
			"HardDeleteWithPrimaryKey": {HandlerPreloadDelete, HandlerSearchWithPrimaryKey, HandlerHardDelete},
			"SoftDeleteWithPrimaryKey": {HandlerPreloadDelete, HandlerSearchWithPrimaryKey, HandlerSoftDelete},
		},
		DefaultModelHandlerChain: map[reflect.Type]map[string]HandlersChain{},
		ToyKernel: ToyKernel{
			CacheModels:       map[reflect.Type]*Model{},
			CacheMiddleModels: map[reflect.Type]*Model{},
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
	if t.debug {
		toyBrick = toyBrick.Debug()
	}
	return toyBrick
}

func (t *Toy) MiddleModel(v, sv interface{}) *ToyBrick {
	vType := LoopTypeIndirect(reflect.ValueOf(v).Type())
	svType := LoopTypeIndirect(reflect.ValueOf(sv).Type())
	model, subModel := t.GetModel(vType), t.GetModel(svType)
	middleModel := NewMiddleModel(model, subModel)
	return NewToyBrick(t, middleModel)
}

func (t *Toy) ModelHandlers(option string, model *Model) HandlersChain {
	handlers := make(HandlersChain, 0, len(t.DefaultHandlerChain[option])+len(t.DefaultModelHandlerChain[model.ReflectType][option]))
	handlers = append(handlers, t.DefaultModelHandlerChain[model.ReflectType][option]...)
	handlers = append(handlers, t.DefaultHandlerChain[option]...)
	return handlers
}

func (t *Toy) Close() error {
	if t == nil {
		return nil
	}
	return t.db.Close()
}

func (t *Toy) BelongToPreload(model *Model, field Field) *BelongToPreload {
	_type := LoopTypeIndirect(field.StructField().Type)
	if _type.Kind() != reflect.Struct {
		return nil
	}
	if subModel := t.CacheModels[_type]; subModel != nil {
		if relationField, ok := model.Association[BelongToWith][field.Name()]; ok {
			return t.BelongToBind(model, subModel, field, relationField)
		}
		if relationField := model.GetFieldWithName(GetBelongsIDFieldName(subModel, field)); relationField != nil {
			return t.BelongToBind(model, subModel, field, relationField)
		}
	}

	return nil
}

func (t *Toy) OneToOnePreload(model *Model, field Field) *OneToOnePreload {

	_type := LoopTypeIndirect(field.StructField().Type)
	if _type.Kind() != reflect.Struct {
		return nil
	}
	if subModel := t.CacheModels[_type]; subModel != nil {
		if relationField, ok := subModel.Association[OneToOneWith][field.Name()]; ok {
			return t.OneToOneBind(model, subModel, field, relationField)
		}
		if relationField := subModel.GetFieldWithName(GetRelationFieldName(model)); relationField != nil {
			return t.OneToOneBind(model, subModel, field, relationField)
		}
	}
	return nil
}

func (t *Toy) OneToManyPreload(model *Model, field Field) *OneToManyPreload {
	_type := LoopTypeIndirect(field.StructField().Type)
	if _type.Kind() == reflect.Slice {
		elemType := LoopTypeIndirect(_type.Elem())
		if subModel, ok := t.CacheModels[elemType]; ok {
			if relationField, ok := subModel.Association[OneToManyWith][field.Name()]; ok {
				return t.OneToManyBind(model, subModel, field, relationField)
			}
			if relationField := subModel.GetFieldWithName(GetRelationFieldName(model)); relationField != nil {
				return t.OneToManyBind(model, subModel, field, relationField)
			}
		}
	}
	return nil
}

func (t *Toy) ManyToManyPreload(model *Model, field Field, isRight bool) *ManyToManyPreload {
	return t.manyToManyPreloadWithTag(model, field, isRight, `toyorm:"primary key"`)
}

func (t *Toy) manyToManyPreloadWithTag(model *Model, field Field, isRight bool, tag reflect.StructTag) *ManyToManyPreload {

	_type := LoopTypeIndirect(field.StructField().Type)
	if _type.Kind() == reflect.Slice {
		elemType := LoopTypeIndirect(_type.Elem())

		if subModel, ok := t.CacheModels[elemType]; ok {
			middleModel := newMiddleModel(model, subModel, tag)
			relationField := GetMiddleField(model, middleModel, isRight)
			subRelationField := GetMiddleField(subModel, middleModel, !isRight)
			return t.ManyToManyPreloadBind(model, subModel, middleModel, field, relationField, subRelationField)

		}
	}
	return nil
}

func (t *Toy) Join(model *Model, field Field) *Join {
	_type := LoopTypeIndirect(field.StructField().Type)
	subModel := t.GetModel(_type)
	containerName := field.Name()
	if model.Association[JoinWith][containerName] != nil && subModel.Association[JoinWith][containerName] != nil {
		return &Join{
			model,
			subModel,
			field,
			model.Association[JoinWith][containerName],
			subModel.Association[JoinWith][containerName],
		}
	}
	return nil
}

func (t *Toy) BelongToBind(model, subModel *Model, containerField, relationField Field) *BelongToPreload {
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

func (t *Toy) OneToOneBind(model, subModel *Model, containerField, relationField Field) *OneToOnePreload {
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
func (t *Toy) OneToManyBind(model, subModel *Model, containerField, relationField Field) *OneToManyPreload {
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

func (t *Toy) ManyToManyPreloadBind(model, subModel, middleModel *Model, containerField, relationField, subRelationField Field) *ManyToManyPreload {
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
