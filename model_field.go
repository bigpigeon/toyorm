package toyorm

import (
	"fmt"
	"reflect"
	"strings"
)

type Field interface {
	Column
	IsPrimary() bool
	Index() string
	UniqueIndex() string
	Type() reflect.Type
	Attr(string) string
	SqlType() string
}

type ModelField struct {
	Name          string
	sqlType       string
	Offset        uintptr
	PrimaryKey    bool
	Index         string
	UniqueIndex   string
	Ignore        bool
	CommonAttr    map[string]string
	AutoIncrement bool
	Field         reflect.StructField
}

func (m *ModelField) Column() string {
	return m.Name
}

func (m *ModelField) IsPrimary() bool {
	return m.PrimaryKey
}

func NewField(f *reflect.StructField, table_name string) *ModelField {
	field := &ModelField{
		Field:      *f,
		CommonAttr: map[string]string{},
		Name:       SqlNameConvert(f.Name),
		Offset:     f.Offset,
		sqlType:    ToSqlType(f.Type),
	}

	// set attribute by tag
	tags := strings.Split(f.Tag.Get("toyorm"), ";")
	for _, t := range tags {
		keyval := strings.Split(t, ":")
		var key, val string
		switch len(keyval) {
		case 0:
			continue
		case 1:
			key = keyval[0]
		case 2:
			key, val = keyval[0], keyval[1]
		default:
			panic(ErrInvalidTag)
		}
		switch strings.ToLower(key) {
		case "auto_increment", "autoincrement":
			field.AutoIncrement = true
		case "primary key":
			field.PrimaryKey = true
		case "type":
			field.sqlType = val
		case "index":
			if val == "" {
				field.Index = fmt.Sprintf("idx_%s_%s", table_name, strings.ToLower(f.Name))
			} else {
				field.Index = val
			}
		case "unique index":
			if val == "" {
				field.UniqueIndex = fmt.Sprintf("udx_%s_%s", table_name, strings.ToLower(f.Name))
			} else {
				field.UniqueIndex = val
			}
		case "column":
			field.Name = val
		case "-":
			field.Ignore = true
		default:
			field.CommonAttr[key] = val
		}
	}
	if field.Name == "" || field.sqlType == "" {
		field.Ignore = true
	}
	return field
}
