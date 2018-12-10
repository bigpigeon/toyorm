/*
 * Copyright 2018. bigpigeon. All rights reserved.
 * Use of this source code is governed by a MIT style
 * license that can be found in the LICENSE file.
 */

package toyorm

import (
	"io"
	"reflect"
)

type CacheMeta struct {
	TableName string
}

type ToyKernel struct {
	debug bool
	// map[model][container_field_name]
	Dialect Dialect
	Logger  io.Writer
}

// TODO testing thread safe? if not add lock
func (t *ToyKernel) GetModel(val reflect.Value) *Model {
	if val.Kind() != reflect.Struct {
		panic(ErrInvalidModelType("invalid struct type " + val.Type().Name()))
	}
	name := ModelName(val)
	return newModel(val, name)
}

func (t *ToyKernel) SetDebug(debug bool) {
	t.debug = debug
}
