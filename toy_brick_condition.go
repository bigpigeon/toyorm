/*
 * Copyright 2018. bigpigeon. All rights reserved.
 * Use of this source code is governed by a MIT style
 * license that can be found in the LICENSE file.
 */

package toyorm

type ToyBrickAnd struct {
	Brick *ToyBrick
}

func (t ToyBrickAnd) Condition(expr SearchExpr, key FieldSelection, v ...interface{}) *ToyBrick {
	search := t.Brick.condition(expr, key, v...)
	return t.Conditions(search)
}

func (t ToyBrickAnd) ConditionGroup(expr SearchExpr, group interface{}) *ToyBrick {
	search := t.Brick.conditionGroup(expr, group)
	return t.Conditions(search)
}

func (t ToyBrickAnd) Conditions(search SearchList) *ToyBrick {
	return t.Brick.Scope(func(t *ToyBrick) *ToyBrick {
		if len(search) == 0 {
			return t
		}
		newSearch := make(SearchList, len(t.Search), len(t.Search)+len(search))
		copy(newSearch, t.Search)
		newOwnSearch := make([]int, len(t.OwnSearch), len(t.OwnSearch)+len(search))
		copy(newOwnSearch, t.OwnSearch)
		if len(t.Search) != 0 {
			// AND have high priority
			if newSearch[len(newSearch)-1].Type == ExprOr {
				start := len(newSearch) - 1
				newSearch = newSearch[:start]
				newSearch = append(newSearch, search...)
				// add join own information
				for i, s := range newSearch[start:] {
					if s.Type.IsBranch() == false {
						newOwnSearch = append(newOwnSearch, start+i)
					}
				}
				newSearch = append(newSearch, NewSearchBranch(ExprAnd), NewSearchBranch(ExprOr))
			} else {
				start := len(newSearch)
				newSearch = append(newSearch, search...)
				// add join own information
				for i, s := range newSearch[start:] {
					if s.Type.IsBranch() == false {
						newOwnSearch = append(newOwnSearch, start+i)
					}
				}
				newSearch = append(newSearch, NewSearchBranch(ExprAnd))
			}
		} else {
			start := len(newSearch)
			newSearch = append(newSearch, search...)
			// add join own information
			for i, s := range newSearch[start:] {
				if s.Type.IsBranch() == false {
					newOwnSearch = append(newOwnSearch, start+i)
				}
			}
		}
		newt := *t
		newt.Search = newSearch
		newt.OwnSearch = newOwnSearch
		return &newt
	})

}

type ToyBrickOr struct {
	Brick *ToyBrick
}

func (t ToyBrickOr) Condition(expr SearchExpr, key FieldSelection, v ...interface{}) *ToyBrick {
	search := t.Brick.condition(expr, key, v...)
	return t.Conditions(search)
}

func (t ToyBrickOr) ConditionGroup(expr SearchExpr, group interface{}) *ToyBrick {
	search := t.Brick.conditionGroup(expr, group)
	return t.Conditions(search)
}

func (t ToyBrickOr) Conditions(search SearchList) *ToyBrick {
	return t.Brick.Scope(func(t *ToyBrick) *ToyBrick {
		if len(search) == 0 {
			return t
		}
		newSearch := make(SearchList, len(t.Search), len(t.Search)+len(search))
		copy(newSearch, t.Search)
		newOwnSearch := make([]int, len(t.OwnSearch), len(t.OwnSearch)+len(search))
		copy(newOwnSearch, t.OwnSearch)
		if len(newSearch) != 0 {
			start := len(newSearch)
			newSearch = append(newSearch, search...)
			// add join own information
			for i, s := range newSearch[start:] {
				if s.Type.IsBranch() == false {
					newOwnSearch = append(newOwnSearch, start+i)
				}
			}
			newSearch = append(newSearch, NewSearchBranch(ExprOr))
		} else {
			start := len(newSearch)
			newSearch = append(newSearch, search...)
			// add join own information
			for i, s := range newSearch[start:] {
				if s.Type.IsBranch() == false {
					newOwnSearch = append(newOwnSearch, start+i)
				}
			}
		}
		newt := *t
		newt.Search = newSearch
		newt.OwnSearch = newOwnSearch
		return &newt
	})
}
