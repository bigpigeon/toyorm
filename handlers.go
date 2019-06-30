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

func HandlerPreloadInsertOrSave(option string) func(*Context) error {
	return func(ctx *Context) error {
		bindFieldMap := map[string]map[int]int{}
		for fieldName, preload := range ctx.Brick.BelongToPreload {
			preloadBrick := ctx.Brick.MapPreloadBrick[fieldName]
			subRecords := MakeRecordsWithElem(preload.SubModel, ctx.Result.Records.GetFieldAddressType(fieldName))

			// map[i]=>j [i]record.SubData -> [j]subRecord
			bindFieldMap[fieldName] = map[int]int{}
			ctx.Result.SimpleRelation[fieldName] = map[int]int{}
			for i, record := range ctx.Result.Records.GetRecords() {
				if ctx.Brick.ignoreModeSelector[ModePreload].Ignore(record.Field(fieldName)) == false {
					j := subRecords.Len()
					bindFieldMap[fieldName][i] = j
					ctx.Result.SimpleRelation[fieldName][j] = i
					subRecords.Add(record.FieldAddress(fieldName))
				}
			}
			subCtx := preloadBrick.GetContext(option, subRecords)
			ctx.Result.Preload[fieldName] = subCtx.Result
			go subCtx.Start()

		}
		var hasErr bool
		// set model relation field
		for fieldName, preload := range ctx.Brick.BelongToPreload {
			subResult := ctx.Result.Preload[fieldName]
			<-subResult.done
			mainField, subField := preload.RelationField, preload.SubModel.GetOnePrimary()
			if subResult.Err() != nil {
				hasErr = true
			}
			for i, record := range ctx.Result.Records.GetRecords() {
				if j, ok := bindFieldMap[fieldName][i]; ok {
					subRecord := subResult.Records.GetRecord(j)
					record.SetField(mainField.Name(), subRecord.Field(subField.Name()))

				}
			}
		}
		if hasErr {
			return errors.New("Cancel by belong to preload field has error ")
		}
		ctx.Next()
		if err := ctx.Result.Err(); err != nil {
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
			go subCtx.Start()

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
			go subCtx.Start()
		}
		// many to many
		for fieldName, preload := range ctx.Brick.ManyToManyPreload {
			subBrick := ctx.Brick.MapPreloadBrick[fieldName]

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
			go subCtx.Start()
		}

		for fieldName, preload := range ctx.Brick.ManyToManyPreload {
			subResult := ctx.Result.Preload[fieldName]
			<-subResult.done
			if subResult.Err() == nil {
				middleBrick := NewToyBrick(ctx.Brick.Toy, preload.MiddleModel).CopyStatus(ctx.Brick)

				mainField, subField := preload.Model.GetOnePrimary(), preload.SubModel.GetOnePrimary()
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
						subRecord := subResult.Records.GetRecord(subi + offset)
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
				go middleCtx.Start()
			}
		}

		for fieldName := range ctx.Brick.OneToOnePreload {
			<-ctx.Result.Preload[fieldName].done
		}
		for fieldName := range ctx.Brick.OneToManyPreload {
			<-ctx.Result.Preload[fieldName].done
		}
		for fieldName := range ctx.Brick.ManyToManyPreload {
			subResult := ctx.Result.Preload[fieldName]
			if subResult.Err() == nil {
				<-ctx.Result.MiddleModelPreload[fieldName].done
			}
		}
		return nil
	}
}

func HandlerInsertTimeGenerate(ctx *Context) error {
	now := reflect.ValueOf(time.Now())
	records := ctx.Result.Records.GetRecords()
	if createAtField := ctx.Brick.Model.GetFieldWithName("CreatedAt"); createAtField != nil {
		for _, record := range records {
			if fieldValue := record.Field(createAtField.Name()); fieldValue.IsValid() == false || IsZero(fieldValue) {
				record.SetField(createAtField.Name(), now)
			}
		}
	}
	if updateField := ctx.Brick.Model.GetFieldWithName("UpdatedAt"); updateField != nil {
		for _, record := range records {
			record.SetField(updateField.Name(), now)
		}
	}
	return nil
}

func HandlerInsert(ctx *Context) error {

	// current insert
	//setInsertId := len(ctx.Brick.Model.GetPrimary()) == 1 && ctx.Brick.Model.GetOnePrimary().AutoIncrement() == true
	for i, record := range ctx.Result.Records.GetRecords() {
		action := ExecAction{affectData: []int{i}}
		var err error
		action.Exec, err = ctx.Brick.InsertExec(record)
		if err != nil {
			return err
		}
		var executor Executor
		if ctx.Brick.tx != nil {
			executor = ctx.Brick.tx
		} else {
			executor = ctx.Brick.Toy.db
		}
		action.Result, action.Error = ctx.Brick.Toy.Dialect.InsertExecutor(
			executor,
			action.Exec,
			ctx.Brick.debugPrint,
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

func HandlerCasVersionPushOne(ctx *Context) error {
	records := ctx.Result.Records
	casField := ctx.Brick.Model.GetFieldWithName("Cas")
	if casField != nil {
		for _, record := range records.GetRecords() {
			record.SetField(casField.Name(), reflect.ValueOf(record.Field("Cas").Int()+1))
		}
	}
	return nil
}

func HandlerFind(ctx *Context) error {
	var action QueryAction
	var err error
	columns, scannersGen := FindColumnFactory(ctx.Result.Records, ctx.Brick)

	// use template or use default exec
	action.Exec, err = ctx.Brick.FindExec(columns)
	if err != nil {
		return err
	}
	rows, err := ctx.Brick.Query(action.Exec)
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
		record := ctx.Result.Records.Add(elem)
		scanners := scannersGen(record)
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

func HandlerFindOne(ctx *Context) error {
	var action QueryAction
	var err error
	columns, scannersGen := FindColumnFactory(ctx.Result.Records, ctx.Brick)

	// use template or use default exec
	action.Exec, err = ctx.Brick.FindExec(columns)
	if err != nil {
		return err
	}
	row := ctx.Brick.QueryRow(action.Exec)
	// find current data
	record := ctx.Result.Records.GetRecord(0)
	scanners := scannersGen(record)
	err = row.Scan(scanners...)
	action.Error = append(action.Error, err)

	action.affectData = makeRange(0, 1)
	ctx.Result.AddRecord(action)
	return nil
}

func HandlerPreloadContainerCheck(ctx *Context) error {
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

func HandlerPreloadOnJoinFind(ctx *Context) error {
	for name := range ctx.Brick.JoinMap {
		if len(ctx.Brick.SwapMap[name].MapPreloadBrick) != 0 {
			brick := ctx.Brick.Join(name)
			records := MakeRecordsWithElem(brick.Model, ctx.Result.Records.GetFieldAddressType(name))
			for _, mainRecord := range ctx.Result.Records.GetRecords() {
				records.Add(mainRecord.FieldAddress(name))
			}

			joinCtx := NewContext(ctx.handlers[ctx.index+1:], brick, records)
			joinCtx.Start()
			<-joinCtx.Done()

			for preloadName, result := range joinCtx.Result.Preload {
				fieldName := fmt.Sprintf("j_%s_%s", name, preloadName)
				ctx.Result.Preload[fieldName] = result
				if relation, ok := joinCtx.Result.SimpleRelation[preloadName]; ok {
					ctx.Result.SimpleRelation[fieldName] = relation
				} else if relation, ok := joinCtx.Result.MultipleRelation[preloadName]; ok {
					ctx.Result.MultipleRelation[fieldName] = relation
				}
			}
			for joinFieldName, result := range joinCtx.Result.MiddleModelPreload {
				fieldName := fmt.Sprintf("j_%s_%s", name, joinFieldName)
				ctx.Result.MiddleModelPreload[fieldName] = result
				if relation, ok := joinCtx.Result.SimpleRelation[joinFieldName]; ok {
					ctx.Result.SimpleRelation[fieldName] = relation
				} else if relation, ok := joinCtx.Result.MultipleRelation[joinFieldName]; ok {
					ctx.Result.MultipleRelation[fieldName] = relation
				}
			}
		}
	}
	return nil
}

func HandlerPreloadFind(ctx *Context) error {
	belongToGroupList := map[string]ModelGroupBy{}
	for fieldName, preload := range ctx.Brick.BelongToPreload {
		mainField, subField := preload.RelationField, preload.SubModel.GetOnePrimary()
		brick := ctx.Brick.MapPreloadBrick[fieldName]
		belongToGroupList[fieldName] = ctx.Result.Records.GroupBy(mainField.Name())
		delete(belongToGroupList[fieldName], reflect.Zero(mainField.StructField().Type))
		if keys := belongToGroupList[fieldName].Keys(); len(keys) != 0 {
			// the relation condition should have lowest priority
			brick = brick.Where(ExprIn, subField, keys).And().Conditions(brick.Search)
			containerList := reflect.New(reflect.SliceOf(ctx.Result.Records.GetFieldType(fieldName))).Elem()
			//var preloadRecords ModelRecords
			subCtx := brick.find(LoopIndirectAndNew(containerList))
			ctx.Result.Preload[fieldName] = subCtx.Result

		}
	}

	oneToOneGroup := map[string]ModelGroupBy{}
	for fieldName, preload := range ctx.Brick.OneToOnePreload {
		var mainField, subField Field
		mainField, subField = preload.Model.GetOnePrimary(), preload.RelationField
		brick := ctx.Brick.MapPreloadBrick[fieldName]
		oneToOneGroup[fieldName] = ctx.Result.Records.GroupBy(mainField.Name())
		delete(oneToOneGroup[fieldName], reflect.Zero(mainField.StructField().Type))
		if keys := oneToOneGroup[fieldName].Keys(); len(keys) != 0 {
			// the relation condition should have lowest priority
			brick = brick.Where(ExprIn, subField, keys).And().Conditions(brick.Search)
			containerList := reflect.New(reflect.SliceOf(ctx.Result.Records.GetFieldType(fieldName))).Elem()
			//var preloadRecords ModelRecords
			subCtx := brick.find(LoopIndirectAndNew(containerList))
			ctx.Result.Preload[fieldName] = subCtx.Result

		}
	}
	// one to many
	oneToManyGroup := map[string]ModelGroupBy{}
	for fieldName, preload := range ctx.Brick.OneToManyPreload {
		mainField, subField := preload.Model.GetOnePrimary(), preload.RelationField
		brick := ctx.Brick.MapPreloadBrick[fieldName]

		oneToManyGroup[fieldName] = ctx.Result.Records.GroupBy(mainField.Name())
		delete(oneToManyGroup[fieldName], reflect.Zero(mainField.StructField().Type))
		if keys := oneToManyGroup[fieldName].Keys(); len(keys) != 0 {
			// the relation condition should have lowest priority
			brick = brick.Where(ExprIn, subField, keys).And().Conditions(brick.Search)
			containerList := reflect.New(ctx.Result.Records.GetFieldType(fieldName)).Elem()

			subCtx := brick.find(LoopIndirectAndNew(containerList))
			ctx.Result.Preload[fieldName] = subCtx.Result

		}
	}

	manyToManyGroup := map[string]ModelGroupBy{}
	// many to many
	for fieldName, preload := range ctx.Brick.ManyToManyPreload {
		mainPrimary := preload.Model.GetOnePrimary()
		middleBrick := NewToyBrick(ctx.Brick.Toy, preload.MiddleModel).CopyStatus(ctx.Brick)
		// primaryMap: map[model.id]->the model's ModelRecord
		//primaryMap := map[interface{}]ModelRecord{}
		manyToManyGroup[fieldName] = ctx.Result.Records.GroupBy(mainPrimary.Name())
		delete(manyToManyGroup[fieldName], reflect.Zero(mainPrimary.StructField().Type))
		if keys := manyToManyGroup[fieldName].Keys(); len(keys) != 0 {
			// the relation condition should have lowest priority
			middleBrick = middleBrick.Where(ExprIn, preload.RelationField, keys).And().Conditions(middleBrick.Search)
			middleModelElemList := reflect.New(reflect.SliceOf(preload.MiddleModel.ReflectType)).Elem()
			//var middleModelRecords ModelRecords
			middleCtx := middleBrick.find(middleModelElemList)
			ctx.Result.MiddleModelPreload[fieldName] = middleCtx.Result

		}
	}

	manyMiddleGroup := map[string]ModelGroupBy{}
	for fieldName, preload := range ctx.Brick.ManyToManyPreload {
		middleResult := ctx.Result.MiddleModelPreload[fieldName]
		<-middleResult.done
		if err := middleResult.Err(); err != nil {
			continue
		}

		subPrimary := preload.SubModel.GetOnePrimary()

		manyMiddleGroup[fieldName] = middleResult.Records.GroupBy(preload.SubRelationField.Name())
		if subKeys := manyMiddleGroup[fieldName].Keys(); len(subKeys) != 0 {
			brick := ctx.Brick.MapPreloadBrick[fieldName]
			// the relation condition should have lowest priority
			brick = brick.Where(ExprIn, subPrimary, subKeys).And().Conditions(brick.Search)
			containerField := reflect.New(ctx.Result.Records.GetFieldType(fieldName)).Elem()
			//var subRecords ModelRecords
			subCtx := brick.find(LoopIndirectAndNew(containerField))
			ctx.Result.Preload[fieldName] = subCtx.Result
		}
	}

	for fieldName, preload := range ctx.Brick.ManyToManyPreload {
		subResult := ctx.Result.Preload[fieldName]
		middleResult := ctx.Result.MiddleModelPreload[fieldName]
		subPrimary := preload.SubModel.GetOnePrimary()
		<-subResult.done
		if err := subResult.Err(); err != nil {
			continue
		}
		mainPrimary := preload.Model.GetOnePrimary()
		ctx.Result.MultipleRelation[fieldName] = map[int]Pair{}
		subTypeConvert := func(val reflect.Value) reflect.Value {
			return val
		}
		middleCompareType := LoopTypeIndirect(middleResult.Records.GetFieldType(preload.SubRelationField.Name()))
		if LoopTypeIndirect(subResult.Records.GetFieldType(subPrimary.Name())) != middleCompareType {
			subTypeConvert = func(val reflect.Value) reflect.Value {
				return val.Convert(middleCompareType)
			}
		}
		middleTypeConvert := func(val reflect.Value) reflect.Value {
			return val
		}
		mainCompareType := LoopTypeIndirect(ctx.Result.Records.GetFieldType(mainPrimary.Name()))
		if LoopTypeIndirect(middleResult.Records.GetFieldType(preload.RelationField.Name())) != mainCompareType {
			middleTypeConvert = func(val reflect.Value) reflect.Value {
				return val.Convert(mainCompareType)
			}
		}

		for j, subRecord := range subResult.Records.GetRecords() {

			if middleRecords := manyMiddleGroup[fieldName][subTypeConvert(subRecord.Field(subPrimary.Name())).Interface()]; len(middleRecords) != 0 {
				for _, middleRecord := range middleRecords {
					mainRecord := manyToManyGroup[fieldName][middleTypeConvert(middleRecord.Field(preload.RelationField.Name())).Interface()][0]
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

	for fieldName, preload := range ctx.Brick.BelongToPreload {
		subResult := ctx.Result.Preload[fieldName]
		<-subResult.done
		if err := subResult.Err(); err != nil {
			continue
		}
		subField := preload.SubModel.GetOnePrimary()
		// set sub data to container field
		subGroup := subResult.Records.GroupBy(subField.Name())
		ctx.Result.SimpleRelation[fieldName] = map[int]int{}
		for key, records := range belongToGroupList[fieldName] {
			if subRecords := subGroup[key]; len(subRecords) != 0 {
				for _, record := range records {
					record.SetField(preload.ContainerField.Name(), subRecords[0].Source())
					ctx.Result.SimpleRelation[fieldName][subRecords[0].Index] = record.Index
				}
			}
		}
	}

	for fieldName, preload := range ctx.Brick.OneToOnePreload {
		subResult := ctx.Result.Preload[fieldName]
		<-subResult.done
		if err := subResult.Err(); err != nil {
			continue
		}
		// set sub data to container field
		subField := preload.RelationField
		ctx.Result.SimpleRelation[fieldName] = map[int]int{}
		subGroup := subResult.Records.GroupBy(subField.Name())
		for key, records := range oneToOneGroup[fieldName] {
			if subRecords := subGroup[key]; len(subRecords) != 0 {
				for _, record := range records {
					record.SetField(preload.ContainerField.Name(), subRecords[0].Source())
					ctx.Result.SimpleRelation[fieldName][subRecords[0].Index] = record.Index
				}
			}
		}
	}

	for fieldName, preload := range ctx.Brick.OneToManyPreload {
		subResult := ctx.Result.Preload[fieldName]
		<-subResult.done
		if err := subResult.Err(); err != nil {
			continue
		}
		subField := preload.RelationField
		subGroup := subResult.Records.GroupBy(subField.Name())
		ctx.Result.MultipleRelation[fieldName] = map[int]Pair{}
		for key, records := range oneToManyGroup[fieldName] {
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

	return nil
}

func HandlerUpdateTimeGenerate(ctx *Context) error {
	records := ctx.Result.Records
	if updateField := ctx.Brick.Model.GetFieldWithName("UpdatedAt"); updateField != nil {
		current := reflect.ValueOf(time.Now())
		for _, record := range records.GetRecords() {
			record.SetField(updateField.Name(), current)
		}
	}
	return nil
}

func HandlerUpdate(ctx *Context) error {
	for i, record := range ctx.Result.Records.GetRecords() {
		action := ExecAction{affectData: []int{i}}
		var err error
		action.Exec, err = ctx.Brick.UpdateExec(record)
		if err != nil {
			return err
		}

		action.Result, action.Error = ctx.Brick.Exec(action.Exec)
		ctx.Result.AddRecord(action)
	}
	return nil
}

// if have not primary ,try to insert
// else try to replace
func HandlerSave(ctx *Context) error {
	//setInsertId := len(ctx.Brick.Model.GetPrimary()) == 1 && ctx.Brick.Model.GetOnePrimary().AutoIncrement() == true
	var executor Executor
	if ctx.Brick.tx != nil {
		executor = ctx.Brick.tx
	} else {
		executor = ctx.Brick.Toy.db
	}
	for i, record := range ctx.Result.Records.GetRecords() {
		var action ExecAction
		var err error
		action = ExecAction{affectData: []int{i}}
		var useInsert bool
		// if have any zero primary key, use insert ,otherwise use save
		for _, key := range ctx.Brick.Model.PrimaryFields {
			if field := record.Field(key.Name()); field.IsValid() == false || IsZero(field) {
				useInsert = true
			}
		}

		if useInsert {
			action.Exec, err = ctx.Brick.InsertExec(record)
			if err != nil {
				return err
			}
			action.Result, action.Error = ctx.Brick.Toy.Dialect.InsertExecutor(
				executor,
				action.Exec,
				ctx.Brick.debugPrint,
			)

			if action.Error == nil {
				// set primary field value if model has one primary key
				err = setNumberPrimaryKey(ctx, record, action)
				if err != nil {
					return err
				}
			}
		} else {
			action.Exec, err = ctx.Brick.SaveExec(record)
			if err != nil {
				return err
			}
			action.Result, action.Error = ctx.Brick.Toy.Dialect.SaveExecutor(
				executor,
				action.Exec,
				ctx.Brick.debugPrint,
			)
		}

		ctx.Result.AddRecord(action)
	}
	return nil
}

func HandlerUSave(ctx *Context) error {
	//setInsertId := len(ctx.Brick.Model.GetPrimary()) == 1 && ctx.Brick.Model.GetOnePrimary().AutoIncrement() == true
	var executor Executor
	if ctx.Brick.tx != nil {
		executor = ctx.Brick.tx
	} else {
		executor = ctx.Brick.Toy.db
	}
	for i, record := range ctx.Result.Records.GetRecords() {
		var action ExecAction
		var err error
		action = ExecAction{affectData: []int{i}}
		action.Exec, err = ctx.Brick.USaveExec(record)
		if err != nil {
			return err
		}
		action.Result, action.Error = ctx.Brick.Toy.Dialect.SaveExecutor(
			executor,
			action.Exec,
			ctx.Brick.debugPrint,
		)
		ctx.Result.AddRecord(action)
	}

	return nil
}

func HandlerSaveTimeGenerate(ctx *Context) error {
	now := reflect.ValueOf(time.Now())
	records := ctx.Result.Records.GetRecords()
	if createAtField := ctx.Brick.Model.GetFieldWithName("CreatedAt"); createAtField != nil {
		for _, record := range records {
			if fieldValue := record.Field(createAtField.Name()); fieldValue.IsValid() == false || IsZero(fieldValue) {
				record.SetField(createAtField.Name(), now)
			}
		}
	}
	if updateField := ctx.Brick.Model.GetFieldWithName("UpdatedAt"); updateField != nil {
		for _, record := range records {
			record.SetField(updateField.Name(), now)
		}
	}
	return nil
}

func HandlerUSaveTimeGenerate(ctx *Context) error {
	now := reflect.ValueOf(time.Now())
	if updateField := ctx.Brick.Model.GetFieldWithName("UpdatedAt"); updateField != nil {
		for _, record := range ctx.Result.Records.GetRecords() {
			record.SetField(updateField.Name(), now)
		}
	}
	return nil
}

// preload schedule belongTo -> Next() -> oneToOne -> oneToMany -> manyToMany(sub -> middle)
func HandlerCreateTablePreload(option string) func(ctx *Context) error {
	return func(ctx *Context) (err error) {
		for fieldName := range ctx.Brick.BelongToPreload {
			brick := ctx.Brick.MapPreloadBrick[fieldName]
			subCtx := brick.GetContext(option, MakeRecordsWithElem(brick.Model, brick.Model.ReflectType))
			ctx.Result.Preload[fieldName] = subCtx.Result
			subCtx.Start()
			<-subCtx.Result.done
			if err := subCtx.Result.Err(); err != nil {
				return err
			}
		}
		ctx.Next()
		if err := ctx.Result.Err(); err != nil {
			return err
		}
		for fieldName := range ctx.Brick.OneToOnePreload {
			brick := ctx.Brick.MapPreloadBrick[fieldName]
			subCtx := brick.GetContext(option, MakeRecordsWithElem(brick.Model, brick.Model.ReflectType))
			ctx.Result.Preload[fieldName] = subCtx.Result
			subCtx.Start()
			<-subCtx.Result.done
			if err := subCtx.Result.Err(); err != nil {
				return err
			}
		}

		for fieldName := range ctx.Brick.OneToManyPreload {
			brick := ctx.Brick.MapPreloadBrick[fieldName]
			subCtx := brick.GetContext(option, MakeRecordsWithElem(brick.Model, brick.Model.ReflectType))
			ctx.Result.Preload[fieldName] = subCtx.Result
			subCtx.Start()
			<-subCtx.Result.done
			if err := subCtx.Result.Err(); err != nil {
				return err
			}
		}
		for fieldName, preload := range ctx.Brick.ManyToManyPreload {
			{
				brick := ctx.Brick.MapPreloadBrick[fieldName]
				subCtx := brick.GetContext(option, MakeRecordsWithElem(brick.Model, brick.Model.ReflectType))
				ctx.Result.Preload[fieldName] = subCtx.Result
				subCtx.Start()
				<-subCtx.Result.done
				if err := subCtx.Result.Err(); err != nil {
					return err
				}
			}
			// process middle model
			{
				middleModel := preload.MiddleModel
				brick := NewToyBrick(ctx.Brick.Toy, middleModel).CopyStatus(ctx.Brick)
				// copy PreToyBrick
				brick = brick.Scope(func(t *ToyBrick) *ToyBrick {
					newt := *t
					newt.preBrick = PreToyBrick{
						ctx.Brick,
						preload.ContainerField,
					}
					return &newt
				})
				middleCtx := brick.GetContext(option, MakeRecordsWithElem(brick.Model, brick.Model.ReflectType))
				ctx.Result.MiddleModelPreload[fieldName] = middleCtx.Result
				middleCtx.Start()
				<-middleCtx.Result.done
				if err := middleCtx.Result.Err(); err != nil {
					return err
				}
			}
		}
		return nil
	}
}

// preload schedule oneToOne -> oneToMany -> current model -> manyToMany(sub -> middle) -> Next() -> belongTo
func HandlerDropTablePreload(option string) func(ctx *Context) error {
	return func(ctx *Context) (err error) {
		for fieldName := range ctx.Brick.OneToOnePreload {
			brick := ctx.Brick.MapPreloadBrick[fieldName]
			subCtx := brick.GetContext(option, MakeRecordsWithElem(brick.Model, brick.Model.ReflectType))
			ctx.Result.Preload[fieldName] = subCtx.Result
			subCtx.Start()
			<-subCtx.Done()
			if err := subCtx.Result.Err(); err != nil {
				return err
			}

		}
		for fieldName := range ctx.Brick.OneToManyPreload {
			brick := ctx.Brick.MapPreloadBrick[fieldName]
			subCtx := brick.GetContext(option, MakeRecordsWithElem(brick.Model, brick.Model.ReflectType))
			ctx.Result.Preload[fieldName] = subCtx.Result
			subCtx.Start()
			<-subCtx.Done()
			if err := subCtx.Result.Err(); err != nil {
				return err
			}
		}
		for fieldName, preload := range ctx.Brick.ManyToManyPreload {
			// process middle model
			{
				middleModel := preload.MiddleModel
				brick := NewToyBrick(ctx.Brick.Toy, middleModel).CopyStatus(ctx.Brick)
				middleCtx := brick.GetContext(option, MakeRecordsWithElem(brick.Model, brick.Model.ReflectType))
				ctx.Result.MiddleModelPreload[fieldName] = middleCtx.Result
				middleCtx.Start()
				<-middleCtx.Done()
				if err := middleCtx.Result.Err(); err != nil {
					return err
				}
			}
			// process sub model
			{
				brick := ctx.Brick.MapPreloadBrick[fieldName]
				subCtx := brick.GetContext(option, MakeRecordsWithElem(brick.Model, brick.Model.ReflectType))
				ctx.Result.Preload[fieldName] = subCtx.Result
				subCtx.Start()
				<-subCtx.Done()
				if err := subCtx.Result.Err(); err != nil {
					return err
				}
			}
		}
		ctx.Next()
		if err := ctx.Result.Err(); err != nil {
			return err
		}
		for fieldName := range ctx.Brick.BelongToPreload {
			brick := ctx.Brick.MapPreloadBrick[fieldName]
			subCtx := brick.GetContext(option, MakeRecordsWithElem(brick.Model, brick.Model.ReflectType))
			ctx.Result.Preload[fieldName] = subCtx.Result
			subCtx.Start()
			<-subCtx.Done()
			if err := subCtx.Result.Err(); err != nil {
				return err
			}

		}

		return nil
	}
}

func HandlerCreateTable(ctx *Context) error {
	foreign := map[string]ForeignKey{}
	for _, field := range ctx.Brick.Model.GetSqlFields() {
		// this is foreign key, mean it must relationship field with parent or child
		if field.IsForeign() {
			if ctx.Brick.preBrick.Parent != nil {
				parent, containerField := ctx.Brick.preBrick.Parent, ctx.Brick.preBrick.Field
				if preload := parent.OneToOnePreload[containerField.Name()]; preload != nil {
					if preload.RelationField.Name() == field.Name() {
						foreign[field.Name()] = ForeignKey{preload.Model, preload.Model.GetOnePrimary()}
					}
				} else if preload := parent.OneToManyPreload[containerField.Name()]; preload != nil {
					if preload.RelationField.Name() == field.Name() {
						foreign[field.Name()] = ForeignKey{preload.Model, preload.Model.GetOnePrimary()}
					}
				} else if preload := parent.ManyToManyPreload[containerField.Name()]; preload != nil {
					if preload.SubRelationField.Name() == field.Name() {
						foreign[field.Name()] = ForeignKey{preload.SubModel, preload.SubModel.GetOnePrimary()}
					} else if preload.RelationField.Name() == field.Name() {
						foreign[field.Name()] = ForeignKey{preload.Model, preload.Model.GetOnePrimary()}
					}
				}
			}
			// search belong to
			for _, preload := range ctx.Brick.BelongToPreload {
				if preload.RelationField.Name() == field.Name() {
					foreign[field.Name()] = ForeignKey{preload.SubModel, preload.SubModel.GetOnePrimary()}
				}
			}
		}
	}

	execs := ctx.Brick.Toy.Dialect.CreateTable(ctx.Brick.Model, foreign)
	for _, exec := range execs {
		action := ExecAction{Exec: exec}
		action.Result, action.Error = ctx.Brick.Exec(exec)
		ctx.Result.AddRecord(action)
	}

	return nil
}

func HandlerExistTableAbort(ctx *Context) error {
	action := QueryAction{}
	action.Exec = ctx.Brick.Toy.Dialect.HasTable(ctx.Brick.Model)
	var hasTable bool
	err := ctx.Brick.QueryRow(action.Exec).Scan(&hasTable)
	if err != nil {
		action.Error = append(action.Error, err)
	}
	ctx.Result.AddRecord(action)
	if err != nil || hasTable == true {
		ctx.Abort()
	}

	return nil
}

func HandlerDropTable(ctx *Context) (err error) {
	exec := ctx.Brick.Toy.Dialect.DropTable(ctx.Brick.Model)
	action := ExecAction{Exec: exec}
	action.Result, action.Error = ctx.Brick.Exec(exec)
	ctx.Result.AddRecord(action)
	return nil
}

func HandlerNotExistTableAbort(ctx *Context) error {
	action := QueryAction{}
	action.Exec = ctx.Brick.Toy.Dialect.HasTable(ctx.Brick.Model)
	var hasTable bool
	err := ctx.Brick.QueryRow(action.Exec).Scan(&hasTable)
	if err != nil {
		action.Error = append(action.Error, err)
	}
	ctx.Result.AddRecord(action)
	if err != nil || hasTable == false {
		ctx.Abort()
	}
	return nil
}

func HandlerPreloadDelete(ctx *Context) error {
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
		result := preloadBrick.deleteWithPrimaryKey(subRecords)
		ctx.Result.Preload[fieldName] = result
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
		fmt.Printf("sub records %v\n", subRecords.Source().Interface())
		result := preloadBrick.deleteWithPrimaryKey(subRecords)
		ctx.Result.Preload[fieldName] = result

	}
	// many to many
	manyToManySubRecordMap := map[string]ModelRecords{}
	for fieldName, preload := range ctx.Brick.ManyToManyPreload {

		middleBrick := NewToyBrick(ctx.Brick.Toy, preload.MiddleModel).CopyStatus(ctx.Brick)
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
		// if has hard delete element in main field or sub field , need delete it's primary key in middle model
		// avoid to pointing to empty primary keys
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
			result := middleBrick.delete(middleRecords)
			ctx.Result.MiddleModelPreload[fieldName] = result
		}
		manyToManySubRecordMap[fieldName] = subRecords
	}
	var hasErr bool
	for fieldName, preload := range ctx.Brick.ManyToManyPreload {
		mainSoftDelete := preload.Model.GetFieldWithName("DeletedAt") != nil
		subSoftDelete := preload.SubModel.GetFieldWithName("DeletedAt") != nil
		if mainSoftDelete != true || subSoftDelete != true {
			middleResult := ctx.Result.MiddleModelPreload[fieldName]
			<-middleResult.done
			if middleResult.Err() != nil {
				hasErr = true
				continue
			}
		}
		subBrick := ctx.Brick.MapPreloadBrick[fieldName]
		result := subBrick.deleteWithPrimaryKey(manyToManySubRecordMap[fieldName])
		ctx.Result.Preload[fieldName] = result
	}

	for fieldName := range ctx.Brick.ManyToManyPreload {
		subResult := ctx.Result.Preload[fieldName]
		<-subResult.done
		if subResult.Err() != nil {
			hasErr = true
		}
	}

	for fieldName := range ctx.Brick.OneToManyPreload {
		subResult := ctx.Result.Preload[fieldName]
		<-subResult.done
		if subResult.Err() != nil {
			hasErr = true
		}
	}
	for fieldName := range ctx.Brick.OneToOnePreload {
		subResult := ctx.Result.Preload[fieldName]
		<-subResult.done
		if subResult.Err() != nil {
			hasErr = true
		}
	}
	if hasErr == true {
		return errors.New("Cancel by other preload field has error ")
	}
	ctx.Next()
	if err := ctx.Result.Err(); err != nil {
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

		result := preloadBrick.deleteWithPrimaryKey(subRecords)
		ctx.Result.Preload[fieldName] = result
	}
	for fieldName := range ctx.Brick.BelongToPreload {
		subResult := ctx.Result.Preload[fieldName]
		<-subResult.done
	}

	return nil
}

func HandlerSearchWithPrimaryKey(ctx *Context) error {
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

func HandlerHardDelete(ctx *Context) error {
	action := ExecAction{}
	var err error
	action.Exec, err = ctx.Brick.HardDeleteExec(ctx.Result.Records)
	if err != nil {
		return err
	}
	action.Result, action.Error = ctx.Brick.Exec(action.Exec)
	ctx.Result.AddRecord(action)
	return nil
}

//
func HandlerSoftDeleteCheck(ctx *Context) error {
	deletedField := ctx.Brick.Model.GetFieldWithName("DeletedAt")
	if deletedField != nil {
		ctx.Brick = ctx.Brick.Where(ExprNull, deletedField).And().Conditions(ctx.Brick.Search)
	}
	return nil
}

func HandlerSoftDelete(ctx *Context) error {
	action := ExecAction{}
	var err error
	action.Exec, err = ctx.Brick.SoftDeleteExec(ctx.Result.Records)
	if err != nil {
		return err
	}
	action.Result, action.Error = ctx.Brick.Exec(action.Exec)
	ctx.Result.AddRecord(action)
	return nil
}
