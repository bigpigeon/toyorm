/*
 * Copyright 2018. bigpigeon. All rights reserved.
 * Use of this source code is governed by a MIT style
 * license that can be found in the LICENSE file.
 */

package toyorm

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestSameModel(t *testing.T) {
	m1 := NewModel(reflect.TypeOf(TestPreloadTable{}))
	m2 := NewModel(reflect.TypeOf(TestPreloadTable{}))

	m1Value := reflect.ValueOf(*m1)
	m2Value := reflect.ValueOf(*m2)
	mType := m1Value.Type()
	for i := 0; i < mType.NumField(); i++ {
		isSame := reflect.DeepEqual(m1Value.Field(i).Interface(), m2Value.Field(i).Interface())
		t.Logf("compare field %s is same? %v", mType.Field(i).Name, isSame)
		if !isSame {
			t.Fail()
			f1, err := json.MarshalIndent(m1Value.Field(i).Interface(), "", "  ")
			if err != nil {
				t.Errorf("err %s", err)
			}
			f2, err := json.MarshalIndent(m2Value.Field(i).Interface(), "", "  ")
			if err != nil {
				t.Errorf("err %s", err)
			}

			t.Log(string(f1))
			t.Log(string(f2))
		}
	}
}
