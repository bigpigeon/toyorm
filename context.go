/*
 * Copyright 2018. bigpigeon. All rights reserved.
 * Use of this source code is governed by a MIT style
 * license that can be found in the LICENSE file.
 */

package toyorm

import (
	"math"
	"time"
)

const abortIndex int = math.MaxInt8 / 2

type HandlerFunc func(c *Context) error
type HandlersChain []HandlerFunc

type ExecHandlerFunc func(c *Context)
type ExecHandlerFuncChain []ExecHandlerFunc
type Context struct {
	handlers    HandlersChain
	execHandler ExecHandlerFunc
	index       int8
	Brick       *ToyBrick
	Result      *Result
	value       map[interface{}]interface{}
	//err      error
}

func (c *Context) Deadline() (deadline time.Time, ok bool) {
	return
}

func (c *Context) Done() <-chan struct{} {
	return c.Result.done
}

func (c *Context) Err() error {
	return nil
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
			done:               make(chan struct{}),
		},
	}
}

func (c *Context) Value(v interface{}) interface{} {
	return c.value[v]
}

func (c *Context) Start() {
	c.Next()
	close(c.Result.done)
}

func (c *Context) Next() {
	c.index++

	var err error
	for s := int8(len(c.handlers)); c.index < s; c.index++ {
		err = c.handlers[c.index](c)
		if err != nil {
			c.Abort()
		}
	}
	c.Result.err = err
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
