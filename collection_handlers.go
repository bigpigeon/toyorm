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

func CollectionHandlerPreloadInsertOrSave(option string) func(*CollectionContext) error {
	return func(ctx *CollectionContext) error {
		for fieldName, preload := range ctx.Brick.BelongToPreload {
			mainField, subField := preload.RelationField, preload.SubModel.GetOnePrimary()
			preloadBrick := ctx.Brick.MapPreloadBrick[fieldName]
			subRecords := MakeRecordsWithElem(preload.SubModel, ctx.Result.Records.GetFieldAddressType(fieldName))

			// map[i]=>j [i]record.SubData -> [j]subRecord
			bindMap := map[int]int{}
			for i, record := range ctx.Result.Records.GetRecords() {
				if ctx.Brick.ignoreModeSelector[ModePreload].Ignore(record.Field(fieldName)) == false {
					bindMap[i] = subRecords.Len()
					subRecords.Add(record.FieldAddress(fieldName))
				}
			}
			subCtx := preloadBrick.GetContext(option, subRecords)
			ctx.Result.Preload[fieldName] = subCtx.Result
			if err := subCtx.Next(); err != nil {
				return err
			}
			// set model relation field
			ctx.Result.SimpleRelation[fieldName] = map[int]int{}
			for i, record := range ctx.Result.Records.GetRecords() {
				if j, ok := bindMap[i]; ok {
					subRecord := subRecords.GetRecord(j)
					record.SetField(mainField.Name(), subRecord.Field(subField.Name()))
					ctx.Result.SimpleRelation[fieldName][j] = i
				}
			}
		}

		if err := ctx.Next(); err != nil {
			return err
		}
		for fieldName, preload := range ctx.Brick.OneToOnePreload {
			preloadBrick := ctx.Brick.MapPreloadBrick[fieldName]
			mainPos, subPos := preload.Model.GetOnePrimary(), preload.RelationField
			subRecords := MakeRecordsWithElem(preload.SubModel, ctx.Result.Records.GetFieldAddressType(fieldName))
			// set sub model relation field
			ctx.Result.SimpleRelation[fieldName] = map[int]int{}
			for i, record := range ctx.Result.Records.GetRecords() {
				if ctx.Brick.ignoreModeSelector[ModePreload].Ignore(record.Field(fieldName)) == false {
					// it means relation field, result[j].LastInsertId() is id value
					subRecord := subRecords.Add(record.FieldAddress(fieldName))
					ctx.Result.SimpleRelation[fieldName][subRecords.Len()-1] = i
					if primary := record.Field(mainPos.Name()); primary.IsValid() {
						subRecord.SetField(subPos.Name(), primary)
					} else {
						panic("relation field not set")
					}
				}
			}
			subCtx := preloadBrick.GetContext(option, subRecords)
			ctx.Result.Preload[fieldName] = subCtx.Result
			if err := subCtx.Next(); err != nil {
				return err
			}

		}

		// one to many
		for fieldName, preload := range ctx.Brick.OneToManyPreload {
			preloadBrick := ctx.Brick.MapPreloadBrick[fieldName]
			mainField, subField := preload.Model.GetOnePrimary(), preload.RelationField
			elemAddressType := reflect.PtrTo(LoopTypeIndirect(ctx.Result.Records.GetFieldType(fieldName)).Elem())
			subRecords := MakeRecordsWithElem(preload.SubModel, elemAddressType)
			// set sub model relation field
			ctx.Result.MultipleRelation[fieldName] = map[int]Pair{}
			for i, record := range ctx.Result.Records.GetRecords() {
				if primary := record.Field(mainField.Name()); primary.IsValid() {
					rField := LoopIndirect(record.Field(fieldName))
					for subi := 0; subi < rField.Len(); subi++ {
						subRecord := subRecords.Add(rField.Index(subi).Addr())
						ctx.Result.MultipleRelation[fieldName][subRecords.Len()-1] = Pair{i, subi}
						subRecord.SetField(subField.Name(), primary)
					}
				} else {
					return errors.New("some records have not primary")
				}
			}
			subCtx := preloadBrick.GetContext(option, subRecords)
			ctx.Result.Preload[fieldName] = subCtx.Result
			if err := subCtx.Next(); err != nil {
				return err
			}
		}
		// many to many
		for fieldName, preload := range ctx.Brick.ManyToManyPreload {
			subBrick := ctx.Brick.MapPreloadBrick[fieldName]
			middleBrick := NewCollectionBrick(ctx.Brick.Toy, preload.MiddleModel).CopyStatus(ctx.Brick)

			mainField, subField := preload.Model.GetOnePrimary(), preload.SubModel.GetOnePrimary()
			elemAddressType := reflect.PtrTo(LoopTypeIndirect(ctx.Result.Records.GetFieldType(fieldName)).Elem())
			subRecords := MakeRecordsWithElem(preload.SubModel, elemAddressType)

			ctx.Result.MultipleRelation[fieldName] = map[int]Pair{}
			for i, record := range ctx.Result.Records.GetRecords() {
				rField := LoopIndirect(record.Field(fieldName))
				for subi := 0; subi < rField.Len(); subi++ {
					subRecords.Add(rField.Index(subi).Addr())
					ctx.Result.MultipleRelation[fieldName][subRecords.Len()-1] = Pair{i, subi}
				}
			}
			subCtx := subBrick.GetContext(option, subRecords)
			ctx.Result.Preload[fieldName] = subCtx.Result
			if err := subCtx.Next(); err != nil {
				return err
			}

			middleRecords := MakeRecordsWithElem(middleBrick.Model, middleBrick.Model.ReflectType)
			// use to calculate what sub records belong for
			offset := 0
			for _, record := range ctx.Result.Records.GetRecords() {
				primary := record.Field(mainField.Name())
				primary.IsValid()
				if primary.IsValid() == false {
					return errors.New("some records have not primary")
				}
				rField := LoopIndirect(record.Field(fieldName))
				for subi := 0; subi < rField.Len(); subi++ {
					subRecord := subRecords.GetRecord(subi + offset)
					subPrimary := subRecord.Field(subField.Name())
					if subPrimary.IsValid() == false {
						return errors.New("some records have not primary")
					}
					middleRecord := NewRecord(middleBrick.Model, reflect.New(middleBrick.Model.ReflectType).Elem())
					middleRecord.SetField(preload.RelationField.Name(), primary)
					middleRecord.SetField(preload.SubRelationField.Name(), subPrimary)
					middleRecords.Add(middleRecord.Source())
				}
				offset += rField.Len()
			}
			middleCtx := middleBrick.GetContext(option, middleRecords)
			ctx.Result.MiddleModelPreload[fieldName] = middleCtx.Result
			if err := middleCtx.Next(); err != nil {
				return err
			}
		}
		return nil
	}
}

func CollectionHandlerInsertTimeGenerate(ctx *CollectionContext) error {
	records := ctx.Result.Records
	createField := ctx.Brick.Model.GetFieldWithName("CreatedAt")
	updateField := ctx.Brick.Model.GetFieldWithName("UpdatedAt")
	if createField != nil || updateField != nil {
		current := time.Now()
		if createField != nil {
			for _, record := range records.GetRecords() {
				record.SetField(createField.Name(), reflect.ValueOf(current))
			}
		}
		if updateField != nil {
			for _, record := range records.GetRecords() {
				record.SetField(updateField.Name(), reflect.ValueOf(current))
			}
		}
	}
	return nil
}

func CollectionHandlerInsertAssignDbIndex(ctx *CollectionContext) error {
	if ctx.Brick.dbIndex != -1 {
		return nil
	}
	primaryKeyField := ctx.Brick.Model.GetPrimary()
	var getDBIndex func(r ModelRecord) (int, error)
	notPtrElemType := LoopTypeIndirect(ctx.Result.Records.ElemType())
	if _, ok := reflect.Zero(reflect.PtrTo(notPtrElemType)).Interface().(DBValSelector); ok {
		getDBIndex = func(r ModelRecord) (int, error) {
			iface := r.Source().Addr().Interface().(DBValSelector)
			return iface.Select(len(ctx.Brick.Toy.dbs)), nil
		}
	} else if selector := ctx.Brick.selector; selector != nil {
		getDBIndex = func(r ModelRecord) (int, error) {
			var ifaces []interface{}
			for _, field := range primaryKeyField {
				if fieldVal := r.Field(field.Name()); !fieldVal.IsValid() || IsZero(fieldVal) {
					return 0, ErrZeroPrimaryKey{ctx.Brick.Model}
				} else {
					ifaces = append(ifaces, fieldVal.Interface())
				}
			}
			return selector(len(ctx.Brick.Toy.dbs), ifaces...), nil
		}
	} else {
		return ErrCollectionDBSelectorNotFound{}
	}
	dbRecordsMap := map[int]ModelRecords{}
	// dbRecordsMap dbIndexMap[i]  means the position in Records
	dbIndexMap := make([][]int, len(ctx.Brick.Toy.dbs))
	for i, record := range ctx.Result.Records.GetRecords() {
		dbIndex, err := getDBIndex(record)
		if err != nil {
			return err
		}
		if dbRecordsMap[dbIndex] == nil {
			dbRecordsMap[dbIndex] = MakeRecordsWithElem(ctx.Brick.Model, record.Source().Addr().Type())
		}
		dbRecordsMap[dbIndex].Add(record.Source().Addr())
		dbIndexMap[dbIndex] = append(dbIndexMap[dbIndex], i)
	}
	for i, records := range dbRecordsMap {
		dbCtx := NewCollectionContext(ctx.handlers[ctx.index+1:], ctx.Brick.DBIndex(i), records)

		err := dbCtx.Next()
		if err != nil {
			return err
		}
		for _, action := range dbCtx.Result.ActionFlow {
			affectData := action.AffectData()
			for j := 0; j < len(affectData); j++ {
				affectData[j] = dbIndexMap[i][affectData[j]]
			}
			action.SetAffectData(affectData)
			ctx.Result.AddRecord(action)
		}
		//
	}
	ctx.Abort()
	return nil
}

func CollectionHandlerInsert(ctx *CollectionContext) error {
	// current insert
	if ctx.Brick.dbIndex == -1 {
		return ErrDbIndexNotSet{}
	}
	//setInsertId := len(ctx.Brick.Model.GetPrimary()) == 1 && ctx.Brick.Model.GetOnePrimary().AutoIncrement() == true
	for i, record := range ctx.Result.Records.GetRecords() {
		action := CollectionExecAction{affectData: []int{i}, dbIndex: ctx.Brick.dbIndex}
		var err error
		if ctx.Brick.template == nil {
			action.Exec = ctx.Brick.InsertExec(record)
		} else {
			tempMap := DefaultCollectionTemplateExec(ctx.Brick)
			values := ctx.Brick.getFieldValuePairWithRecord(ModeInsert, record).ToValueList()
			tempMap["Columns"] = getColumnExec(columnsValueToColumn(values))
			tempMap["Values"] = getValuesExec(values)
			action.Exec, err = ctx.Brick.Toy.Dialect.TemplateExec(*ctx.Brick.template, tempMap)
			if err != nil {
				return err
			}
		}
		action.Result, action.Error = ctx.Brick.Toy.Dialect.InsertExecutor(
			ctx.Brick.Toy.dbs[action.dbIndex],
			action.Exec,
			ctx.Brick.debugPrint(action.dbIndex),
		)

		if action.Error == nil {
			// set primary field value if model has one primary key
			if len(ctx.Brick.Model.GetPrimary()) == 1 {
				primaryKey := ctx.Brick.Model.GetOnePrimary()
				primaryKeyName := primaryKey.Name()
				if IntKind(primaryKey.StructField().Type.Kind()) {
					// just set not zero primary key
					if fieldValue := record.Field(primaryKeyName); !fieldValue.IsValid() || IsZero(fieldValue) {
						if lastId, err := action.Result.LastInsertId(); err == nil {
							ctx.Result.Records.GetRecord(i).SetField(primaryKeyName, reflect.ValueOf(lastId))
						} else {
							return errors.New(fmt.Sprintf("get (%s) auto increment  failure reason(%s)", ctx.Brick.Model.Name, err))
						}
					}
				}
			}
		}
		ctx.Result.AddRecord(action)
	}
	return nil
}

// preload schedule belongTo -> Next() -> oneToOne -> oneToMany -> manyToMany(sub -> middle)
func CollectionHandlerSimplePreload(option string) func(ctx *CollectionContext) error {
	return func(ctx *CollectionContext) (err error) {
		for fieldName := range ctx.Brick.BelongToPreload {
			brick := ctx.Brick.MapPreloadBrick[fieldName]
			subCtx := brick.GetContext(option, MakeRecordsWithElem(brick.Model, brick.Model.ReflectType))
			ctx.Result.Preload[fieldName] = subCtx.Result
			if err := subCtx.Next(); err != nil {
				return err
			}
		}
		err = ctx.Next()
		if err != nil {
			return err
		}
		for fieldName := range ctx.Brick.OneToOnePreload {
			brick := ctx.Brick.MapPreloadBrick[fieldName]
			subCtx := brick.GetContext(option, MakeRecordsWithElem(brick.Model, brick.Model.ReflectType))
			ctx.Result.Preload[fieldName] = subCtx.Result
			if err := subCtx.Next(); err != nil {
				return err
			}
		}

		for fieldName := range ctx.Brick.OneToManyPreload {
			brick := ctx.Brick.MapPreloadBrick[fieldName]
			subCtx := brick.GetContext(option, MakeRecordsWithElem(brick.Model, brick.Model.ReflectType))
			ctx.Result.Preload[fieldName] = subCtx.Result
			if err := subCtx.Next(); err != nil {
				return err
			}
		}
		for fieldName, preload := range ctx.Brick.ManyToManyPreload {
			{
				brick := ctx.Brick.MapPreloadBrick[fieldName]
				subCtx := brick.GetContext(option, MakeRecordsWithElem(brick.Model, brick.Model.ReflectType))
				ctx.Result.Preload[fieldName] = subCtx.Result
				if err := subCtx.Next(); err != nil {
					return err
				}
			}
			// process middle model
			{
				middleModel := preload.MiddleModel
				brick := NewCollectionBrick(ctx.Brick.Toy, middleModel).CopyStatus(ctx.Brick)
				middleCtx := brick.GetContext(option, MakeRecordsWithElem(brick.Model, brick.Model.ReflectType))
				ctx.Result.MiddleModelPreload[fieldName] = middleCtx.Result
				if err := middleCtx.Next(); err != nil {
					return err
				}
			}
		}
		return nil
	}
}

// preload schedule oneToOne -> oneToMany -> current model -> manyToMany(sub -> middle) -> Next() -> belongTo
func CollectionHandlerDropTablePreload(option string) func(ctx *CollectionContext) error {
	return func(ctx *CollectionContext) (err error) {
		for fieldName := range ctx.Brick.OneToOnePreload {
			brick := ctx.Brick.MapPreloadBrick[fieldName]
			subCtx := brick.GetContext(option, MakeRecordsWithElem(brick.Model, brick.Model.ReflectType))
			ctx.Result.Preload[fieldName] = subCtx.Result
			if err := subCtx.Next(); err != nil {
				return err
			}

		}
		for fieldName := range ctx.Brick.OneToManyPreload {
			brick := ctx.Brick.MapPreloadBrick[fieldName]
			subCtx := brick.GetContext(option, MakeRecordsWithElem(brick.Model, brick.Model.ReflectType))
			ctx.Result.Preload[fieldName] = subCtx.Result
			if err := subCtx.Next(); err != nil {
				return err
			}
		}
		for fieldName, preload := range ctx.Brick.ManyToManyPreload {
			// process middle model
			{
				middleModel := preload.MiddleModel
				brick := NewCollectionBrick(ctx.Brick.Toy, middleModel).CopyStatus(ctx.Brick)
				middleCtx := brick.GetContext(option, MakeRecordsWithElem(brick.Model, brick.Model.ReflectType))
				ctx.Result.MiddleModelPreload[fieldName] = middleCtx.Result
				if err := middleCtx.Next(); err != nil {
					return err
				}
			}
			// process sub model
			{
				brick := ctx.Brick.MapPreloadBrick[fieldName]
				subCtx := brick.GetContext(option, MakeRecordsWithElem(brick.Model, brick.Model.ReflectType))
				ctx.Result.Preload[fieldName] = subCtx.Result
				if err := subCtx.Next(); err != nil {
					return err
				}
			}
		}
		err = ctx.Next()
		if err != nil {
			return err
		}
		for fieldName := range ctx.Brick.BelongToPreload {
			brick := ctx.Brick.MapPreloadBrick[fieldName]
			subCtx := brick.GetContext(option, MakeRecordsWithElem(brick.Model, brick.Model.ReflectType))
			ctx.Result.Preload[fieldName] = subCtx.Result
			if err := subCtx.Next(); err != nil {
				return err
			}

		}

		return nil
	}
}

// assign after handlers to all db
func CollectionHandlerAssignToAllDb(ctx *CollectionContext) error {
	if ctx.Brick.dbIndex != -1 {
		return nil
	}
	for i := range ctx.Brick.Toy.dbs {
		dbCtx := NewCollectionContext(ctx.handlers[ctx.index+1:], ctx.Brick.DBIndex(i), ctx.Result.Records)
		err := dbCtx.Next()
		if err != nil {
			return err
		}
		ctx.Result.ActionFlow = append(ctx.Result.ActionFlow, dbCtx.Result.ActionFlow...)
	}
	ctx.Abort()
	return nil
}

func CollectionHandlerCreateTable(ctx *CollectionContext) error {
	if ctx.Brick.dbIndex == -1 {
		return ErrDbIndexNotSet{}
	}
	execs := ctx.Brick.Toy.Dialect.CreateTable(ctx.Brick.Model, map[string]ForeignKey{})
	for _, exec := range execs {
		action := CollectionExecAction{Exec: exec, dbIndex: ctx.Brick.dbIndex}
		action.Result, action.Error = ctx.Brick.Exec(exec, action.dbIndex)
		ctx.Result.AddRecord(action)
	}
	return nil
}

func CollectionHandlerExistTableAbort(ctx *CollectionContext) error {
	if ctx.Brick.dbIndex == -1 {
		return ErrDbIndexNotSet{}
	}
	action := CollectionQueryAction{dbIndex: ctx.Brick.dbIndex}
	action.Exec = ctx.Brick.Toy.Dialect.HasTable(ctx.Brick.Model)
	var hasTable bool
	err := ctx.Brick.QueryRow(action.Exec, ctx.Brick.dbIndex).Scan(&hasTable)
	if err != nil {
		action.Error = append(action.Error, err)
	}
	ctx.Result.AddRecord(action)
	if err != nil || hasTable == true {
		ctx.Abort()
	}

	return nil
}

func CollectionHandlerDropTable(ctx *CollectionContext) (err error) {
	if ctx.Brick.dbIndex == -1 {
		return ErrDbIndexNotSet{}
	}
	exec := ctx.Brick.Toy.Dialect.DropTable(ctx.Brick.Model)
	action := CollectionExecAction{Exec: exec, dbIndex: ctx.Brick.dbIndex}
	action.Result, action.Error = ctx.Brick.Exec(exec, action.dbIndex)
	ctx.Result.AddRecord(action)
	return nil
}

func CollectionHandlerNotExistTableAbort(ctx *CollectionContext) error {
	if ctx.Brick.dbIndex == -1 {
		return ErrDbIndexNotSet{}
	}
	action := CollectionQueryAction{}
	action.Exec = ctx.Brick.Toy.Dialect.HasTable(ctx.Brick.Model)
	var hasTable bool
	err := ctx.Brick.QueryRow(action.Exec, ctx.Brick.dbIndex).Scan(&hasTable)
	if err != nil {
		action.Error = append(action.Error, err)
	}
	ctx.Result.AddRecord(action)
	if err != nil || hasTable == false {
		ctx.Abort()
	}
	return nil
}

func CollectionHandlerPreloadContainerCheck(ctx *CollectionContext) error {
	for fieldName, preload := range ctx.Brick.BelongToPreload {
		if fieldType := ctx.Result.Records.GetFieldType(fieldName); fieldType == nil {
			return errors.New(fmt.Sprintf("struct missing %s field", fieldName))
		} else {
			subRecords := MakeRecordsWithElem(preload.SubModel, fieldType)
			subPrimaryFieldName := preload.SubModel.GetOnePrimary().Name()
			if relationFieldType := subRecords.GetFieldType(subPrimaryFieldName); relationFieldType == nil {
				return errors.New(fmt.Sprintf("struct of the %s field missing %s field", fieldName, subPrimaryFieldName))
			}
		}
		if fieldType := ctx.Result.Records.GetFieldType(preload.RelationField.Name()); fieldType == nil {
			return errors.New(fmt.Sprintf("struct missing %s field", preload.RelationField.Name()))
		}
	}
	var needPrimaryKey bool
	for fieldName, preload := range ctx.Brick.OneToOnePreload {
		needPrimaryKey = true
		if fieldType := ctx.Result.Records.GetFieldType(fieldName); fieldType == nil {
			return errors.New(fmt.Sprintf("struct missing %s field", fieldName))
		} else {
			subRecords := MakeRecordsWithElem(preload.SubModel, fieldType)
			if relationFieldType := subRecords.GetFieldType(preload.RelationField.Name()); relationFieldType == nil {
				return errors.New(fmt.Sprintf("struct of the %s field missing %s field", fieldName, preload.RelationField.Name()))
			}
		}
	}
	for fieldName, preload := range ctx.Brick.OneToManyPreload {
		needPrimaryKey = true
		if fieldType := ctx.Result.Records.GetFieldType(fieldName); fieldType == nil {
			return errors.New(fmt.Sprintf("struct missing %s field", fieldName))
		} else {
			subRecords := MakeRecordsWithElem(preload.SubModel, fieldType)
			if relationFieldType := subRecords.GetFieldType(preload.RelationField.Name()); relationFieldType == nil {
				return errors.New(fmt.Sprintf("struct of the %s field missing %s field", fieldName, preload.RelationField.Name()))
			}
		}
	}
	for fieldName, preload := range ctx.Brick.ManyToManyPreload {
		needPrimaryKey = true
		if fieldType := ctx.Result.Records.GetFieldType(fieldName); fieldType == nil {
			return errors.New(fmt.Sprintf("struct missing %s field", fieldName))
		} else {
			subRecords := MakeRecordsWithElem(preload.SubModel, fieldType)
			subPrimaryFieldName := preload.SubModel.GetOnePrimary().Name()
			if relationFieldType := subRecords.GetFieldType(subPrimaryFieldName); relationFieldType == nil {
				return errors.New(fmt.Sprintf("struct of the %s field missing %s field", fieldName, subPrimaryFieldName))
			}
		}
	}
	if needPrimaryKey {
		primaryName := ctx.Brick.Model.GetOnePrimary().Name()
		if primaryType := ctx.Result.Records.GetFieldType(primaryName); primaryType == nil {
			return errors.New(fmt.Sprintf("struct missing %s field", primaryName))
		}
	}
	return nil
}

func CollectionHandlerSaveTimeGenerate(ctx *CollectionContext) error {
	now := reflect.ValueOf(time.Now())
	if createAtField := ctx.Brick.Model.GetFieldWithName("CreatedAt"); createAtField != nil {
		for _, record := range ctx.Result.Records.GetRecords() {
			if fieldValue := record.Field(createAtField.Name()); fieldValue.IsValid() == false || IsZero(fieldValue) {
				record.SetField(createAtField.Name(), now)
			}
		}
	}
	if updateField := ctx.Brick.Model.GetFieldWithName("UpdatedAt"); updateField != nil {
		for _, record := range ctx.Result.Records.GetRecords() {
			record.SetField(updateField.Name(), now)
		}
	}
	return nil
}

func CollectionHandlerSave(ctx *CollectionContext) error {
	if ctx.Brick.dbIndex == -1 {
		return ErrDbIndexNotSet{}
	}
	//setInsertId := len(ctx.Brick.Model.GetPrimary()) == 1 && ctx.Brick.Model.GetOnePrimary().AutoIncrement() == true
	for i, record := range ctx.Result.Records.GetRecords() {
		var action CollectionExecAction
		var err error
		action = CollectionExecAction{affectData: []int{i}, dbIndex: ctx.Brick.dbIndex}
		var useInsert bool
		// if have any primary key, use insert ,other use update
		for _, key := range ctx.Brick.Model.PrimaryFields {
			if field := record.Field(key.Name()); field.IsValid() == false || IsZero(field) {
				useInsert = true
			}
		}

		if ctx.Brick.template == nil {
			if useInsert {
				action.Exec = ctx.Brick.InsertExec(record)
				action.Result, action.Error = ctx.Brick.Toy.Dialect.InsertExecutor(
					ctx.Brick.Toy.dbs[action.dbIndex],
					action.Exec,
					ctx.Brick.debugPrint(action.dbIndex),
				)

				if action.Error == nil {
					// set primary field value if model has one primary key
					if len(ctx.Brick.Model.GetPrimary()) == 1 {
						primaryKey := ctx.Brick.Model.GetOnePrimary()
						primaryKeyName := primaryKey.Name()
						if IntKind(primaryKey.StructField().Type.Kind()) {
							// just set not zero primary key
							if fieldValue := record.Field(primaryKeyName); !fieldValue.IsValid() || IsZero(fieldValue) {
								if lastId, err := action.Result.LastInsertId(); err == nil {
									ctx.Result.Records.GetRecord(i).SetField(primaryKeyName, reflect.ValueOf(lastId))
								} else {
									return errors.New(fmt.Sprintf("get (%s) auto increment  failure reason(%s)", ctx.Brick.Model.Name, err))
								}
							}
						}
					}

				}
			} else {
				action.Exec = ctx.Brick.SaveExec(record)
				action.Result, action.Error = ctx.Brick.Toy.Dialect.SaveExecutor(
					ctx.Brick.Toy.dbs[action.dbIndex],
					action.Exec,
					ctx.Brick.debugPrint(action.dbIndex),
				)
			}
		} else {
			tempMap := DefaultCollectionTemplateExec(ctx.Brick)
			values := ctx.Brick.getFieldValuePairWithRecord(ModeSave, record).ToValueList()
			tempMap["Columns"] = getColumnExec(columnsValueToColumn(values))
			tempMap["Values"] = getValuesExec(values)
			tempMap["UpdateValues"] = getUpdateValuesExec(values)
			action.Exec, err = ctx.Brick.Toy.Dialect.TemplateExec(*ctx.Brick.template, tempMap)
			if err != nil {
				return err
			}
		}

		ctx.Result.AddRecord(action)
	}
	return nil
}

func CollectionHandlerUSave(ctx *CollectionContext) error {
	if ctx.Brick.dbIndex == -1 {
		return ErrDbIndexNotSet{}
	}
	notIgnoreBrick := ctx.Brick.IgnoreMode(ModeDefault, IgnoreNo)
	for i, record := range ctx.Result.Records.GetRecords() {
		var action CollectionExecAction
		var err error
		action = CollectionExecAction{affectData: []int{i}}
		var useInsert bool
		ConditionKeyVal := map[string]interface{}{}
		// if have any zero primary key,, return error ,otherwise use update
		for _, field := range ctx.Brick.Model.PrimaryFields {
			if fieldVal := record.Field(field.Name()); fieldVal.IsValid() == false || IsZero(fieldVal) {
				useInsert = true
			} else {
				ConditionKeyVal[field.Name()] = fieldVal.Interface()
			}
		}
		if useInsert {
			action.Exec = DefaultExec{"nil sql", nil}
			action.Error = ErrNilPrimaryKey{}
		} else {
			// cas process
			if casField := ctx.Brick.Model.GetFieldWithName("Cas"); casField != nil {
				ConditionKeyVal[casField.Name()] = record.Field(casField.Name()).Int() - 1
			}
			brick := notIgnoreBrick.WhereGroup(ExprAnd, ConditionKeyVal)
			if ctx.Brick.template == nil {
				action.Exec = brick.UpdateExec(record)
				action.Result, action.Error = ctx.Brick.Exec(action.Exec, brick.dbIndex)
			} else {
				tempMap := DefaultCollectionTemplateExec(brick)
				values := ctx.Brick.getFieldValuePairWithRecord(ModeSave, record).ToValueList()
				tempMap["UpdateValues"] = getUpdateValuesExec(values)
				action.Exec, err = ctx.Brick.Toy.Dialect.TemplateExec(*ctx.Brick.template, tempMap)
				if err != nil {
					return err
				}
			}
		}

		ctx.Result.AddRecord(action)
	}
	return nil
}

func CollectionHandlerPreloadFind(ctx *CollectionContext) error {
	ctx.Next()
	for fieldName, preload := range ctx.Brick.BelongToPreload {
		mainField, subField := preload.RelationField, preload.SubModel.GetOnePrimary()
		brick := ctx.Brick.MapPreloadBrick[fieldName]

		mainGroup := ctx.Result.Records.GroupBy(mainField.Name())

		delete(mainGroup, reflect.Zero(mainField.StructField().Type))
		if keys := mainGroup.Keys(); len(keys) != 0 {
			// the relation condition should have lowest priority
			brick = brick.Where(ExprIn, subField, keys).And().Conditions(brick.Search)
			containerList := reflect.New(reflect.SliceOf(ctx.Result.Records.GetFieldType(fieldName))).Elem()
			//var preloadRecords ModelRecords
			subCtx, err := brick.find(LoopIndirectAndNew(containerList))
			ctx.Result.Preload[fieldName] = subCtx.Result
			if err != nil {
				return err
			}
			// set sub data to container field
			subGroup := subCtx.Result.Records.GroupBy(subField.Name())
			ctx.Result.SimpleRelation[fieldName] = map[int]int{}
			for key, records := range mainGroup {
				if subRecords := subGroup[key]; len(subRecords) != 0 {
					for _, record := range records {
						record.SetField(preload.ContainerField.Name(), subRecords[0].Source())
						ctx.Result.SimpleRelation[fieldName][subRecords[0].Index] = record.Index
					}
				}
			}
		}
	}
	for fieldName, preload := range ctx.Brick.OneToOnePreload {
		var mainField, subField Field
		mainField, subField = preload.Model.GetOnePrimary(), preload.RelationField
		brick := ctx.Brick.MapPreloadBrick[fieldName]

		mainGroup := ctx.Result.Records.GroupBy(mainField.Name())
		delete(mainGroup, reflect.Zero(mainField.StructField().Type))
		if keys := mainGroup.Keys(); len(keys) != 0 {
			// the relation condition should have lowest priority
			brick = brick.Where(ExprIn, subField, keys).And().Conditions(brick.Search)
			containerList := reflect.New(reflect.SliceOf(ctx.Result.Records.GetFieldType(fieldName))).Elem()
			//var preloadRecords ModelRecords
			subCtx, err := brick.find(LoopIndirectAndNew(containerList))
			ctx.Result.Preload[fieldName] = subCtx.Result
			if err != nil {
				return err
			}
			// set sub data to container field
			ctx.Result.SimpleRelation[fieldName] = map[int]int{}
			subGroup := subCtx.Result.Records.GroupBy(subField.Name())
			for key, records := range mainGroup {
				if subRecords := subGroup[key]; len(subRecords) != 0 {
					for _, record := range records {
						record.SetField(preload.ContainerField.Name(), subRecords[0].Source())
						ctx.Result.SimpleRelation[fieldName][subRecords[0].Index] = record.Index
					}
				}
			}
		}
	}
	// one to many
	for fieldName, preload := range ctx.Brick.OneToManyPreload {
		mainField, subField := preload.Model.GetOnePrimary(), preload.RelationField
		brick := ctx.Brick.MapPreloadBrick[fieldName]

		mainGroup := ctx.Result.Records.GroupBy(mainField.Name())
		delete(mainGroup, reflect.Zero(mainField.StructField().Type))
		if keys := mainGroup.Keys(); len(keys) != 0 {
			// the relation condition should have lowest priority
			brick = brick.Where(ExprIn, subField, keys).And().Conditions(brick.Search)
			containerList := reflect.New(ctx.Result.Records.GetFieldType(fieldName)).Elem()

			subCtx, err := brick.find(LoopIndirectAndNew(containerList))
			ctx.Result.Preload[fieldName] = subCtx.Result
			if err != nil {
				return err
			}
			subGroup := subCtx.Result.Records.GroupBy(subField.Name())

			ctx.Result.MultipleRelation[fieldName] = map[int]Pair{}
			for key, records := range mainGroup {
				if subRecords := subGroup[key]; len(subRecords) != 0 {
					for _, record := range records {
						container := record.Field(preload.ContainerField.Name())
						containerIndirect := LoopIndirectAndNew(container)
						for j, subRecord := range subRecords {
							containerIndirect.Set(SafeAppend(containerIndirect, subRecord.Source()))
							ctx.Result.MultipleRelation[fieldName][subRecord.Index] = Pair{record.Index, j}
						}
					}
				}
			}
		}
	}
	// many to many
	for fieldName, preload := range ctx.Brick.ManyToManyPreload {
		mainPrimary, subPrimary := preload.Model.GetOnePrimary(), preload.SubModel.GetOnePrimary()
		middleBrick := NewCollectionBrick(ctx.Brick.Toy, preload.MiddleModel).CopyStatus(ctx.Brick)

		// primaryMap: map[model.id]->the model's ModelRecord
		//primaryMap := map[interface{}]ModelRecord{}
		mainGroup := ctx.Result.Records.GroupBy(mainPrimary.Name())
		if keys := mainGroup.Keys(); len(keys) != 0 {
			// the relation condition should have lowest priority
			middleBrick = middleBrick.Where(ExprIn, preload.RelationField, keys).And().Conditions(middleBrick.Search)
			middleModelElemList := reflect.New(reflect.SliceOf(preload.MiddleModel.ReflectType)).Elem()
			//var middleModelRecords ModelRecords
			middleCtx, err := middleBrick.find(middleModelElemList)
			ctx.Result.MiddleModelPreload[fieldName] = middleCtx.Result
			if err != nil {
				return err
			}
			middleGroup := middleCtx.Result.Records.GroupBy(preload.SubRelationField.Name())
			if subKeys := middleGroup.Keys(); len(subKeys) != 0 {
				brick := ctx.Brick.MapPreloadBrick[fieldName]
				// the relation condition should have lowest priority
				brick = brick.Where(ExprIn, subPrimary, subKeys).And().Conditions(brick.Search)
				containerField := reflect.New(ctx.Result.Records.GetFieldType(fieldName)).Elem()
				//var subRecords ModelRecords
				subCtx, err := brick.find(LoopIndirectAndNew(containerField))
				ctx.Result.Preload[fieldName] = subCtx.Result
				if err != nil {
					return err
				}

				ctx.Result.MultipleRelation[fieldName] = map[int]Pair{}
				for j, subRecord := range subCtx.Result.Records.GetRecords() {
					if middleRecords := middleGroup[subRecord.Field(subPrimary.Name()).Interface()]; len(middleRecords) != 0 {
						for _, middleRecord := range middleRecords {
							mainRecord := mainGroup[middleRecord.Field(preload.RelationField.Name()).Interface()][0]
							name := preload.ContainerField.Name()
							container := mainRecord.Field(name)
							containerIndirect := LoopIndirectAndNew(container)
							subi := containerIndirect.Len()
							containerIndirect.Set(SafeAppend(containerIndirect, subRecord.Source()))
							ctx.Result.MultipleRelation[fieldName][j] = Pair{mainRecord.Index, subi}
						}
					}
				}

			}

		}
	}
	return nil
}

func CollectionHandlerFind(ctx *CollectionContext) error {
	if ctx.Brick.dbIndex == -1 {
		return ErrDbIndexNotSet{}
	}
	action := CollectionQueryAction{
		dbIndex: ctx.Brick.dbIndex,
	}
	var err error
	// use template or use default exec
	if ctx.Brick.template == nil {
		action.Exec = ctx.Brick.FindExec(ctx.Result.Records)
	} else {
		tempMap := DefaultCollectionTemplateExec(ctx.Brick)
		tempMap["Columns"] = getColumnExec(ctx.Brick.getSelectFields(ctx.Result.Records).ToColumnList())
		action.Exec, err = ctx.Brick.Toy.Dialect.TemplateExec(*ctx.Brick.template, tempMap)
		if err != nil {
			return err
		}
	}
	rows, err := ctx.Brick.Query(action.Exec, action.dbIndex)
	if err != nil {
		action.Error = append(action.Error, err)
		ctx.Result.AddRecord(action)
		return err
	}
	defer rows.Close()
	// find current data
	min := ctx.Result.Records.Len()
	for rows.Next() {
		elem := reflect.New(ctx.Result.Records.ElemType()).Elem()
		ctx.Result.Records.Len()
		record := ctx.Result.Records.Add(elem)

		var scanners []interface{}
		for _, field := range ctx.Brick.getScanFields(ctx.Result.Records) {
			value := record.Field(field.Name())
			scanners = append(scanners, value.Addr().Interface())
		}
		err := rows.Scan(scanners...)
		action.Error = append(action.Error, err)
	}
	if err := rows.Err(); err != nil {
		action.Error = append(action.Error, err)
	}
	max := ctx.Result.Records.Len()
	action.affectData = makeRange(min, max)
	ctx.Result.AddRecord(action)
	return nil
}

func CollectionHandlerFindOneAssignDbIndex(ctx *CollectionContext) error {
	if ctx.Brick.dbIndex != -1 {
		return nil
	}
	for i := range ctx.Brick.Toy.dbs {
		currentLen := ctx.Result.Records.Len()
		dbCtx := NewCollectionContext(ctx.handlers[ctx.index+1:], ctx.Brick.DBIndex(i), ctx.Result.Records)
		err := dbCtx.Next()
		if err != nil {
			fmt.Printf("not row %v\n", err)
			return err
		}

		ctx.Result.ActionFlow = append(ctx.Result.ActionFlow, dbCtx.Result.ActionFlow...)
		if ctx.Result.Records.Len() > currentLen {
			break
		}
	}
	ctx.Abort()
	return nil
}

func CollectionHandlerFindOne(ctx *CollectionContext) error {
	if ctx.Brick.dbIndex == -1 {
		return ErrDbIndexNotSet{}
	}
	action := CollectionQueryAction{
		dbIndex: ctx.Brick.dbIndex,
	}
	var err error
	// use template or use default exec
	if ctx.Brick.template == nil {
		action.Exec = ctx.Brick.FindExec(ctx.Result.Records)
	} else {
		tempMap := DefaultCollectionTemplateExec(ctx.Brick)
		tempMap["Columns"] = getColumnExec(ctx.Brick.getSelectFields(ctx.Result.Records).ToColumnList())
		action.Exec, err = ctx.Brick.Toy.Dialect.TemplateExec(*ctx.Brick.template, tempMap)
		if err != nil {
			return err
		}
	}
	rows, err := ctx.Brick.Query(action.Exec, action.dbIndex)
	if err != nil {
		action.Error = append(action.Error, err)
		ctx.Result.AddRecord(action)
		return err
	}
	defer rows.Close()
	// find current data
	min := ctx.Result.Records.Len()
	if rows.Next() {
		elem := reflect.New(ctx.Result.Records.ElemType()).Elem()
		ctx.Result.Records.Len()
		record := ctx.Result.Records.Add(elem)

		var scanners []interface{}
		for _, field := range ctx.Brick.getScanFields(ctx.Result.Records) {
			value := record.Field(field.Name())
			scanners = append(scanners, value.Addr().Interface())
		}
		err := rows.Scan(scanners...)
		action.Error = append(action.Error, err)
	}
	max := ctx.Result.Records.Len()
	action.affectData = makeRange(min, max)
	ctx.Result.AddRecord(action)
	return nil
}

func CollectionHandlerSoftDeleteCheck(ctx *CollectionContext) error {
	deletedField := ctx.Brick.Model.GetFieldWithName("DeletedAt")
	if deletedField != nil {
		ctx.Brick = ctx.Brick.Where(ExprNull, deletedField).And().Conditions(ctx.Brick.Search)
	}
	return nil
}

func CollectionHandlerUpdateTimeGenerate(ctx *CollectionContext) error {
	records := ctx.Result.Records
	if updateField := ctx.Brick.Model.GetFieldWithName("UpdatedAt"); updateField != nil {
		current := reflect.ValueOf(time.Now())
		for _, record := range records.GetRecords() {
			record.SetField(updateField.Name(), current)
		}
	}
	return nil
}

func CollectionHandlerUpdate(ctx *CollectionContext) error {
	if ctx.Brick.dbIndex == -1 {
		return ErrDbIndexNotSet{}
	}
	for i, record := range ctx.Result.Records.GetRecords() {
		action := CollectionExecAction{affectData: []int{i}, dbIndex: ctx.Brick.dbIndex}
		var err error
		if ctx.Brick.template == nil {
			action.Exec = ctx.Brick.UpdateExec(record)
		} else {
			tempMap := DefaultCollectionTemplateExec(ctx.Brick)
			values := ctx.Brick.getFieldValuePairWithRecord(ModeUpdate, record).ToValueList()
			tempMap["Columns"] = getColumnExec(columnsValueToColumn(values))
			tempMap["Values"] = getUpdateValuesExec(values)
			action.Exec, err = ctx.Brick.Toy.Dialect.TemplateExec(*ctx.Brick.template, tempMap)
			if err != nil {
				return err
			}
		}
		action.Result, action.Error = ctx.Brick.Exec(action.Exec, action.dbIndex)
		ctx.Result.AddRecord(action)
	}
	return nil
}

func HandlerCollectionPreloadDelete(ctx *CollectionContext) error {
	for fieldName, preload := range ctx.Brick.OneToOnePreload {
		preloadBrick := ctx.Brick.MapPreloadBrick[fieldName]
		subRecords := MakeRecordsWithElem(preload.SubModel, ctx.Result.Records.GetFieldAddressType(fieldName))
		mainSoftDelete := preload.Model.GetFieldWithName("DeletedAt") != nil
		subSoftDelete := preload.SubModel.GetFieldWithName("DeletedAt") != nil
		// set sub model relation field
		for _, record := range ctx.Result.Records.GetRecords() {
			// it means relation field, result[j].LastInsertId() is id value
			subRecords.Add(record.FieldAddress(fieldName))
		}
		// if main model is hard delete need set relationship field set zero if sub model is soft delete
		if mainSoftDelete == false && subSoftDelete == true {
			deletedAtField := preloadBrick.Model.GetFieldWithName("DeletedAt")
			preloadBrick = preloadBrick.bindDefaultFields(preload.RelationField, deletedAtField)
		}
		result, err := preloadBrick.deleteWithPrimaryKey(subRecords)
		ctx.Result.Preload[fieldName] = result
		if err != nil {
			return err
		}
	}

	// one to many
	for fieldName, preload := range ctx.Brick.OneToManyPreload {
		preloadBrick := ctx.Brick.MapPreloadBrick[fieldName]
		mainSoftDelete := preload.Model.GetFieldWithName("DeletedAt") != nil
		subSoftDelete := preload.SubModel.GetFieldWithName("DeletedAt") != nil
		elemAddressType := reflect.PtrTo(LoopTypeIndirect(ctx.Result.Records.GetFieldType(fieldName)).Elem())
		subRecords := MakeRecordsWithElem(preload.SubModel, elemAddressType)
		for _, record := range ctx.Result.Records.GetRecords() {
			rField := LoopIndirect(record.Field(fieldName))
			for subi := 0; subi < rField.Len(); subi++ {
				subRecords.Add(rField.Index(subi).Addr())
			}
		}
		// model relationship field set zero
		if mainSoftDelete == false && subSoftDelete == true {
			deletedAtField := preloadBrick.Model.GetFieldWithName("DeletedAt")
			preloadBrick = preloadBrick.bindDefaultFields(preload.RelationField, deletedAtField)
		}
		result, err := preloadBrick.deleteWithPrimaryKey(subRecords)
		ctx.Result.Preload[fieldName] = result
		if err != nil {
			return err
		}
	}
	// many to many
	for fieldName, preload := range ctx.Brick.ManyToManyPreload {
		subBrick := ctx.Brick.MapPreloadBrick[fieldName]
		middleBrick := NewCollectionBrick(ctx.Brick.Toy, preload.MiddleModel).CopyStatus(ctx.Brick)
		mainField, subField := preload.Model.GetOnePrimary(), preload.SubModel.GetOnePrimary()
		mainSoftDelete := preload.Model.GetFieldWithName("DeletedAt") != nil
		subSoftDelete := preload.SubModel.GetFieldWithName("DeletedAt") != nil

		elemAddressType := reflect.PtrTo(LoopTypeIndirect(ctx.Result.Records.GetFieldType(fieldName)).Elem())
		subRecords := MakeRecordsWithElem(preload.SubModel, elemAddressType)

		for _, record := range ctx.Result.Records.GetRecords() {
			rField := LoopIndirect(record.Field(fieldName))
			for subi := 0; subi < rField.Len(); subi++ {
				subRecords.Add(rField.Index(subi).Addr())
			}
		}

		middleRecords := MakeRecordsWithElem(middleBrick.Model, middleBrick.Model.ReflectType)
		// use to calculate what sub records belong for
		offset := 0
		for _, record := range ctx.Result.Records.GetRecords() {
			primary := record.Field(mainField.Name())
			if primary.IsValid() == false {
				return errors.New("some records have not primary key")
			}
			rField := LoopIndirect(record.Field(fieldName))
			for subi := 0; subi < rField.Len(); subi++ {
				subRecord := subRecords.GetRecord(subi + offset)
				subPrimary := subRecord.Field(subField.Name())
				if subPrimary.IsValid() == false {
					return errors.New("some records have not primary key")
				}
				middleRecord := NewRecord(middleBrick.Model, reflect.New(middleBrick.Model.ReflectType).Elem())
				middleRecord.SetField(preload.RelationField.Name(), primary)
				middleRecord.SetField(preload.SubRelationField.Name(), subPrimary)
				middleRecords.Add(middleRecord.Source())
			}
			offset += rField.Len()
		}

		// delete middle model data
		var primaryFields []Field
		if mainSoftDelete == false {
			primaryFields = append(primaryFields, middleBrick.Model.GetPrimary()[0])
		}
		if subSoftDelete == false {
			primaryFields = append(primaryFields, middleBrick.Model.GetPrimary()[1])
		}
		if len(primaryFields) != 0 {
			conditions := middleBrick.Search
			middleBrick = middleBrick.Conditions(nil)
			for _, primaryField := range primaryFields {
				primarySetType := reflect.MapOf(primaryField.StructField().Type, reflect.TypeOf(struct{}{}))
				primarySet := reflect.MakeMap(primarySetType)
				for _, record := range middleRecords.GetRecords() {
					primarySet.SetMapIndex(record.Field(primaryField.Name()), reflect.ValueOf(struct{}{}))
				}
				var primaryKeys = reflect.New(reflect.SliceOf(primaryField.StructField().Type)).Elem()
				for _, k := range primarySet.MapKeys() {
					primaryKeys = reflect.Append(primaryKeys, k)
				}
				middleBrick = middleBrick.Where(ExprIn, primaryField, primaryKeys.Interface()).
					Or().Conditions(middleBrick.Search)
			}
			middleBrick = middleBrick.And().Conditions(conditions)
			result, err := middleBrick.delete(middleRecords)
			ctx.Result.MiddleModelPreload[fieldName] = result
			if err != nil {
				return err
			}
		}

		result, err := subBrick.deleteWithPrimaryKey(subRecords)
		ctx.Result.Preload[fieldName] = result
		if err != nil {
			return err
		}
	}

	if err := ctx.Next(); err != nil {
		return err
	}

	for fieldName, preload := range ctx.Brick.BelongToPreload {
		preloadBrick := ctx.Brick.MapPreloadBrick[fieldName]
		subRecords := MakeRecordsWithElem(preload.SubModel, ctx.Result.Records.GetFieldAddressType(fieldName))
		for _, record := range ctx.Result.Records.GetRecords() {
			subRecords.Add(record.FieldAddress(fieldName))
		}

		mainSoftDelete := preload.Model.GetFieldWithName("DeletedAt") != nil
		subSoftDelete := preload.SubModel.GetFieldWithName("DeletedAt") != nil
		if mainSoftDelete == false && subSoftDelete == true {
			deletedAtField := preloadBrick.Model.GetFieldWithName("DeletedAt")
			preloadBrick = preloadBrick.bindDefaultFields(preload.RelationField, deletedAtField)
		}

		result, err := preloadBrick.deleteWithPrimaryKey(subRecords)
		ctx.Result.Preload[fieldName] = result
		if err != nil {
			return err
		}

	}
	return nil
}

func HandlerCollectionSearchWithPrimaryKey(ctx *CollectionContext) error {
	var primaryKeys []interface{}
	primaryField := ctx.Brick.Model.GetOnePrimary()
	for _, record := range ctx.Result.Records.GetRecords() {
		primaryKeys = append(primaryKeys, record.Field(primaryField.Name()).Interface())
	}
	if len(primaryKeys) == 0 {
		ctx.Abort()
		return nil
	}
	ctx.Brick = ctx.Brick.Where(ExprIn, primaryField, primaryKeys).And().Conditions(ctx.Brick.Search)
	return nil
}

func HandlerCollectionHardDelete(ctx *CollectionContext) error {
	if ctx.Brick.dbIndex == -1 {
		return ErrDbIndexNotSet{}
	}
	action := CollectionExecAction{dbIndex: ctx.Brick.dbIndex}
	action.Exec = ctx.Brick.DeleteExec()
	action.Result, action.Error = ctx.Brick.Exec(action.Exec, action.dbIndex)
	ctx.Result.AddRecord(action)
	return nil
}

//
func HandlerCollectionSoftDeleteCheck(ctx *CollectionContext) error {
	deletedField := ctx.Brick.Model.GetFieldWithName("DeletedAt")
	if deletedField != nil {
		ctx.Brick = ctx.Brick.Where(ExprNull, deletedField).And().Conditions(ctx.Brick.Search)
	}
	return nil
}

func HandlerCollectionSoftDelete(ctx *CollectionContext) error {
	if ctx.Brick.dbIndex == -1 {
		return ErrDbIndexNotSet{}
	}

	now := time.Now()
	value := reflect.New(ctx.Brick.Model.ReflectType).Elem()
	record := NewStructRecord(ctx.Brick.Model, value)
	record.SetField("DeletedAt", reflect.ValueOf(now))
	bindFields := []interface{}{"DeletedAt"}
	for _, preload := range ctx.Brick.BelongToPreload {
		subSoftDelete := preload.SubModel.GetFieldWithName("DeletedAt") != nil
		if subSoftDelete == false {
			rField := preload.RelationField
			bindFields = append(bindFields, rField.Name())
			record.SetField(rField.Name(), reflect.Zero(rField.StructField().Type))
		}
	}
	action := CollectionExecAction{dbIndex: ctx.Brick.dbIndex}
	ctx.Brick = ctx.Brick.BindFields(ModeUpdate, bindFields...)
	action.Exec = ctx.Brick.UpdateExec(record)
	action.Result, action.Error = ctx.Brick.Exec(action.Exec, action.dbIndex)
	ctx.Result.AddRecord(action)
	return nil
}

func HandlerCollectionCasVersionPushOne(ctx *CollectionContext) error {
	records := ctx.Result.Records
	casField := ctx.Brick.Model.GetFieldWithName("Cas")
	if casField != nil {
		for _, record := range records.GetRecords() {
			record.SetField(casField.Name(), reflect.ValueOf(record.Field("Cas").Int()+1))
		}
	}
	return nil
}
