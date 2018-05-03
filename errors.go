/*
 * Copyright 2018. bigpigeon. All rights reserved.
 * Use of this source code is governed by a MIT style
 * license that can be found in the LICENSE file.
 */

package toyorm

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidTag        = errors.New("invalid tag")
	ErrInvalidSearchTree = errors.New("invalid search tree")
	ErrNotMatchDialect   = errors.New("not match dialect")
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

type ErrRepeatField struct {
	ModelName string
	FieldName string
}

func (e ErrRepeatField) Error() string {
	return fmt.Sprintf("model '%s' have repeat field '%s'", e.ModelName, e.FieldName)
}

type ErrSameColumnName struct {
	ModelName    string
	Same         string
	OldFieldName string
	NewFieldName string
}

func (e ErrSameColumnName) Error() string {
	return fmt.Sprintf("model '%s' have same sql column %s in field '%s', '%s'", e.ModelName, e.Same, e.OldFieldName, e.NewFieldName)
}

type ErrCollectionExec map[int]error

func (e ErrCollectionExec) Error() string {
	var s string
	for k, v := range e {
		s += fmt.Sprintf("[%d] %s;", k, v)
	}
	return s
}

type ErrCollectionQuery map[int]error

func (e ErrCollectionQuery) Error() string {
	var s string
	for k, v := range e {
		s += fmt.Sprintf("[%d] %s;", k, v)
	}
	return s
}

type ErrCollectionQueryRow map[int]error

func (e ErrCollectionQueryRow) Error() string {
	var s string
	for k, v := range e {
		s += fmt.Sprintf("[%d] %s;", k, v)
	}
	return s
}

type ErrCollectionDBSelectorNotFound struct {
}

func (e ErrCollectionDBSelectorNotFound) Error() string {
	return "db selector not found"
}

type ErrCollectionClose map[int]error

func (e ErrCollectionClose) Error() string {
	var s string
	for k, v := range e {
		s += fmt.Sprintf("[%d] %s;", k, v)
	}
	return s
}

type ErrDbIndexNotSet struct{}

func (e ErrDbIndexNotSet) Error() string {
	return "db index not set"
}

type ErrZeroPrimaryKey struct{ Model *Model }

func (e ErrZeroPrimaryKey) Error() string {
	return fmt.Sprintf("%s have zero primary key", e.Model.Name)
}

type ErrNotSupportRowsAffected struct{}

func (e ErrNotSupportRowsAffected) Error() string {
	return "not support rows affected method"
}

type ErrLastInsertId struct{}

func (e ErrLastInsertId) Error() string {
	return "cannot scan last insert id"
}

type ErrTemplateExecInvalidWord struct {
	Word string
}

func (e ErrTemplateExecInvalidWord) Error() string {
	return fmt.Sprintf("template exec have invalid word %s", e.Word)
}

type ErrCannotSet struct {
	Operation string
}

func (e ErrCannotSet) Error() string {
	return fmt.Sprintf("%s can't be set", e.Operation)
}

type ErrModelDuplicateAssociation struct {
	Model string
	Type  AssociationType
	Name  string
}

func (e ErrModelDuplicateAssociation) Error() string {
	return fmt.Sprintf("model %s have duplicate %s in field %s tag", e.Model, e.Type, e.Name)
}
