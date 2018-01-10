package toyorm

import (
	"reflect"
)

type Dialect interface {
	HasTable(*Model) ExecValue
	CreateTable(*Model) []ExecValue
	ToSqlType(reflect.Type) string
}

type DefaultDialect struct{}
