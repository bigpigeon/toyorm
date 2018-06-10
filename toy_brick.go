/*
 * Copyright 2018. bigpigeon. All rights reserved.
 * Use of this source code is governed by a MIT style
 * license that can be found in the LICENSE file.
 */

package toyorm

import (
	"database/sql"
	"fmt"
	"reflect"
)

type PreToyBrick struct {
	Parent *ToyBrick
	Field  Field
}

type PreJoinSwap struct {
	Swap    *JoinSwap
	PreSwap *PreJoinSwap
	Model   *Model
	Field   Field
}

type ToyBrick struct {
	Toy             *Toy
	preBrick        PreToyBrick
	MapPreloadBrick map[string]*ToyBrick
	//Error             []error
	debug bool
	tx    *sql.Tx

	orderBy  FieldList
	Search   SearchList
	offset   int
	limit    int
	groupBy  FieldList
	template *BasicExec
	// use join to association query in one command
	preSwap    *PreJoinSwap
	OwnOrderBy []int
	OwnGroupBy []int
	OwnSearch  []int
	alias      string
	SwapMap    map[string]*JoinSwap
	JoinMap    map[string]*Join

	BrickCommon
}

func NewToyBrick(toy *Toy, model *Model) *ToyBrick {
	return &ToyBrick{
		Toy:             toy,
		MapPreloadBrick: map[string]*ToyBrick{},
		SwapMap:         map[string]*JoinSwap{},
		JoinMap:         map[string]*Join{},
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
func (t *ToyBrick) RightValuePreload(fv FieldSelection) *ToyBrick {
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
func (t *ToyBrick) Preload(fv FieldSelection) *ToyBrick {
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

func (t *ToyBrick) CopyJoinSwap() map[string]*JoinSwap {
	newMap := make(map[string]*JoinSwap, len(t.SwapMap))
	for k, v := range t.SwapMap {
		newMap[k] = v
	}
	return newMap
}

func (t *ToyBrick) CopyJoin() map[string]*Join {
	newMap := make(map[string]*Join, len(t.JoinMap))
	for k, v := range t.JoinMap {
		newMap[k] = v
	}
	return newMap
}

// use join to association query
func (t *ToyBrick) Join(fv FieldSelection) *ToyBrick {
	return t.Scope(func(t *ToyBrick) *ToyBrick {
		if t.alias == "" {
			t = t.Alias("m")
		}
		field := t.Model.fieldSelect(fv)

		if join := t.JoinMap[field.Name()]; join != nil {
			newt := *t
			currentJoinSwap := joinSwap(t.SwapMap[field.Name()], &newt)
			newt.Model = join.SubModel
			newt.preSwap = &PreJoinSwap{currentJoinSwap, t.preSwap, t.Model, field}
			return &newt
		} else if join := t.Toy.Join(t.Model, field); join != nil {
			newt := *t
			newt.Model = join.SubModel
			swap := NewJoinSwap(fmt.Sprintf("%s_%d", t.alias, len(t.JoinMap)))
			currentJoinSwap := joinSwap(swap, &newt)

			// add field to pre swap
			currentJoinSwap.JoinMap = t.CopyJoin()
			currentJoinSwap.JoinMap[field.Name()] = join
			currentJoinSwap.SwapMap = t.CopyJoinSwap()
			currentJoinSwap.SwapMap[field.Name()] = swap

			newt.preSwap = &PreJoinSwap{currentJoinSwap, t.preSwap, t.Model, field}
			return &newt
		} else {
			panic(ErrInvalidPreloadField{t.Model.ReflectType.Name(), field.Name()})
		}

	})
}

func (t *ToyBrick) Swap() *ToyBrick {
	if t.preSwap == nil {
		panic("parent swap is nil")
	}
	newt := *t
	field := t.preSwap.Field
	newt.Model = t.preSwap.Model
	currentJoinSwap := joinSwap(t.preSwap.Swap, &newt)
	newt.SwapMap = newt.CopyJoinSwap()
	newt.SwapMap[field.Name()] = currentJoinSwap

	newt.preSwap = newt.preSwap.PreSwap
	return &newt
}

func (t *ToyBrick) Alias(alias string) *ToyBrick {
	return t.Scope(func(t *ToyBrick) *ToyBrick {
		newt := *t
		newt.alias = alias
		// reassignment all order by
		if len(t.OwnOrderBy) != 0 {
			newt.orderBy = make(FieldList, len(t.orderBy))
			copy(newt.orderBy, t.orderBy)
			for _, i := range t.OwnOrderBy {
				newt.orderBy[i] = newt.orderBy[i].Source().ToColumnAlias(alias)
			}
		}

		// reassignment all group by
		if len(t.OwnGroupBy) != 0 {
			newt.groupBy = make(FieldList, len(t.groupBy))
			copy(newt.groupBy, t.groupBy)
			for _, i := range t.OwnGroupBy {
				newt.groupBy[i] = newt.groupBy[i].Source().ToColumnAlias(alias)
			}
		}

		// reassignment all search
		if len(t.OwnSearch) != 0 {
			newt.Search = make(SearchList, len(t.Search))
			copy(newt.Search, t.Search)
			for _, i := range t.OwnSearch {
				newt.Search[i].Val = newt.Search[i].Val.Source().
					ToColumnAlias(alias).ToFieldValue(newt.Search[i].Val.Value())

			}
		}

		return &newt
	})
}

func (t *ToyBrick) CopyMapPreloadBrick() map[string]*ToyBrick {
	preloadBrick := map[string]*ToyBrick{}
	for k, v := range t.MapPreloadBrick {
		preloadBrick[k] = v
	}
	return preloadBrick
}

func (t *ToyBrick) CustomBelongToPreload(container, relationship FieldSelection, args ...interface{}) *ToyBrick {
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

func (t *ToyBrick) CustomOneToOnePreload(container FieldSelection, relationship interface{}, args ...interface{}) *ToyBrick {
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

func (t *ToyBrick) CustomOneToManyPreload(container FieldSelection, relationship interface{}, args ...interface{}) *ToyBrick {
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

func (t *ToyBrick) CustomManyToManyPreload(middleStruct interface{}, container FieldSelection, relation, subRelation interface{}, args ...interface{}) *ToyBrick {
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
func (t *ToyBrick) BindDefaultFields(args ...FieldSelection) *ToyBrick {
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

func (t *ToyBrick) condition(expr SearchExpr, key FieldSelection, args ...interface{}) SearchList {
	var value reflect.Value
	if len(args) == 1 {
		value = reflect.ValueOf(args[0])
	} else {
		value = reflect.ValueOf(args)
	}
	mField := t.Model.fieldSelect(key)

	search := SearchList{}.Condition(mField.ToColumnAlias(t.alias).ToFieldValue(value), expr, ExprAnd)

	return search
}

func (t *ToyBrick) conditionGroup(expr SearchExpr, group interface{}) SearchList {
	switch expr {
	case ExprAnd, ExprOr:
		var search SearchList
		keyValue := LoopIndirect(reflect.ValueOf(group))
		record := NewRecord(t.Model, keyValue)
		pairs := t.getFieldValuePairWithRecord(ModeCondition, record)
		for _, pair := range pairs {
			search = search.Condition(pair, ExprEqual, expr)
		}
		// avoid "or" condition effected by priority
		if expr == ExprOr {
			search = append(search, NewSearchBranch(ExprIgnore))
		}

		return search
	}
	panic("invalid expr")
}

// where will clean old condition
func (t *ToyBrick) Where(expr SearchExpr, key FieldSelection, v ...interface{}) *ToyBrick {
	return t.Conditions(t.condition(expr, key, v...))
}

// expr only support And/Or , group must be struct data or map[string]interface{}/map[uintptr]interface{}
func (t *ToyBrick) WhereGroup(expr SearchExpr, group interface{}) *ToyBrick {
	return t.Conditions(t.conditionGroup(expr, group))
}

func (t *ToyBrick) Conditions(search SearchList) *ToyBrick {
	return t.Scope(func(t *ToyBrick) *ToyBrick {
		newt := *t.CleanOwnSearch()
		if len(search) == 0 {
			newt.Search = nil
			return &newt
		}
		newSearch := make(SearchList, len(search), len(search)+1)
		copy(newSearch, search)
		// Avoid "or" condition effected by priority
		newSearch = append(newSearch, NewSearchBranch(ExprIgnore))
		newt.Search = newSearch
		for i, s := range newt.Search {
			if s.Type.IsBranch() == false {
				newt.OwnSearch = append(newt.OwnSearch, i)
			}
		}
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

func (t *ToyBrick) OrderBy(vList ...FieldSelection) *ToyBrick {
	return t.Scope(func(t *ToyBrick) *ToyBrick {
		newt := *t.CleanOwnOrderBy()
		newt.orderBy = nil
		for i, v := range vList {
			if column, ok := v.(Field); ok {
				newt.orderBy = append(newt.orderBy, column)
			} else {
				newt.orderBy = append(newt.orderBy, t.Model.fieldSelect(v).ToColumnAlias(t.alias))
			}
			newt.OwnOrderBy = append(newt.OwnOrderBy, i)
		}
		return &newt
	})
}

func (t *ToyBrick) GroupBy(vList ...FieldSelection) *ToyBrick {
	return t.Scope(func(t *ToyBrick) *ToyBrick {
		// remove old model group by data
		newt := *t.CleanOwnGroupBy()
		newt.groupBy = nil
		for i, v := range vList {
			if column, ok := v.(Field); ok {
				newt.groupBy = append(newt.groupBy, column)
			} else {
				newt.groupBy = append(newt.groupBy, t.Model.fieldSelect(v).ToColumnAlias(t.alias))
			}
			newt.OwnGroupBy = append(newt.OwnGroupBy, i)
		}
		return &newt
	})
}

func (t *ToyBrick) Template(temp string, args ...interface{}) *ToyBrick {
	return t.Scope(func(t *ToyBrick) *ToyBrick {
		newt := *t
		if temp == "" && len(args) == 0 {
			newt.template = nil
		} else {
			newt.template = &BasicExec{temp, args}
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
		return nil, ErrCannotSet{"v"}
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

func (t *ToyBrick) debugPrint(exec ExecValue, err error) {
	if t.debug {
		if err != nil {
			fmt.Fprintf(t.Toy.Logger, "use tx: %p, query:%s  args:%s faiure reason %s\n", t.tx, exec.Query(), exec.JsonArgs(), err)
		} else {
			fmt.Fprintf(t.Toy.Logger, "use tx: %p, query:%s  args:%s\n", t.tx, exec.Query(), exec.JsonArgs())
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

	t.debugPrint(exec, err)
	return
}

func (t *ToyBrick) Query(exec ExecValue) (rows *sql.Rows, err error) {
	query := exec.Query()
	if t.tx == nil {
		rows, err = t.Toy.db.Query(query, exec.Args()...)
	} else {
		rows, err = t.tx.Query(query, exec.Args()...)
	}
	t.debugPrint(exec, err)
	return
}

func (t *ToyBrick) QueryRow(exec ExecValue) (row *sql.Row) {
	query := exec.Query()
	if t.tx == nil {
		row = t.Toy.db.QueryRow(query, exec.Args()...)
	} else {
		row = t.tx.QueryRow(query, exec.Args()...)
	}
	t.debugPrint(exec, nil)
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
	t.debugPrint(&DefaultExec{query: query}, err)
	return stmt, err
}

func (t *ToyBrick) CountExec() (exec ExecValue) {
	exec = t.Toy.Dialect.CountExec(t.Model)
	cExec := t.ConditionExec()
	exec = exec.Append(" "+cExec.Source(), cExec.Args()...)

	return
}

func (t *ToyBrick) ConditionExec() ExecValue {
	return t.Toy.Dialect.ConditionExec(t.Search, t.limit, t.offset, t.orderBy.ToColumnList(), t.groupBy.ToColumnList())
}

func (t *ToyBrick) FindExec(columns []Column) ExecValue {

	exec := t.Toy.Dialect.FindExec(t.Model, columns, t.alias)
	jExec := t.Toy.Dialect.JoinExec(joinSwap(nil, t))
	exec = exec.Append(" "+jExec.Source(), jExec.Args()...)
	cExec := t.ConditionExec()
	exec = exec.Append(" "+cExec.Source(), cExec.Args()...)
	return exec
}

func (t *ToyBrick) UpdateExec(record ModelRecord) ExecValue {
	exec := t.Toy.Dialect.UpdateExec(t.Model, t.getFieldValuePairWithRecord(ModeUpdate, record).ToValueList())
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
	exec := t.Toy.Dialect.InsertExec(t.Model, recorders.ToValueList())
	cExec := t.ConditionExec()
	exec = exec.Append(" "+cExec.Source(), cExec.Args()...)
	return exec
}

func (t *ToyBrick) SaveExec(record ModelRecord) ExecValue {
	recorders := t.getFieldValuePairWithRecord(ModeReplace, record)
	exec := t.Toy.Dialect.SaveExec(t.Model, recorders.ToNameValueList())
	cExec := t.ConditionExec()
	exec = exec.Append(" "+cExec.Source(), cExec.Args()...)
	return exec
}

func (t *ToyBrick) getFieldValuePairWithRecord(mode Mode, record ModelRecord) FieldValueList {
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
					if mField.IsPrimary() && IsZero(fieldValue) {

					} else {
						columnValues = append(columnValues, mField.ToColumnAlias(t.alias).ToFieldValue(fieldValue))
					}
				}
			}
		}
	} else {
		for _, mField := range fields {
			if fieldValue := record.Field(mField.Name()); fieldValue.IsValid() {
				columnValues = append(columnValues, mField.ToColumnAlias(t.alias).ToFieldValue(fieldValue))
			}
		}
	}
	return columnValues
}

func (t *ToyBrick) getSelectFields(records ModelRecordFieldTypes) FieldList {
	var fields []Field
	if len(t.FieldsSelector[ModeSelect]) > 0 {
		fields = t.FieldsSelector[ModeSelect]
	} else if len(t.FieldsSelector[ModeDefault]) > 0 {
		fields = t.FieldsSelector[ModeDefault]
	} else {
		fields = t.Model.GetSqlFields()
	}
	if t.alias != "" {
		aliasFields := make([]Field, len(fields))
		for i := range fields {
			aliasFields[i] = fields[i].ToColumnAlias(t.alias)
		}
		fields = aliasFields
	}
	return getFieldsWithRecords(fields, records)
}

func (t *ToyBrick) CleanOwnOrderBy() *ToyBrick {
	newt := *t
	newt.OwnOrderBy = nil
	newt.SwapMap = t.CopyJoinSwap()
	for name, swap := range newt.SwapMap {
		if swap.OwnOrderBy != nil {
			newt.SwapMap[name] = newt.SwapMap[name].Copy()
			newt.SwapMap[name].OwnOrderBy = nil
		}
	}
	return &newt
}

func (t *ToyBrick) CleanOwnGroupBy() *ToyBrick {
	newt := *t
	newt.OwnGroupBy = nil
	newt.SwapMap = t.CopyJoinSwap()
	for name, swap := range newt.SwapMap {
		if swap.OwnGroupBy != nil {
			newt.SwapMap[name] = newt.SwapMap[name].Copy()
			newt.SwapMap[name].OwnGroupBy = nil
		}
	}
	return &newt
}

func (t *ToyBrick) CleanOwnSearch() *ToyBrick {
	newt := *t
	newt.OwnSearch = nil
	newt.SwapMap = t.CopyJoinSwap()
	for name, swap := range newt.SwapMap {
		if swap.OwnSearch != nil {
			newt.SwapMap[name] = newt.SwapMap[name].Copy()
			newt.SwapMap[name].OwnSearch = nil
		}
	}
	return &newt
}
