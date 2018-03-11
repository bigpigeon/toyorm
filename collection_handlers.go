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

			middleRecords := MakeRecordsWithElem(middleBrick.model, middleBrick.model.ReflectType)
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
					middleRecord := NewRecord(middleBrick.model, reflect.New(middleBrick.model.ReflectType).Elem())
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
	createField := ctx.Brick.model.GetFieldWithName("CreatedAt")
	updateField := ctx.Brick.model.GetFieldWithName("UpdatedAt")
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
	if ctx.dbIndex != -1 {
		return nil
	}
	primaryKeyField := ctx.Brick.model.GetOnePrimary()
	var getDBIndex func(r ModelRecord) int
	notPtrElemType := LoopTypeIndirect(ctx.Result.Records.ElemType())
	if _, ok := reflect.Zero(reflect.PtrTo(notPtrElemType)).Interface().(DBValSelector); ok {
		getDBIndex = func(r ModelRecord) int {
			iface := r.Source().Addr().Interface().(DBValSelector)
			return iface.Select(len(ctx.Brick.Toy.dbs))
		}
	} else if selector := ctx.Brick.selector; selector != nil {
		getDBIndex = func(r ModelRecord) int {
			return selector(r.Field(primaryKeyField.Name()).Interface(), len(ctx.Brick.Toy.dbs))
		}
	} else {
		return ErrCollectionDBSelectorNotFound{}
	}
	dbRecordsMap := map[int]ModelRecords{}
	// dbRecordsMap dbIndexMap[i]  means the position in Records
	dbIndexMap := make([][]int, len(ctx.Brick.Toy.dbs))
	for i, record := range ctx.Result.Records.GetRecords() {
		dbIndex := getDBIndex(record)
		if dbRecordsMap[dbIndex] == nil {
			dbRecordsMap[dbIndex] = MakeRecordsWithElem(ctx.Brick.model, record.Source().Addr().Type())
		}
		dbRecordsMap[dbIndex].Add(record.Source().Addr())
		dbIndexMap[dbIndex] = append(dbIndexMap[dbIndex], i)
	}
	for i, records := range dbRecordsMap {
		dbCtx := NewCollectionContext(ctx.handlers[ctx.index+1:], ctx.Brick, records)
		dbCtx.dbIndex = i
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
	if ctx.dbIndex == -1 {
		return ErrDbIndexNotSet{}
	}
	for i, record := range ctx.Result.Records.GetRecords() {
		action := CollectionExecAction{affectData: []int{i}, dbIndex: ctx.dbIndex}
		action.Exec = ctx.Brick.InsertExec(record)
		action.Result, action.Error = ctx.Brick.Exec(action.Exec, action.dbIndex)
		if action.Error == nil {
			// set primary field value if model is autoincrement
			if len(ctx.Brick.model.GetPrimary()) == 1 && ctx.Brick.model.GetOnePrimary().AutoIncrement() == true {
				if lastId, err := action.Result.LastInsertId(); err == nil {
					ctx.Result.Records.GetRecord(i).SetField(ctx.Brick.model.GetOnePrimary().Name(), reflect.ValueOf(lastId))
				} else {
					return errors.New(fmt.Sprintf("get (%s) auto increment  failure reason(%s)", ctx.Brick.model.Name, err))
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
			subCtx := brick.GetContext(option, MakeRecordsWithElem(brick.model, brick.model.ReflectType))
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
			subCtx := brick.GetContext(option, MakeRecordsWithElem(brick.model, brick.model.ReflectType))
			ctx.Result.Preload[fieldName] = subCtx.Result
			if err := subCtx.Next(); err != nil {
				return err
			}
		}

		for fieldName := range ctx.Brick.OneToManyPreload {
			brick := ctx.Brick.MapPreloadBrick[fieldName]
			subCtx := brick.GetContext(option, MakeRecordsWithElem(brick.model, brick.model.ReflectType))
			ctx.Result.Preload[fieldName] = subCtx.Result
			if err := subCtx.Next(); err != nil {
				return err
			}
		}
		for fieldName, preload := range ctx.Brick.ManyToManyPreload {
			{
				brick := ctx.Brick.MapPreloadBrick[fieldName]
				subCtx := brick.GetContext(option, MakeRecordsWithElem(brick.model, brick.model.ReflectType))
				ctx.Result.Preload[fieldName] = subCtx.Result
				if err := subCtx.Next(); err != nil {
					return err
				}
			}
			// process middle model
			{
				middleModel := preload.MiddleModel
				brick := NewCollectionBrick(ctx.Brick.Toy, middleModel).CopyStatus(ctx.Brick)
				middleCtx := brick.GetContext(option, MakeRecordsWithElem(brick.model, brick.model.ReflectType))
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
			subCtx := brick.GetContext(option, MakeRecordsWithElem(brick.model, brick.model.ReflectType))
			ctx.Result.Preload[fieldName] = subCtx.Result
			if err := subCtx.Next(); err != nil {
				return err
			}

		}
		for fieldName := range ctx.Brick.OneToManyPreload {
			brick := ctx.Brick.MapPreloadBrick[fieldName]
			subCtx := brick.GetContext(option, MakeRecordsWithElem(brick.model, brick.model.ReflectType))
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
				middleCtx := brick.GetContext(option, MakeRecordsWithElem(brick.model, brick.model.ReflectType))
				ctx.Result.MiddleModelPreload[fieldName] = middleCtx.Result
				if err := middleCtx.Next(); err != nil {
					return err
				}
			}
			// process sub model
			{
				brick := ctx.Brick.MapPreloadBrick[fieldName]
				subCtx := brick.GetContext(option, MakeRecordsWithElem(brick.model, brick.model.ReflectType))
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
			subCtx := brick.GetContext(option, MakeRecordsWithElem(brick.model, brick.model.ReflectType))
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
	if ctx.dbIndex != -1 {
		return nil
	}
	for i := range ctx.Brick.Toy.dbs {
		dbCtx := NewCollectionContext(ctx.handlers[ctx.index+1:], ctx.Brick, ctx.Result.Records)
		dbCtx.dbIndex = i
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
	if ctx.dbIndex == -1 {
		return ErrDbIndexNotSet{}
	}
	execs := ctx.Brick.Toy.Dialect.CreateTable(ctx.Brick.model)
	for _, exec := range execs {
		action := CollectionExecAction{Exec: exec, dbIndex: ctx.dbIndex}
		action.Result, action.Error = ctx.Brick.Exec(exec, action.dbIndex)
		ctx.Result.AddRecord(action)
	}
	return nil
}

func CollectionHandlerExistTableAbort(ctx *CollectionContext) error {
	if ctx.dbIndex == -1 {
		return ErrDbIndexNotSet{}
	}
	action := CollectionQueryAction{dbIndex: ctx.dbIndex}
	action.Exec = ctx.Brick.Toy.Dialect.HasTable(ctx.Brick.model)
	var hasTable bool
	err := ctx.Brick.QueryRow(action.Exec, ctx.dbIndex).Scan(&hasTable)
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
	if ctx.dbIndex == -1 {
		return ErrDbIndexNotSet{}
	}
	exec := ctx.Brick.Toy.Dialect.DropTable(ctx.Brick.model)
	action := CollectionExecAction{Exec: exec, dbIndex: ctx.dbIndex}
	action.Result, action.Error = ctx.Brick.Exec(exec, action.dbIndex)
	ctx.Result.AddRecord(action)
	return nil
}

func CollectionHandlerNotExistTableAbort(ctx *CollectionContext) error {
	if ctx.dbIndex == -1 {
		return ErrDbIndexNotSet{}
	}
	action := CollectionQueryAction{}
	action.Exec = ctx.Brick.Toy.Dialect.HasTable(ctx.Brick.model)
	var hasTable bool
	err := ctx.Brick.QueryRow(action.Exec, ctx.dbIndex).Scan(&hasTable)
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
		primaryName := ctx.Brick.model.GetOnePrimary().Name()
		if primaryType := ctx.Result.Records.GetFieldType(primaryName); primaryType == nil {
			return errors.New(fmt.Sprintf("struct missing %s field", primaryName))
		}
	}
	return nil
}

func CollectionHandlerSaveTimeGenerate(ctx *CollectionContext) error {
	if ctx.dbIndex == -1 {
		return ErrDbIndexNotSet{}
	}
	createdAtField := ctx.Brick.model.GetFieldWithName("CreatedAt")
	deletedAtField := ctx.Brick.model.GetFieldWithName("DeletedAt")
	now := reflect.ValueOf(time.Now())

	var timeFields []Field
	var defaultFieldValue []reflect.Value
	if createdAtField != nil {
		timeFields = append(timeFields, createdAtField)
		defaultFieldValue = append(defaultFieldValue, now)
	}
	if deletedAtField != nil {
		timeFields = append(timeFields, deletedAtField)
		defaultFieldValue = append(defaultFieldValue, reflect.Zero(deletedAtField.StructField().Type))
	}

	if ctx.Result.Records.Len() > 0 && len(timeFields) > 0 {
		primaryField := ctx.Brick.model.GetOnePrimary()
		brick := ctx.Brick.bindFields(ModeDefault, append([]Field{primaryField}, timeFields...)...)
		primaryKeys := reflect.MakeSlice(reflect.SliceOf(primaryField.StructField().Type), 0, ctx.Result.Records.Len())
		action := CollectionQueryAction{dbIndex: ctx.dbIndex}

		for i, record := range ctx.Result.Records.GetRecords() {
			pri := record.Field(primaryField.Name())
			if pri.IsValid() && IsZero(pri) == false {
				primaryKeys = reflect.Append(primaryKeys, pri)
				action.affectData = append(action.affectData, i)
			}
		}

		if primaryKeys.Len() > 0 {
			action.Exec = brick.Where(ExprIn, primaryField, primaryKeys.Interface()).FindExec(ctx.Result.Records)

			rows, err := brick.Query(action.Exec, action.dbIndex)
			if err != nil {
				action.Error = append(action.Error, err)

				ctx.Result.AddRecord(action)
				return nil
			}
			defer rows.Close()
			var mapElemTypeFields []reflect.StructField
			{
				for _, f := range timeFields {
					mapElemTypeFields = append(mapElemTypeFields, f.StructField())
				}
			}
			mapElemType := reflect.StructOf(mapElemTypeFields)
			primaryKeysMap := reflect.MakeMap(reflect.MapOf(primaryField.StructField().Type, mapElemType))

			// find all createtime
			for rows.Next() {
				id := reflect.New(primaryField.StructField().Type)
				timeFieldValues := reflect.New(mapElemType).Elem()
				scaners := []interface{}{id.Interface()}
				for i := 0; i < timeFieldValues.NumField(); i++ {
					scaners = append(scaners, timeFieldValues.Field(i).Addr().Interface())
				}
				err := rows.Scan(scaners...)
				if err != nil {
					action.Error = append(action.Error, err)
				}
				primaryKeysMap.SetMapIndex(id.Elem(), timeFieldValues)
			}

			ctx.Result.AddRecord(action)
			for _, record := range ctx.Result.Records.GetRecords() {
				pri := record.Field(primaryField.Name())
				fields := primaryKeysMap.MapIndex(pri)
				if fields.IsValid() {
					for i := 0; i < fields.NumField(); i++ {
						field := fields.Field(i)
						if field.IsValid() && IsZero(field) == false {
							record.SetField(timeFields[i].Name(), field)
						} else if IsZero(record.Field(timeFields[i].Name())) {
							record.SetField(timeFields[i].Name(), defaultFieldValue[i])
						}
					}
				} else {
					for i := 0; i < len(timeFields); i++ {
						if IsZero(record.Field(timeFields[i].Name())) {
							record.SetField(timeFields[i].Name(), defaultFieldValue[i])
						}
					}
				}
			}
		} else {
			for _, record := range ctx.Result.Records.GetRecords() {
				for i := 0; i < len(timeFields); i++ {
					if IsZero(record.Field(timeFields[i].Name())) {
						record.SetField(timeFields[i].Name(), defaultFieldValue[i])
					}
				}
			}
		}
	}
	if updateField := ctx.Brick.model.GetFieldWithName("UpdatedAt"); updateField != nil {
		for _, record := range ctx.Result.Records.GetRecords() {
			record.SetField(updateField.Name(), now)
		}
	}
	return nil
}

func CollectionHandlerSave(ctx *CollectionContext) error {
	if ctx.dbIndex == -1 {
		return ErrDbIndexNotSet{}
	}
	for i, record := range ctx.Result.Records.GetRecords() {
		primaryFields := ctx.Brick.model.GetPrimary()
		var tryInsert bool
		for _, primaryField := range primaryFields {
			pkeyFieldValue := record.Field(primaryField.Name())
			if pkeyFieldValue.IsValid() == false || IsZero(pkeyFieldValue) {
				tryInsert = true
				break
			}
		}
		action := CollectionExecAction{affectData: []int{i}, dbIndex: ctx.dbIndex}
		if tryInsert {

			action.Exec = ctx.Brick.InsertExec(record)
			action.Result, action.Error = ctx.Brick.Exec(action.Exec, action.dbIndex)
			if action.Error == nil {
				// set primary field value if model is autoincrement
				if len(ctx.Brick.model.GetPrimary()) == 1 && ctx.Brick.model.GetOnePrimary().AutoIncrement() == true {
					if lastId, err := action.Result.LastInsertId(); err == nil {
						ctx.Result.Records.GetRecord(i).SetField(ctx.Brick.model.GetOnePrimary().Name(), reflect.ValueOf(lastId))
					} else {
						return errors.New(fmt.Sprintf("get (%s) auto increment  failure reason(%s)", ctx.Brick.model.Name, err))
					}
				}
			}
		} else {
			action.Exec = ctx.Brick.ReplaceExec(record)
			action.Result, action.Error = ctx.Brick.Exec(action.Exec, action.dbIndex)
		}
		ctx.Result.AddRecord(action)
	}
	return nil
}

func CollectionHandlerPreloadFind(ctx *CollectionContext) error {
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

func CollectionHandlerFindAssignDbIndex(ctx *CollectionContext) error {
	for i := range ctx.Brick.Toy.dbs {
		dbCtx := NewCollectionContext(ctx.handlers[ctx.index+1:], ctx.Brick, ctx.Result.Records)
		dbCtx.dbIndex = i
		err := dbCtx.Next()
		if err != nil {
			return err
		}
		for _, action := range dbCtx.Result.ActionFlow {
			ctx.Result.AddRecord(action)
		}
	}
	ctx.Abort()
	return nil
}

func CollectionHandlerFind(ctx *CollectionContext) error {
	if ctx.dbIndex == -1 {
		return ErrDbIndexNotSet{}
	}
	action := CollectionQueryAction{
		Exec:    ctx.Brick.FindExec(ctx.Result.Records),
		dbIndex: ctx.dbIndex,
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
	max := ctx.Result.Records.Len()
	action.affectData = makeRange(min, max)
	ctx.Result.AddRecord(action)
	return nil
}

func CollectionHandlerFindOneAssignDbIndex(ctx *CollectionContext) error {
	for i := range ctx.Brick.Toy.dbs {
		currentLen := ctx.Result.Records.Len()
		dbCtx := NewCollectionContext(ctx.handlers[ctx.index+1:], ctx.Brick, ctx.Result.Records)
		dbCtx.dbIndex = i
		err := dbCtx.Next()
		if err != nil {
			return err
		}
		for _, action := range dbCtx.Result.ActionFlow {
			ctx.Result.AddRecord(action)
		}
		if ctx.Result.Records.Len() > currentLen {
			break
		}
	}
	ctx.Abort()
	return nil
}

func CollectionHandlerFindOne(ctx *CollectionContext) error {
	if ctx.dbIndex == -1 {
		return ErrDbIndexNotSet{}
	}
	action := CollectionQueryAction{
		Exec:    ctx.Brick.FindExec(ctx.Result.Records),
		dbIndex: ctx.dbIndex,
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
	deletedField := ctx.Brick.model.GetFieldWithName("DeletedAt")
	if deletedField != nil {
		ctx.Brick = ctx.Brick.Where(ExprNull, deletedField).And().Conditions(ctx.Brick.Search)
	}
	return nil
}

func CollectionHandlerUpdateTimeGenerate(ctx *CollectionContext) error {
	records := ctx.Result.Records
	if updateField := ctx.Brick.model.GetFieldWithName("UpdatedAt"); updateField != nil {
		current := reflect.ValueOf(time.Now())
		for _, record := range records.GetRecords() {
			record.SetField(updateField.Name(), current)
		}
	}
	return nil
}

func CollectionHandlerUpdateAssignDbIndex(ctx *CollectionContext) error {
	for i := range ctx.Brick.Toy.dbs {
		dbCtx := NewCollectionContext(ctx.handlers[ctx.index+1:], ctx.Brick, ctx.Result.Records)
		dbCtx.dbIndex = i
		err := dbCtx.Next()
		if err != nil {
			return err
		}
		for _, action := range dbCtx.Result.ActionFlow {
			ctx.Result.AddRecord(action)
		}
	}
	ctx.Abort()
	return nil
}

func CollectionHandlerUpdate(ctx *CollectionContext) error {
	if ctx.dbIndex == -1 {
		return ErrDbIndexNotSet{}
	}
	for i, record := range ctx.Result.Records.GetRecords() {
		action := CollectionExecAction{Exec: ctx.Brick.UpdateExec(record), affectData: []int{i}, dbIndex: ctx.dbIndex}
		action.Result, action.Error = ctx.Brick.Exec(action.Exec, action.dbIndex)
		ctx.Result.AddRecord(action)
	}
	return nil
}
