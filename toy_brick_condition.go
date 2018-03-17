/*
 * Copyright 2018. bigpigeon. All rights reserved.
 * Use of this source code is governed by a MIT style
 * license that can be found in the LICENSE file.
 */

package toyorm

type ToyBrickAnd struct {
	Brick *ToyBrick
}

func (t ToyBrickAnd) Condition(expr SearchExpr, key interface{}, v ...interface{}) *ToyBrick {
	search := t.Brick.condition(expr, key, v...)
	return t.Conditions(search)
}

func (t ToyBrickAnd) Conditions(search SearchList) *ToyBrick {
	return t.Brick.Scope(func(t *ToyBrick) *ToyBrick {
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

type ToyBrickOr struct {
	Brick *ToyBrick
}

func (t ToyBrickOr) Condition(expr SearchExpr, key interface{}, v ...interface{}) *ToyBrick {
	search := t.Brick.condition(expr, key, v...)
	return t.Conditions(search)
}

func (t ToyBrickOr) Conditions(search SearchList) *ToyBrick {
	return t.Brick.Scope(func(t *ToyBrick) *ToyBrick {
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
