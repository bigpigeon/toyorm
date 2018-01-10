package toyorm

type ToyBrickAnd struct {
	Brick *ToyBrick
}

func (t ToyBrickAnd) Condition(expr SearchExpr, v ...interface{}) *ToyBrick {
	return t.Brick.Scope(func(t *ToyBrick) *ToyBrick {
		*t = *t
		search := t.condition(expr, v...)
		if len(t.Search) != 0 {
			// make and have high priority
			if t.Search[len(t.Search)-1].Type == ExprOr {
				t.Search = t.Search[:len(t.Search)-1]
				t.Search = append(t.Search, search...)
				t.Search = append(t.Search, NewSearchBranch(ExprAnd), NewSearchBranch(ExprOr))
			} else {
				t.Search = append(t.Search, search...)
				t.Search = append(t.Search, NewSearchBranch(ExprAnd))
			}
		} else {
			t.Search = append(t.Search, search...)
		}
		return t
	})
}

func (t ToyBrickAnd) Conditions(search SearchList) *ToyBrick {
	return t.Brick.Scope(func(t *ToyBrick) *ToyBrick {
		*t = *t
		if len(search) == 0 {
			return t
		} else if len(t.Search) != 0 {
			// make and have high priority
			if t.Search[len(t.Search)-1].Type == ExprOr {
				t.Search = t.Search[:len(t.Search)-1]
				t.Search = append(t.Search, search...)
				t.Search = append(t.Search, NewSearchBranch(ExprAnd), NewSearchBranch(ExprOr))
			} else {
				t.Search = append(t.Search, search...)
				t.Search = append(t.Search, NewSearchBranch(ExprAnd))
			}
		} else {
			t.Search = append(t.Search, search...)
		}
		return t
	})

}

type ToyBrickOr struct {
	Brick *ToyBrick
}

func (t ToyBrickOr) Condition(expr SearchExpr, v ...interface{}) *ToyBrick {
	return t.Brick.Scope(func(t *ToyBrick) *ToyBrick {
		*t = *t
		search := t.condition(expr, v...)
		if len(t.Search) != 0 {
			t.Search = append(t.Search, search...)
			t.Search = append(t.Search, NewSearchBranch(ExprOr))
		} else {
			t.Search = append(t.Search, search...)
		}
		return t
	})
}

func (t ToyBrickOr) Conditions(search SearchList) *ToyBrick {
	return t.Brick.Scope(func(t *ToyBrick) *ToyBrick {
		*t = *t
		if len(search) == 0 {
			return t
		} else if len(t.Search) != 0 {
			t.Search = append(t.Search, search...)
			t.Search = append(t.Search, NewSearchBranch(ExprOr))
		} else {
			t.Search = append(t.Search, search...)
		}
		return t
	})
}
