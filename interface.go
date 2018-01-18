package toyorm

import (
	"reflect"
)

type Column interface {
	Column() string
}

type ColumnValue interface {
	Column
	Value() reflect.Value
}
