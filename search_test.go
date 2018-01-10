package toyorm

import (
	"reflect"
	"testing"
)

func TestSearchToExecValue(t *testing.T) {
	{
		model := NewModel(reflect.TypeOf(TestSearchTable{}), TestDB.Dialect)
		t1 := NewSearchTree(NewSearchBranch(ExprOr)).Fill(
			NewSearchTree(NewSearchBranch(ExprAnd)).Fill(
				NewSearchTree(NewSearchLeaf(ExprEqual, model.GetFieldWithName("A"), "22")),
				NewSearchTree(NewSearchLeaf(ExprEqual, model.GetFieldWithName("B"), "33")),
			),
			NewSearchTree(NewSearchBranch(ExprAnd)).Fill(
				NewSearchTree(NewSearchLeaf(ExprEqual, model.GetFieldWithName("C"), "55")),
				NewSearchTree(NewSearchLeaf(ExprEqual, model.GetFieldWithName("D"), "66")),
			),
		)
		t.Log(t1.ToStack().ToExecValue())
	}
	{
		model := NewModel(reflect.TypeOf(TestSearchTable{}), TestDB.Dialect)
		t2 := NewSearchTree(NewSearchBranch(ExprOr)).Fill(
			NewSearchTree(NewSearchBranch(ExprNot)).Fill(
				NewSearchTree(NewSearchLeaf(ExprEqual, model.GetFieldWithName("A"), "22")),
				nil,
			),
			NewSearchTree(NewSearchBranch(ExprAnd)).Fill(
				NewSearchTree(NewSearchLeaf(ExprEqual, model.GetFieldWithName("B"), "55")),
				NewSearchTree(NewSearchLeaf(ExprEqual, model.GetFieldWithName("C"), "66")),
			),
		)

		t2List := t2.ToStack()
		t2List = t2List.Condition(model.GetFieldWithName("D"), "909", ExprEqual, ExprAnd)
		t.Logf("%#v", t2List.ToExecValue())
	}
}
