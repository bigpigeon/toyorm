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
	CacheModels       map[reflect.Type]map[CacheMeta]*Model
	CacheMiddleModels map[reflect.Type]map[CacheMeta]*Model
	debug             bool
	// map[model][container_field_name]
	Dialect Dialect
	Logger  io.Writer
}

// TODO testing thread safe? if not add lock
func (t *ToyKernel) GetModel(val reflect.Value) *Model {
	name := ModelName(val)
	typ := val.Type()
	if t.CacheModels[typ] == nil {
		t.CacheModels[typ] = map[CacheMeta]*Model{}
	}
	if model, ok := t.CacheModels[typ][CacheMeta{name}]; ok == false {
		model = newModel(val, name)
		t.CacheModels[typ][CacheMeta{name}] = model
	}
	return t.CacheModels[typ][CacheMeta{name}]
}

func (t *ToyKernel) SetDebug(debug bool) {
	t.debug = debug
}
