package toyorm

import (
	"fmt"
	"reflect"
	"strings"
)

type SearchExpr int

const (
	ExprAnd = SearchExpr(iota)
	ExprOr
	ExprNot
	ExprEqual
	ExprNotEqual
	ExprGreater
	ExprGreaterEqual
	ExprLess
	ExprLessEqual
	ExprBetween
	ExprNotBetween
	ExprIn
	ExprNotIn
	ExprLike
	ExprNotLike
	ExprNull
	ExprNotNull
)

func (op SearchExpr) IsBranch() bool {
	return op == ExprAnd || op == ExprOr || op == ExprNot
}

type SearchObjVal struct {
	Field *ModelField
	Val   interface{}
}

type SearchCell struct {
	Type SearchExpr
	Val  *SearchObjVal
}

func NewSearchBranch(op SearchExpr) SearchCell {
	return SearchCell{
		op,
		nil,
	}
}

func NewSearchLeaf(op SearchExpr, field *ModelField, val interface{}) SearchCell {
	return SearchCell{
		op,
		&SearchObjVal{field, val},
	}
}

func (s SearchList) Condition(field *ModelField, val interface{}, expr, linkExpr SearchExpr) SearchList {
	if len(s) == 0 {
		return SearchList{NewSearchLeaf(expr, field, val)}
	}
	newS := make(SearchList, len(s))
	copy(newS, s)
	newS = append(newS, NewSearchLeaf(expr, field, val), NewSearchBranch(linkExpr))
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

func (s SearchList) ToExecValue() ExecValue {
	var stack []ExecValue
	for i := 0; i < len(s); i++ {
		if s[i].Type.IsBranch() {
			var exec ExecValue
			switch s[i].Type {
			case ExprAnd:
				if len(stack) < 2 {
					panic(ErrInvalidSearchTree)
				}
				last1, last2 := stack[len(stack)-1], stack[len(stack)-2]
				stack = stack[:len(stack)-2]
				exec.Query = fmt.Sprintf("%s AND %s", last2.Query, last1.Query)
				exec.Args = append(exec.Args, last2.Args...)
				exec.Args = append(exec.Args, last1.Args...)
			case ExprOr:
				if len(stack) < 2 {
					panic(ErrInvalidSearchTree)
				}
				last1, last2 := stack[len(stack)-1], stack[len(stack)-2]
				stack = stack[:len(stack)-2]
				exec.Query = fmt.Sprintf("(%s OR %s)", last2.Query, last1.Query)
				exec.Args = append(exec.Args, last2.Args...)
				exec.Args = append(exec.Args, last1.Args...)
			case ExprNot:
				if len(stack) < 1 {
					panic(ErrInvalidSearchTree)
				}
				last := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				exec.Query = fmt.Sprintf("NOT(%s)", last.Query)
				exec.Args = append(exec.Args, last.Args...)
			default:
				panic(ErrInvalidType)
			}
			stack = append(stack, exec)
		} else {
			exec := ExecValue{}
			switch s[i].Type {
			case ExprEqual:
				exec.Query = fmt.Sprintf("%s = ?", s[i].Val.Field.Name)
				exec.Args = append(exec.Args, s[i].Val.Val)
			case ExprNotEqual:
				exec.Query = fmt.Sprintf("%s <> ?", s[i].Val.Field.Name)
				exec.Args = append(exec.Args, s[i].Val.Val)
			case ExprGreater:
				exec.Query = fmt.Sprintf("%s > ?", s[i].Val.Field.Name)
				exec.Args = append(exec.Args, s[i].Val.Val)
			case ExprGreaterEqual:
				exec.Query = fmt.Sprintf("%s >= ?", s[i].Val.Field.Name)
				exec.Args = append(exec.Args, s[i].Val.Val)
			case ExprLess:
				exec.Query = fmt.Sprintf("%s < ?", s[i].Val.Field.Name)
				exec.Args = append(exec.Args, s[i].Val.Val)
			case ExprLessEqual:
				exec.Query = fmt.Sprintf("%s <= ?", s[i].Val.Field.Name)
				exec.Args = append(exec.Args, s[i].Val.Val)
			case ExprBetween:
				exec.Query = fmt.Sprintf("%s NOT BETWEEN ? AND ?", s[i].Val.Field.Name)
				vv := reflect.ValueOf(s[i].Val.Val)
				exec.Args = append(exec.Args, vv.Index(0).Interface(), vv.Index(1).Interface())
			case ExprNotBetween:
				exec.Query = fmt.Sprintf("%s NOT BETWEEN ? AND ?", s[i].Val.Field.Name)
				vv := reflect.ValueOf(s[i].Val.Val)
				exec.Args = append(exec.Args, vv.Index(0).Interface(), vv.Index(1).Interface())
			case ExprIn:
				vv := reflect.ValueOf(s[i].Val.Val)
				questionMarks := strings.TrimSuffix(strings.Repeat("?,", vv.Len()), ",")
				exec.Query = fmt.Sprintf("%s IN (%s)", s[i].Val.Field.Name, questionMarks)
				for i := 0; i < vv.Len(); i++ {
					exec.Args = append(exec.Args, vv.Index(i).Interface())
				}
			case ExprNotIn:
				vv := reflect.ValueOf(s[i].Val.Val)
				questionMarks := strings.TrimSuffix(strings.Repeat("?,", vv.Len()), ",")
				exec.Query = fmt.Sprintf("%s NOT IN (%s)", s[i].Val.Field.Name, questionMarks)
				for i := 0; i < vv.Len(); i++ {
					exec.Args = append(exec.Args, vv.Index(i).Interface())
				}
			case ExprLike:
				exec.Query = fmt.Sprintf("%s LIKE ?", s[i].Val.Field.Name)
				exec.Args = append(exec.Args, s[i].Val.Val)
			case ExprNotLike:
				exec.Query = fmt.Sprintf("%s NOT LIKE ?", s[i].Val.Field.Name)
				exec.Args = append(exec.Args, s[i].Val.Val)
			case ExprNull:
				exec.Query = fmt.Sprintf("%s IS NULL", s[i].Val.Field.Name)
			case ExprNotNull:
				exec.Query = fmt.Sprintf("%s IS NOT NULL", s[i].Val.Field.Name)
			}
			stack = append(stack, exec)
		}
	}
	return stack[0]
}
