/*
 * Copyright 2018. bigpigeon. All rights reserved.
 * Use of this source code is governed by a MIT style
 * license that can be found in the LICENSE file.
 */

package toyorm

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
)

type PreCollectionBrick struct {
	Parent *CollectionBrick
	Field  Field
}

type CollectionBrick struct {
	Toy             *ToyCollection
	preBrick        PreCollectionBrick
	MapPreloadBrick map[string]*CollectionBrick

	//Error             []error
	debug bool
	//tx    *sql.Tx

	//orderBy []Column
	Search SearchList
	//offset int
	//limit  int
	//groupBy []Column
	template *BasicExec

	selector DBPrimarySelector
	dbIndex  int
	BrickCommon
}

func NewCollectionBrick(toy *ToyCollection, model *Model) *CollectionBrick {
	return &CollectionBrick{
		Toy: toy,

		MapPreloadBrick: map[string]*CollectionBrick{},
		selector:        dbPrimaryKeySelector,
		dbIndex:         -1,
		BrickCommon: BrickCommon{
			Model:             model,
			BelongToPreload:   map[string]*BelongToPreload{},
			OneToOnePreload:   map[string]*OneToOnePreload{},
			OneToManyPreload:  map[string]*OneToManyPreload{},
			ManyToManyPreload: map[string]*ManyToManyPreload{},
			ignoreModeSelector: [ModeEnd]IgnoreMode{
				ModeInsert:    IgnoreNo,
				ModeSave:      IgnoreNo,
				ModeUpdate:    IgnoreZero,
				ModeCondition: IgnoreZero,
				ModePreload:   IgnoreZero,
			},
		},
	}
}

func (t *CollectionBrick) And() CollectionBrickAnd {
	return CollectionBrickAnd{t}
}

func (t *CollectionBrick) Or() CollectionBrickOr {
	return CollectionBrickOr{t}
}

func (t *CollectionBrick) Clone() *CollectionBrick {
	newt := &CollectionBrick{
		Toy: t.Toy,
	}
	return newt
}

func (t *CollectionBrick) Scope(fn func(*CollectionBrick) *CollectionBrick) *CollectionBrick {
	ret := fn(t)
	return ret
}

func (t *CollectionBrick) CopyStatus(statusBrick *CollectionBrick) *CollectionBrick {
	newt := *t
	//newt.tx = statusBrick.tx
	newt.debug = statusBrick.debug
	newt.ignoreModeSelector = t.ignoreModeSelector

	return &newt
}

// return it parent CollectionBrick
// it will panic when the parent CollectionBrick is nil
func (t *CollectionBrick) Enter() *CollectionBrick {
	return t.Scope(func(t *CollectionBrick) *CollectionBrick {
		newt := *t.preBrick.Parent
		newt.MapPreloadBrick = map[string]*CollectionBrick{}
		for k, v := range t.preBrick.Parent.MapPreloadBrick {
			newt.MapPreloadBrick[k] = v
		}
		t.preBrick.Parent = &newt
		newt.MapPreloadBrick[t.preBrick.Field.Name()] = t
		return t.preBrick.Parent
	})
}

// this module is get preload which is right middle field name in many-to-many mode
// it only use for sub model type is same with main model type
// e.g
// User{
//     ID int `toyorm:"primary key"`
//     Friend []User
// }
// now the main model middle field name is L_UserID, sub model middle field name is R_UserID
// if you want to get preload with main model middle field name == R_UserID use RightValuePreload
func (t *CollectionBrick) RightValuePreload(fv interface{}) *CollectionBrick {
	return t.Scope(func(t *CollectionBrick) *CollectionBrick {
		field := t.Model.fieldSelect(fv)

		subModel := t.Toy.GetModel(LoopDiveSliceAndPtr(field.FieldValue()))
		newSubt := NewCollectionBrick(t.Toy, subModel).CopyStatus(t)

		newt := *t
		newt.MapPreloadBrick = t.CopyMapPreloadBrick()
		newt.MapPreloadBrick[field.Name()] = newSubt
		newSubt.preBrick = PreCollectionBrick{&newt, field}
		if preload := newt.Toy.ManyToManyPreload(newt.Model, field, true); preload != nil {
			newt.ManyToManyPreload = t.CopyManyToManyPreload()
			newt.ManyToManyPreload[field.Name()] = preload
		} else {
			panic(ErrInvalidPreloadField{t.Model.ReflectType.Name(), field.Name()})
		}
		return newSubt
	})
}

// return
func (t *CollectionBrick) Preload(fv interface{}) *CollectionBrick {
	return t.Scope(func(t *CollectionBrick) *CollectionBrick {
		field := t.Model.fieldSelect(fv)
		//if subBrick, ok := t.MapPreloadBrick[field.Name()]; ok {
		//	return subBrick
		//}
		subModel := t.Toy.GetModel(LoopDiveSliceAndPtr(field.FieldValue()))
		newSubt := NewCollectionBrick(t.Toy, subModel).CopyStatus(t)

		newt := *t
		newt.MapPreloadBrick = t.CopyMapPreloadBrick()
		newt.MapPreloadBrick[field.Name()] = newSubt
		newSubt.preBrick = PreCollectionBrick{&newt, field}
		if preload := newt.Toy.BelongToPreload(newt.Model, field); preload != nil {
			newt.BelongToPreload = t.CopyBelongToPreload()
			newt.BelongToPreload[field.Name()] = preload
		} else if preload := newt.Toy.OneToOnePreload(newt.Model, field); preload != nil {
			newt.OneToOnePreload = t.CopyOneToOnePreload()
			newt.OneToOnePreload[field.Name()] = preload
		} else if preload := newt.Toy.OneToManyPreload(newt.Model, field); preload != nil {
			newt.OneToManyPreload = t.CopyOneToManyPreload()
			for k, v := range t.OneToManyPreload {
				newt.OneToManyPreload[k] = v
			}
			newt.OneToManyPreload[field.Name()] = preload
		} else if preload := newt.Toy.ManyToManyPreload(newt.Model, field, false); preload != nil {
			newt.ManyToManyPreload = t.CopyManyToManyPreload()
			newt.ManyToManyPreload[field.Name()] = preload
		} else {
			panic(ErrInvalidPreloadField{t.Model.ReflectType.Name(), field.Name()})
		}
		return newSubt
	})
}

func (t *CollectionBrick) CopyMapPreloadBrick() map[string]*CollectionBrick {
	preloadBrick := map[string]*CollectionBrick{}
	for k, v := range t.MapPreloadBrick {
		preloadBrick[k] = v
	}
	return preloadBrick
}

func (t *CollectionBrick) CustomOneToOnePreload(container, relationship interface{}, args ...interface{}) *CollectionBrick {
	containerField := t.Model.fieldSelect(container)
	var subModel *Model
	if len(args) > 0 {
		subModel = t.Toy.GetModel(LoopDivePtr(reflect.ValueOf(args[0])))
	} else {
		subModel = t.Toy.GetModel(LoopDivePtr(containerField.FieldValue()))
	}
	relationshipField := subModel.fieldSelect(relationship)
	preload := t.Toy.OneToOneBind(t.Model, subModel, containerField, relationshipField)
	if preload == nil {
		panic(ErrInvalidPreloadField{t.Model.ReflectType.Name(), containerField.Name()})
	}

	newSubt := NewCollectionBrick(t.Toy, subModel).CopyStatus(t)
	newt := *t
	newt.MapPreloadBrick = t.CopyMapPreloadBrick()
	newt.MapPreloadBrick[containerField.Name()] = newSubt
	newSubt.preBrick = PreCollectionBrick{&newt, containerField}

	newt.OneToOnePreload = t.CopyOneToOnePreload()
	newt.OneToOnePreload[containerField.Name()] = preload
	return newSubt
}

func (t *CollectionBrick) CustomBelongToPreload(container, relationship interface{}, args ...interface{}) *CollectionBrick {
	containerField, relationshipField := t.Model.fieldSelect(container), t.Model.fieldSelect(relationship)
	var subModel *Model
	if len(args) > 0 {
		subModel = t.Toy.GetModel(LoopDivePtr(reflect.ValueOf(args[0])))
	} else {
		subModel = t.Toy.GetModel(LoopDivePtr(containerField.FieldValue()))
	}
	preload := t.Toy.BelongToBind(t.Model, subModel, containerField, relationshipField)
	if preload == nil {
		panic(ErrInvalidPreloadField{t.Model.ReflectType.Name(), containerField.Name()})
	}

	newSubt := NewCollectionBrick(t.Toy, subModel).CopyStatus(t)
	newt := *t
	newt.MapPreloadBrick = t.CopyMapPreloadBrick()
	newt.MapPreloadBrick[containerField.Name()] = newSubt
	newSubt.preBrick = PreCollectionBrick{&newt, containerField}

	newt.BelongToPreload = t.CopyBelongToPreload()
	newt.BelongToPreload[containerField.Name()] = preload
	return newSubt
}

func (t *CollectionBrick) CustomOneToManyPreload(container, relationship interface{}, args ...interface{}) *CollectionBrick {
	containerField := t.Model.fieldSelect(container)
	var subModel *Model
	if len(args) > 0 {
		subModel = t.Toy.GetModel(LoopDiveSliceAndPtr(reflect.ValueOf(args[0])))
	} else {
		subModel = t.Toy.GetModel(LoopDiveSliceAndPtr(containerField.FieldValue()))
	}
	relationshipField := subModel.fieldSelect(relationship)
	preload := t.Toy.OneToManyBind(t.Model, subModel, containerField, relationshipField)
	if preload == nil {
		panic(ErrInvalidPreloadField{t.Model.ReflectType.Name(), containerField.Name()})
	}

	newSubt := NewCollectionBrick(t.Toy, subModel).CopyStatus(t)
	newt := *t
	newt.MapPreloadBrick = t.CopyMapPreloadBrick()
	newt.MapPreloadBrick[containerField.Name()] = newSubt
	newSubt.preBrick = PreCollectionBrick{&newt, containerField}

	newt.OneToManyPreload = t.CopyOneToManyPreload()
	newt.OneToManyPreload[containerField.Name()] = preload
	return newSubt
}

func (t *CollectionBrick) CustomManyToManyPreload(middleStruct, container, relation, subRelation interface{}, args ...interface{}) *CollectionBrick {
	containerField := t.Model.fieldSelect(container)
	var subModel *Model
	if len(args) > 0 {
		subModel = t.Toy.GetModel(LoopDiveSliceAndPtr(reflect.ValueOf(args[0])))
	} else {
		subModel = t.Toy.GetModel(LoopDiveSliceAndPtr(containerField.FieldValue()))
	}
	middleModel := t.Toy.GetModel(LoopDiveSliceAndPtr(reflect.ValueOf(middleStruct)))
	relationField, subRelationField := middleModel.fieldSelect(relation), middleModel.fieldSelect(subRelation)
	preload := t.Toy.ManyToManyPreloadBind(t.Model, subModel, middleModel, containerField, relationField, subRelationField)
	if preload == nil {
		panic(ErrInvalidPreloadField{t.Model.ReflectType.Name(), containerField.Name()})
	}

	newSubt := NewCollectionBrick(t.Toy, subModel).CopyStatus(t)
	newt := *t
	newt.MapPreloadBrick = t.CopyMapPreloadBrick()
	newt.MapPreloadBrick[containerField.Name()] = newSubt
	newSubt.preBrick = PreCollectionBrick{&newt, containerField}

	newt.ManyToManyPreload = t.CopyManyToManyPreload()
	newt.ManyToManyPreload[containerField.Name()] = preload
	return newSubt
}

func (t *CollectionBrick) BindFields(mode Mode, args ...interface{}) *CollectionBrick {
	return t.Scope(func(t *CollectionBrick) *CollectionBrick {
		var fields []Field
		for _, v := range args {
			fields = append(fields, t.Model.fieldSelect(v))
		}
		newt := *t

		newt.FieldsSelector[mode] = fields
		return &newt
	})
}

func (t *CollectionBrick) BindDefaultFields(args ...interface{}) *CollectionBrick {
	return t.Scope(func(t *CollectionBrick) *CollectionBrick {
		var fields []Field
		for _, v := range args {
			fields = append(fields, t.Model.fieldSelect(v))
		}
		return t.bindDefaultFields(fields...)
	})
}

func (t *CollectionBrick) bindDefaultFields(fields ...Field) *CollectionBrick {
	return t.Scope(func(t *CollectionBrick) *CollectionBrick {
		newt := *t
		newt.FieldsSelector[ModeDefault] = fields
		return &newt
	})
}

func (t *CollectionBrick) bindFields(mode Mode, fields ...Field) *CollectionBrick {
	return t.Scope(func(t *CollectionBrick) *CollectionBrick {
		newt := *t

		newt.FieldsSelector[mode] = fields
		return &newt
	})
}

func (t *CollectionBrick) Debug() *CollectionBrick {
	return t.Scope(func(t *CollectionBrick) *CollectionBrick {
		newt := *t
		newt.debug = true
		return &newt
	})
}

func (t *CollectionBrick) Selector(selector DBPrimarySelector) *CollectionBrick {
	return t.Scope(func(t *CollectionBrick) *CollectionBrick {
		newt := *t
		newt.selector = selector
		return &newt
	})
}

func (t *CollectionBrick) DBIndex(i int) *CollectionBrick {
	return t.Scope(func(t *CollectionBrick) *CollectionBrick {
		newt := *t
		newt.dbIndex = i
		return &newt
	})
}

func (t *CollectionBrick) Template(temp string, args ...interface{}) *CollectionBrick {
	return t.Scope(func(t *CollectionBrick) *CollectionBrick {
		newt := *t
		if temp == "" && len(args) == 0 {
			newt.template = nil
		} else {
			newt.template = &BasicExec{temp, args}
		}
		return &newt
	})
}

func (t *CollectionBrick) IgnoreMode(s Mode, ignore IgnoreMode) *CollectionBrick {
	newt := *t
	newt.ignoreModeSelector[s] = ignore
	return &newt
}

func (t *CollectionBrick) GetContext(option string, records ModelRecords) *CollectionContext {
	handlers := t.Toy.ModelHandlers(option, t.Model)
	ctx := NewCollectionContext(handlers, t, records)
	//ctx.Next()
	return ctx
}

func (t *CollectionBrick) insert(records ModelRecords) (*Result, error) {
	handlers := t.Toy.ModelHandlers("Insert", t.Model)
	ctx := NewCollectionContext(handlers, t, records)
	return ctx.Result, ctx.Next()
}

func (t *CollectionBrick) save(records ModelRecords) (*Result, error) {
	handlers := t.Toy.ModelHandlers("Save", t.Model)
	ctx := NewCollectionContext(handlers, t, records)
	return ctx.Result, ctx.Next()
}

func (t *CollectionBrick) usave(records ModelRecords) (*Result, error) {
	handlers := t.Toy.ModelHandlers("USave", t.Model)
	ctx := NewCollectionContext(handlers, t, records)
	return ctx.Result, ctx.Next()
}

func (t *CollectionBrick) deleteWithPrimaryKey(records ModelRecords) (*Result, error) {
	if field := t.Model.GetFieldWithName("DeletedAt"); field != nil {
		return t.softDeleteWithPrimaryKey(records)
	} else {
		return t.hardDeleteWithPrimaryKey(records)
	}
}

func (t *CollectionBrick) softDeleteWithPrimaryKey(records ModelRecords) (*Result, error) {
	handlers := t.Toy.ModelHandlers("SoftDeleteWithPrimaryKey", t.Model)
	ctx := NewCollectionContext(handlers, t, records)
	return ctx.Result, ctx.Next()
}

func (t *CollectionBrick) hardDeleteWithPrimaryKey(records ModelRecords) (*Result, error) {
	handlers := t.Toy.ModelHandlers("HardDeleteWithPrimaryKey", t.Model)
	ctx := NewCollectionContext(handlers, t, records)
	return ctx.Result, ctx.Next()
}

func (t *CollectionBrick) delete(records ModelRecords) (*Result, error) {
	if field := t.Model.GetFieldWithName("DeletedAt"); field != nil {
		return t.softDelete(records)
	} else {
		return t.hardDelete(records)
	}
}

func (t *CollectionBrick) softDelete(records ModelRecords) (*Result, error) {
	handlers := t.Toy.ModelHandlers("SoftDelete", t.Model)
	ctx := NewCollectionContext(handlers, t, records)
	return ctx.Result, ctx.Next()
}

func (t *CollectionBrick) hardDelete(records ModelRecords) (*Result, error) {
	handlers := t.Toy.ModelHandlers("HardDelete", t.Model)
	ctx := NewCollectionContext(handlers, t, records)
	return ctx.Result, ctx.Next()
}

func (t *CollectionBrick) find(value reflect.Value) (*CollectionContext, error) {
	if value.Kind() == reflect.Slice {
		records := NewRecords(t.Model, value)
		ctx := NewCollectionContext(t.Toy.ModelHandlers("Find", t.Model), t, records)
		return ctx, ctx.Next()
	} else {
		vList := reflect.New(reflect.SliceOf(value.Type())).Elem()
		records := NewRecords(t.Model, vList)
		ctx := NewCollectionContext(t.Toy.ModelHandlers("FindOne", t.Model), t, records)
		err := ctx.Next()
		if vList.Len() == 0 {
			if err == nil {
				err = sql.ErrNoRows
			}
			return ctx, err
		}
		value.Set(vList.Index(0))
		return ctx, err
	}
}

func (t *CollectionBrick) CreateTable() (*Result, error) {
	ctx := t.GetContext("CreateTable", MakeRecordsWithElem(t.Model, t.Model.ReflectType))
	return ctx.Result, ctx.Next()
}

func (t *CollectionBrick) CreateTableIfNotExist() (*Result, error) {
	ctx := t.GetContext("CreateTableIfNotExist", MakeRecordsWithElem(t.Model, t.Model.ReflectType))
	return ctx.Result, ctx.Next()
}

func (t *CollectionBrick) DropTable() (*Result, error) {
	ctx := t.GetContext("DropTable", MakeRecordsWithElem(t.Model, t.Model.ReflectType))
	return ctx.Result, ctx.Next()
}

func (t *CollectionBrick) DropTableIfExist() (*Result, error) {
	ctx := t.GetContext("DropTableIfExist", MakeRecordsWithElem(t.Model, t.Model.ReflectType))
	return ctx.Result, ctx.Next()
}

func (t *CollectionBrick) HasTable() ([]bool, error) {
	set := make([]bool, len(t.Toy.dbs))
	exec := t.Toy.Dialect.HasTable(t.Model)
	errs := ErrCollectionQueryRow{}
	for i := range t.Toy.dbs {
		err := t.QueryRow(exec, i).Scan(&set[i])
		if err != nil {
			errs[i] = err
		}
	}
	if len(errs) != 0 {
		return set, errs
	}
	return set, nil
}

func (t *CollectionBrick) Count() (count int, err error) {
	exec := t.CountExec()
	countCount := 0
	errs := ErrCollectionQueryRow{}
	for i := range t.Toy.dbs {
		var count int
		err := t.QueryRow(exec, i).Scan(&count)
		if err != nil {
			errs[i] = err
		}
		countCount += count
	}
	if len(errs) != 0 {
		return countCount, errs
	}
	return countCount, nil
}

// insert can receive three type data
// struct
// map[offset]interface{}
// map[int]interface{}
// insert is difficult that have preload data
func (t *CollectionBrick) Insert(v interface{}) (*Result, error) {
	vValue := LoopIndirect(reflect.ValueOf(v))
	var records ModelRecords
	switch vValue.Kind() {
	case reflect.Slice:
		records = NewRecords(t.Model, vValue)
		return t.insert(records)
	default:
		var records ModelRecords
		if vValue.CanAddr() {
			records = MakeRecordsWithElem(t.Model, vValue.Addr().Type())
			records.Add(vValue.Addr())
		} else {
			records = MakeRecordsWithElem(t.Model, vValue.Type())
			records.Add(vValue)
		}
		return t.insert(records)
	}
}

func (t *CollectionBrick) Find(v interface{}) (*Result, error) {
	vValue := LoopIndirectAndNew(reflect.ValueOf(v))
	if vValue.CanSet() == false {
		return nil, errors.New("find value cannot be set")
	}
	ctx, err := t.find(vValue)
	return ctx.Result, err
}

func (t *CollectionBrick) Update(v interface{}) (*Result, error) {
	vValue := LoopIndirect(reflect.ValueOf(v))
	vValueList := reflect.MakeSlice(reflect.SliceOf(vValue.Type()), 0, 1)
	vValueList = reflect.Append(vValueList, vValue)
	handlers := t.Toy.ModelHandlers("Update", t.Model)
	ctx := NewCollectionContext(handlers, t, NewRecords(t.Model, vValueList))
	return ctx.Result, ctx.Next()
}

func (t *CollectionBrick) Save(v interface{}) (*Result, error) {
	vValue := LoopIndirect(reflect.ValueOf(v))

	switch vValue.Kind() {
	case reflect.Slice:
		records := NewRecords(t.Model, vValue)
		return t.save(records)
	default:
		records := MakeRecordsWithElem(t.Model, vValue.Addr().Type())
		records.Add(vValue.Addr())
		return t.save(records)
	}
}

func (t *CollectionBrick) USave(v interface{}) (*Result, error) {
	vValue := LoopIndirect(reflect.ValueOf(v))

	switch vValue.Kind() {
	case reflect.Slice:
		records := NewRecords(t.Model, vValue)
		return t.usave(records)
	default:
		records := MakeRecordsWithElem(t.Model, vValue.Addr().Type())
		records.Add(vValue.Addr())
		return t.usave(records)
	}
}

func (t *CollectionBrick) Delete(v interface{}) (*Result, error) {
	vValue := LoopIndirect(reflect.ValueOf(v))
	var records ModelRecords
	switch vValue.Kind() {
	case reflect.Slice:
		records = NewRecords(t.Model, vValue)
		return t.deleteWithPrimaryKey(records)
	default:
		records = MakeRecordsWithElem(t.Model, vValue.Addr().Type())
		records.Add(vValue.Addr())
		return t.deleteWithPrimaryKey(records)
	}
}

func (t *CollectionBrick) DeleteWithConditions() (*Result, error) {
	return t.delete(nil)
}

func (t *CollectionBrick) debugPrint(i int) func(ExecValue, error) {
	return func(exec ExecValue, err error) {
		if t.debug {
			if err != nil {
				fmt.Fprintf(t.Toy.Logger, "db[%d] query:%s, args:%s faiure reason %s\n", i, exec.Query(), exec.JsonArgs(), err)
			} else {
				fmt.Fprintf(t.Toy.Logger, "db[%d] query:%s, args:%s\n", i, exec.Query(), exec.JsonArgs())
			}
		}
	}

}

func (t *CollectionBrick) Exec(exec ExecValue, i int) (sql.Result, error) {
	query := exec.Query()
	result, err := t.Toy.dbs[i].Exec(query, exec.Args()...)
	t.debugPrint(i)(exec, err)

	return result, err
}

func (t *CollectionBrick) Query(exec ExecValue, i int) (*sql.Rows, error) {
	query := exec.Query()
	rows, err := t.Toy.dbs[i].Query(query, exec.Args()...)
	t.debugPrint(i)(exec, err)

	return rows, err
}

func (t *CollectionBrick) QueryRow(exec ExecValue, i int) *sql.Row {
	query := exec.Query()
	row := t.Toy.dbs[i].QueryRow(query, exec.Args()...)
	t.debugPrint(i)(exec, nil)
	return row
}

func (t *CollectionBrick) CountExec() (exec ExecValue) {
	exec = t.Toy.Dialect.CountExec(t.Model, "")
	cExec := t.ConditionExec()
	exec = exec.Append(" "+cExec.Source(), cExec.Args()...)
	return
}

func (t *CollectionBrick) ConditionExec() ExecValue {
	return t.Toy.Dialect.ConditionExec(t.Search, 0, 0, nil, nil)
}

func (t *CollectionBrick) FindExec(records ModelRecordFieldTypes) ExecValue {
	exec := t.Toy.Dialect.FindExec(t.Model, t.getSelectFields(records).ToColumnList(), "")

	cExec := t.ConditionExec()
	exec = exec.Append(" "+cExec.Source(), cExec.Args()...)
	return exec
}

func (t *CollectionBrick) UpdateExec(record ModelRecord) ExecValue {
	exec := t.Toy.Dialect.UpdateExec(t.Model, t.getFieldValuePairWithRecord(ModeUpdate, record).ToValueList())
	cExec := t.ConditionExec()
	exec = exec.Append(" "+cExec.Source(), cExec.Args()...)
	return exec
}

func (t *CollectionBrick) DeleteExec() ExecValue {
	exec := t.Toy.Dialect.DeleteExec(t.Model)
	cExec := t.ConditionExec()
	exec = exec.Append(" "+cExec.Source(), cExec.Args()...)
	return exec
}

func (t *CollectionBrick) InsertExec(record ModelRecord) (ExecValue, error) {
	recorders := t.getFieldValuePairWithRecord(ModeInsert, record)
	cExec := t.Toy.Dialect.ConditionBasicExec(t.Search, 0, 0, nil, nil)
	exec, err := t.Toy.Dialect.InsertExec(t.template, t.Model, recorders, cExec)
	if err != nil {
		return nil, err
	}
	return exec, nil
}

func (t *CollectionBrick) SaveExec(record ModelRecord) (ExecValue, error) {
	recorders := t.getFieldValuePairWithRecord(ModeSave, record)
	cExec := t.Toy.Dialect.ConditionBasicExec(t.Search, 0, 0, nil, nil)
	exec, err := t.Toy.Dialect.SaveExec(t.template, t.Model, recorders, cExec)
	if err != nil {
		return nil, err
	}
	return exec, nil
}

// collection not support table alias
func (t *CollectionBrick) getFieldValuePairWithRecord(mode Mode, record ModelRecord) FieldValueList {
	var fields []Field
	if len(t.FieldsSelector[mode]) > 0 {
		fields = t.FieldsSelector[mode]
	} else if len(t.FieldsSelector[ModeDefault]) > 0 {
		fields = t.FieldsSelector[ModeDefault]
	}

	var useIgnoreMode bool
	if len(fields) == 0 {
		fields = t.Model.GetSqlFields()
		useIgnoreMode = record.IsVariableContainer() == false
	}
	var columnValues FieldValueList
	if useIgnoreMode {
		for _, mField := range fields {
			if fieldValue := record.Field(mField.Name()); fieldValue.IsValid() {
				if t.ignoreModeSelector[mode].Ignore(fieldValue) == false {
					// move primary key check to dialect
					//if mField.IsPrimary() && IsZero(fieldValue) {
					//} else {
					//}
					columnValues = append(columnValues, mField.ToFieldValue(fieldValue))
				}
			}
		}
	} else {
		for _, mField := range fields {
			if fieldValue := record.Field(mField.Name()); fieldValue.IsValid() {
				columnValues = append(columnValues, mField.ToFieldValue(fieldValue))
			}
		}
	}
	return columnValues
}

func (t *BrickCommon) getSelectFields(records ModelRecordFieldTypes) FieldList {
	var fields []Field
	if len(t.FieldsSelector[ModeSelect]) > 0 {
		fields = t.FieldsSelector[ModeSelect]
	} else if len(t.FieldsSelector[ModeDefault]) > 0 {
		fields = t.FieldsSelector[ModeDefault]
	} else {
		fields = t.Model.GetSqlFields()
	}
	return getFieldsWithRecords(fields, records)
}
