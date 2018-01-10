package toyorm

import (
	"errors"
)

var (
	ErrInvalidType         = errors.New("invalid type")
	ErrInvalidTag          = errors.New("invalid tag")
	ErrInvalidPreloadField = errors.New("invalid preload field")
	ErrInvalidSearchTree   = errors.New("invalid search tree")
	ErrNotMatchDialect     = errors.New("not match dialect")
	ErrNeedSourceTable     = errors.New("this function need set table")
	ErrCannotSetValue      = errors.New("unable to set value")
	ErrZeroRecord          = errors.New("zero record found")
	ErrBindFieldsFailure   = errors.New("bind fields failure")
	ErrRepeatFieldName     = errors.New("this table have repeat name field")
)
