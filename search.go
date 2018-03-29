/*
 * Copyright 2018. bigpigeon. All rights reserved.
 * Use of this source code is governed by a MIT style
 * license that can be found in the LICENSE file.
 */

package toyorm

type SearchExpr string

const (
	ExprIgnore       = ""
	ExprAnd          = "AND"
	ExprOr           = "OR"
	ExprNot          = "NOT"
	ExprEqual        = "="
	ExprNotEqual     = "<>"
	ExprGreater      = ">"
	ExprGreaterEqual = ">="
	ExprLess         = "<"
	ExprLessEqual    = "<="
	ExprBetween      = "BETWEEN"
	ExprNotBetween   = "NOT BETWEEN"
	ExprIn           = "IN"
	ExprNotIn        = "NOT IN"
	ExprLike         = "LIKE"
	ExprNotLike      = "NOT LIKE"
	ExprNull         = "NULL"
	ExprNotNull      = "NOT NULL"
)

func (op SearchExpr) IsBranch() bool {
	return op == ExprAnd || op == ExprOr || op == ExprNot || op == ExprIgnore
}

type SearchObjVal struct {
	Field *modelField
	Val   interface{}
}

type SearchCell struct {
	Type SearchExpr
	Val  ColumnValue
}

func NewSearchBranch(op SearchExpr) SearchCell {
	return SearchCell{
		op,
		nil,
	}
}

func NewSearchLeaf(op SearchExpr, columnValue ColumnValue) SearchCell {
	return SearchCell{
		op,
		columnValue,
	}
}

func (s SearchList) Condition(columnValue ColumnValue, expr, linkExpr SearchExpr) SearchList {
	if len(s) == 0 {
		return SearchList{NewSearchLeaf(expr, columnValue)}
	}
	newS := make(SearchList, len(s))
	copy(newS, s)
	newS = append(newS, NewSearchLeaf(expr, columnValue), NewSearchBranch(linkExpr))
	return newS
}

//func (s SearchCell) IsBranch() bool {
//	return s.Type == ExprAnd || s.Type == ExprOr
//}

type SearchList []SearchCell

type SearchTree struct {
	Val   SearchCell
	Left  *SearchTree
	Right *SearchTree
}

func NewSearchTree(val SearchCell) *SearchTree {
	return &SearchTree{
		Val: val,
	}
}

//func NewSearchList(val ...SearchCell) SearchList{
//
//}

func (s *SearchTree) Fill(left, right *SearchTree) *SearchTree {
	return &SearchTree{
		Val:   s.Val,
		Left:  left,
		Right: right,
	}
}

func (s *SearchTree) ToStack() SearchList {
	stack := SearchList{}
	if s.Left != nil {
		stack = append(stack, s.Left.ToStack()...)
	}
	if s.Right != nil {
		stack = append(stack, s.Right.ToStack()...)
	}
	stack = append(stack, s.Val)
	return stack
}
