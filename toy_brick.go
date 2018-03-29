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

type PreToyBrick struct {
	Parent *ToyBrick
	Field  Field
}

type ToyBrick struct {
	Toy             *Toy
	preBrick        PreToyBrick
	MapPreloadBrick map[string]*ToyBrick
	//Error             []error
	debug bool
	tx    *sql.Tx

	orderBy []Column
	Search  SearchList
	offset  int
	limit   int
	groupBy []Column

	BrickCommon
}

func NewToyBrick(toy *Toy, model *Model) *ToyBrick {
	return &ToyBrick{
		Toy:             toy,
		MapPreloadBrick: map[string]*ToyBrick{},

		BrickCommon: BrickCommon{
			Model:             model,
			BelongToPreload:   map[string]*BelongToPreload{},
			OneToOnePreload:   map[string]*OneToOnePreload{},
			OneToManyPreload:  map[string]*OneToManyPreload{},
			ManyToManyPreload: map[string]*ManyToManyPreload{},
			ignoreModeSelector: [ModeEnd]IgnoreMode{
				ModeInsert:    IgnoreNo,
				ModeReplace:   IgnoreNo,
				ModeUpdate:    IgnoreZero,
				ModeCondition: IgnoreZero,
				ModePreload:   IgnoreZero,
			},
		},
	}
}

func (t *ToyBrick) And() ToyBrickAnd {
	return ToyBrickAnd{t}
}

func (t *ToyBrick) Or() ToyBrickOr {
	return ToyBrickOr{t}
}

func (t *ToyBrick) Scope(fn func(*ToyBrick) *ToyBrick) *ToyBrick {
	ret := fn(t)
	return ret
}

func (t *ToyBrick) CopyStatus(statusBrick *ToyBrick) *ToyBrick {
	newt := *t
	newt.tx = statusBrick.tx
	newt.debug = statusBrick.debug
	newt.ignoreModeSelector = t.ignoreModeSelector

	return &newt
}

// return it parent ToyBrick
// it will panic when the parent ToyBrick is nil
func (t *ToyBrick) Enter() *ToyBrick {
	return t.Scope(func(t *ToyBrick) *ToyBrick {
		newt := *t.preBrick.Parent
		newt.MapPreloadBrick = map[string]*ToyBrick{}
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
func (t *ToyBrick) RightValuePreload(fv interface{}) *ToyBrick {
	return t.Scope(func(t *ToyBrick) *ToyBrick {
		field := t.Model.fieldSelect(fv)

		subModel := t.Toy.GetModel(LoopTypeIndirectSliceAndPtr(field.StructField().Type))
		newSubt := NewToyBrick(t.Toy, subModel).CopyStatus(t)

		newt := *t
		newt.MapPreloadBrick = t.CopyMapPreloadBrick()
		newt.MapPreloadBrick[field.Name()] = newSubt
		newSubt.preBrick = PreToyBrick{&newt, field}
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
func (t *ToyBrick) Preload(fv interface{}) *ToyBrick {
	return t.Scope(func(t *ToyBrick) *ToyBrick {
		field := t.Model.fieldSelect(fv)
		//if subBrick, ok := t.MapPreloadBrick[field.Name()]; ok {
		//	return subBrick
		//}
		subModel := t.Toy.GetModel(LoopTypeIndirectSliceAndPtr(field.StructField().Type))
		newSubt := NewToyBrick(t.Toy, subModel).CopyStatus(t)

		newt := *t
		newt.MapPreloadBrick = t.CopyMapPreloadBrick()
		newt.MapPreloadBrick[field.Name()] = newSubt
		newSubt.preBrick = PreToyBrick{&newt, field}
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

func (t *ToyBrick) CopyMapPreloadBrick() map[string]*ToyBrick {
	preloadBrick := map[string]*ToyBrick{}
	for k, v := range t.MapPreloadBrick {
		preloadBrick[k] = v
	}
	return preloadBrick
}

func (t *ToyBrick) CustomOneToOnePreload(container, relationship interface{}, args ...interface{}) *ToyBrick {
	containerField := t.Model.fieldSelect(container)
	var subModel *Model
	if len(args) > 0 {
		subModel = t.Toy.GetModel(LoopTypeIndirect(reflect.TypeOf(args[0])))
	} else {
		subModel = t.Toy.GetModel(LoopTypeIndirect(containerField.StructField().Type))
	}
	relationshipField := subModel.fieldSelect(relationship)
	preload := t.Toy.OneToOneBind(t.Model, subModel, containerField, relationshipField)
	if preload == nil {
		panic(ErrInvalidPreloadField{t.Model.ReflectType.Name(), containerField.Name()})
	}

	newSubt := NewToyBrick(t.Toy, subModel).CopyStatus(t)
	newt := *t
	newt.MapPreloadBrick = t.CopyMapPreloadBrick()
	newt.MapPreloadBrick[containerField.Name()] = newSubt
	newSubt.preBrick = PreToyBrick{&newt, containerField}

	newt.OneToOnePreload = t.CopyOneToOnePreload()
	newt.OneToOnePreload[containerField.Name()] = preload
	return newSubt
}

func (t *ToyBrick) CustomBelongToPreload(container, relationship interface{}, args ...interface{}) *ToyBrick {
	containerField, relationshipField := t.Model.fieldSelect(container), t.Model.fieldSelect(relationship)
	var subModel *Model
	if len(args) > 0 {
		subModel = t.Toy.GetModel(LoopTypeIndirect(reflect.TypeOf(args[0])))
	} else {
		subModel = t.Toy.GetModel(LoopTypeIndirect(containerField.StructField().Type))
	}
	preload := t.Toy.BelongToBind(t.Model, subModel, containerField, relationshipField)
	if preload == nil {
		panic(ErrInvalidPreloadField{t.Model.ReflectType.Name(), containerField.Name()})
	}

	newSubt := NewToyBrick(t.Toy, subModel).CopyStatus(t)
	newt := *t
	newt.MapPreloadBrick = t.CopyMapPreloadBrick()
	newt.MapPreloadBrick[containerField.Name()] = newSubt
	newSubt.preBrick = PreToyBrick{&newt, containerField}

	newt.BelongToPreload = t.CopyBelongToPreload()
	newt.BelongToPreload[containerField.Name()] = preload
	return newSubt
}

func (t *ToyBrick) CustomOneToManyPreload(container, relationship interface{}, args ...interface{}) *ToyBrick {
	containerField := t.Model.fieldSelect(container)
	var subModel *Model
	if len(args) > 0 {
		subModel = t.Toy.GetModel(LoopTypeIndirect(reflect.TypeOf(args[0])))
	} else {
		subModel = t.Toy.GetModel(LoopTypeIndirectSliceAndPtr(containerField.StructField().Type))
	}
	relationshipField := subModel.fieldSelect(relationship)
	preload := t.Toy.OneToManyBind(t.Model, subModel, containerField, relationshipField)
	if preload == nil {
		panic(ErrInvalidPreloadField{t.Model.ReflectType.Name(), containerField.Name()})
	}

	newSubt := NewToyBrick(t.Toy, subModel).CopyStatus(t)
	newt := *t
	newt.MapPreloadBrick = t.CopyMapPreloadBrick()
	newt.MapPreloadBrick[containerField.Name()] = newSubt
	newSubt.preBrick = PreToyBrick{&newt, containerField}

	newt.OneToManyPreload = t.CopyOneToManyPreload()
	newt.OneToManyPreload[containerField.Name()] = preload
	return newSubt
}

func (t *ToyBrick) CustomManyToManyPreload(middleStruct, container, relation, subRelation interface{}, args ...interface{}) *ToyBrick {
	containerField := t.Model.fieldSelect(container)
	var subModel *Model
	if len(args) > 0 {
		subModel = t.Toy.GetModel(LoopTypeIndirect(reflect.TypeOf(args[0])))
	} else {
		subModel = t.Toy.GetModel(LoopTypeIndirectSliceAndPtr(containerField.StructField().Type))
	}
	middleModel := t.Toy.GetModel(LoopTypeIndirect(reflect.TypeOf(middleStruct)))
	relationField, subRelationField := middleModel.fieldSelect(relation), middleModel.fieldSelect(subRelation)
	preload := t.Toy.ManyToManyPreloadBind(t.Model, subModel, middleModel, containerField, relationField, subRelationField)
	if preload == nil {
		panic(ErrInvalidPreloadField{t.Model.ReflectType.Name(), containerField.Name()})
	}

	newSubt := NewToyBrick(t.Toy, subModel).CopyStatus(t)
	newt := *t
	newt.MapPreloadBrick = t.CopyMapPreloadBrick()
	newt.MapPreloadBrick[containerField.Name()] = newSubt
	newSubt.preBrick = PreToyBrick{&newt, containerField}

	newt.ManyToManyPreload = t.CopyManyToManyPreload()
	newt.ManyToManyPreload[containerField.Name()] = preload
	return newSubt
}

func (t *ToyBrick) BindFields(mode Mode, args ...interface{}) *ToyBrick {
	return t.Scope(func(t *ToyBrick) *ToyBrick {
		var fields []Field
		for _, v := range args {
			fields = append(fields, t.Model.fieldSelect(v))
		}
		newt := *t

		newt.FieldsSelector[mode] = fields
		return &newt
	})
}
func (t *ToyBrick) BindDefaultFields(args ...interface{}) *ToyBrick {
	return t.Scope(func(t *ToyBrick) *ToyBrick {
		var fields []Field
		for _, v := range args {
			fields = append(fields, t.Model.fieldSelect(v))
		}
		return t.bindDefaultFields(fields...)
	})
}

func (t *ToyBrick) bindDefaultFields(fields ...Field) *ToyBrick {
	return t.Scope(func(t *ToyBrick) *ToyBrick {
		newt := *t
		newt.FieldsSelector[ModeDefault] = fields
		return &newt
	})
}

func (t *ToyBrick) bindFields(mode Mode, fields ...Field) *ToyBrick {
	return t.Scope(func(t *ToyBrick) *ToyBrick {
		newt := *t

		newt.FieldsSelector[mode] = fields
		return &newt
	})
}

func (t *ToyBrick) condition(expr SearchExpr, key interface{}, args ...interface{}) SearchList {
	search := SearchList{}
	switch expr {
	case ExprAnd, ExprOr:
		keyValue := LoopIndirect(reflect.ValueOf(key))
		record := NewRecord(t.Model, keyValue)
		pairs := t.getFieldValuePairWithRecord(ModeCondition, record)
		for _, pair := range pairs {
			search = search.Condition(pair, ExprEqual, expr)
		}
		// avoid "or" condition effected by priority
		if expr == ExprOr {
			search = append(search, NewSearchBranch(ExprIgnore))
		}
	default:
		var value reflect.Value
		if len(args) == 1 {
			value = reflect.ValueOf(args[0])
		} else {
			value = reflect.ValueOf(args)
		}
		mField := t.Model.fieldSelect(key)

		search = search.Condition(&modelFieldValue{mField, value}, expr, ExprAnd)
	}
	return search
}

// where will clean old condition
func (t *ToyBrick) Where(expr SearchExpr, key interface{}, v ...interface{}) *ToyBrick {
	return t.Scope(func(t *ToyBrick) *ToyBrick {
		newt := *t
		newt.Search = t.condition(expr, key, v...)
		return &newt
	})
}

func (t *ToyBrick) Conditions(search SearchList) *ToyBrick {
	return t.Scope(func(t *ToyBrick) *ToyBrick {
		if len(search) == 0 {
			newt := *t
			newt.Search = nil
			return &newt
		}
		newSearch := make(SearchList, len(search), len(search)+1)
		copy(newSearch, search)
		// to protect search priority
		newSearch = append(newSearch, NewSearchBranch(ExprIgnore))

		newt := *t
		newt.Search = newSearch
		return &newt
	})
}

func (t *ToyBrick) Limit(i int) *ToyBrick {
	return t.Scope(func(t *ToyBrick) *ToyBrick {
		newt := *t
		newt.limit = i
		return &newt
	})
}

func (t *ToyBrick) Offset(i int) *ToyBrick {
	return t.Scope(func(t *ToyBrick) *ToyBrick {
		newt := *t
		newt.offset = i
		return &newt
	})
}

func (t *ToyBrick) OrderBy(vList ...interface{}) *ToyBrick {
	return t.Scope(func(t *ToyBrick) *ToyBrick {
		newt := *t
		newt.orderBy = nil
		for _, v := range vList {
			if column, ok := v.(Column); ok {
				newt.orderBy = append(newt.orderBy, column)
			} else {
				newt.orderBy = append(newt.orderBy, t.Model.fieldSelect(v))
			}
		}
		return &newt
	})
}

func (t *ToyBrick) GroupBy(vList ...interface{}) *ToyBrick {
	return t.Scope(func(t *ToyBrick) *ToyBrick {
		newt := *t
		newt.groupBy = nil
		for _, v := range vList {
			field := newt.Model.fieldSelect(v)
			newt.groupBy = append(newt.groupBy, field)
		}
		return &newt
	})
}

func (t *ToyBrick) Begin() *ToyBrick {
	return t.Scope(func(t *ToyBrick) *ToyBrick {
		newt := *t
		tx, err := newt.Toy.db.Begin()
		if err != nil {
			panic(err)
		}
		newt.tx = tx
		return &newt
	})
}

func (t *ToyBrick) Commit() error {
	return t.tx.Commit()
}

func (t *ToyBrick) Rollback() error {
	return t.tx.Rollback()
}

func (t *ToyBrick) Debug() *ToyBrick {
	return t.Scope(func(t *ToyBrick) *ToyBrick {
		newt := *t
		newt.debug = true
		return &newt
	})
}

func (t *ToyBrick) IgnoreMode(s Mode, ignore IgnoreMode) *ToyBrick {
	newt := *t
	newt.ignoreModeSelector[s] = ignore
	return &newt
}

func (t *ToyBrick) GetContext(option string, records ModelRecords) *Context {
	handlers := t.Toy.ModelHandlers(option, t.Model)
	ctx := NewContext(handlers, t, records)
	//ctx.Next()
	return ctx
}

func (t *ToyBrick) insert(records ModelRecords) (*Result, error) {
	handlers := t.Toy.ModelHandlers("Insert", t.Model)
	ctx := NewContext(handlers, t, records)
	return ctx.Result, ctx.Next()
}

func (t *ToyBrick) save(records ModelRecords) (*Result, error) {
	handlers := t.Toy.ModelHandlers("Save", t.Model)
	ctx := NewContext(handlers, t, records)
	return ctx.Result, ctx.Next()
}

func (t *ToyBrick) deleteWithPrimaryKey(records ModelRecords) (*Result, error) {
	if field := t.Model.GetFieldWithName("DeletedAt"); field != nil {
		return t.softDeleteWithPrimaryKey(records)
	} else {
		return t.hardDeleteWithPrimaryKey(records)
	}
}

func (t *ToyBrick) softDeleteWithPrimaryKey(records ModelRecords) (*Result, error) {
	handlers := t.Toy.ModelHandlers("SoftDeleteWithPrimaryKey", t.Model)
	ctx := NewContext(handlers, t, records)
	return ctx.Result, ctx.Next()
}

func (t *ToyBrick) hardDeleteWithPrimaryKey(records ModelRecords) (*Result, error) {
	handlers := t.Toy.ModelHandlers("HardDeleteWithPrimaryKey", t.Model)
	ctx := NewContext(handlers, t, records)
	return ctx.Result, ctx.Next()
}

func (t *ToyBrick) delete(records ModelRecords) (*Result, error) {
	if field := t.Model.GetFieldWithName("DeletedAt"); field != nil {
		return t.softDelete(records)
	} else {
		return t.hardDelete(records)
	}
}

func (t *ToyBrick) softDelete(records ModelRecords) (*Result, error) {
	handlers := t.Toy.ModelHandlers("SoftDelete", t.Model)
	ctx := NewContext(handlers, t, records)
	return ctx.Result, ctx.Next()
}

func (t *ToyBrick) hardDelete(records ModelRecords) (*Result, error) {
	handlers := t.Toy.ModelHandlers("HardDelete", t.Model)
	ctx := NewContext(handlers, t, records)
	return ctx.Result, ctx.Next()
}

func (t *ToyBrick) find(value reflect.Value) (*Context, error) {
	handlers := t.Toy.ModelHandlers("Find", t.Model)
	if value.Kind() == reflect.Slice {
		records := NewRecords(t.Model, value)
		ctx := NewContext(handlers, t, records)
		return ctx, ctx.Next()
	} else {
		vList := reflect.New(reflect.SliceOf(value.Type())).Elem()
		records := NewRecords(t.Model, vList)
		ctx := NewContext(handlers, t.Limit(1), records)
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

func (t *ToyBrick) CreateTable() (*Result, error) {
	ctx := t.GetContext("CreateTable", MakeRecordsWithElem(t.Model, t.Model.ReflectType))
	return ctx.Result, ctx.Next()
}

func (t *ToyBrick) CreateTableIfNotExist() (*Result, error) {
	ctx := t.GetContext("CreateTableIfNotExist", MakeRecordsWithElem(t.Model, t.Model.ReflectType))
	return ctx.Result, ctx.Next()
}

func (t *ToyBrick) DropTable() (*Result, error) {
	ctx := t.GetContext("DropTable", MakeRecordsWithElem(t.Model, t.Model.ReflectType))
	return ctx.Result, ctx.Next()
}

func (t *ToyBrick) DropTableIfExist() (*Result, error) {
	ctx := t.GetContext("DropTableIfExist", MakeRecordsWithElem(t.Model, t.Model.ReflectType))
	return ctx.Result, ctx.Next()
}

func (t *ToyBrick) HasTable() (b bool, err error) {
	exec := t.Toy.Dialect.HasTable(t.Model)
	err = t.QueryRow(exec).Scan(&b)
	return b, err
}

func (t *ToyBrick) Count() (count int, err error) {
	exec := t.CountExec()
	err = t.QueryRow(exec).Scan(&count)
	return count, err
}

// insert can receive three type data
// struct
// map[offset]interface{}
// map[int]interface{}
// insert is difficult that have preload data
func (t *ToyBrick) Insert(v interface{}) (*Result, error) {
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

func (t *ToyBrick) Find(v interface{}) (*Result, error) {
	vValue := LoopIndirectAndNew(reflect.ValueOf(v))
	if vValue.CanSet() == false {
		return nil, errors.New("find value cannot be set")
	}
	ctx, err := t.find(vValue)
	return ctx.Result, err
}

func (t *ToyBrick) Update(v interface{}) (*Result, error) {
	vValue := LoopIndirect(reflect.ValueOf(v))
	vValueList := reflect.MakeSlice(reflect.SliceOf(vValue.Type()), 0, 1)
	vValueList = reflect.Append(vValueList, vValue)
	handlers := t.Toy.ModelHandlers("Update", t.Model)
	ctx := NewContext(handlers, t, NewRecords(t.Model, vValueList))
	return ctx.Result, ctx.Next()
}

func (t *ToyBrick) Save(v interface{}) (*Result, error) {
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

func (t *ToyBrick) Delete(v interface{}) (*Result, error) {
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

func (t *ToyBrick) DeleteWithConditions() (*Result, error) {
	return t.delete(nil)
}

func (t *ToyBrick) debugPrint(query string, args string, err error) {
	if t.debug {
		if err != nil {
			fmt.Fprintf(t.Toy.Logger, "use tx: %p, query:%s  args:%s faiure reason %s\n", t.tx, query, args, err)
		} else {
			fmt.Fprintf(t.Toy.Logger, "use tx: %p, query:%s  args:%s\n", t.tx, query, args)
		}
	}
}

func (t *ToyBrick) Exec(exec ExecValue) (result sql.Result, err error) {
	query := exec.Query()
	if t.tx == nil {
		result, err = t.Toy.db.Exec(query, exec.Args()...)
	} else {
		result, err = t.tx.Exec(query, exec.Args()...)
	}

	t.debugPrint(query, exec.JsonArgs(), err)
	return
}

func (t *ToyBrick) Query(exec ExecValue) (rows *sql.Rows, err error) {
	query := exec.Query()
	if t.tx == nil {
		rows, err = t.Toy.db.Query(query, exec.Args()...)
	} else {
		rows, err = t.tx.Query(query, exec.Args()...)
	}
	t.debugPrint(query, exec.JsonArgs(), err)
	return
}

func (t *ToyBrick) QueryRow(exec ExecValue) (row *sql.Row) {
	query := exec.Query()
	if t.tx == nil {
		row = t.Toy.db.QueryRow(query, exec.Args()...)
	} else {
		row = t.tx.QueryRow(query, exec.Args()...)
	}
	t.debugPrint(query, exec.JsonArgs(), nil)
	return
}

func (t *ToyBrick) Prepare(query string) (*sql.Stmt, error) {
	var stmt *sql.Stmt
	var err error
	if t.tx == nil {
		stmt, err = t.Toy.db.Prepare(query)
	} else {
		stmt, err = t.tx.Prepare(query)
	}
	t.debugPrint(query, "", err)
	return stmt, err
}

func (t *ToyBrick) CountExec() (exec ExecValue) {
	exec = t.Toy.Dialect.CountExec(t.Model)
	cExec := t.ConditionExec()
	exec = exec.Append(" "+cExec.Source(), cExec.Args()...)

	return
}

func (t *ToyBrick) ConditionExec() ExecValue {
	return t.Toy.Dialect.ConditionExec(t.Search, t.limit, t.offset, t.orderBy, t.groupBy)
}

func (t *ToyBrick) FindExec(records ModelRecordFieldTypes) ExecValue {
	var columns []Column
	for _, mField := range t.getSelectFields(records) {
		columns = append(columns, mField)
	}

	exec := t.Toy.Dialect.FindExec(t.Model, columns)

	cExec := t.ConditionExec()
	exec = exec.Append(" "+cExec.Source(), cExec.Args()...)
	return exec
}

func (t *ToyBrick) UpdateExec(record ModelRecord) ExecValue {
	exec := t.Toy.Dialect.UpdateExec(t.Model, t.getFieldValuePairWithRecord(ModeUpdate, record))
	cExec := t.ConditionExec()
	exec = exec.Append(" "+cExec.Source(), cExec.Args()...)
	return exec
}

func (t *ToyBrick) DeleteExec() ExecValue {
	exec := t.Toy.Dialect.DeleteExec(t.Model)
	cExec := t.ConditionExec()
	exec = exec.Append(" "+cExec.Source(), cExec.Args()...)
	return exec
}

func (t *ToyBrick) InsertExec(record ModelRecord) ExecValue {
	recorders := t.getFieldValuePairWithRecord(ModeInsert, record)
	exec := t.Toy.Dialect.InsertExec(t.Model, recorders)
	cExec := t.ConditionExec()
	exec = exec.Append(" "+cExec.Source(), cExec.Args()...)
	return exec
}

func (t *ToyBrick) ReplaceExec(record ModelRecord) ExecValue {
	recorders := t.getFieldValuePairWithRecord(ModeReplace, record)
	exec := t.Toy.Dialect.ReplaceExec(t.Model, recorders)
	cExec := t.ConditionExec()
	exec = exec.Append(" "+cExec.Source(), cExec.Args()...)
	return exec
}
