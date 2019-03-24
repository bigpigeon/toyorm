/*
 * Copyright 2019. bigpigeon. All rights reserved.
 * Use of this source code is governed by a MIT style
 * license that can be found in the LICENSE file.
 */

package toyorm

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestTemplateExecWord(t *testing.T) {
	{
		exec, err := GetTemplateExec(BasicExec{
			query: "Select $ModelName $db $1",
		}, map[string]BasicExec{
			"ModelName": {"user", nil},
			"db":        {"toyorm", nil},
			"1":         {"args1", nil},
		})
		assert.Nil(t, err)
		t.Log(exec.query)
		assert.Equal(t, exec.query, "Select user toyorm args1")
	}
	{
		exec, err := GetTemplateExec(BasicExec{
			query: "Select $ModelName-$db_$1",
		}, map[string]BasicExec{
			"ModelName": {"user", nil},
			"db":        {"toyorm", nil},
			"1":         {"args1", nil},
		})
		assert.Nil(t, err)
		t.Log(exec.query)
		assert.Equal(t, exec.query, "Select user-toyorm_args1")
	}
	{
		_, err := GetTemplateExec(BasicExec{
			query: "Select $-",
		}, map[string]BasicExec{})
		assert.Equal(t, err, ErrTemplateExecInvalidWord{"$"})
	}
	{
		exec, err := GetTemplateExec(BasicExec{
			query: "Select $User-",
		}, map[string]BasicExec{})
		assert.NoError(t, err)
		assert.Equal(t, exec.query, "Select -")
	}

	{
		exec, err := GetTemplateExec(BasicExec{
			query: "Select \\$User\\",
		}, map[string]BasicExec{})
		assert.Nil(t, err)
		t.Log(exec.query)
		assert.Equal(t, exec.query, "Select $User\\")
	}
	{
		_, err := GetTemplateExec(BasicExec{
			query: "Select $",
		}, map[string]BasicExec{})
		assert.Equal(t, err, ErrTemplateExecInvalidWord{"$"})
	}
}

func TestTemplateConditionExecWord(t *testing.T) {
	{
		exec, err := GetTemplateExec(BasicExec{
			query: "$(Cas){WHERE $Cas}",
		}, map[string]BasicExec{
			"Cas": {
				"cas = ?",
				[]interface{}{5},
			},
		})
		assert.Nil(t, err)
		t.Log(exec.query)
		assert.Equal(t, exec.query, "WHERE cas = ?")
		assert.Equal(t, exec.args, []interface{}{5})
	}
	{
		base := BasicExec{
			query: "$(Cas){WHERE $Cas}{do nothing}",
		}
		exec, err := GetTemplateExec(base, map[string]BasicExec{
			"Cas": {
				"cas = ?",
				[]interface{}{5},
			},
		})
		assert.Nil(t, err)
		t.Log(exec.query)
		assert.Equal(t, exec.query, "WHERE cas = ?")
		assert.Equal(t, exec.args, []interface{}{5})

		exec, err = GetTemplateExec(base, map[string]BasicExec{})
		assert.Nil(t, err)
		t.Log(exec.query)
		assert.Equal(t, exec.query, "do nothing")
		assert.Equal(t, len(exec.args), 0)
	}
	{
		exec, err := GetTemplateExec(BasicExec{
			query: "$(Cas){WHERE $Cas}{$(Limit){LIMIT $Limit}}",
		}, map[string]BasicExec{
			"Limit": {
				"5",
				nil,
			},
		})
		assert.Nil(t, err)
		t.Log(exec.query)
		assert.Equal(t, exec.query, "LIMIT 5")
		assert.Equal(t, len(exec.args), 0)
	}
}

func TestCheckTemplateCondition(t *testing.T) {
	const charSeq = "abcdefghi"
	toExec := func(args ...bool) map[string]BasicExec { // must a,b,c,d,e
		exec := map[string]BasicExec{}
		for i, arg := range args {
			if arg {
				exec[string(charSeq[i])] = BasicExec{}
			}
		}
		return exec
	}
	allPossible := func(n uint) [][]bool {
		result := make([][]bool, 1<<n)
		for i := 0; i < (1 << n); i++ {
			ret := make([]bool, n)
			for j := uint(0); j < n; j++ {
				ret[j] = (i >> j & 1) == 1
			}
			result[i] = ret
		}
		return result
	}
	{
		var a, b, c bool
		baseCond := "a&(b|c)"
		checkFn := func() bool { return a && (b || c) }
		for _, v := range allPossible(3) {
			a, b, c = v[0], v[1], v[2]
			t.Log(v)
			result, err := checkTempCondition(baseCond, toExec(a, b, c))
			require.NoError(t, err)
			assert.Equal(t, result, checkFn())
		}
	}

	{
		var a, b, c bool
		checkFn := func() bool { return a || b && c }
		baseCond := "a|b&c"

		for _, v := range allPossible(3) {
			a, b, c = v[0], v[1], v[2]
			t.Log(v)
			result, err := checkTempCondition(baseCond, toExec(a, b, c))
			require.NoError(t, err)
			assert.Equal(t, result, checkFn())
		}
	}
	{
		var a, b, c, d, e, f bool
		baseCond := "a|b&c&(d|e&f)"
		checkFn := func() bool { return a || b && c && (d || e && f) }
		for _, v := range allPossible(6) {
			a, b, c, d, e, f = v[0], v[1], v[2], v[3], v[4], v[5]
			t.Log(v)
			result, err := checkTempCondition(baseCond, toExec(a, b, c, d, e, f))
			require.NoError(t, err)
			assert.Equal(t, result, checkFn())
		}
	}
	{
		var a, b, c, d, e, f bool
		baseCond := "a|(b&c|d)|e&f"
		checkFn := func() bool { return a || (b && c || d) || e && f }
		for _, v := range allPossible(6) {
			a, b, c, d, e, f = v[0], v[1], v[2], v[3], v[4], v[5]
			t.Log(v)
			result, err := checkTempCondition(baseCond, toExec(a, b, c, d, e, f))
			require.NoError(t, err)
			assert.Equal(t, result, checkFn())
		}
	}
}
