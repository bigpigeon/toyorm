/*
 * Copyright 2018. bigpigeon. All rights reserved.
 * Use of this source code is governed by a MIT style
 * license that can be found in the LICENSE file.
 */

package toyorm

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
	. "unsafe"
)

func TestSqlNameConvert(t *testing.T) {
	assert.Equal(t, SqlNameConvert("UserDetail"), "user_detail")
	assert.Equal(t, SqlNameConvert("OneToOne"), "one_to_one")
	assert.Equal(t, SqlNameConvert("_UserDetail"), "_user_detail")
	assert.Equal(t, SqlNameConvert("userDetail"), "user_detail")
	assert.Equal(t, SqlNameConvert("UserDetailID"), "user_detail_id")
	assert.Equal(t, SqlNameConvert("NameHTTPtest"), "name_http_test")
	assert.Equal(t, SqlNameConvert("IDandValue"), "id_and_value")
	assert.Equal(t, SqlNameConvert("toyorm.User.field"), "toyorm_user_field")
}

func TestSafeAppend(t *testing.T) {
	{
		l1 := []int{}
		var v int = 1
		var pv *int = &v
		var ppv **int = &pv
		list := SafeAppend(reflect.ValueOf(l1), reflect.ValueOf(v), reflect.ValueOf(pv), reflect.ValueOf(ppv)).Interface()
		assert.Equal(t, list, []int{1, 1, 1})
	}
	{
		l2 := []*int{}
		var v int = 1
		var pv *int = &v
		var ppv **int = &pv
		list := SafeAppend(reflect.ValueOf(l2), reflect.ValueOf(v), reflect.ValueOf(pv), reflect.ValueOf(ppv)).Interface()
		for _, v := range list.([]*int) {
			assert.Equal(t, *v, 1)
		}
	}

}

func TestTemplateExec(t *testing.T) {
	{
		exec, err := GetTemplateExec(BasicExec{
			query: "Select * From $ModelName $Conditions Or id = ? $Limit",
			args:  []interface{}{2},
		}, map[string]BasicExec{
			"ModelName":  {"user", nil},
			"Conditions": {"WHERE age > ? AND id = ?", []interface{}{20, 1}},
			"Limit":      {"Limit = ?", []interface{}{10}},
		})
		assert.Nil(t, err)
		t.Log(exec.query, exec.args)
		assert.Equal(t, exec.query, "Select * From user WHERE age > ? AND id = ? Or id = ? Limit = ?")
		assert.Equal(t, exec.args, []interface{}{20, 1, 2, 10})
	}
}

func TestFindColumnFactory(t *testing.T) {
	var tab TestJoinTable
	var priceTab TestJoinPriceTable
	brick := TestDB.Model(&tab).
		Join(Offsetof(tab.NameJoin)).Swap().
		Join(Offsetof(tab.PriceJoin)).Join(Offsetof(priceTab.StarJoin)).Swap().Swap()
	record := NewRecord(brick.Model, reflect.ValueOf(&TestJoinTable{}).Elem())
	columns, scannersGen := FindColumnFactory(record, brick)
	var colStrList []string

	for _, c := range columns {
		colStrList = append(colStrList, c.Column())
	}
	t.Log(colStrList)
	scanners := scannersGen(record)
	var scanTypes []reflect.Type
	for _, scanner := range scanners {
		scanTypes = append(scanTypes, reflect.TypeOf(scanner).Elem())
	}
	t.Log(scanTypes)
	assert.Equal(t, len(columns), len(scanners))
}

func TestLoopGetElemAndPtr(t *testing.T) {
	type TestData struct {
		Name string
	}
	data := []*TestData{
		nil,
	}
	val := LoopDiveSliceAndPtr(reflect.ValueOf(data))
	t.Log(val.Interface())
	assert.Equal(t, val.Type(), reflect.TypeOf(TestData{}))
	data = []*TestData{
		{Name: "bigpigeon"},
	}
	val = LoopDiveSliceAndPtr(reflect.ValueOf(data))
	t.Log(val.Interface())
	assert.Equal(t, val.Type(), reflect.TypeOf(TestData{}))
}

func TestConditionToReversePolandExpr(t *testing.T) {
	{
		cond := "a | b & c & (d & e | f) | g"
		expr, err := conditionToReversePolandExprDebug(cond, map[string]BasicExec{}, true)
		require.NoError(t, err)
		t.Log(string(expr))
		assert.Equal(t, string(expr), "abcde&f|&&g||")
	}
	{
		cond := "a & (b | c | (d | e))"
		expr, err := conditionToReversePolandExprDebug(cond, map[string]BasicExec{}, true)
		require.NoError(t, err)
		t.Log(string(expr))
		assert.Equal(t, string(expr), "abcde|||&")
	}
	{
		cond := "a|b&c&(d|e&f)"
		expr, err := conditionToReversePolandExprDebug(cond, map[string]BasicExec{}, true)
		require.NoError(t, err)
		t.Log(string(expr))
		assert.Equal(t, string(expr), "abcdef&|&&|")
	}
	{
		cond := "a & (|b | c | (d | e))"
		_, err := conditionToReversePolandExprDebug(cond, map[string]BasicExec{}, true)
		t.Log(err)
		require.Equal(t, err, ErrInvalidConditionWord{"a & (|", cond})
	}

	{
		cond := "a & o (b | c | (d | e))"
		_, err := conditionToReversePolandExprDebug(cond, map[string]BasicExec{}, true)
		t.Log(err)
		require.Equal(t, err, ErrInvalidConditionWord{"a & o (b", cond})
	}
	{
		cond := "a && b"
		_, err := conditionToReversePolandExprDebug(cond, map[string]BasicExec{}, true)
		t.Log(err)
		require.Equal(t, err, ErrInvalidConditionWord{"a &&", cond})
	}

}
