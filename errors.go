package toyorm

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidTag        = errors.New("invalid tag")
	ErrInvalidSearchTree = errors.New("invalid search tree")
	ErrNotMatchDialect   = errors.New("not match dialect")
	ErrNeedSourceTable   = errors.New("this function need set table")
	ErrCannotSetValue    = errors.New("unable to set value")
	ErrZeroRecord        = errors.New("zero record found")
	ErrBindFieldsFailure = errors.New("bind fields failure")
	ErrRepeatFieldName   = errors.New("this table have repeat name field")
)

type ErrInvalidModelType string

func (e ErrInvalidModelType) Error() string {
	return "invalid model type " + string(e)
}

type ErrInvalidModelName struct{}

func (e ErrInvalidModelName) Error() string {
	return "invalid model name"
}

type ErrInvalidPreloadField struct {
	ModelName string
	FieldName string
}

func (e ErrInvalidPreloadField) Error() string {
	return fmt.Sprintf("invalid preload field %s.%s", e.ModelName, e.FieldName)
}

type ErrInvalidRecordType struct {
}

func (e ErrInvalidRecordType) Error() string {
	return fmt.Sprintf("record type must be the struct or map[string]interface{} or map[uintptr]interface{}")
}
