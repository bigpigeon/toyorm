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
				NewSearchTree(NewSearchLeaf(ExprEqual, &BrickColumnValue{BrickColumn{"", model.GetFieldWithName("A").Column()}, reflect.ValueOf("22")})),
				NewSearchTree(NewSearchLeaf(ExprEqual, &BrickColumnValue{BrickColumn{"", model.GetFieldWithName("B").Column()}, reflect.ValueOf("33")})),
			),
			NewSearchTree(NewSearchBranch(ExprAnd)).Fill(
				NewSearchTree(NewSearchLeaf(ExprEqual, &BrickColumnValue{BrickColumn{"", model.GetFieldWithName("C").Column()}, reflect.ValueOf("55")})),
				NewSearchTree(NewSearchLeaf(ExprEqual, &BrickColumnValue{BrickColumn{"", model.GetFieldWithName("D").Column()}, reflect.ValueOf("66")})),
			),
		)
		t.Log(dialect.SearchExec(t1.ToStack()))
	}
	{
		model := NewModel(reflect.TypeOf(TestSearchTable{}))
		t2 := NewSearchTree(NewSearchBranch(ExprOr)).Fill(
			NewSearchTree(NewSearchBranch(ExprNot)).Fill(
				NewSearchTree(NewSearchLeaf(ExprEqual, &BrickColumnValue{BrickColumn{"", model.GetFieldWithName("A").Column()}, reflect.ValueOf("22")})),
				nil,
			),
			NewSearchTree(NewSearchBranch(ExprAnd)).Fill(
				NewSearchTree(NewSearchLeaf(ExprEqual, &BrickColumnValue{BrickColumn{"", model.GetFieldWithName("B").Column()}, reflect.ValueOf("55")})),
				NewSearchTree(NewSearchLeaf(ExprEqual, &BrickColumnValue{BrickColumn{"", model.GetFieldWithName("C").Column()}, reflect.ValueOf("66")})),
			),
		)

		t2List := t2.ToStack()
		t2List = t2List.Condition(&BrickColumnValue{BrickColumn{"", model.GetFieldWithName("D").Column()}, reflect.ValueOf("909")}, ExprEqual, ExprAnd)
		t.Logf("%#v", dialect.SearchExec(t2List))
	}

	{
		model := NewModel(reflect.TypeOf(TestSearchTable{}))
		t1 := NewSearchTree(NewSearchBranch(ExprOr)).Fill(
			NewSearchTree(NewSearchBranch(ExprAnd)).Fill(
				NewSearchTree(NewSearchLeaf(ExprEqual, &BrickColumnValue{BrickColumn{"t1", model.GetFieldWithName("A").Column()}, reflect.ValueOf("22")})),
				NewSearchTree(NewSearchLeaf(ExprEqual, &BrickColumnValue{BrickColumn{"t1", model.GetFieldWithName("B").Column()}, reflect.ValueOf("33")})),
			),
			NewSearchTree(NewSearchBranch(ExprAnd)).Fill(
				NewSearchTree(NewSearchLeaf(ExprEqual, &BrickColumnValue{BrickColumn{"t2", model.GetFieldWithName("C").Column()}, reflect.ValueOf("55")})),
				NewSearchTree(NewSearchLeaf(ExprEqual, &BrickColumnValue{BrickColumn{"t2", model.GetFieldWithName("D").Column()}, reflect.ValueOf("66")})),
			),
		)
		t.Log(dialect.SearchExec(t1.ToStack()))
	}
}
