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
			"Insert":                   {HandlerPreloadContainerCheck, HandlerPreloadInsertOrSave("Insert"), HandlerCasVersionPushOne, HandlerInsertTimeGenerate, HandlerInsert},
			"Find":                     {HandlerPreloadContainerCheck, HandlerSoftDeleteCheck, HandlerFind, HandlerPreloadOnJoinFind, HandlerPreloadFind},
			"Update":                   {HandlerSoftDeleteCheck, HandlerUpdateTimeGenerate, HandlerUpdate},
			"Save":                     {HandlerPreloadContainerCheck, HandlerPreloadInsertOrSave("Save"), HandlerCasVersionPushOne, HandlerSaveTimeGenerate, HandlerSave},
			"USave":                    {HandlerPreloadContainerCheck, HandlerPreloadInsertOrSave("USave"), HandlerCasVersionPushOne, HandlerUSaveTimeGenerate, HandlerUSave},
			"HardDelete":               {HandlerPreloadDelete, HandlerHardDelete},
			"SoftDelete":               {HandlerPreloadDelete, HandlerSoftDelete},
			"HardDeleteWithPrimaryKey": {HandlerPreloadDelete, HandlerSearchWithPrimaryKey, HandlerHardDelete},
			"SoftDeleteWithPrimaryKey": {HandlerPreloadDelete, HandlerSearchWithPrimaryKey, HandlerSoftDelete},
		},
		DefaultModelHandlerChain: map[reflect.Type]map[string]HandlersChain{},
		ToyKernel: ToyKernel{
			Dialect: dialect,
			Logger:  os.Stdout,
		},
	}, nil
}

func (t *Toy) Model(v interface{}) *ToyBrick {
	vVal := LoopDivePtr(reflect.ValueOf(v))
	model := t.GetModel(vVal)
	toyBrick := NewToyBrick(t, model)
	if t.debug {
		toyBrick = toyBrick.Debug()
	}
	return toyBrick
}

func (t *Toy) MiddleModel(v, sv interface{}) *ToyBrick {
	vVal := LoopDivePtr(reflect.ValueOf(v))
	svVal := LoopDivePtr(reflect.ValueOf(sv))
	model, subModel := t.GetModel(vVal), t.GetModel(svVal)
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
	val := LoopDivePtr(field.FieldValue())
	if val.Kind() != reflect.Struct {
		return nil
	}
	if subModel := t.GetModel(val); subModel != nil {
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

	val := LoopDivePtr(field.FieldValue())
	if val.Kind() != reflect.Struct {
		return nil
	}
	if subModel := t.GetModel(val); subModel != nil {
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
	val := LoopDivePtr(field.FieldValue())
	if val.Kind() == reflect.Slice {
		elemVal := LoopDiveSliceAndPtr(val)
		if subModel := t.GetModel(elemVal); subModel != nil {
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

	val := LoopDivePtr(field.FieldValue())
	if val.Kind() == reflect.Slice {
		elemVal := LoopDiveSliceAndPtr(val)
		if subModel := t.GetModel(elemVal); subModel != nil {
			middleModel := newMiddleModel(model, subModel, tag)
			relationField := GetMiddleField(model, middleModel, isRight)
			subRelationField := GetMiddleField(subModel, middleModel, !isRight)
			return t.ManyToManyPreloadBind(model, subModel, middleModel, field, relationField, subRelationField)

		}
	}
	return nil
}

func (t *Toy) Join(model *Model, field Field) *Join {
	val := LoopDiveSliceAndPtr(field.FieldValue())
	subModel := t.GetModel(val)
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

	return &ManyToManyPreload{
		Model:            model,
		SubModel:         subModel,
		MiddleModel:      middleModel,
		ContainerField:   containerField,
		RelationField:    relationField,
		SubRelationField: subRelationField,
	}
}

func (t *Toy) DB() *sql.DB {
	return t.db
}
