/*
 * Copyright 2018. bigpigeon. All rights reserved.
 * Use of this source code is governed by a MIT style
 * license that can be found in the LICENSE file.
 */

package toyorm

import (
	"encoding/json"
)

type ExecValue struct {
	Query string
	Args  []interface{}
}

func (e ExecValue) JsonArgs() string {
	s, err := json.Marshal(e.Args)
	if err != nil {
		panic(err)
	}
	return string(s)
}
