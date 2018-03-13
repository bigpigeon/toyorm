package toyorm

import (
	"math"
)

const abortIndex int = math.MaxInt8 / 2

type HandlerFunc func(c *Context) error
type HandlersChain []HandlerFunc
type Context struct {
	handlers HandlersChain
	index    int8
	Brick    *ToyBrick
	Result   *Result
	value    map[interface{}]interface{}
	//err      error
}

func NewContext(handlers HandlersChain, brick *ToyBrick, columns ModelRecords) *Context {
	return &Context{
		handlers: handlers,
		index:    -1,
		Brick:    brick,
		Result: &Result{
			Records:            columns,
			Preload:            map[string]*Result{},
			RecordsActions:     map[int][]int{},
			MiddleModelPreload: map[string]*Result{},
			SimpleRelation:     map[string]map[int]int{},
			MultipleRelation:   map[string]map[int]Pair{},
		},
	}
}

func (c *Context) Value(v interface{}) interface{} {
	return c.value[v]
}

func (c *Context) Next() error {
	c.index++
	//handlerNames := make([]string, len(c.handlers))
	//for i, h := range c.handlers {
	//	handlerNames[i] = runtime.FuncForPC(reflect.ValueOf(h).Pointer()).column()
	//}
	var err error
	for s := int8(len(c.handlers)); c.index < s; c.index++ {
		err = c.handlers[c.index](c)
		if err != nil {
			c.Abort()
		}
	}
	return err
}

func (c *Context) IsAborted() bool {
	return len(c.handlers) >= abortIndex
}

func (c *Context) Abort() {
	c.index = int8(abortIndex)
}

type CollectionHandlerFunc func(c *CollectionContext) error
type CollectionHandlersChain []CollectionHandlerFunc
type CollectionContext struct {
	handlers CollectionHandlersChain
	index    int8
	Brick    *CollectionBrick
	Result   *Result
	value    map[interface{}]interface{}
	//err      error
}

func NewCollectionContext(handlers CollectionHandlersChain, brick *CollectionBrick, columns ModelRecords) *CollectionContext {
	return &CollectionContext{
		handlers: handlers,
		index:    -1,
		Brick:    brick,
		Result: &Result{
			Records:            columns,
			Preload:            map[string]*Result{},
			RecordsActions:     map[int][]int{},
			MiddleModelPreload: map[string]*Result{},
			SimpleRelation:     map[string]map[int]int{},
			MultipleRelation:   map[string]map[int]Pair{},
		},
	}
}

func (c *CollectionContext) Value(v interface{}) interface{} {
	return c.value[v]
}

func (c *CollectionContext) Next() error {
	c.index++
	//handlerNames := make([]string, len(c.handlers))
	//for i, h := range c.handlers {
	//	handlerNames[i] = runtime.FuncForPC(reflect.ValueOf(h).Pointer()).column()
	//}
	var err error
	for s := int8(len(c.handlers)); c.index < s; c.index++ {
		err = c.handlers[c.index](c)
		if err != nil {
			c.Abort()
		}
	}
	return err
}

func (c *CollectionContext) IsAborted() bool {
	return len(c.handlers) >= abortIndex
}

func (c *CollectionContext) Abort() {
	c.index = int8(abortIndex)
}
