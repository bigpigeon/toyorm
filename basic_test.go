/*
 * Copyright 2018. bigpigeon. All rights reserved.
 * Use of this source code is governed by a MIT style
 * license that can be found in the LICENSE file.
 */

package toyorm

import (
	"github.com/stretchr/testify/assert"
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

func TestTemplateExecWord(t *testing.T) {
	{
		exec, err := getTemplateExec(BasicExec{
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
		exec, err := getTemplateExec(BasicExec{
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
		_, err := getTemplateExec(BasicExec{
			query: "Select $-",
		}, map[string]BasicExec{})
		assert.Equal(t, err, ErrTemplateExecInvalidWord{"$"})
	}
	{
		_, err := getTemplateExec(BasicExec{
			query: "Select $User-",
		}, map[string]BasicExec{})
		assert.Equal(t, err, ErrTemplateExecInvalidWord{"$User"})
	}

	{
		exec, err := getTemplateExec(BasicExec{
			query: "Select \\$User",
		}, map[string]BasicExec{})
		assert.Nil(t, err)
		t.Log(exec.query)
		assert.Equal(t, exec.query, "Select $User")
	}
}

func TestTemplateExec(t *testing.T) {
	{
		exec, err := getTemplateExec(BasicExec{
			query: "Select * From $ModelName $Condition Or id = ? $Limit",
			args:  []interface{}{2},
		}, map[string]BasicExec{
			"ModelName": {"user", nil},
			"Condition": {"WHERE age > ? AND id = ?", []interface{}{20, 1}},
			"Limit":     {"Limit = ?", []interface{}{10}},
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

}
