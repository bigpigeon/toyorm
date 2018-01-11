package toyorm

import (
	"errors"
	"fmt"
	"reflect"
	"time"
)

func HandlerPreloadInsertOrSave(option string) func(*Context) error {
	return func(ctx *Context) error {
		for mField, preload := range ctx.Brick.OneToOnePreload {
			if preload.IsBelongTo == true {
				mainField, subField := preload.RelationField, preload.SubModel.GetOnePrimary()
				preloadBrick := ctx.Brick.Preload(mField)
				subRecords := MakeRecordsWithElem(preload.SubModel, ctx.Result.Records.GetFieldAddressType(mField))
				for _, record := range ctx.Result.Records.GetRecords() {
					subRecords.Add(record.FieldAddress(mField))
				}
				subCtx := preloadBrick.GetContext(option, subRecords)
				ctx.Result.Preload[mField] = subCtx.Result
				if err := subCtx.Next(); err != nil {
					return err
				}
				// set model relation field
				for jdx, record := range ctx.Result.Records.GetRecords() {
					subRecord := subRecords.GetRecord(jdx)
					record.SetField(mainField, subRecord.Field(subField))
				}
			}
		}

		if err := ctx.Next(); err != nil {
			return err
		}
		for mField, preload := range ctx.Brick.OneToOnePreload {
			if preload.IsBelongTo == false {
				preloadBrick := ctx.Brick.Preload(mField)
				mainPos, subPos := preload.Model.GetOnePrimary(), preload.RelationField
				subRecords := MakeRecordsWithElem(preload.SubModel, ctx.Result.Records.GetFieldAddressType(mField))
				// set sub model relation field
				for i, record := range ctx.Result.Records.GetRecords() {
					// it means relation field, result[j].LastInsertId() is id value
					subRecords.Add(record.FieldAddress(mField))
					if primary := record.Field(mainPos); primary.IsValid() {
						subRecords.GetRecord(i).SetField(subPos, primary)
					} else {
						panic("relation field not set")
					}
				}
				subCtx := preloadBrick.GetContext(option, subRecords)
				ctx.Result.Preload[mField] = subCtx.Result
				if err := subCtx.Next(); err != nil {
					return err
				}
			}
		}

		// one to many
		for mField, preload := range ctx.Brick.OneToManyPreload {
			preloadBrick := ctx.Brick.Preload(mField)
			mainField, subField := preload.Model.GetOnePrimary(), preload.RelationField
			elemAddressType := reflect.PtrTo(LoopTypeIndirect(ctx.Result.Records.GetFieldType(mField)).Elem())
			subRecords := MakeRecordsWithElem(preload.SubModel, elemAddressType)
			// reset sub model relation field

			for _, record := range ctx.Result.Records.GetRecords() {
				if primary := record.Field(mainField); primary.IsValid() {
					rField := LoopIndirect(record.Field(mField))
					for subi := 0; subi < rField.Len(); subi++ {
						subRecord := subRecords.Add(rField.Index(subi).Addr())
						subRecord.SetField(subField, primary)
					}
				} else {
					return errors.New("some records have not primary")
				}
			}
			subCtx := preloadBrick.GetContext(option, subRecords)
			ctx.Result.Preload[mField] = subCtx.Result
			if err := subCtx.Next(); err != nil {
				return err
			}
		}
		// many to many
		for mField, preload := range ctx.Brick.ManyToManyPreload {
			subBrick := ctx.Brick.Preload(mField)
			middleBrick := NewToyBrick(ctx.Brick.Toy, preload.MiddleModel).CopyStatus(ctx.Brick)

			mainField, subField := preload.Model.GetOnePrimary(), preload.SubModel.GetOnePrimary()
			middleMainField, middleSubField :=
				GetMiddleField(preload.Model, preload.MiddleModel, preload.IsRight),
				GetMiddleField(preload.SubModel, preload.MiddleModel, !preload.IsRight)
			elemAddressType := reflect.PtrTo(LoopTypeIndirect(ctx.Result.Records.GetFieldType(mField)).Elem())
			subRecords := MakeRecordsWithElem(preload.SubModel, elemAddressType)

			for _, record := range ctx.Result.Records.GetRecords() {
				rField := LoopIndirect(record.Field(mField))
				for subi := 0; subi < rField.Len(); subi++ {
					subRecords.Add(rField.Index(subi).Addr())
				}
			}
			subCtx := subBrick.GetContext(option, subRecords)
			ctx.Result.Preload[mField] = subCtx.Result
			if err := subCtx.Next(); err != nil {
				return err
			}

			middleRecords := MakeRecordsWithElem(middleBrick.model, middleBrick.model.ReflectType)
			// use to calculate what sub records belong for
			offset := 0
			for _, record := range ctx.Result.Records.GetRecords() {
				primary := record.Field(mainField)
				primary.IsValid()
				if primary.IsValid() == false {
					return errors.New("some records have not primary")
				}
				rField := LoopIndirect(record.Field(mField))
				for subi := 0; subi < rField.Len(); subi++ {
					subRecord := subRecords.GetRecord(subi + offset)
					subPrimary := subRecord.Field(subField)
					if subPrimary.IsValid() == false {
						return errors.New("some records have not primary")
					}
					middleValue := reflect.New(middleBrick.model.ReflectType).Elem()
					middleValue.Field(middleBrick.model.FieldsPos[middleMainField]).Set(primary)
					middleValue.Field(middleBrick.model.FieldsPos[middleSubField]).Set(subPrimary)
					middleRecords.Add(middleValue)
				}
				offset += rField.Len()
			}
			middleCtx := middleBrick.GetContext(option, middleRecords)
			ctx.Result.MiddleModelPreload[mField] = middleCtx.Result
			if err := middleCtx.Next(); err != nil {
				return err
			}
		}
		return nil
	}
}

func HandlerInsertTimeGenerate(ctx *Context) error {
	records := ctx.Result.Records
	if field, ok := ctx.Brick.model.NameFields["CreatedAt"]; ok {
		current := time.Now()
		for _, record := range records.GetRecords() {
			record.SetField(field, reflect.ValueOf(current))
		}
	}
	if mField, ok := ctx.Brick.model.NameFields["UpdatedAt"]; ok {
		current := time.Now()
		for _, record := range records.GetRecords() {
			record.SetField(mField, reflect.ValueOf(current))
		}
	}
	return nil
}

func HandlerInsert(ctx *Context) error {
	// current insert

	for i, record := range ctx.Result.Records.GetRecords() {
		action := ExecAction{}
		action.Exec = ctx.Brick.InsertExec(record)
		action.Result, action.Error = ctx.Brick.Exec(action.Exec.Query, action.Exec.Args...)
		if action.Error == nil {
			// set primary field value if model is autoincrement
			if len(ctx.Brick.model.GetPrimary()) == 1 && ctx.Brick.model.GetOnePrimary().AutoIncrement == true {
				if lastId, err := action.Result.LastInsertId(); err == nil {
					ctx.Result.Records.GetRecord(i).SetField(ctx.Brick.model.GetOnePrimary(), reflect.ValueOf(lastId))
				} else {
					return errors.New(fmt.Sprintf("get (%s) auto increment  failure reason(%s)", ctx.Brick.model.Name, err))
				}
			}
		}
		ctx.Result.AddExecRecord(action, i)
	}
	return nil
}

func HandlerFind(ctx *Context) error {
	action := QueryAction{
		Exec: ctx.Brick.FindExec(ctx.Result.Records),
	}
	rows, err := ctx.Brick.Query(action.Exec.Query, action.Exec.Args...)
	if err != nil {
		return err
	}
	// find current data
	min := ctx.Result.Records.Len()
	for rows.Next() {
		elem := reflect.New(ctx.Result.Records.ElemType()).Elem()
		ctx.Result.Records.Len()
		record := ctx.Result.Records.Add(elem)
		var scanners []interface{}
		for _, field := range ctx.Brick.getScanFields(ctx.Result.Records) {
			value := record.Field(field)
			scanners = append(scanners, value.Addr().Interface())
		}
		err := rows.Scan(scanners...)
		action.Error = append(action.Error, err)
	}
	max := ctx.Result.Records.Len()
	ctx.Result.AddQueryRecord(action, makeRange(min, max)...)
	return nil
}

func HandlerPreloadFind(ctx *Context) error {
	records := ctx.Result.Records
	for mField, preload := range ctx.Brick.OneToOnePreload {
		var mainField, subField *ModelField
		// select fields from subtable where ... and subtable.id = table.subtableID
		// select fields from subtable where ... and subtable.tableID = table.ID
		if preload.IsBelongTo {
			mainField, subField = preload.RelationField, preload.SubModel.GetOnePrimary()
		} else {
			mainField, subField = preload.Model.GetOnePrimary(), preload.RelationField
		}
		brick := ctx.Brick.MapPreloadBrick[mField]

		keys := reflect.New(reflect.SliceOf(records.GetFieldType(mainField))).Elem()
		for _, record := range records.GetRecords() {
			keys = SafeAppend(keys, record.Field(mainField))
		}
		// the relation condition should have lowest priority
		brick = brick.Where(ExprIn, subField, keys.Interface()).And().Conditions(brick.Search)
		containerList := reflect.New(reflect.SliceOf(records.GetFieldType(mField))).Elem()
		//var preloadRecords ModelRecords
		subCtx, err := brick.find(LoopIndirectAndNew(containerList))
		ctx.Result.Preload[mField] = subCtx.Result
		if err != nil {
			return err
		}
		// use to map preload model relation field
		fieldMapKeyType := LoopTypeIndirect(subCtx.Result.Records.GetFieldType(subField))
		fieldMapType := reflect.MapOf(fieldMapKeyType, records.GetFieldType(mField))
		fieldMap := reflect.MakeMap(fieldMapType)
		for i, pRecord := range subCtx.Result.Records.GetRecords() {
			fieldMapKey := LoopIndirect(pRecord.Field(subField))
			if fieldMapKey.IsValid() {
				fieldMap.SetMapIndex(fieldMapKey, containerList.Index(i))
			}
		}
		for _, record := range records.GetRecords() {
			key := record.Field(mainField)
			if preloadMatchValue := fieldMap.MapIndex(key); preloadMatchValue.IsValid() {
				record.Field(mField).Set(preloadMatchValue)
			}
		}
	}
	// one to many
	for mField, preload := range ctx.Brick.OneToManyPreload {
		mainField, subField := preload.Model.GetOnePrimary(), preload.RelationField
		brick := ctx.Brick.MapPreloadBrick[mField]

		keys := reflect.New(reflect.SliceOf(records.GetFieldType(mainField))).Elem()
		for _, fieldValue := range records.GetRecords() {
			keys = SafeAppend(keys, fieldValue.Field(mainField))
		}
		// the relation condition should have lowest priority
		brick = brick.Where(ExprIn, subField, keys.Interface()).And().Conditions(brick.Search)
		containerList := reflect.New(records.GetFieldType(mField)).Elem()
		//var preloadRecords ModelRecords
		subCtx, err := brick.find(LoopIndirectAndNew(containerList))
		ctx.Result.Preload[mField] = subCtx.Result
		if err != nil {
			return err
		}
		// fieldMap:  map[submodel.id]->submodel
		fieldMapKeyType := LoopTypeIndirect(subCtx.Result.Records.GetFieldType(subField))
		fieldMapValueType := records.GetFieldType(mField)
		fieldMapType := reflect.MapOf(fieldMapKeyType, fieldMapValueType)
		fieldMap := reflect.MakeMap(fieldMapType)
		for i, pRecord := range subCtx.Result.Records.GetRecords() {
			fieldMapKey := LoopIndirect(pRecord.Field(subField))
			if fieldMapKey.IsValid() {
				currentListField := fieldMap.MapIndex(fieldMapKey)
				if currentListField.IsValid() == false {
					currentListField = reflect.MakeSlice(fieldMapValueType, 0, 1)
				}
				fieldMap.SetMapIndex(fieldMapKey, SafeAppend(currentListField, containerList.Index(i)))
			}
		}
		for _, record := range records.GetRecords() {
			key := record.Field(mainField)
			if preloadMatchValue := fieldMap.MapIndex(key); preloadMatchValue.IsValid() {
				record.Field(preload.ContainerField).Set(preloadMatchValue)
			}
		}
	}
	// many to many
	for mField, preload := range ctx.Brick.ManyToManyPreload {
		mainPrimary, subPrimary := preload.Model.GetOnePrimary(), preload.SubModel.GetOnePrimary()
		middlePrimaryLeft, middlePrimaryRight :=
			GetMiddleField(preload.Model, preload.MiddleModel, preload.IsRight),
			GetMiddleField(preload.SubModel, preload.MiddleModel, !preload.IsRight)
		middleBrick := NewToyBrick(ctx.Brick.Toy, preload.MiddleModel).CopyStatus(ctx.Brick)

		// primaryMap: map[model.id]->the model's ModelRecord
		primaryMap := map[interface{}]ModelRecord{}
		keys := reflect.New(reflect.SliceOf(middlePrimaryLeft.Field.Type)).Elem()
		for _, record := range records.GetRecords() {
			keys = SafeAppend(keys, record.Field(mainPrimary))
			primaryMap[record.Field(mainPrimary).Interface()] = record
		}
		// the relation condition should have lowest priority
		middleBrick = middleBrick.Where(ExprIn, middlePrimaryLeft, keys.Interface()).And().Conditions(middleBrick.Search)
		middleModelElemList := reflect.New(reflect.SliceOf(preload.MiddleModel.ReflectType)).Elem()
		//var middleModelRecords ModelRecords
		middleCtx, err := middleBrick.find(middleModelElemList)
		ctx.Result.MiddleModelPreload[mField] = middleCtx.Result
		if err != nil {
			return err
		}
		// middle model records
		if middleCtx.Result.Records.Len() == 0 {
			continue
		}
		// primaryMap

		// subPrimaryMap:  map[submodel.id]->[]the model's ModelRecord
		subPrimaryMap := map[interface{}][]ModelRecord{}
		middlePrimary2Keys := reflect.New(reflect.SliceOf(middlePrimaryRight.Field.Type)).Elem()
		for _, pRecord := range middleCtx.Result.Records.GetRecords() {
			subPrimaryMapKey := LoopIndirect(pRecord.Field(middlePrimaryRight))
			subPrimaryMapValue := LoopIndirect(pRecord.Field(middlePrimaryLeft))
			subPrimaryMap[subPrimaryMapKey.Interface()] =
				append(subPrimaryMap[subPrimaryMapKey.Interface()], primaryMap[subPrimaryMapValue.Interface()])
		}
		for key, _ := range subPrimaryMap {
			middlePrimary2Keys = reflect.Append(middlePrimary2Keys, reflect.ValueOf(key))
		}
		brick := ctx.Brick.MapPreloadBrick[mField]
		// the relation condition should have lowest priority
		brick = brick.Where(ExprIn, subPrimary, middlePrimary2Keys.Interface()).And().Conditions(brick.Search)
		containerField := reflect.New(records.GetFieldType(mField)).Elem()
		//var subRecords ModelRecords
		subCtx, err := brick.find(LoopIndirectAndNew(containerField))
		ctx.Result.Preload[mField] = subCtx.Result
		if err != nil {
			return err
		}
		for i, subRecord := range subCtx.Result.Records.GetRecords() {
			records := subPrimaryMap[subRecord.Field(subPrimary).Interface()]
			for _, record := range records {
				record.SetField(mField, reflect.Append(record.Field(mField), containerField.Index(i)))
			}
		}
	}
	return nil
}

func HandlerUpdateTimeGenerate(ctx *Context) error {
	records := ctx.Result.Records
	if mField, ok := ctx.Brick.model.NameFields["UpdatedAt"]; ok {
		current := reflect.ValueOf(time.Now())
		for _, record := range records.GetRecords() {
			record.SetField(mField, current)
		}
	}
	return nil
}

func HandlerUpdate(ctx *Context) error {
	for i, record := range ctx.Result.Records.GetRecords() {
		action := ExecAction{Exec: ctx.Brick.UpdateExec(record)}
		action.Result, action.Error = ctx.Brick.Exec(action.Exec.Query, action.Exec.Args...)
		ctx.Result.AddExecRecord(action, i)
	}
	return nil
}

// if have not primary ,try to insert
// else try to replace
func HandlerSave(ctx *Context) error {
	for i, record := range ctx.Result.Records.GetRecords() {
		primaryFields := ctx.Brick.model.GetPrimary()
		var tryInsert bool
		for _, primaryField := range primaryFields {
			pkeyFieldValue := record.Field(primaryField)
			if pkeyFieldValue.IsValid() == false || IsZero(pkeyFieldValue) {
				tryInsert = true
				break
			}
		}
		var action ExecAction
		if tryInsert {
			action = ExecAction{}
			action.Exec = ctx.Brick.InsertExec(record)
			action.Result, action.Error = ctx.Brick.Exec(action.Exec.Query, action.Exec.Args...)
			if action.Error == nil {
				// set primary field value if model is autoincrement
				if len(ctx.Brick.model.GetPrimary()) == 1 && ctx.Brick.model.GetOnePrimary().AutoIncrement == true {
					if lastId, err := action.Result.LastInsertId(); err == nil {
						ctx.Result.Records.GetRecord(i).SetField(ctx.Brick.model.GetOnePrimary(), reflect.ValueOf(lastId))
					} else {
						return errors.New(fmt.Sprintf("get (%s) auto increment  failure reason(%s)", ctx.Brick.model.Name, err))
					}
				}
			}
		} else {
			action = ExecAction{}
			action.Exec = ctx.Brick.ReplaceExec(record)
			action.Result, action.Error = ctx.Brick.Exec(action.Exec.Query, action.Exec.Args...)
		}
		ctx.Result.AddExecRecord(action, i)
	}
	return nil
}

func HandlerSaveTimeProcess(ctx *Context) error {
	createField := ctx.Brick.model.NameFields["CreatedAt"]
	now := reflect.ValueOf(time.Now())
	if ctx.Result.Records.Len() > 0 && createField != nil {
		primaryField := ctx.Brick.model.GetOnePrimary()
		brick := ctx.Brick.BindFields(primaryField, createField)
		primaryKeys := reflect.MakeSlice(reflect.SliceOf(primaryField.Field.Type), 0, ctx.Result.Records.Len())
		action := QueryAction{}
		var tryFindTimeIndex []int

		for i, record := range ctx.Result.Records.GetRecords() {
			pri := record.Field(primaryField)
			if pri.IsValid() && IsZero(pri) == false {
				primaryKeys = reflect.Append(primaryKeys, pri)
			}
			tryFindTimeIndex = append(tryFindTimeIndex, i)
		}

		action.Exec = brick.Where(ExprIn, primaryField, primaryKeys.Interface()).FindExec(ctx.Result.Records)

		rows, err := brick.Query(action.Exec.Query, action.Exec.Args...)
		if err != nil {
			action.Error = append(action.Error, err)
			ctx.Result.AddQueryRecord(action, tryFindTimeIndex...)
			return nil
		}
		primaryKeysMap := reflect.MakeMap(reflect.MapOf(primaryField.Field.Type, createField.Field.Type))

		// find all createtime
		for rows.Next() {
			id := reflect.New(primaryField.Field.Type)
			createAt := reflect.New(createField.Field.Type)
			err := rows.Scan(id.Interface(), createAt.Interface())
			if err != nil {
				action.Error = append(action.Error, err)
			}
			primaryKeysMap.SetMapIndex(id.Elem(), createAt.Elem())
		}
		for _, record := range ctx.Result.Records.GetRecords() {
			pri := record.Field(primaryField)
			if createAt := primaryKeysMap.MapIndex(pri); createAt.IsValid() && IsZero(createAt) == false {
				record.SetField(createField, createAt)
			} else {
				record.SetField(createField, now)
			}
		}
		ctx.Result.AddQueryRecord(action, tryFindTimeIndex...)
	}
	if mField, ok := ctx.Brick.model.NameFields["UpdatedAt"]; ok {
		for _, record := range ctx.Result.Records.GetRecords() {
			record.SetField(mField, now)
		}
	}
	return nil
}

func HandlerNotRecordPreload(option string) func(ctx *Context) error {
	return func(ctx *Context) (err error) {
		for mField, _ := range ctx.Brick.OneToOnePreload {
			brick := ctx.Brick.MapPreloadBrick[mField]
			subCtx := brick.GetContext(option, MakeRecordsWithElem(brick.model, brick.model.ReflectType))
			ctx.Result.Preload[mField] = subCtx.Result
			if err := subCtx.Next(); err != nil {
				return err
			}
		}
		for mField, _ := range ctx.Brick.OneToManyPreload {
			brick := ctx.Brick.MapPreloadBrick[mField]
			subCtx := brick.GetContext(option, MakeRecordsWithElem(brick.model, brick.model.ReflectType))
			ctx.Result.Preload[mField] = subCtx.Result
			if err := subCtx.Next(); err != nil {
				return err
			}
		}
		for mField, preload := range ctx.Brick.ManyToManyPreload {
			{
				brick := ctx.Brick.MapPreloadBrick[mField]
				subCtx := brick.GetContext(option, MakeRecordsWithElem(brick.model, brick.model.ReflectType))
				ctx.Result.Preload[mField] = subCtx.Result
				if err := subCtx.Next(); err != nil {
					return err
				}
			}
			// process middle model
			{
				middleModel := preload.MiddleModel
				brick := NewToyBrick(ctx.Brick.Toy, middleModel).CopyStatus(ctx.Brick)
				middleCtx := brick.GetContext(option, MakeRecordsWithElem(brick.model, brick.model.ReflectType))
				ctx.Result.MiddleModelPreload[mField] = middleCtx.Result
				if err := middleCtx.Next(); err != nil {
					return err
				}
			}
		}
		return nil
	}
}

func HandlerCreateTable(ctx *Context) error {
	execs := ctx.Brick.CreateTableExec(ctx.Brick.Toy.Dialect)
	for _, exec := range execs {
		action := ExecAction{Exec: exec}
		action.Result, action.Error = ctx.Brick.Exec(exec.Query, exec.Args...)
		ctx.Result.AddExecRecord(action)
	}
	return nil
}

func HandlerExistTableAbort(ctx *Context) error {
	action := QueryAction{}
	action.Exec = ctx.Brick.HasTableExec(ctx.Brick.Toy.Dialect)
	var hasTable bool
	err := ctx.Brick.QueryRow(action.Exec.Query, action.Exec.Args...).Scan(&hasTable)
	if err != nil {
		action.Error = append(action.Error, err)
	}
	ctx.Result.AddQueryRecord(action)
	if err != nil || hasTable == true {
		ctx.Abort()
	}

	return nil
}

func HandlerDropTable(ctx *Context) (err error) {
	exec := ctx.Brick.DropTableExec()
	action := ExecAction{Exec: exec}
	action.Result, action.Error = ctx.Brick.Exec(exec.Query, exec.Args...)
	ctx.Result.AddExecRecord(action)
	return nil
}

func HandlerNotExistTableAbort(ctx *Context) error {
	action := QueryAction{}
	action.Exec = ctx.Brick.HasTableExec(ctx.Brick.Toy.Dialect)
	var hasTable bool
	err := ctx.Brick.QueryRow(action.Exec.Query, action.Exec.Args...).Scan(&hasTable)
	if err != nil {
		action.Error = append(action.Error, err)
	}
	ctx.Result.AddQueryRecord(action)
	if err != nil || hasTable == false {
		ctx.Abort()
	}
	return nil
}

func HandlerPreloadDelete(ctx *Context) error {
	for mField, preload := range ctx.Brick.OneToOnePreload {
		if preload.IsBelongTo == false {
			preloadBrick := ctx.Brick.Preload(mField)
			subRecords := MakeRecordsWithElem(preload.SubModel, ctx.Result.Records.GetFieldAddressType(mField))
			mainSoftDelete := preload.Model.GetFieldWithName("DeletedAt") != nil
			subSoftDelete := preload.SubModel.GetFieldWithName("DeletedAt") != nil
			// set sub model relation field
			for _, record := range ctx.Result.Records.GetRecords() {
				// it means relation field, result[j].LastInsertId() is id value
				subRecords.Add(record.FieldAddress(mField))
			}
			// if main model is hard delete need set relationship field set zero if sub model is soft delete
			if mainSoftDelete == false && subSoftDelete == true {
				deletedAtField := preloadBrick.model.GetFieldWithName("DeletedAt")
				preloadBrick = preloadBrick.BindFields(preload.RelationField, deletedAtField)
			}
			result, err := preloadBrick.deleteWithPrimaryKey(subRecords)
			ctx.Result.Preload[mField] = result
			if err != nil {
				return err
			}
		}
	}

	// one to many
	for mField, preload := range ctx.Brick.OneToManyPreload {
		preloadBrick := ctx.Brick.Preload(mField)
		mainSoftDelete := preload.Model.GetFieldWithName("DeletedAt") != nil
		subSoftDelete := preload.SubModel.GetFieldWithName("DeletedAt") != nil
		elemAddressType := reflect.PtrTo(LoopTypeIndirect(ctx.Result.Records.GetFieldType(mField)).Elem())
		subRecords := MakeRecordsWithElem(preload.SubModel, elemAddressType)
		for _, record := range ctx.Result.Records.GetRecords() {
			rField := LoopIndirect(record.Field(mField))
			for subi := 0; subi < rField.Len(); subi++ {
				subRecords.Add(rField.Index(subi).Addr())
			}
		}
		// model relationship field set zero
		if mainSoftDelete == false && subSoftDelete == true {
			deletedAtField := preloadBrick.model.GetFieldWithName("DeletedAt")
			preloadBrick = preloadBrick.BindFields(preload.RelationField, deletedAtField)
		}
		result, err := preloadBrick.deleteWithPrimaryKey(subRecords)
		ctx.Result.Preload[mField] = result
		if err != nil {
			return err
		}
	}
	// many to many
	for mField, preload := range ctx.Brick.ManyToManyPreload {
		subBrick := ctx.Brick.Preload(mField)
		middleBrick := NewToyBrick(ctx.Brick.Toy, preload.MiddleModel).CopyStatus(ctx.Brick)
		mainField, subField := preload.Model.GetOnePrimary(), preload.SubModel.GetOnePrimary()
		middleMainField, middleSubField :=
			GetMiddleField(preload.Model, preload.MiddleModel, preload.IsRight),
			GetMiddleField(preload.SubModel, preload.MiddleModel, !preload.IsRight)
		mainSoftDelete := preload.Model.GetFieldWithName("DeletedAt") != nil
		subSoftDelete := preload.SubModel.GetFieldWithName("DeletedAt") != nil

		elemAddressType := reflect.PtrTo(LoopTypeIndirect(ctx.Result.Records.GetFieldType(mField)).Elem())
		subRecords := MakeRecordsWithElem(preload.SubModel, elemAddressType)

		for _, record := range ctx.Result.Records.GetRecords() {
			rField := LoopIndirect(record.Field(mField))
			for subi := 0; subi < rField.Len(); subi++ {
				subRecords.Add(rField.Index(subi).Addr())
			}
		}

		middleRecords := MakeRecordsWithElem(middleBrick.model, middleBrick.model.ReflectType)
		// use to calculate what sub records belong for
		offset := 0
		for _, record := range ctx.Result.Records.GetRecords() {
			primary := record.Field(mainField)
			primary.IsValid()
			if primary.IsValid() == false {
				return errors.New("some records have not primary key")
			}
			rField := LoopIndirect(record.Field(mField))
			for subi := 0; subi < rField.Len(); subi++ {
				subRecord := subRecords.GetRecord(subi + offset)
				subPrimary := subRecord.Field(subField)
				if subPrimary.IsValid() == false {
					return errors.New("some records have not primary key")
				}
				middleValue := reflect.New(middleBrick.model.ReflectType).Elem()
				middleValue.Field(middleBrick.model.FieldsPos[middleMainField]).Set(primary)
				middleValue.Field(middleBrick.model.FieldsPos[middleSubField]).Set(subPrimary)
				middleRecords.Add(middleValue)
			}
			offset += rField.Len()
		}
		conditions := middleBrick.Search
		middleBrick = middleBrick.Conditions(nil)

		// delete middle model data
		var primaryFields []*ModelField
		if mainSoftDelete == false {
			primaryFields = append(primaryFields, middleBrick.model.PrimaryFields[0])
		}
		if subSoftDelete == false {
			primaryFields = append(primaryFields, middleBrick.model.PrimaryFields[1])
		}
		if len(primaryFields) != 0 {
			for _, primaryField := range primaryFields {
				primarySetType := reflect.MapOf(primaryField.Field.Type, reflect.TypeOf(struct{}{}))
				primarySet := reflect.MakeMap(primarySetType)
				for _, record := range middleRecords.GetRecords() {
					primarySet.SetMapIndex(record.Field(primaryField), reflect.ValueOf(struct{}{}))
				}
				var primaryKeys = reflect.New(reflect.SliceOf(primaryField.Field.Type)).Elem()
				for _, k := range primarySet.MapKeys() {
					primaryKeys = reflect.Append(primaryKeys, k)
				}
				middleBrick = middleBrick.Where(ExprIn, primaryField, primaryKeys.Interface()).
					Or().Conditions(middleBrick.Search)
			}
			middleBrick = middleBrick.And().Conditions(conditions)
			result, err := middleBrick.delete(middleRecords)
			ctx.Result.MiddleModelPreload[mField] = result
			if err != nil {
				return err
			}
		}

		result, err := subBrick.deleteWithPrimaryKey(subRecords)
		ctx.Result.Preload[mField] = result
		if err != nil {
			return err
		}
	}
	if err := ctx.Next(); err != nil {
		return err
	}

	for mField, preload := range ctx.Brick.OneToOnePreload {
		if preload.IsBelongTo == true {
			preloadBrick := ctx.Brick.Preload(mField)
			subRecords := MakeRecordsWithElem(preload.SubModel, ctx.Result.Records.GetFieldAddressType(mField))
			for _, record := range ctx.Result.Records.GetRecords() {
				subRecords.Add(record.FieldAddress(mField))
			}

			mainSoftDelete := preload.Model.GetFieldWithName("DeletedAt") != nil
			subSoftDelete := preload.SubModel.GetFieldWithName("DeletedAt") != nil
			if mainSoftDelete == false && subSoftDelete == true {
				deletedAtField := preloadBrick.model.GetFieldWithName("DeletedAt")
				preloadBrick = preloadBrick.BindFields(preload.RelationField, deletedAtField)
			}

			result, err := preloadBrick.deleteWithPrimaryKey(subRecords)
			ctx.Result.Preload[mField] = result
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func HandlerHardDelete(ctx *Context) error {
	action := ExecAction{}
	action.Exec = ctx.Brick.DeleteExec()
	action.Result, action.Error = ctx.Brick.Exec(action.Exec.Query, action.Exec.Args...)
	ctx.Result.AddExecRecord(action)
	return nil
}

//
func HandlerSoftDeleteCheck(ctx *Context) error {
	mField := ctx.Brick.model.GetFieldWithName("DeletedAt")
	if mField != nil {
		ctx.Brick = ctx.Brick.Where(ExprNull, mField).And().Conditions(ctx.Brick.Search)
	}
	return nil
}

func HandlerSoftDelete(ctx *Context) error {
	action := ExecAction{}
	now := time.Now()
	value := reflect.New(ctx.Brick.model.ReflectType).Elem()
	record := NewStructRecord(ctx.Brick.model, value)
	record.SetField(ctx.Brick.model.GetFieldWithName("DeletedAt"), reflect.ValueOf(now))

	action.Exec = ctx.Brick.UpdateExec(record)
	action.Result, action.Error = ctx.Brick.Exec(action.Exec.Query, action.Exec.Args...)
	ctx.Result.AddExecRecord(action)
	return nil
}
