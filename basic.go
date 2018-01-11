package toyorm

import (
	"database/sql"
	"fmt"
	"reflect"
)

type NilType struct{}

type QueryFunc func(string, ...interface{}) (*sql.Rows, error)

/*
Name convert rule:
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

type SelectField struct {
	ModelField
	Value interface{}
}

func LoopTypeIndirect(_type reflect.Type) reflect.Type {
	for _type.Kind() == reflect.Ptr {
		_type = _type.Elem()
	}
	return _type
}

func LoopTypeIndirectSliceAndPtr(_type reflect.Type) reflect.Type {
	for _type.Kind() == reflect.Ptr || _type.Kind() == reflect.Slice {
		_type = _type.Elem()
	}
	return _type
}

func LoopIndirect(vValue reflect.Value) reflect.Value {
	for vValue.Kind() == reflect.Ptr {
		vValue = vValue.Elem()
	}
	return vValue
}

func LoopIndirectAndNew(vValue reflect.Value) reflect.Value {
	for vValue.Kind() == reflect.Ptr {
		if vValue.IsNil() {
			vValue.Set(reflect.New(vValue.Type().Elem()))
		}
		vValue = vValue.Elem()
	}
	return vValue
}

func IsZero(v reflect.Value) bool {
	return v.Interface() == reflect.Zero(v.Type()).Interface()
}

func SafeMapSet(m, k, x reflect.Value) {
	eType, xType := m.Type().Elem(), x.Type()
	if eType != xType {
		m.SetMapIndex(k, x.Convert(eType))
	} else {
		m.SetMapIndex(k, x)
	}
}

func SafeSet(v, x reflect.Value) {
	vType, xType := v.Type(), x.Type()
	if vType != xType {
		v.Set(x.Convert(vType))
	} else {
		v.Set(x)
	}
}

//TODO performance optimize
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

//func GetBindFields(model *Model, _type reflect.Type) []*ModelField {
//	var fields []*ModelField
//	if _type == model.ReflectType {
//		for _, mField := range model.SqlFields {
//			fields = append(fields, mField)
//		}
//	} else if _type.Kind() == reflect.Struct {
//		structFieldList := GetStructFields(_type)
//		for i := 0; i < len(structFieldList); i++ {
//			field := structFieldList[i]
//			idx, ok := model.NameFields[field.Name]
//			if ok && model.GetPosField(i).Ignore == false {
//				// process with preload or field
//				fields = append(fields, idx)
//			}
//		}
//	} else {
//		panic("cannot bind fields with a not struct type")
//	}
//	return fields
//}

// generate a default field name with relation model
func GetRelationFieldName(subModel *Model) string {
	return subModel.ReflectType.Name() + subModel.GetOnePrimary().Field.Name
}

func GetBelongsIDFieldName(subModel *Model, containerField *ModelField) string {
	return containerField.Field.Name + subModel.GetOnePrimary().Field.Name
}

func GetMiddleField(model, middleModel *Model, leftOrRight bool) *ModelField {
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

func getFieldsWithRecords(fields []*ModelField, records ModelRecordFieldTypes) []*ModelField {
	var selectFields []*ModelField
	for _, field := range fields {
		if _type := records.GetFieldType(field); _type != nil {
			selectFields = append(selectFields, field)
		}
	}
	return selectFields
}

// use for insert/update/
func getModelFieldAndValues(ignoreMode IgnoreMode, record ModelRecord, currentFields, defaultFields, sqlFields []*ModelField) []struct {
	Field *ModelField
	Value reflect.Value
} {
	var fields []*ModelField
	if len(currentFields) > 0 {
		fields = defaultFields
	} else if len(defaultFields) > 0 {
		fields = defaultFields
	}

	var useIgnoreMode bool
	if len(fields) == 0 {
		fields = sqlFields
		useIgnoreMode = record.IsVariableContainer() == false
	}
	var pairs []struct {
		Field *ModelField
		Value reflect.Value
	}
	if useIgnoreMode {
		for _, mField := range fields {
			if fieldValue := record.Field(mField); fieldValue.IsValid() {
				if ignoreMode.Ignore(fieldValue) == false {
					pairs = append(pairs, struct {
						Field *ModelField
						Value reflect.Value
					}{mField, fieldValue})
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
