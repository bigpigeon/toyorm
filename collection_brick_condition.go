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

func (t CollectionBrickAnd) Condition(expr SearchExpr, key interface{}, v ...interface{}) *CollectionBrick {
	search := t.Brick.condition(expr, key, v...)
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

func (t CollectionBrickOr) Condition(expr SearchExpr, key interface{}, v ...interface{}) *CollectionBrick {
	search := t.Brick.condition(expr, key, v...)
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

func (t *CollectionBrick) condition(expr SearchExpr, key interface{}, args ...interface{}) SearchList {
	search := SearchList{}
	switch expr {
	case ExprAnd, ExprOr:
		keyValue := LoopIndirect(reflect.ValueOf(key))
		record := NewRecord(t.Model, keyValue)
		pairs := t.getFieldValuePairWithRecord(ModeCondition, record)
		for _, pair := range pairs {
			search = search.Condition(pair, ExprEqual, expr)
		}
		if expr == ExprOr {
			search = append(search, NewSearchBranch(ExprIgnore))
		}
	default:
		var value reflect.Value
		if len(args) == 1 {
			value = reflect.ValueOf(args[0])
		} else {
			value = reflect.ValueOf(args)
		}
		mField := t.Model.fieldSelect(key)

		search = search.Condition(&modelFieldValue{mField, value}, expr, ExprAnd)
	}
	return search
}

// where will clean old condition
func (t *CollectionBrick) Where(expr SearchExpr, key interface{}, v ...interface{}) *CollectionBrick {
	return t.Scope(func(t *CollectionBrick) *CollectionBrick {
		newt := *t
		newt.Search = t.condition(expr, key, v...)
		return &newt
	})
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
