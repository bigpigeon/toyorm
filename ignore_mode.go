package toyorm

import (
	"reflect"
)

type IgnoreMode int

const (
	IgnoreFalse = IgnoreMode(1 << iota)
	IgnoreZeroInt
	IgnoreZeroFloat
	IgnoreZeroComplex
	IgnoreNilString
	IgnoreNilPoint
	IgnoreZeroLen
	IgnoreNullStruct
	IgnoreNil  = IgnoreNilPoint | IgnoreZeroLen
	IgnoreZero = IgnoreFalse | IgnoreZeroInt | IgnoreZeroFloat | IgnoreZeroComplex | IgnoreNilString | IgnoreNilPoint | IgnoreZeroLen | IgnoreNullStruct
)

func (ignoreMode IgnoreMode) Ignore(v reflect.Value) (ignore bool) {
	switch v.Kind() {
	case reflect.Bool:
		ignore = (ignoreMode&IgnoreFalse) != 0 && IsZero(v)
	case reflect.Uintptr, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		ignore = (ignoreMode&IgnoreZeroInt) != 0 && IsZero(v)
	case reflect.Float32, reflect.Float64:
		ignore = (ignoreMode&IgnoreZeroFloat) != 0 && IsZero(v)
	case reflect.Complex64, reflect.Complex128:
		ignore = (ignoreMode&IgnoreZeroComplex) != 0 && IsZero(v)
	case reflect.String:
		ignore = (ignoreMode&IgnoreNilString) != 0 && IsZero(v)
	case reflect.Ptr, reflect.Interface:
		ignore = (ignoreMode&IgnoreNilPoint) != 0 && IsZero(v)
	case reflect.Array:
		ignore = (ignoreMode&IgnoreZeroLen) != 0 && IsZero(v)
	case reflect.Map:
		ignore = ((ignoreMode&IgnoreNilPoint) != 0 && v.IsNil()) || ((ignoreMode&IgnoreZeroLen) != 0 && v.Len() == 0)
	case reflect.Slice:
		ignore = ((ignoreMode&IgnoreNilPoint) != 0 && v.IsNil()) || ((ignoreMode&IgnoreZeroLen) != 0 && v.Len() == 0)
	case reflect.Struct:
		ignore = (ignoreMode&IgnoreNullStruct) != 0 && IsZero(v)
	}
	return
}
