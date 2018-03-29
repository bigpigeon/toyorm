/*
 * Copyright 2018. bigpigeon. All rights reserved.
 * Use of this source code is governed by a MIT style
 * license that can be found in the LICENSE file.
 */

package toyorm

import (
	"reflect"
	"testing"
)

func TestSearchToExecValue(t *testing.T) {
	dialect := DefaultDialect{}
	{
		model := NewModel(reflect.TypeOf(TestSearchTable{}))
		t1 := NewSearchTree(NewSearchBranch(ExprOr)).Fill(
			NewSearchTree(NewSearchBranch(ExprAnd)).Fill(
				NewSearchTree(NewSearchLeaf(ExprEqual, &modelFieldValue{model.GetFieldWithName("A"), reflect.ValueOf("22")})),
				NewSearchTree(NewSearchLeaf(ExprEqual, &modelFieldValue{model.GetFieldWithName("B"), reflect.ValueOf("33")})),
			),
			NewSearchTree(NewSearchBranch(ExprAnd)).Fill(
				NewSearchTree(NewSearchLeaf(ExprEqual, &modelFieldValue{model.GetFieldWithName("C"), reflect.ValueOf("55")})),
				NewSearchTree(NewSearchLeaf(ExprEqual, &modelFieldValue{model.GetFieldWithName("D"), reflect.ValueOf("66")})),
			),
		)
		t.Log(dialect.SearchExec(t1.ToStack()))
	}
	{
		model := NewModel(reflect.TypeOf(TestSearchTable{}))
		t2 := NewSearchTree(NewSearchBranch(ExprOr)).Fill(
			NewSearchTree(NewSearchBranch(ExprNot)).Fill(
				NewSearchTree(NewSearchLeaf(ExprEqual, &modelFieldValue{model.GetFieldWithName("A"), reflect.ValueOf("22")})),
				nil,
			),
			NewSearchTree(NewSearchBranch(ExprAnd)).Fill(
				NewSearchTree(NewSearchLeaf(ExprEqual, &modelFieldValue{model.GetFieldWithName("B"), reflect.ValueOf("55")})),
				NewSearchTree(NewSearchLeaf(ExprEqual, &modelFieldValue{model.GetFieldWithName("C"), reflect.ValueOf("66")})),
			),
		)

		t2List := t2.ToStack()
		t2List = t2List.Condition(&modelFieldValue{model.GetFieldWithName("D"), reflect.ValueOf("909")}, ExprEqual, ExprAnd)
		t.Logf("%#v", dialect.SearchExec(t2List))
	}
}
