/*
 * Copyright 2018. bigpigeon. All rights reserved.
 * Use of this source code is governed by a MIT style
 * license that can be found in the LICENSE file.
 */

package toyorm

import (
	"bytes"
	"encoding/json"
	"strconv"
)

type ExecValue interface {
	Source() string
	Query() string
	Args() []interface{}
	Append(query string, args ...interface{}) ExecValue
	JsonArgs() string
}

// not process source str, Source() == Query()
type DefaultExec struct {
	query string
	args  []interface{}
}

func (e DefaultExec) Source() string {
	return e.query
}

func (e DefaultExec) Query() string {
	return e.query
}

func (e DefaultExec) Args() []interface{} {
	return e.args
}

func (e DefaultExec) Append(query string, args ...interface{}) ExecValue {
	e.query += query
	e.args = append(e.args, args...)
	return e
}

func (e DefaultExec) JsonArgs() string {
	s, err := json.Marshal(e.args)
	if err != nil {
		panic(err)
	}
	return string(s)
}

// when call Query() method, all '?' in query will replace to '$1','$2'...
type QToSExec struct {
	DefaultExec
}

// go-bug DefaultExec will return ExecValue(DefaultExec), so  must to implement this func
func (e QToSExec) Append(query string, args ...interface{}) ExecValue {
	e.query += query
	e.args = append(e.args, args...)
	return e
}

func (e QToSExec) Query() string {
	data := []byte(e.query)
	buff := bytes.Buffer{}
	isEscaping := false
	pNum := 1 // number of placeholder
	pre, i := 0, 0
	for ; i < len(data); i++ {
		switch e.query[i] {
		case '?':
			if isEscaping == false {
				buff.Write(data[pre:i])
				buff.Write(append([]byte{'$'}, []byte(strconv.Itoa(pNum))...))
				pre = i + 1
				pNum++
			} else {
				buff.Write(data[pre : i-1])
				buff.WriteByte(data[i])
				pre = i + 1
			}
		case '\\':
			isEscaping = true
			continue
		}
		isEscaping = false
	}
	buff.Write(data[pre:i])
	return buff.String()
}
