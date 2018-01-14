package toyorm

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

type ToyBrickRelationship struct {
	Parent *ToyBrick
	Field  *ModelField
}

type ToyBrick struct {
	Toy               *Toy
	relationship      ToyBrickRelationship
	MapPreloadBrick   map[*ModelField]*ToyBrick
	OneToOnePreload   map[*ModelField]*OneToOnePreload
	OneToManyPreload  map[*ModelField]*OneToManyPreload
	ManyToManyPreload map[*ModelField]*ManyToManyPreload
	//Error             []error
	debug bool
	tx    *sql.Tx

	model   *Model
	orderBy []*ModelField
	Search  SearchList
	offset  int
	limit   int
	// use by update/insert/replace data when source value is struct
	//ignoreMode IgnoreMode
	// use by SELECT fields FROM TABLE/INSERT INFO table(fields)/UPDATE table SET field...
	// if len(Fields) == 0, use model.SqlFields
	//Fields []*ModelField
	// TODO use by find data if fields number not equal scan value number
	FieldsSelector     map[Mode][]*ModelField
	ignoreModeSelector map[Mode]IgnoreMode

	// use for SELECT Fields... FROM table, if nil or len = 0, use Fields
	//SelectField []*ModelField
	// use for INSERT / REPLACE INTO(InsertField), if nil or len = 0, use Fields
	//InsertField []*ModelField
	// use for UPDATE table SET field1=xx,field2=xx, if nil or len = 0, use Fields
	//UpdateFields []*ModelField
	// use for scan, if nil or len = 0, use Fields
	//ScanFields []*ModelField
	// use for brick.Where(ExprAnd/ExprOr, data)
	//ConditionFields []*ModelField
}

func NewToyBrick(toy *Toy, model *Model) *ToyBrick {
	return &ToyBrick{
		Toy:               toy,
		model:             model,
		MapPreloadBrick:   map[*ModelField]*ToyBrick{},
		OneToOnePreload:   map[*ModelField]*OneToOnePreload{},
		OneToManyPreload:  map[*ModelField]*OneToManyPreload{},
		ManyToManyPreload: map[*ModelField]*ManyToManyPreload{},
		ignoreModeSelector: map[Mode]IgnoreMode{
			ModeInsert:    IgnoreNo,
			ModeReplace:   IgnoreNo,
			ModeCondition: IgnoreZero,
			ModeUpdate:    IgnoreZero,
		},
	}
}

func (t *ToyBrick) And() ToyBrickAnd {
	return ToyBrickAnd{t}
}

func (t *ToyBrick) Or() ToyBrickOr {
	return ToyBrickOr{t}
}

func (t *ToyBrick) Clone() *ToyBrick {
	newt := &ToyBrick{
		Toy: t.Toy,
	}
	return newt
}

func (t *ToyBrick) Scope(fn func(*ToyBrick) *ToyBrick) *ToyBrick {
	ret := fn(t)
	return ret
}

func (t *ToyBrick) CopyStatus(statusBrick *ToyBrick) *ToyBrick {
	newt := *t
	newt.tx = statusBrick.tx
	newt.debug = statusBrick.debug
	newt.ignoreModeSelector = map[Mode]IgnoreMode{}
	for k, v := range t.ignoreModeSelector {
		newt.ignoreModeSelector[k] = v
	}
	return &newt
}

// return it parent ToyBrick
// it will panic when the parent ToyBrick is nil
func (t *ToyBrick) Enter() *ToyBrick {
	return t.Scope(func(t *ToyBrick) *ToyBrick {
		newt := *t.relationship.Parent
		newt.MapPreloadBrick = map[*ModelField]*ToyBrick{}
		for k, v := range t.relationship.Parent.MapPreloadBrick {
			newt.MapPreloadBrick[k] = v
		}
		t.relationship.Parent = &newt
		newt.MapPreloadBrick[t.relationship.Field] = t
		return t.relationship.Parent
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
		field := t.model.fieldSelect(fv)
		if subBrick, ok := t.MapPreloadBrick[field]; ok {
			return subBrick
		}
		subModel := t.Toy.GetModel(LoopTypeIndirectSliceAndPtr(field.Field.Type))
		newSubt := NewToyBrick(t.Toy, subModel).CopyStatus(t)

		newt := *t
		newt.MapPreloadBrick = t.CopyMapPreloadBrick()
		newt.MapPreloadBrick[field] = newSubt
		newSubt.relationship = ToyBrickRelationship{&newt, field}
		if preload := newt.Toy.ManyToManyPreload(newt.model, field, true); preload != nil {
			newt.ManyToManyPreload = t.CopyManyToManyPreload()
			newt.ManyToManyPreload[field] = preload
		} else {
			panic(fmt.Sprintf("invalid preload field '%s'", field.Field.Name))
		}
		return newSubt
	})
}

// return
func (t *ToyBrick) Preload(fv interface{}) *ToyBrick {
	return t.Scope(func(t *ToyBrick) *ToyBrick {
		field := t.model.fieldSelect(fv)
		if subBrick, ok := t.MapPreloadBrick[field]; ok {
			return subBrick
		}
		subModel := t.Toy.GetModel(LoopTypeIndirectSliceAndPtr(field.Field.Type))
		newSubt := NewToyBrick(t.Toy, subModel).CopyStatus(t)

		newt := *t
		newt.MapPreloadBrick = t.CopyMapPreloadBrick()
		newt.MapPreloadBrick[field] = newSubt
		newSubt.relationship = ToyBrickRelationship{&newt, field}
		if preload := newt.Toy.OneToOnePreload(newt.model, field); preload != nil {
			newt.OneToOnePreload = t.CopyOneToOnePreload()
			newt.OneToOnePreload[field] = preload
		} else if preload := newt.Toy.OneToManyPreload(newt.model, field); preload != nil {
			newt.OneToManyPreload = t.CopyOneToManyPreload()
			for k, v := range t.OneToManyPreload {
				newt.OneToManyPreload[k] = v
			}
			newt.OneToManyPreload[field] = preload
		} else if preload := newt.Toy.ManyToManyPreload(newt.model, field, false); preload != nil {
			newt.ManyToManyPreload = t.CopyManyToManyPreload()
			newt.ManyToManyPreload[field] = preload
		} else {
			panic(fmt.Sprintf("invalid preload field '%s'", field.Field.Name))
		}
		return newSubt
	})
}

func (t *ToyBrick) CopyMapPreloadBrick() map[*ModelField]*ToyBrick {
	preloadBrick := map[*ModelField]*ToyBrick{}
	for k, v := range t.MapPreloadBrick {
		preloadBrick[k] = v
	}
	return preloadBrick
}

func (t *ToyBrick) CopyOneToOnePreload() map[*ModelField]*OneToOnePreload {
	preloadMap := map[*ModelField]*OneToOnePreload{}
	for k, v := range t.OneToOnePreload {
		preloadMap[k] = v
	}
	return preloadMap
}

func (t *ToyBrick) CopyOneToManyPreload() map[*ModelField]*OneToManyPreload {
	preloadMap := map[*ModelField]*OneToManyPreload{}
	for k, v := range t.OneToManyPreload {
		preloadMap[k] = v
	}
	return preloadMap
}

func (t *ToyBrick) CopyManyToManyPreload() map[*ModelField]*ManyToManyPreload {
	preloadMap := map[*ModelField]*ManyToManyPreload{}
	for k, v := range t.ManyToManyPreload {
		preloadMap[k] = v
	}
	return preloadMap
}

func (t *ToyBrick) CustomOneToOnePreload(container, relationship interface{}, args ...interface{}) *ToyBrick {
	containerField := t.model.fieldSelect(container)
	var subModel *Model
	if len(args) > 0 {
		subModel = t.Toy.GetModel(LoopTypeIndirect(reflect.TypeOf(args[0])))
	} else {
		subModel = t.Toy.GetModel(LoopTypeIndirect(containerField.Field.Type))
	}
	relationshipField := subModel.fieldSelect(relationship)
	preload := t.Toy.OneToOneBind(t.model, subModel, containerField, relationshipField, false)
	if preload == nil {
		panic(fmt.Sprintf("invalid preload field '%s'", containerField.Field.Name))
	}

	newSubt := NewToyBrick(t.Toy, subModel).CopyStatus(t)
	newt := *t
	newt.MapPreloadBrick = t.CopyMapPreloadBrick()
	newt.MapPreloadBrick[containerField] = newSubt
	newSubt.relationship = ToyBrickRelationship{&newt, containerField}

	newt.OneToOnePreload = t.CopyOneToOnePreload()
	newt.OneToOnePreload[containerField] = preload
	return newSubt
}

func (t *ToyBrick) CustomBelongToPreload(container, relationship interface{}, args ...interface{}) *ToyBrick {
	containerField, relationshipField := t.model.fieldSelect(container), t.model.fieldSelect(relationship)
	var subModel *Model
	if len(args) > 0 {
		subModel = t.Toy.GetModel(LoopTypeIndirect(reflect.TypeOf(args[0])))
	} else {
		subModel = t.Toy.GetModel(LoopTypeIndirect(containerField.Field.Type))
	}
	preload := t.Toy.OneToOneBind(t.model, subModel, containerField, relationshipField, false)
	if preload == nil {
		panic(fmt.Sprintf("invalid preload field '%s'", containerField.Field.Name))
	}

	newSubt := NewToyBrick(t.Toy, subModel).CopyStatus(t)
	newt := *t
	newt.MapPreloadBrick = t.CopyMapPreloadBrick()
	newt.MapPreloadBrick[containerField] = newSubt
	newSubt.relationship = ToyBrickRelationship{&newt, containerField}

	newt.OneToOnePreload = t.CopyOneToOnePreload()
	newt.OneToOnePreload[containerField] = preload
	return newSubt
}

func (t *ToyBrick) CustomOneToManyPreload(container, relationship interface{}, args ...interface{}) *ToyBrick {
	containerField := t.model.fieldSelect(container)
	var subModel *Model
	if len(args) > 0 {
		subModel = t.Toy.GetModel(LoopTypeIndirect(reflect.TypeOf(args[0])))
	} else {
		subModel = t.Toy.GetModel(LoopTypeIndirectSliceAndPtr(containerField.Field.Type))
	}
	relationshipField := subModel.fieldSelect(relationship)
	preload := t.Toy.OneToManyBind(t.model, subModel, containerField, relationshipField)
	if preload == nil {
		panic(fmt.Sprintf("invalid preload field '%s'", containerField.Field.Name))
	}

	newSubt := NewToyBrick(t.Toy, subModel).CopyStatus(t)
	newt := *t
	newt.MapPreloadBrick = t.CopyMapPreloadBrick()
	newt.MapPreloadBrick[containerField] = newSubt
	newSubt.relationship = ToyBrickRelationship{&newt, containerField}

	newt.OneToManyPreload = t.CopyOneToManyPreload()
	newt.OneToManyPreload[containerField] = preload
	return newSubt
}

func (t *ToyBrick) CustomManyToManyPreload(container, middleStruct, relation, subRelation interface{}, args ...interface{}) *ToyBrick {
	containerField := t.model.fieldSelect(container)
	var subModel *Model
	if len(args) > 0 {
		subModel = t.Toy.GetModel(LoopTypeIndirect(reflect.TypeOf(args[0])))
	} else {
		subModel = t.Toy.GetModel(LoopTypeIndirectSliceAndPtr(containerField.Field.Type))
	}
	middleModel := t.Toy.GetModel(LoopTypeIndirect(reflect.TypeOf(middleStruct)))
	relationField, subRelationField := middleModel.fieldSelect(relation), middleModel.fieldSelect(subRelation)
	preload := t.Toy.ManyToManyPreloadBind(t.model, subModel, middleModel, containerField, relationField, subRelationField)
	if preload == nil {
		panic(fmt.Sprintf("invalid preload field '%s'", containerField.Field.Name))
	}

	newSubt := NewToyBrick(t.Toy, subModel).CopyStatus(t)
	newt := *t
	newt.MapPreloadBrick = t.CopyMapPreloadBrick()
	newt.MapPreloadBrick[containerField] = newSubt
	newSubt.relationship = ToyBrickRelationship{&newt, containerField}

	newt.ManyToManyPreload = t.CopyManyToManyPreload()
	newt.ManyToManyPreload[containerField] = preload
	return newSubt
}

func (t *ToyBrick) BindFields(mode string, args ...interface{}) *ToyBrick {
	return t.Scope(func(t *ToyBrick) *ToyBrick {
		var fields []*ModelField
		for _, v := range args {
			fields = append(fields, t.model.fieldSelect(v))
		}
		newt := *t
		newt.FieldsSelector = map[Mode][]*ModelField{}
		for k, v := range t.FieldsSelector {
			newt.FieldsSelector[k] = v
		}
		newt.FieldsSelector[Mode(mode)] = fields
		return &newt
	})
}
func (t *ToyBrick) BindDefaultFields(args ...interface{}) *ToyBrick {
	return t.Scope(func(t *ToyBrick) *ToyBrick {
		var fields []*ModelField
		for _, v := range args {
			fields = append(fields, t.model.fieldSelect(v))
		}
		newt := *t
		newt.FieldsSelector = map[Mode][]*ModelField{}
		for k, v := range t.FieldsSelector {
			newt.FieldsSelector[k] = v
		}
		newt.FieldsSelector[ModeDefault] = fields
		return &newt
	})
}

func (t *ToyBrick) bindDefaultFields(fields ...*ModelField) *ToyBrick {
	return t.Scope(func(t *ToyBrick) *ToyBrick {
		newt := *t
		newt.FieldsSelector = map[Mode][]*ModelField{}
		for k, v := range t.FieldsSelector {
			newt.FieldsSelector[k] = v
		}
		newt.FieldsSelector[ModeDefault] = fields
		return &newt
	})
}

func (t *ToyBrick) bindFields(mode Mode, fields ...*ModelField) *ToyBrick {
	return t.Scope(func(t *ToyBrick) *ToyBrick {
		newt := *t
		newt.FieldsSelector = map[Mode][]*ModelField{}
		for k, v := range t.FieldsSelector {
			newt.FieldsSelector[k] = v
		}
		newt.FieldsSelector[Mode(mode)] = fields
		return &newt
	})
}

func (t *ToyBrick) condition(expr SearchExpr, v ...interface{}) SearchList {
	search := SearchList{}
	switch expr {
	case ExprAnd, ExprOr:
		value := v[0]
		vValue := LoopIndirect(reflect.ValueOf(value))
		switch vValue.Kind() {
		case reflect.Struct:
			record := NewStructRecord(t.model, vValue)
			pairs := t.getFieldValuePairWithRecord(ModeCondition, record)
			for _, pair := range pairs {
				search = search.Condition(pair.Field, pair.Value.Interface(), ExprEqual, expr)
			}
		default:
			if m, ok := vValue.Interface().(map[string]interface{}); ok {
				for name, mField := range t.model.NameFields {
					if iface, ok := m[name]; ok {
						search = search.Condition(mField, iface, ExprEqual, expr)
					}
				}
			} else {
				panic("and/or expr must have a struct/map args")
			}
		}
	default:
		var key, value interface{}
		if len(v) == 2 {
			key, value = v[0], v[1]
		} else if len(v) == 1 {
			key = v[0]
		} else {
			panic("error number args")
		}
		kidx := t.model.fieldSelect(key)
		search = search.Condition(kidx, value, expr, ExprAnd)
	}
	return search
}

// where will clean old condition
func (t *ToyBrick) Where(expr SearchExpr, v ...interface{}) *ToyBrick {
	return t.Scope(func(t *ToyBrick) *ToyBrick {
		newt := *t
		newt.Search = t.condition(expr, v...)
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
			newt.orderBy = append(newt.orderBy, newt.model.fieldSelect(v))
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

func (t *ToyBrick) IgnoreMode(s string, ignore IgnoreMode) *ToyBrick {
	newt := *t
	newt.ignoreModeSelector = map[Mode]IgnoreMode{}
	for k, v := range t.ignoreModeSelector {
		newt.ignoreModeSelector[k] = v
	}
	newt.ignoreModeSelector[Mode(s)] = ignore
	return &newt
}

func (t *ToyBrick) GetContext(option string, records ModelRecords) *Context {
	handlers := t.Toy.ModelHandlers(option, t.model)
	ctx := NewContext(handlers, t, records)
	//ctx.Next()
	return ctx
}

func (t *ToyBrick) insert(records ModelRecords) (*Result, error) {
	handlers := t.Toy.ModelHandlers("Insert", t.model)
	ctx := NewContext(handlers, t, records)
	return ctx.Result, ctx.Next()
}

func (t *ToyBrick) save(records ModelRecords) (*Result, error) {
	handlers := t.Toy.ModelHandlers("Save", t.model)
	ctx := NewContext(handlers, t, records)
	return ctx.Result, ctx.Next()
}

func (t *ToyBrick) deleteWithPrimaryKey(records ModelRecords) (*Result, error) {
	var primaryKeys []interface{}
	primaryField := t.model.GetOnePrimary()
	for _, record := range records.GetRecords() {
		primaryKeys = append(primaryKeys, record.Field(primaryField).Interface())
	}
	newt := t.Where(ExprIn, primaryField, primaryKeys).And().Conditions(t.Search)
	return newt.delete(records)
}

func (t *ToyBrick) delete(records ModelRecords) (*Result, error) {
	if field := t.model.GetFieldWithName("DeletedAt"); field != nil {
		return t.softDelete(records)
	} else {
		return t.hardDelete(records)
	}
}

func (t *ToyBrick) softDelete(records ModelRecords) (*Result, error) {
	handlers := t.Toy.ModelHandlers("SoftDelete", t.model)
	ctx := NewContext(handlers, t, records)
	return ctx.Result, ctx.Next()
}

func (t *ToyBrick) hardDelete(records ModelRecords) (*Result, error) {
	handlers := t.Toy.ModelHandlers("HardDelete", t.model)
	ctx := NewContext(handlers, t, records)
	return ctx.Result, ctx.Next()
}

func (t *ToyBrick) find(value reflect.Value) (*Context, error) {
	handlers := t.Toy.ModelHandlers("Find", t.model)

	if value.Kind() == reflect.Slice {
		records := NewRecords(t.model, value)
		ctx := NewContext(handlers, t, records)
		return ctx, ctx.Next()
	} else {
		vList := reflect.New(reflect.SliceOf(value.Type())).Elem()
		records := NewRecords(t.model, vList)
		ctx := NewContext(handlers, t.Limit(1), records)
		err := ctx.Next()
		if vList.Len() == 0 {
			if err == nil {
				err = errors.New("record not found")
			}
			return ctx, err
		}
		value.Set(vList.Index(0))
		return ctx, err
	}
}

func (t *ToyBrick) CreateTable() (*Result, error) {
	ctx := t.GetContext("CreateTable", MakeRecordsWithElem(t.model, t.model.ReflectType))
	return ctx.Result, ctx.Next()
}

func (t *ToyBrick) CreateTableIfNotExist() (*Result, error) {
	ctx := t.GetContext("CreateTableIfNotExist", MakeRecordsWithElem(t.model, t.model.ReflectType))
	return ctx.Result, ctx.Next()
}

func (t *ToyBrick) DropTable() (*Result, error) {
	ctx := t.GetContext("DropTable", MakeRecordsWithElem(t.model, t.model.ReflectType))
	return ctx.Result, ctx.Next()
}

func (t *ToyBrick) DropTableIfExist() (*Result, error) {
	ctx := t.GetContext("DropTableIfExist", MakeRecordsWithElem(t.model, t.model.ReflectType))
	return ctx.Result, ctx.Next()
}

func (t *ToyBrick) HasTable() (b bool, err error) {
	exec := t.HasTableExec(t.Toy.Dialect)
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
		records = NewRecords(t.model, vValue)
		return t.insert(records)
	default:
		records := MakeRecordsWithElem(t.model, vValue.Addr().Type())
		records.Add(vValue.Addr())
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
	handlers := t.Toy.ModelHandlers("Update", t.model)
	ctx := NewContext(handlers, t, NewRecords(t.model, vValueList))
	return ctx.Result, ctx.Next()
}

func (t *ToyBrick) Save(v interface{}) (*Result, error) {
	vValue := LoopIndirect(reflect.ValueOf(v))

	switch vValue.Kind() {
	case reflect.Slice:
		records := NewRecords(t.model, vValue)
		return t.save(records)
	default:
		records := MakeRecordsWithElem(t.model, vValue.Addr().Type())
		records.Add(vValue.Addr())
		return t.save(records)
	}
}

func (t *ToyBrick) Delete(v interface{}) (*Result, error) {
	vValue := LoopIndirect(reflect.ValueOf(v))
	var records ModelRecords
	switch vValue.Kind() {
	case reflect.Slice:
		records = NewRecords(t.model, vValue)
		return t.deleteWithPrimaryKey(records)
	default:
		records = MakeRecordsWithElem(t.model, vValue.Addr().Type())
		records.Add(vValue.Addr())
		return t.deleteWithPrimaryKey(records)
	}
}

func (t *ToyBrick) DeleteWithConditions() (*Result, error) {
	return t.delete(nil)
}

func (t *ToyBrick) Exec(exec ExecValue) (result sql.Result, err error) {
	if t.tx == nil {
		result, err = t.Toy.db.Exec(exec.Query, exec.Args...)
	} else {
		result, err = t.tx.Exec(exec.Query, exec.Args...)
	}

	if t.debug {
		if err != nil {
			fmt.Fprintf(t.Toy.Logger, "use tx: %p, query:%s, args:%s faiure reason %s\n", t.tx, exec.Query, exec.JsonArgs(), err)
		} else {
			fmt.Fprintf(t.Toy.Logger, "use tx: %p, query:%s, args:%s\n", t.tx, exec.Query, exec.JsonArgs())
		}
	}
	return
}

func (t *ToyBrick) Query(exec ExecValue) (rows *sql.Rows, err error) {
	if t.tx == nil {
		rows, err = t.Toy.db.Query(exec.Query, exec.Args...)
	} else {
		rows, err = t.tx.Query(exec.Query, exec.Args...)
	}
	if t.debug {
		if err != nil {
			fmt.Fprintf(t.Toy.Logger, "use tx: %p, query:%s, args:%s faiure reason %s\n", t.tx, exec.Query, exec.JsonArgs(), err)
		} else {
			fmt.Fprintf(t.Toy.Logger, "use tx: %p, query:%s, args:%s\n", t.tx, exec.Query, exec.JsonArgs())
		}
	}
	return
}

func (t *ToyBrick) QueryRow(exec ExecValue) (row *sql.Row) {
	if t.tx == nil {
		row = t.Toy.db.QueryRow(exec.Query, exec.Args...)
	} else {
		row = t.tx.QueryRow(exec.Query, exec.Args...)
	}
	if t.debug {

		fmt.Fprintf(t.Toy.Logger, "use tx: %p, query:%s, args:%s\n", t.tx, exec.Query, exec.JsonArgs())
	}
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
	if t.debug {
		if err != nil {
			fmt.Fprintf(t.Toy.Logger, "use tx: %p, stmt query:%s, error: %s\n", t.tx, query, err)
		} else {
			fmt.Fprintf(t.Toy.Logger, "use tx: %p, stmt query:%s\n", t.tx, query)
		}
	}
	return stmt, err
}

// TODO all exec method move to dialect
func (t *ToyBrick) DropTableExec() (exec ExecValue) {
	return ExecValue{fmt.Sprintf("DROP TABLE %s", t.model.Name), nil}
}

func (t *ToyBrick) CreateTableExec(dia Dialect) (execlist []ExecValue) {
	return dia.CreateTable(t.model)
}

func (t *ToyBrick) HasTableExec(dialect Dialect) (exec ExecValue) {
	return dialect.HasTable(t.model)
}

func (t *ToyBrick) CountExec() (exec ExecValue) {
	exec.Query = fmt.Sprintf("SELECT count(*) FROM %s", t.model.Name)
	cExec := t.ConditionExec()
	exec.Query += " " + cExec.Query
	exec.Args = append(exec.Args, cExec.Args...)
	return
}

func (s *ToyBrick) ConditionExec() (exec ExecValue) {
	if len(s.Search) > 0 {
		searchExec := s.Search.ToExecValue()
		exec.Query += "WHERE " + searchExec.Query
		exec.Args = append(exec.Args, searchExec.Args...)
	}
	if s.limit != 0 {
		exec.Query += fmt.Sprintf(" LIMIT %d", s.limit)
	}
	if s.offset != 0 {
		exec.Query += fmt.Sprintf(" OFFSET %d", s.offset)
	}
	if len(s.orderBy) > 0 {
		exec.Query += fmt.Sprintf(" ORDER BY ")
		__list := []string{}
		for _, field := range s.orderBy {
			__list = append(__list, field.Name)
		}
		exec.Query += strings.Join(__list, ",")
	}
	return
}

func (t *ToyBrick) FindExec(records ModelRecords) (exec ExecValue) {
	var _list []string
	for _, mField := range t.getSelectFields(records) {
		_list = append(_list, mField.Name)
	}
	exec.Query = fmt.Sprintf("SELECT %s FROM %s", strings.Join(_list, ","), t.model.Name)
	cExec := t.ConditionExec()
	exec.Query += " " + cExec.Query
	exec.Args = append(exec.Args, cExec.Args...)
	return
}

func (t *ToyBrick) UpdateExec(record ModelRecord) (exec ExecValue) {
	var recordList []string
	pairs := t.getFieldValuePairWithRecord(ModeUpdate, record)
	for _, pair := range pairs {
		recordList = append(recordList, pair.Field.Name+"=?")
		exec.Args = append(exec.Args, pair.Value.Interface())
	}
	exec.Query = fmt.Sprintf("UPDATE %s SET %s", t.model.Name, strings.Join(recordList, ","))
	cExec := t.ConditionExec()
	exec.Query += " " + cExec.Query
	exec.Args = append(exec.Args, cExec.Args...)

	return
}

func (t *ToyBrick) DeleteExec() (exec ExecValue) {
	exec.Query = fmt.Sprintf("DELETE FROM %s", t.model.Name)
	cExec := t.ConditionExec()
	exec.Query += " " + cExec.Query
	exec.Args = append(exec.Args, cExec.Args...)
	return
}

func (t *ToyBrick) InsertExec(record ModelRecord) (exec ExecValue) {
	fieldStr := ""
	qStr := ""
	exec = ExecValue{}
	{
		__list := []string{}
		__qlist := []string{}

		pairs := t.getFieldValuePairWithRecord(ModeInsert, record)
		for _, pair := range pairs {
			__list = append(__list, pair.Field.Name)
			__qlist = append(__qlist, "?")
			exec.Args = append(exec.Args, pair.Value.Interface())
		}
		fieldStr += strings.Join(__list, ",")
		qStr += strings.Join(__qlist, ",")
	}
	exec.Query = fmt.Sprintf("INSERT INTO %s(%s) VALUES(%s)", t.model.Name, fieldStr, qStr)
	cExec := t.ConditionExec()
	exec.Query += " " + cExec.Query
	exec.Args = append(exec.Args, cExec.Args...)
	return
}

func (t *ToyBrick) ReplaceExec(record ModelRecord) (exec ExecValue) {
	fieldStr := ""
	qStr := ""
	exec = ExecValue{}
	{
		__list := []string{}
		__qlist := []string{}

		pairs := t.getFieldValuePairWithRecord(ModeReplace, record)
		for _, pair := range pairs {
			__list = append(__list, pair.Field.Name)
			__qlist = append(__qlist, "?")
			exec.Args = append(exec.Args, pair.Value.Interface())
		}
		fieldStr += strings.Join(__list, ",")
		qStr += strings.Join(__qlist, ",")
	}
	exec.Query = fmt.Sprintf("REPLACE INTO %s(%s) VALUES(%s)", t.model.Name, fieldStr, qStr)
	cExec := t.ConditionExec()
	exec.Query += " " + cExec.Query
	exec.Args = append(exec.Args, cExec.Args...)
	return
}

func (t *ToyBrick) getFieldValuePairWithRecord(mode Mode, record ModelRecord) []struct {
	Field *ModelField
	Value reflect.Value
} {
	var fields []*ModelField
	if len(t.FieldsSelector[mode]) > 0 {
		fields = t.FieldsSelector[mode]
	} else if len(t.FieldsSelector[ModeDefault]) > 0 {
		fields = t.FieldsSelector[ModeDefault]
	}

	var useIgnoreMode bool
	if len(fields) == 0 {
		fields = t.model.SqlFields
		useIgnoreMode = record.IsVariableContainer() == false
	}
	var pairs []struct {
		Field *ModelField
		Value reflect.Value
	}
	if useIgnoreMode {
		for _, mField := range fields {
			if fieldValue := record.Field(mField); fieldValue.IsValid() {
				if t.ignoreModeSelector[mode].Ignore(fieldValue) == false {
					// TODO have a special mode is when primary key is zero, need to ignore
					if mField.PrimaryKey && IsZero(fieldValue) {

					} else {
						pairs = append(pairs, struct {
							Field *ModelField
							Value reflect.Value
						}{mField, fieldValue})
					}
				}
			}
		}
	} else {
		for _, mField := range fields {
			if fieldValue := record.Field(mField); fieldValue.IsValid() {
				pairs = append(pairs, struct {
					Field *ModelField
					Value reflect.Value
				}{mField, fieldValue})
			}
		}
	}
	return pairs
}

func (t *ToyBrick) getSelectFields(records ModelRecordFieldTypes) []*ModelField {
	var fields []*ModelField
	if len(t.FieldsSelector[ModeSelect]) > 0 {
		fields = t.FieldsSelector[ModeSelect]
	} else if len(t.FieldsSelector[ModeDefault]) > 0 {
		fields = t.FieldsSelector[ModeDefault]
	} else {
		fields = t.model.SqlFields
	}
	return getFieldsWithRecords(fields, records)
}

func (t *ToyBrick) getScanFields(records ModelRecordFieldTypes) []*ModelField {
	var fields []*ModelField
	if len(t.FieldsSelector[ModeScan]) > 0 {
		fields = t.FieldsSelector[ModeScan]
	} else if len(t.FieldsSelector[ModeDefault]) > 0 {
		fields = t.FieldsSelector[ModeDefault]
	} else {
		fields = t.model.SqlFields
	}
	return getFieldsWithRecords(fields, records)
}

// use for order by
func (t *ToyBrick) ToDesc(v interface{}) *ModelField {
	field := t.model.fieldSelect(v)
	newField := *field
	newField.Name = field.Name + " DESC"
	return &newField
}
