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
	"sort"
	"strings"
	"time"
)

/*
column convert rule:
User => user
UserName => user_name
UserID => user_id
OneToOne => one_to_one
*/
func SqlNameConvert(name string) string {
	if len(name) == 0 {
		panic("error length name string")
	}
	convert := []byte{}

	var lowerCount, upperCount uint32
	for i := 0; i < len(name); i++ {
		a := name[i]
		if a >= 'A' && a <= 'Z' {
			if lowerCount >= 1 {
				convert = append(convert, '_', a-'A'+'a')
			} else {
				convert = append(convert, a-'A'+'a')
			}
			upperCount++
			lowerCount = 0
		} else if a >= 'a' && a <= 'z' {
			if upperCount > 1 {
				convert = append(convert, '_', a)
			} else {
				convert = append(convert, a)
			}
			upperCount = 0
			lowerCount++
		} else if a == '.' {
			lowerCount, upperCount = 0, 0
			convert = append(convert, '_')
		} else {
			lowerCount, upperCount = 0, 0
			convert = append(convert, a)
		}
	}
	return string(convert)
}

func getStructFieldLen(vType reflect.Type) int {
	sum := vType.NumField()
	for i := 0; i < vType.NumField(); i++ {
		field := vType.Field(i)
		if field.Anonymous {
			sum += getStructFieldLen(LoopTypeIndirect(field.Type))
		}
	}
	return sum
}

func getStructValueFields(value reflect.Value, fieldList *[]reflect.Value) {
	vtype := value.Type()
	for i := 0; i < value.NumField(); i++ {
		vfield := value.Field(i)
		sfield := vtype.Field(i)
		if sfield.Anonymous {
			getStructValueFields(vfield, fieldList)
		} else {
			*fieldList = append(*fieldList, vfield)
		}
	}
}

// get all Value with struct field and its embed struct field
func GetStructValueFields(value reflect.Value) []reflect.Value {
	vtype := value.Type()
	// opt allocation

	fieldList := make([]reflect.Value, 0, getStructFieldLen(vtype))
	getStructValueFields(value, &fieldList)
	return fieldList
}

// get all StructField with struct field and its embed field
func GetStructFields(_type reflect.Type) []reflect.StructField {
	fieldList := []reflect.StructField{}
	for i := 0; i < _type.NumField(); i++ {
		sfield := _type.Field(i)
		if sfield.Anonymous {
			embedFieldList := GetStructFields(sfield.Type)
			fieldList = append(fieldList, embedFieldList...)
		} else {
			fieldList = append(fieldList, sfield)
		}
	}
	return fieldList
}

// loop to get ptr type elem when its type is not ptr/slice
func LoopTypeIndirect(_type reflect.Type) reflect.Type {
	for _type.Kind() == reflect.Ptr {
		_type = _type.Elem()
	}
	return _type
}

// loop to get ptr/slice type elem when its type is not ptr/slice
func LoopTypeIndirectSliceAndPtr(_type reflect.Type) reflect.Type {
	for _type.Kind() == reflect.Ptr || _type.Kind() == reflect.Slice {
		_type = _type.Elem()
	}
	return _type
}

// loop to get ptr value elem when its type is not ptr
func LoopIndirect(vValue reflect.Value) reflect.Value {
	for vValue.Kind() == reflect.Ptr {
		vValue = vValue.Elem()
	}
	return vValue
}

// loop to get ptr value elem when its type is not ptr and if value is zero, set a new one
func LoopIndirectAndNew(vValue reflect.Value) reflect.Value {
	for vValue.Kind() == reflect.Ptr {
		if vValue.IsNil() {
			vValue.Set(reflect.New(vValue.Type().Elem()))
		}
		vValue = vValue.Elem()
	}
	return vValue
}

func LoopDivePtr(vValue reflect.Value) reflect.Value {
	for vValue.Kind() == reflect.Ptr {
		if vValue.IsNil() {
			vValue = reflect.Zero(vValue.Type().Elem())
		} else {
			vValue = vValue.Elem()
		}
	}

	return vValue
}

// loop to get ptr value elem
// if its type is ptr, get it's elem
// if it's type is slice get it's first elem or zero value's elem type
// loop get will not change the source object
func LoopDiveSliceAndPtr(vValue reflect.Value) reflect.Value {
	for vValue.Kind() == reflect.Ptr || vValue.Kind() == reflect.Slice {
		if vValue.Kind() == reflect.Ptr {
			if vValue.IsNil() {
				vValue = reflect.Zero(vValue.Type().Elem())
			} else {
				vValue = vValue.Elem()
			}
		} else {
			if vValue.Len() > 0 {
				vValue = vValue.Index(0)
			} else {
				vValue = reflect.Zero(vValue.Type().Elem())
			}
		}
	}
	return vValue
}

// to check value is zero
func IsZero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Slice:
		return v.IsNil()
	case reflect.Map:
		return v.IsNil()
	default:
		if v.Type().Comparable() == false {
			return reflect.Zero(v.Type()).String() == v.String()
		} else {
			return v.Interface() == reflect.Zero(v.Type()).Interface()
		}
	}
}

// if value type is different with current value type ,try to convert it
func safeSet(v, x reflect.Value) {
	vType, xType := v.Type(), x.Type()

	if vType != xType {
		// try to convert v

		if vType.Kind() == reflect.Ptr && xType.Kind() != reflect.Ptr {
			v = LoopIndirectAndNew(v)
			vType = v.Type()
		}
		v.Set(x.Convert(vType))
	} else {
		v.Set(x)
	}
}

// if value type is ptr to l elem type , try to get its elem and append to l
// if l elem type is ptr to x type , try to  make x ptr and append to l
func SafeAppend(l reflect.Value, x ...reflect.Value) reflect.Value {
	tPtrElem := l.Type().Elem()
	canAppend := make([]reflect.Value, 0, len(x))
	for _, e := range x {
		if e.Type() == tPtrElem {
			canAppend = append(canAppend, e)
		} else {
			tElem := tPtrElem
			ptrDeep := 0
			for tElem.Kind() == reflect.Ptr {
				tElem = tElem.Elem()
				ptrDeep++
			}
			e = LoopIndirect(e)
			if tElem == e.Type() {
				ePtr := e
				for ptrDeep != 0 {
					ptrDeep--
					if e.CanAddr() {
						ePtr = e.Addr()
					} else {
						ePtr = reflect.New(e.Type())
						ePtr.Elem().Set(e)
					}
				}
				canAppend = append(canAppend, ePtr)
			} else {
				panic(fmt.Sprintf("cannot assignable %s to %s", e.Type(), tElem))
			}
		}
	}
	return reflect.Append(l, canAppend...)
}

// generate a default field name with relation model
func GetRelationFieldName(subModel *Model) string {
	return subModel.ReflectType.Name() + subModel.GetOnePrimary().Name()
}

func GetBelongsIDFieldName(subModel *Model, containerField Field) string {
	return containerField.Name() + subModel.GetOnePrimary().Name()
}

func GetMiddleField(model, middleModel *Model, leftOrRight bool) Field {
	// try to find field with name
	if modelField := middleModel.GetFieldWithName(GetRelationFieldName(model)); modelField != nil {
		return modelField
	}
	b2i := map[bool]int{false: 0, true: 1}
	return middleModel.GetPosField(b2i[leftOrRight])
}

func makeRange(min, max int) (l []int) {
	for min < max {
		l = append(l, min)
		min++
	}
	return
}

func getFieldsWithRecords(fields []Field, records ModelRecordFieldTypes) []Field {
	var selectFields []Field
	for _, field := range fields {
		if _type := records.GetFieldType(field.Name()); _type != nil {
			selectFields = append(selectFields, field)
		}
	}
	return selectFields
}

func ToSqlType(_type reflect.Type) (sqlType string) {
	switch _type.Kind() {
	case reflect.Ptr:
		sqlType = ToSqlType(_type.Elem())
	case reflect.Bool:
		sqlType = "BOOLEAN"
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Uint8, reflect.Uint16, reflect.Uint32:
		sqlType = "INTEGER"
	case reflect.Int64, reflect.Uint64, reflect.Int, reflect.Uint:
		sqlType = "BIGINT"
	case reflect.Float32, reflect.Float64:
		sqlType = "FLOAT"
	case reflect.String:
		sqlType = "VARCHAR(255)"
	case reflect.Struct:
		if _, ok := reflect.New(_type).Elem().Interface().(time.Time); ok {
			sqlType = "TIMESTAMP"
		} else if _, ok := reflect.New(_type).Elem().Interface().(sql.NullBool); ok {
			sqlType = "BOOLEAN"
		} else if _, ok := reflect.New(_type).Elem().Interface().(sql.NullInt64); ok {
			sqlType = "BIGINT"
		} else if _, ok := reflect.New(_type).Elem().Interface().(sql.NullString); ok {
			sqlType = "VARCHAR(255)"
		} else if _, ok := reflect.New(_type).Elem().Interface().(sql.NullFloat64); ok {
			sqlType = "FLOAT"
		} else if _, ok := reflect.New(_type).Elem().Interface().(sql.RawBytes); ok {
			sqlType = "VARCHAR(255)"
		}
	default:
		if _, ok := reflect.New(_type).Elem().Interface().([]byte); ok {
			sqlType = "VARCHAR(255)"
		}
	}
	return
}

// get model name with type
func ModelName(val reflect.Value) string {
	var modelName string

	if v, ok := val.Interface().(tabler); ok {
		modelName = v.TableName()
	} else {
		canModelName := false
		if val.CanAddr() {
			if v, ok := val.Addr().Interface().(tabler); ok {
				modelName = v.TableName()
				canModelName = true
			}
		}
		if canModelName == false {
			if typeName := val.Type().Name(); typeName != "" {
				modelName = SqlNameConvert(val.Type().Name())
			}
		}
	}
	if modelName == "" {
		panic(ErrInvalidModelName{})
	}
	return modelName
}

func DefaultTemplateExec(brick *ToyBrick) map[string]BasicExec {
	cExec := brick.ConditionExec()
	result := map[string]BasicExec{
		"ModelName":  {brick.Model.Name, nil},
		"Conditions": {cExec.Query(), cExec.Args()},
	}
	for _, field := range brick.Model.GetSqlFields() {
		// add field name placeholder exec
		result["FN-"+field.Name()] = BasicExec{field.Column(), nil}
		// add field offset placeholder exec
		result[fmt.Sprintf("0x%x", field.Offset())] = BasicExec{field.Column(), nil}
	}
	return result
}

func DefaultCollectionTemplateExec(brick *CollectionBrick) map[string]BasicExec {
	cExec := brick.ConditionExec()
	result := map[string]BasicExec{
		"ModelName":  {brick.Model.Name, nil},
		"Conditions": {cExec.Query(), cExec.Args()},
		"DBIndex":    {fmt.Sprintf("%d", brick.dbIndex), nil},
	}
	for _, field := range brick.Model.GetSqlFields() {
		// add field name placeholder exec
		result["FN-"+field.Name()] = BasicExec{field.Column(), nil}
		// add field offset placeholder exec
		result[fmt.Sprintf("FO-%d", field.Offset())] = BasicExec{field.Column(), nil}
	}
	return result
}

func columnsValueToColumn(values []ColumnValue) []Column {
	columns := make([]Column, len(values))
	for i := range values {
		columns[i] = values[i]
	}
	return columns
}

func getColumnExec(columns []Column) BasicExec {
	var _list []string
	for _, column := range columns {
		_list = append(_list, column.Column())
	}
	return BasicExec{strings.Join(_list, ","), nil}
}

func getInsertFieldValue(model *Model, values FieldValueList) FieldValueList {
	valMap := map[string]FieldValue{}
	for i, f := range values {
		valMap[f.Name()] = values[i]
	}
	var result FieldValueList

	for _, field := range model.GetSqlFields() {
		var val FieldValue
		if v, ok := valMap[field.Name()]; ok {
			val = v
		} else {
			val = field.ToFieldValue(reflect.Zero(field.StructField().Type))
		}
		if IsZero(val.Value()) {
			if val.IsPrimary() || val.AutoIncrement() || val.Default() != "" {
				continue
			}
		}
		result = append(result, val)
	}
	return result
}

func getSaveArgs(model *Model, columnValues FieldValueList) DialectSaveArgs {
	valueMap := map[string]FieldValue{}
	for _, c := range columnValues {
		valueMap[c.Name()] = c
	}
	args := DialectSaveArgs{}
	for _, field := range model.GetSqlFields() {
		var val FieldValue
		needInsert, needSave := true, true
		if v, ok := valueMap[field.Name()]; ok {
			val = v
		} else {
			needSave = false
			val = field.ToFieldValue(reflect.Zero(field.StructField().Type))
		}
		if val.IsPrimary() {
			args.PrimaryFields = append(args.PrimaryFields, val)
		}
		// insert field detect
		if IsZero(val.Value()) {
			if val.AutoIncrement() || val.IsPrimary() || val.Default() != "" {
				needInsert = false
			}
		}
		if name := field.Name(); name == "CreatedAt" {
			args.CreatedAtField = val
			needSave = false
		}
		if name := field.Name(); name == "UpdatedAt" {
			args.UpdatedAtField = val
		}
		if field.Name() == "Cas" {
			args.CasField = val
		}
		if needInsert {
			args.InsertFieldList = append(args.InsertFieldList, val)
		}
		if needSave {
			args.SaveFieldList = append(args.SaveFieldList, val)
		}
	}
	return args
}

// e.g return BasicExec{query: "?,?,?" args:[1,2,3]}
func getValuesExec(values []ColumnValue) BasicExec {
	var args []interface{}
	var qList []string
	for _, value := range values {
		qList = append(qList, "?")
		args = append(args, value.Value().Interface())
	}

	return BasicExec{strings.Join(qList, ","), args}
}

// e.g return BasicExec{query: "a = ?,b = ?,c = ?" args:[1,2,3]}
func getUpdateValuesExec(values []ColumnValue) BasicExec {
	var args []interface{}
	var _list []string

	for _, value := range values {
		_list = append(_list, value.Column()+" = ?")
		args = append(args, value.Value().Interface())
	}

	return BasicExec{strings.Join(_list, ","), args}
}

func joinSwap(swap *JoinSwap, brick *ToyBrick) *JoinSwap {
	currentJoinSwap := JoinSwap{
		OwnOrderBy:        brick.OwnOrderBy,
		OwnGroupBy:        brick.OwnGroupBy,
		OwnSearch:         brick.OwnSearch,
		Alias:             brick.alias,
		FieldsSelector:    brick.FieldsSelector,
		SwapMap:           brick.SwapMap,
		JoinMap:           brick.JoinMap,
		MapPreloadBrick:   brick.MapPreloadBrick,
		BelongToPreload:   brick.BelongToPreload,
		OneToOnePreload:   brick.OneToOnePreload,
		OneToManyPreload:  brick.OneToManyPreload,
		ManyToManyPreload: brick.ManyToManyPreload,
	}
	if swap != nil {
		brick.OwnOrderBy = swap.OwnOrderBy
		brick.OwnGroupBy = swap.OwnGroupBy
		brick.OwnSearch = swap.OwnSearch
		brick.alias = swap.Alias
		brick.FieldsSelector = swap.FieldsSelector
		brick.SwapMap = swap.SwapMap
		brick.JoinMap = swap.JoinMap
		brick.MapPreloadBrick = swap.MapPreloadBrick
		brick.BelongToPreload = swap.BelongToPreload
		brick.OneToOnePreload = swap.OneToOnePreload
		brick.OneToManyPreload = swap.OneToManyPreload
		brick.ManyToManyPreload = swap.ManyToManyPreload
	}
	return &currentJoinSwap
}

// get columns and scanner generator
func FindColumnFactory(fieldTypes ModelRecordFieldTypes, brick *ToyBrick) ([]Column, func(ModelRecord) []interface{}) {
	columns := brick.getSelectFields(fieldTypes).ToColumnList()
	names := make([]string, 0, len(brick.JoinMap))
	for name := range brick.JoinMap {
		names = append(names, name)
	}
	sort.Strings(names)
	nameFnMap := map[string]func(records ModelRecord) []interface{}{}
	for _, name := range names {
		joinBrick := brick.Join(name)
		subRecord := MakeRecord(brick.JoinMap[name].SubModel, LoopTypeIndirect(fieldTypes.GetFieldType(name)))

		var subColumns []Column
		subColumns, nameFnMap[name] = FindColumnFactory(subRecord, joinBrick)
		columns = append(columns, subColumns...)
	}
	var fn func(records ModelRecord) []interface{}
	fn = func(record ModelRecord) []interface{} {
		var scanners []interface{}
		for _, field := range brick.getScanFields(record) {
			value := record.FieldAddress(field.Name())
			scanners = append(scanners, value.Interface())
		}
		for _, name := range names {
			subRecord := NewRecord(brick.JoinMap[name].SubModel, LoopIndirectAndNew(record.Field(name)))
			scanners = append(scanners, nameFnMap[name](subRecord)...)
		}

		return scanners
	}
	return columns, fn
}

func CollectionFindColumnFactory(fieldTypes ModelRecordFieldTypes, brick *CollectionBrick) ([]Column, func(ModelRecord) []interface{}) {
	columns := brick.getSelectFields(fieldTypes).ToColumnList()

	var fn func(records ModelRecord) []interface{}
	fn = func(record ModelRecord) []interface{} {
		var scanners []interface{}
		for _, field := range brick.getScanFields(record) {
			value := record.FieldAddress(field.Name())
			scanners = append(scanners, value.Interface())
		}
		return scanners
	}
	return columns, fn
}

func IntKind(kind reflect.Kind) bool {
	return kind >= reflect.Int && kind <= reflect.Uint64
}

func GetSaveValues(model *Model, columnValues FieldValueList) (FieldValueList, FieldValue) {
	valueMap := map[string]FieldValue{}
	for _, c := range columnValues {
		valueMap[c.Name()] = c
	}
	var cas FieldValue
	values := make(FieldValueList, 0, len(columnValues))
	for _, r := range model.GetSqlFields() {
		if name := r.Name(); name == "CreatedAt" {
			continue
		}
		if r.Name() == "Cas" {
			cas = valueMap[r.Name()]
		}

		if v, ok := valueMap[r.Name()]; ok {
			if IsZero(v.Value()) {
				if r.AutoIncrement() {
					continue
				}
			}
			values = append(values, v)
		}
	}
	return values, cas
}

func setNumberPrimaryKey(ctx *Context, record ModelRecord, action ExecAction) error {
	// set primary field value if model has one primary key
	if len(ctx.Brick.Model.GetPrimary()) == 1 {
		primaryKey := ctx.Brick.Model.GetOnePrimary()
		primaryKeyName := primaryKey.Name()
		if IntKind(primaryKey.StructField().Type.Kind()) {
			// just set not zero primary key
			if fieldValue := record.Field(primaryKeyName); !fieldValue.IsValid() || IsZero(fieldValue) {
				if lastId, err := action.Result.LastInsertId(); err == nil {
					record.SetField(primaryKeyName, reflect.ValueOf(lastId))
				} else {
					return errors.New(fmt.Sprintf("get (%s) auto increment  failure reason(%s)", ctx.Brick.Model.Name, err))
				}
			}
		}
	}
	return nil
}

type CharStack []byte

func (s *CharStack) Push(b byte) {
	*s = append(*s, b)
}

func (s *CharStack) Pop() byte {
	b := (*s)[len(*s)-1]
	*s = (*s)[:len(*s)-1]
	return b
}

func conditionToReversePolandExprDebug(cond string, execs map[string]BasicExec, debug bool) (CharStack, error) {
	var stack CharStack
	var operationStack CharStack
	okToAsii := map[bool]byte{false: '0', true: '1'}
	type wordNeed bool
	const needOp wordNeed = true
	const needWord wordNeed = false
	need := needWord
	for i := 0; i < len(cond); {
		switch cond[i] {
		case ' ':
			i++
		case '|', '&':
			if need != needOp {
				return nil, ErrInvalidConditionWord{cond[:i+1], cond}
			}
			switch cond[i] {
			case '|':
				endOf := len(operationStack)
				for j := len(operationStack) - 1; j >= 0; j-- {
					if operationStack[j] == '&' {
						stack = append(stack, operationStack[j])
						endOf = j
					} else {
						break
					}
				}
				operationStack = operationStack[:endOf]
				operationStack = append(operationStack, '|')
			case '&':
				operationStack = append(operationStack, '&')
			}
			need = needWord
			i++
		case '(', ')':
			switch cond[i] {
			case '(':
				operationStack = append(operationStack, '(')
			case ')':
				endOf := len(operationStack)
				for j := len(operationStack) - 1; j >= 0; j-- {
					endOf = j
					if operationStack[j] != '(' {
						stack = append(stack, operationStack[j])
					} else {
						break
					}
				}
				operationStack = operationStack[:endOf]
			}
			i++
		default:
			if need != needWord {
				return nil, ErrInvalidConditionWord{cond[:i+1], cond}
			}
			if e := findWord(cond[i:]); e != 0 {
				if debug {
					if e != 1 {
						return nil, errors.New("cannot use more than one character word in debug mode")
					}
					stack = append(stack, cond[i])
				} else {
					_, ok := execs[cond[i:i+e]]
					stack = append(stack, okToAsii[ok])
				}
				i += e
				need = needOp
			} else {
				return nil, ErrInvalidConditionWord{cond[:i+1], cond}
			}
		}

	}
	for j := len(operationStack) - 1; j >= 0; j-- {
		stack = append(stack, operationStack[j])
	}

	return stack, nil
}

//  a | b & c & (d & e | f) | g => a b c d e & f | & & g | |
//  a & (b | c | (d | e)) => a b c d e | | | &
func conditionToReversePolandExpr(cond string, execs map[string]BasicExec) (CharStack, error) {
	return conditionToReversePolandExprDebug(cond, execs, false)
}

func getInsertColumnExecAndValue(model *Model, values FieldValueList) (BasicExec, BasicExec) {
	var _list []string
	var args []interface{}
	var qList []string
	for _, val := range values {
		_list = append(_list, val.Column())
		qList = append(qList, "?")
		args = append(args, val.Value().Interface())
	}
	return BasicExec{strings.Join(_list, ","), nil},
		BasicExec{strings.Join(qList, ","), args}
}
