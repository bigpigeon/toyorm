/*
 * Copyright 2018. bigpigeon. All rights reserved.
 * Use of this source code is governed by a MIT style
 * license that can be found in the LICENSE file.
 */

package toyorm

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestQToSExecQuery(t *testing.T) {
	{
		exec := QToSExec{
			DefaultExec{
				query: "Update set id = ?,name = ?",
				args:  []interface{}{2, "ben"},
			},
		}
		query := exec.Query()
		t.Log(query)
		assert.Equal(t, query, "Update set id = $1,name = $2")
	}
	{
		exec := QToSExec{
			DefaultExec{
				query: "Select * From name \\?= ?",
				args:  []interface{}{2},
			},
		}
		query := exec.Query()
		t.Log(query)
		assert.Equal(t, query, "Select * From name ?= $1")
	}
}
