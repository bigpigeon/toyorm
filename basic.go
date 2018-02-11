package toyorm

import (
	"database/sql"
	"fmt"
	"reflect"
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

// get all Value with struct field and its embed struct field
func GetStructValueFields(value reflect.Value) []reflect.Value {
	vtype := value.Type()
	fieldList := []reflect.Value{}
	for i := 0; i < value.NumField(); i++ {
		vfield := value.Field(i)
		sfield := vtype.Field(i)
		if sfield.Anonymous {
			embedFieldList := GetStructValueFields(vfield)
			fieldList = append(fieldList, embedFieldList...)
		} else {
			fieldList = append(fieldList, vfield)
		}
	}
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

// to check value is zero
func IsZero(v reflect.Value) bool {
	if v.Kind() == reflect.Struct && v.Type().Comparable() == false {
		return reflect.Zero(v.Type()).String() == v.String()
	} else {
		return v.Interface() == reflect.Zero(v.Type()).Interface()
	}
}

// if map value type is different with current value type ,try to convert it
func SafeMapSet(m, k, x reflect.Value) {
	eType, xType := m.Type().Elem(), x.Type()
	if eType != xType {
		m.SetMapIndex(k, x.Convert(eType))
	} else {
		m.SetMapIndex(k, x)
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
func ModelName(_type reflect.Type) string {
	var modelName string
	if v, ok := reflect.New(_type).Interface().(tabler); ok {
		modelName = v.TableName()
	} else {
		modelName = SqlNameConvert(_type.Name())
	}
	return modelName
}
