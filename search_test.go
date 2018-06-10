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
				NewSearchTree(NewSearchLeaf(ExprEqual, model.GetFieldWithName("A").ToFieldValue(reflect.ValueOf("22")))),
				NewSearchTree(NewSearchLeaf(ExprEqual, model.GetFieldWithName("B").ToFieldValue(reflect.ValueOf("33")))),
			),
			NewSearchTree(NewSearchBranch(ExprAnd)).Fill(
				NewSearchTree(NewSearchLeaf(ExprEqual, model.GetFieldWithName("C").ToFieldValue(reflect.ValueOf("55")))),
				NewSearchTree(NewSearchLeaf(ExprEqual, model.GetFieldWithName("D").ToFieldValue(reflect.ValueOf("66")))),
			),
		)
		t.Log(dialect.SearchExec(t1.ToStack()))
	}
	{
		model := NewModel(reflect.TypeOf(TestSearchTable{}))
		t2 := NewSearchTree(NewSearchBranch(ExprOr)).Fill(
			NewSearchTree(NewSearchBranch(ExprNot)).Fill(
				NewSearchTree(NewSearchLeaf(ExprEqual, model.GetFieldWithName("A").ToFieldValue(reflect.ValueOf("22")))),
				nil,
			),
			NewSearchTree(NewSearchBranch(ExprAnd)).Fill(
				NewSearchTree(NewSearchLeaf(ExprEqual, model.GetFieldWithName("B").ToFieldValue(reflect.ValueOf("55")))),
				NewSearchTree(NewSearchLeaf(ExprEqual, model.GetFieldWithName("C").ToFieldValue(reflect.ValueOf("66")))),
			),
		)

		t2List := t2.ToStack()
		t2List = t2List.Condition(model.GetFieldWithName("D").ToFieldValue(reflect.ValueOf("909")), ExprEqual, ExprAnd)
		t.Logf("%#v", dialect.SearchExec(t2List))
	}

	{
		model := NewModel(reflect.TypeOf(TestSearchTable{}))
		t1 := NewSearchTree(NewSearchBranch(ExprOr)).Fill(
			NewSearchTree(NewSearchBranch(ExprAnd)).Fill(
				NewSearchTree(NewSearchLeaf(ExprEqual, model.GetFieldWithName("A").ToColumnAlias("t1").ToFieldValue(reflect.ValueOf("22")))),
				NewSearchTree(NewSearchLeaf(ExprEqual, model.GetFieldWithName("B").ToColumnAlias("t1").ToFieldValue(reflect.ValueOf("33")))),
			),
			NewSearchTree(NewSearchBranch(ExprAnd)).Fill(
				NewSearchTree(NewSearchLeaf(ExprEqual, model.GetFieldWithName("C").ToColumnAlias("t2").ToFieldValue(reflect.ValueOf("55")))),
				NewSearchTree(NewSearchLeaf(ExprEqual, model.GetFieldWithName("D").ToColumnAlias("t2").ToFieldValue(reflect.ValueOf("66")))),
			),
		)
		t.Log(dialect.SearchExec(t1.ToStack()))
		t.Logf("%#v\n", t1.ToStack())
	}
}
