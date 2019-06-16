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

	orderBy     FieldList
	Search      SearchList
	offset      int
	limit       int
	groupBy     FieldList
	templateMap map[TempMode]*BasicExec
	// use join to association query in one command
	preSwap     *PreJoinSwap
	OwnOrderBy  []int
	OwnGroupBy  []int
	OwnSearch   []int
	alias       string
	SwapMap     map[string]*JoinSwap
	JoinMap     map[string]*Join
	objMustAddr bool // TODO maybe not a good way
	rsync       bool

	BrickCommon
}

func NewToyBrick(toy *Toy, model *Model) *ToyBrick {
	return &ToyBrick{
		Toy:             toy,
		MapPreloadBrick: map[string]*ToyBrick{},
		SwapMap:         map[string]*JoinSwap{},
		JoinMap:         map[string]*Join{},
		templateMap:     map[TempMode]*BasicExec{},
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

		subModel := t.Toy.GetModel(LoopDiveSliceAndPtr(field.FieldValue()))
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
		subModel := t.Toy.GetModel(LoopDiveSliceAndPtr(field.FieldValue()))
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
		subModel = t.Toy.GetModel(LoopDivePtr(reflect.ValueOf(args[0])))
	} else {
		subModel = t.Toy.GetModel(LoopDivePtr(containerField.FieldValue()))
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
		subModel = t.Toy.GetModel(LoopDivePtr(reflect.ValueOf(args[0])))
	} else {
		subModel = t.Toy.GetModel(LoopDivePtr(containerField.FieldValue()))
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
		subModel = t.Toy.GetModel(LoopDiveSliceAndPtr(reflect.ValueOf(args[0])))
	} else {
		subModel = t.Toy.GetModel(LoopDiveSliceAndPtr(containerField.FieldValue()))
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
		newt := t.CleanOwnSearch()
		if len(search) == 0 {
			newt.Search = nil
			return newt
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
		return newt
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
			field := t.Model.fieldSelect(v)
			newt.orderBy = append(newt.orderBy, field.ToColumnAlias(t.alias))
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
			field := t.Model.fieldSelect(v)
			newt.groupBy = append(newt.groupBy, field.ToColumnAlias(t.alias))
			newt.OwnGroupBy = append(newt.OwnGroupBy, i)
		}
		return &newt
	})
}

// use custom template sql replace default sql, it will replace all mode template sql
func (t *ToyBrick) Template(temp string, args ...interface{}) *ToyBrick {
	return t.Scope(func(t *ToyBrick) *ToyBrick {
		newt := *t
		newt.templateMap = map[TempMode]*BasicExec{}
		for k, v := range t.templateMap {
			newt.templateMap[k] = v
		}
		if temp == "" && len(args) == 0 {
			newt.templateMap[TempDefault] = nil
		} else {
			newt.templateMap[TempDefault] = &BasicExec{temp, args}
		}
		return &newt
	})
}

// use custom template sql replace default sql
func (t *ToyBrick) TemplateMode(mode TempMode, temp string, args ...interface{}) *ToyBrick {
	return t.Scope(func(t *ToyBrick) *ToyBrick {
		newt := *t
		newt.templateMap = map[TempMode]*BasicExec{}
		for k, v := range t.templateMap {
			newt.templateMap[k] = v
		}
		if temp == "" && len(args) == 0 {
			newt.templateMap[mode] = nil
		} else {
			newt.templateMap[mode] = &BasicExec{temp, args}
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

func (t *ToyBrick) insert(records ModelRecords) *Result {
	handlers := t.Toy.ModelHandlers("Insert", t.Model)
	ctx := NewContext(handlers, t, records)
	go ctx.Start()
	return ctx.Result
}

func (t *ToyBrick) save(records ModelRecords) *Result {
	handlers := t.Toy.ModelHandlers("Save", t.Model)
	ctx := NewContext(handlers, t, records)
	go ctx.Start()
	return ctx.Result
}

func (t *ToyBrick) usave(records ModelRecords) *Result {
	handlers := t.Toy.ModelHandlers("USave", t.Model)
	ctx := NewContext(handlers, t, records)
	go ctx.Start()
	return ctx.Result
}

func (t *ToyBrick) deleteWithPrimaryKey(records ModelRecords) *Result {
	if field := t.Model.GetFieldWithName("DeletedAt"); field != nil {
		return t.softDeleteWithPrimaryKey(records)
	} else {
		return t.hardDeleteWithPrimaryKey(records)
	}
}

func (t *ToyBrick) softDeleteWithPrimaryKey(records ModelRecords) *Result {
	handlers := t.Toy.ModelHandlers("SoftDeleteWithPrimaryKey", t.Model)
	ctx := NewContext(handlers, t, records)
	go ctx.Start()
	return ctx.Result
}

func (t *ToyBrick) hardDeleteWithPrimaryKey(records ModelRecords) *Result {
	handlers := t.Toy.ModelHandlers("HardDeleteWithPrimaryKey", t.Model)
	ctx := NewContext(handlers, t, records)
	go ctx.Start()
	return ctx.Result
}

func (t *ToyBrick) delete(records ModelRecords) *Result {
	if field := t.Model.GetFieldWithName("DeletedAt"); field != nil {
		return t.softDelete(records)
	} else {
		return t.hardDelete(records)
	}
}

func (t *ToyBrick) softDelete(records ModelRecords) *Result {
	handlers := t.Toy.ModelHandlers("SoftDelete", t.Model)
	ctx := NewContext(handlers, t, records)
	go ctx.Start()
	return ctx.Result
}

func (t *ToyBrick) hardDelete(records ModelRecords) *Result {
	handlers := t.Toy.ModelHandlers("HardDelete", t.Model)
	ctx := NewContext(handlers, t, records)
	go ctx.Start()
	return ctx.Result
}

func (t *ToyBrick) find(value reflect.Value) *Context {
	if value.Kind() == reflect.Slice {
		handlers := t.Toy.ModelHandlers("Find", t.Model)
		records := NewRecords(t.Model, value)
		ctx := NewContext(handlers, t, records)
		go ctx.Start()
		return ctx
	} else {
		handlers := t.Toy.ModelHandlers("FindOne", t.Model)
		vList := reflect.New(reflect.SliceOf(value.Type())).Elem()
		vList = reflect.Append(vList, value)
		records := NewRecords(t.Model, vList)

		ctx := NewContext(handlers, t.Limit(1), records)
		go ctx.Start()
		//err := ctx.Next()
		//if vList.Len() == 0 {
		//	if err == nil {
		//		err = sql.ErrNoRows
		//	}
		//	return ctx, err
		//}
		//value.Set(vList.Index(0))
		return ctx
	}
}

func (t *ToyBrick) CreateTable() *Result {
	ctx := t.GetContext("CreateTable", MakeRecordsWithElem(t.Model, t.Model.ReflectType))
	go ctx.Start()
	return ctx.Result
}

func (t *ToyBrick) CreateTableIfNotExist() *Result {
	ctx := t.GetContext("CreateTableIfNotExist", MakeRecordsWithElem(t.Model, t.Model.ReflectType))
	go ctx.Start()
	return ctx.Result
}

func (t *ToyBrick) DropTable() *Result {
	ctx := t.GetContext("DropTable", MakeRecordsWithElem(t.Model, t.Model.ReflectType))
	go ctx.Start()
	return ctx.Result
}

func (t *ToyBrick) DropTableIfExist() *Result {
	ctx := t.GetContext("DropTableIfExist", MakeRecordsWithElem(t.Model, t.Model.ReflectType))
	go ctx.Start()
	return ctx.Result
}

func (t *ToyBrick) HasTable() (b bool, err error) {
	exec := t.Toy.Dialect.HasTable(t.Model)
	err = t.QueryRow(exec).Scan(&b)
	return b, err
}

func (t *ToyBrick) Count() (count int, err error) {
	exec, err := t.CountExec()
	if err != nil {
		return 0, err
	}
	err = t.QueryRow(exec).Scan(&count)
	return count, err
}

// insert can receive three type data
// struct
// map[offset]interface{}
// map[int]interface{}
// insert is difficult that have preload data
func (t *ToyBrick) Insert(v interface{}) *Result {
	vValue := LoopIndirect(reflect.ValueOf(v))
	if t.objMustAddr && vValue.CanAddr() == false {
		panic("object must can addr")
	}
	var result *Result
	switch vValue.Kind() {
	case reflect.Slice:
		records := NewRecords(t.Model, vValue)
		result = t.insert(records)
	default:
		var records ModelRecords
		if vValue.CanAddr() {
			records = MakeRecordsWithElem(t.Model, vValue.Addr().Type())
			records.Add(vValue.Addr())
		} else {
			records = MakeRecordsWithElem(t.Model, vValue.Type())
			records.Add(vValue)
		}
		result = t.insert(records)
	}
	if t.rsync == false {
		<-result.done
	}
	return result
}

func (t *ToyBrick) Find(v interface{}) *Result {
	vValue := LoopIndirectAndNew(reflect.ValueOf(v))
	if t.objMustAddr && vValue.CanAddr() == false {
		panic("object must can addr")
	}
	ctx := t.find(vValue)
	if t.rsync == false {
		<-ctx.Result.done
	}
	return ctx.Result
}

func (t *ToyBrick) Update(v interface{}) *Result {
	vValue := LoopIndirect(reflect.ValueOf(v))
	if t.objMustAddr && vValue.CanAddr() == false {
		panic("object must can addr")
	}
	vValueList := reflect.MakeSlice(reflect.SliceOf(vValue.Type()), 0, 1)
	vValueList = reflect.Append(vValueList, vValue)
	handlers := t.Toy.ModelHandlers("Update", t.Model)
	ctx := NewContext(handlers, t, NewRecords(t.Model, vValueList))
	go ctx.Start()
	if t.rsync == false {
		<-ctx.Result.done
	}
	return ctx.Result
}

func (t *ToyBrick) Save(v interface{}) *Result {
	vValue := LoopIndirect(reflect.ValueOf(v))
	if t.objMustAddr && vValue.CanAddr() == false {
		panic("object must can addr")
	}
	var result *Result
	switch vValue.Kind() {
	case reflect.Slice:
		records := NewRecords(t.Model, vValue)
		result = t.save(records)
	default:
		records := MakeRecordsWithElem(t.Model, vValue.Addr().Type())
		records.Add(vValue.Addr())
		result = t.save(records)
	}
	if t.rsync == false {
		<-result.done
	}
	return result
}

// save with exist data
func (t *ToyBrick) USave(v interface{}) *Result {
	vValue := LoopIndirect(reflect.ValueOf(v))
	if t.objMustAddr && vValue.CanAddr() == false {
		panic("object must can addr")
	}
	var result *Result
	switch vValue.Kind() {
	case reflect.Slice:
		records := NewRecords(t.Model, vValue)
		result = t.usave(records)
	default:
		records := MakeRecordsWithElem(t.Model, vValue.Addr().Type())
		records.Add(vValue.Addr())
		result = t.usave(records)
	}
	if t.rsync == false {
		<-result.done
	}
	return result
}

func (t *ToyBrick) Delete(v interface{}) *Result {
	vValue := LoopIndirect(reflect.ValueOf(v))
	if t.objMustAddr && vValue.CanAddr() == false {
		panic("object must can addr")
	}
	var result *Result

	switch vValue.Kind() {
	case reflect.Slice:
		records := NewRecords(t.Model, vValue)
		result = t.deleteWithPrimaryKey(records)
	default:
		records := MakeRecordsWithElem(t.Model, vValue.Addr().Type())
		records.Add(vValue.Addr())
		result = t.deleteWithPrimaryKey(records)
	}
	if t.rsync == false {
		<-result.done
	}
	return result
}

func (t *ToyBrick) DeleteWithConditions() *Result {
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

func (t *ToyBrick) CountExec() (ExecValue, error) {
	// TODO move to handlers
	deletedField := t.Model.GetFieldWithName("DeletedAt")
	if deletedField != nil {
		t = t.Where(ExprNull, deletedField).And().Conditions(t.Search)
	}
	condition := DialectConditionArgs{
		t.Search,
		t.limit, t.offset,
		t.orderBy.ToColumnList(), t.groupBy.ToColumnList(),
	}
	exec, err := t.Toy.Dialect.FindExec(t.templateSelect(TempFind), t.Model, DialectFindArgs{
		Columns: []Column{CountColumn{}},
		Swap:    joinSwap(nil, t),
	}, condition)
	if err != nil {
		return nil, err
	}

	return exec, nil
}

func (t *ToyBrick) ConditionExec() ExecValue {
	return t.Toy.Dialect.ConditionExec(t.Search, t.limit, t.offset, t.orderBy.ToColumnList(), t.groupBy.ToColumnList())
}

func (t *ToyBrick) FindExec(columns []Column) (ExecValue, error) {
	condition := DialectConditionArgs{
		t.Search,
		t.limit, t.offset,
		t.orderBy.ToColumnList(), t.groupBy.ToColumnList(),
	}
	exec, err := t.Toy.Dialect.FindExec(t.templateSelect(TempFind), t.Model, DialectFindArgs{
		Columns: columns,
		Swap:    joinSwap(nil, t),
	}, condition)
	if err != nil {
		return nil, err
	}

	return exec, nil
}

func (t *ToyBrick) UpdateExec(record ModelRecord) (ExecValue, error) {
	condition := DialectConditionArgs{
		t.Search,
		t.limit, t.offset,
		t.orderBy.ToColumnList(), t.groupBy.ToColumnList(),
	}
	recorders := t.getFieldValuePairWithRecord(ModeUpdate, record)
	return t.Toy.Dialect.UpdateExec(t.templateSelect(TempUpdate), t.Model, DialectUpdateArgs{recorders}, condition)
}

func (t *ToyBrick) SoftDeleteExec(records ModelRecords) (ExecValue, error) {
	condition := DialectConditionArgs{
		t.Search,
		t.limit, t.offset,
		t.orderBy.ToColumnList(), t.groupBy.ToColumnList(),
	}
	delArgs := getDeleteArgs(t.Model, t.BelongToPreload, records)
	return t.Toy.Dialect.SoftDeleteExec(t.templateSelect(TempUpdate), t.Model, delArgs, condition)
}

func (t *ToyBrick) HardDeleteExec(records ModelRecords) (ExecValue, error) {
	condition := DialectConditionArgs{
		t.Search,
		t.limit, t.offset,
		t.orderBy.ToColumnList(), t.groupBy.ToColumnList(),
	}
	delArgs := getDeleteArgs(t.Model, t.BelongToPreload, records)
	return t.Toy.Dialect.HardDeleteExec(t.templateSelect(TempDelete), t.Model, delArgs, condition)
}

func (t *ToyBrick) InsertExec(record ModelRecord) (ExecValue, error) {
	recorders := t.getFieldValuePairWithRecord(ModeInsert, record)
	condition := DialectConditionArgs{
		t.Search,
		0, 0, nil, nil,
	}
	save := getSaveArgs(t.Model, recorders)
	exec, err := t.Toy.Dialect.InsertExec(t.templateSelect(TempInsert), t.Model, save, condition)
	if err != nil {
		return nil, err
	}
	return exec, nil
}

func (t *ToyBrick) SaveExec(record ModelRecord) (ExecValue, error) {
	recorders := t.getFieldValuePairWithRecord(ModeSave, record)
	condition := DialectConditionArgs{
		t.Search,
		0, 0, nil, nil,
	}
	save := getSaveArgs(t.Model, recorders)
	exec, err := t.Toy.Dialect.SaveExec(t.templateSelect(TempSave), t.Model, save, condition)
	if err != nil {
		return nil, err
	}
	return exec, nil
}

func (t *ToyBrick) USaveExec(record ModelRecord) (ExecValue, error) {
	recorders := t.getFieldValuePairWithRecord(ModeSave, record)
	condition := DialectConditionArgs{
		t.Search,
		0, 0, nil, nil,
	}
	save := getSaveArgs(t.Model, recorders)
	exec, err := t.Toy.Dialect.USaveExec(t.templateSelect(TempUSave), t.Model, t.alias, save, condition)
	if err != nil {
		return nil, err
	}
	return exec, nil
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
					// move primary key check to dialect
					//if mField.IsPrimary() && IsZero(fieldValue) {
					//
					//} else {
					//	columnValues = append(columnValues, mField.ToColumnAlias(t.alias).ToFieldValue(fieldValue))
					//}
					columnValues = append(columnValues, mField.ToColumnAlias(t.alias).ToFieldValue(fieldValue))
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

func (t *ToyBrick) templateSelect(mode TempMode) *BasicExec {
	if temp := t.templateMap[mode]; temp != nil {
		return temp
	}
	return t.templateMap[TempDefault]
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
	return t.Scope(func(t *ToyBrick) *ToyBrick {
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
	})
}

func (t *ToyBrick) CleanOwnSearch() *ToyBrick {
	return t.Scope(func(t *ToyBrick) *ToyBrick {
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
	})
}

func (t *ToyBrick) Rsync() *ToyBrick {
	return t.Scope(func(t *ToyBrick) *ToyBrick {
		newt := *t
		newt.rsync = true
		return &newt
	})
}
