/*
 * Copyright 2018. bigpigeon. All rights reserved.
 * Use of this source code is governed by a MIT style
 * license that can be found in the LICENSE file.
 */

package toyorm

import (
	"bytes"
	"database/sql"
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
	if v.Kind() == reflect.Struct && v.Type().Comparable() == false {
		return reflect.Zero(v.Type()).String() == v.String()
	} else {
		return v.Interface() == reflect.Zero(v.Type()).Interface()
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

// placeholder format:
// $Name   ok
// $name   ok
// $1Name  ok
// $User-Name  ok
// $user_name  ok
// $User-  no, the placeholder are interception as $User
// $user_name_ no, the placeholder are interception as $user_name
// $-   error, placeholder is null
func getTemplateExec(exec BasicExec, execs map[string]BasicExec) (BasicExec, error) {
	buff := bytes.Buffer{}
	var pre, i int
	var args []interface{}
	isEscaping := false
	qNum := 0
	for i < len(exec.query) {
		switch exec.query[i] {
		case '$':
			if isEscaping == false {
				buff.WriteString(exec.query[pre:i])
				i++
				pre = i
				end := i
				for i < len(exec.query) {
					c := exec.query[i]
					i++
					if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') {
						end = i
					} else if c == '-' || c == '_' {

					} else {
						break
					}
				}
				word := exec.query[pre:end]
				if word == "" {
					return BasicExec{}, ErrTemplateExecInvalidWord{"$"}
				} else if match, ok := execs[word]; ok {
					buff.WriteString(match.query)
					args = append(args, match.args...)
				} else {
					return BasicExec{}, ErrTemplateExecInvalidWord{"$" + word}
				}

				pre, i = end, end
			} else {
				buff.WriteString(exec.query[pre : i-1])
				buff.WriteByte(exec.query[i])

				pre, i = i+1, i+1
			}
		case '?':
			args = append(args, exec.args[qNum])
			qNum++
			i++
		case '\\':
			i++
			isEscaping = true
			continue
		default:
			i++
		}
		isEscaping = false
	}
	args = append(args, exec.args[qNum:]...)

	buff.WriteString(exec.query[pre:i])
	return BasicExec{buff.String(), args}, nil
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

func insertValuesFormat(model *Model, columnValues []ColumnNameValue) (string, string, []interface{}) {
	valueMap := map[string]reflect.Value{}
	for _, c := range columnValues {
		valueMap[c.Name()] = c.Value()
	}
	fieldBuff := bytes.Buffer{}
	qBytes := make([]byte, 0, len(model.GetSqlFields())*2)
	args := make([]interface{}, 0, len(model.GetSqlFields()))

	for _, r := range model.GetSqlFields() {
		var val reflect.Value
		if v, ok := valueMap[r.Name()]; ok {
			val = v
		} else if r.AutoIncrement() || r.IsPrimary() {
			continue
		} else {
			val = reflect.Zero(r.StructField().Type)
		}
		fieldBuff.WriteString(r.Column())
		fieldBuff.WriteByte(',')
		qBytes = append(qBytes, '?', ',')
		args = append(args, val.Interface())
	}
	fieldBytes := fieldBuff.Bytes()
	// last buff must be ,
	fieldBytes = fieldBytes[:len(fieldBytes)-1]
	qBytes = qBytes[:len(qBytes)-1]
	return string(fieldBytes), string(qBytes), args
}
