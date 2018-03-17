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
