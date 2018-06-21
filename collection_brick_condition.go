/*
 * Copyright 2018. bigpigeon. All rights reserved.
 * Use of this source code is governed by a MIT style
 * license that can be found in the LICENSE file.
 */

package toyorm

import (
	"reflect"
)

type CollectionBrickAnd struct {
	Brick *CollectionBrick
}

func (t CollectionBrickAnd) Condition(expr SearchExpr, key FieldSelection, v ...interface{}) *CollectionBrick {
	search := t.Brick.condition(expr, key, v...)
	return t.Conditions(search)
}

func (t CollectionBrickAnd) ConditionGroup(expr SearchExpr, group interface{}) *CollectionBrick {
	search := t.Brick.conditionGroup(expr, group)
	return t.Conditions(search)
}

func (t CollectionBrickAnd) Conditions(search SearchList) *CollectionBrick {
	return t.Brick.Scope(func(t *CollectionBrick) *CollectionBrick {
		if len(search) == 0 {
			return t
		}
		newSearch := make(SearchList, len(t.Search), len(t.Search)+len(search))
		copy(newSearch, t.Search)
		if len(t.Search) != 0 {
			// AND have high priority
			if newSearch[len(newSearch)-1].Type == ExprOr {
				newSearch = newSearch[:len(newSearch)-1]
				newSearch = append(newSearch, search...)
				newSearch = append(newSearch, NewSearchBranch(ExprAnd), NewSearchBranch(ExprOr))
			} else {
				newSearch = append(newSearch, search...)
				newSearch = append(newSearch, NewSearchBranch(ExprAnd))
			}
		} else {
			newSearch = append(newSearch, search...)
		}
		newt := *t
		newt.Search = newSearch
		return &newt
	})

}

type CollectionBrickOr struct {
	Brick *CollectionBrick
}

func (t CollectionBrickOr) Condition(expr SearchExpr, key FieldSelection, v ...interface{}) *CollectionBrick {
	search := t.Brick.condition(expr, key, v...)
	return t.Conditions(search)
}

func (t CollectionBrickOr) ConditionGroup(expr SearchExpr, group interface{}) *CollectionBrick {
	search := t.Brick.conditionGroup(expr, group)
	return t.Conditions(search)
}

func (t CollectionBrickOr) Conditions(search SearchList) *CollectionBrick {
	return t.Brick.Scope(func(t *CollectionBrick) *CollectionBrick {
		if len(search) == 0 {
			return t
		}
		newSearch := make(SearchList, len(t.Search), len(t.Search)+len(search))
		copy(newSearch, t.Search)
		if len(newSearch) != 0 {
			newSearch = append(newSearch, search...)
			newSearch = append(newSearch, NewSearchBranch(ExprOr))
		} else {
			newSearch = append(newSearch, search...)
		}
		newt := *t
		newt.Search = newSearch
		return &newt
	})
}

func (t *CollectionBrick) condition(expr SearchExpr, key FieldSelection, args ...interface{}) SearchList {
	var value reflect.Value
	if len(args) == 1 {
		value = reflect.ValueOf(args[0])
	} else {
		value = reflect.ValueOf(args)
	}
	mField := t.Model.fieldSelect(key)

	search := SearchList{}.Condition(mField.ToFieldValue(value), expr, ExprAnd)

	return search
}

func (t *CollectionBrick) conditionGroup(expr SearchExpr, group interface{}) SearchList {
	switch expr {
	case ExprAnd, ExprOr:
		var search SearchList
		keyValue := LoopIndirect(reflect.ValueOf(group))
		record := NewRecord(t.Model, keyValue)
		pairs := t.getFieldValuePairWithRecord(ModeCondition, record)
		for _, pair := range pairs {
			search = search.Condition(pair, ExprEqual, expr)
		}
		// avoid "or" condition effected by priority
		if expr == ExprOr {
			search = append(search, NewSearchBranch(ExprIgnore))
		}

		return search
	}
	panic("invalid expr")
}

// where will clean old condition
func (t *CollectionBrick) Where(expr SearchExpr, key FieldSelection, v ...interface{}) *CollectionBrick {
	return t.Conditions(t.condition(expr, key, v...))
}

// expr only support And/Or , group must be struct data or map[string]interface{}/map[uintptr]interface{}
func (t *CollectionBrick) WhereGroup(expr SearchExpr, group interface{}) *CollectionBrick {
	return t.Conditions(t.conditionGroup(expr, group))
}

func (t *CollectionBrick) Conditions(search SearchList) *CollectionBrick {
	return t.Scope(func(t *CollectionBrick) *CollectionBrick {
		if len(search) == 0 {
			newt := *t
			newt.Search = nil
			return &newt
		}
		newSearch := make(SearchList, len(search), len(search)+1)
		copy(newSearch, search)
		// to protect search priority
		newSearch = append(newSearch, NewSearchBranch(ExprIgnore))

		newt := *t
		newt.Search = newSearch
		return &newt
	})
}
