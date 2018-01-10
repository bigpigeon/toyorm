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
			Preload:            map[*ModelField]*Result{},
			RecordsActions:     map[int][]SqlAction{},
			MiddleModelPreload: map[*ModelField]*Result{},
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
	//	handlerNames[i] = runtime.FuncForPC(reflect.ValueOf(h).Pointer()).Name()
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
