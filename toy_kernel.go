package toyorm

import (
	"io"
	"reflect"
)

type ToyKernel struct {
	CacheModels              map[reflect.Type]*Model
	CacheMiddleModels        map[reflect.Type]*Model
	CacheReverseMiddleModels map[reflect.Type]*Model
	// map[model][container_field_name]

	belongToPreload   map[*Model]map[string]*BelongToPreload
	oneToOnePreload   map[*Model]map[string]*OneToOnePreload
	oneToManyPreload  map[*Model]map[string]*OneToManyPreload
	manyToManyPreload map[*Model]map[string]map[bool]*ManyToManyPreload
	Dialect           Dialect
	Logger            io.Writer
}

// TODO testing thread safe? if not add lock
func (t *ToyKernel) GetModel(_type reflect.Type) *Model {
	if model, ok := t.CacheModels[_type]; ok == false {
		model = NewModel(_type)
		t.CacheModels[_type] = model
	}
	return t.CacheModels[_type]
}

func (t *ToyKernel) GetMiddleModel(_type reflect.Type) *Model {
	if model, ok := t.CacheModels[_type]; ok == false {
		model = NewModel(_type)
		t.CacheModels[_type] = model
	}
	return t.CacheModels[_type]
}
