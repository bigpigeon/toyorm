/*
 * Copyright 2018. bigpigeon. All rights reserved.
 * Use of this source code is governed by a MIT style
 * license that can be found in the LICENSE file.
 */

package toyorm

type Join struct {
	Model     *Model
	SubModel  *Model
	Container Field
	OnMain    Field
	OnSub     Field
}

// join will replace attribute
type JoinSwap struct {
	OwnOrderBy        []int
	OwnGroupBy        []int
	OwnSearch         []int
	Alias             string
	FieldsSelector    [ModeEnd][]Field
	SwapMap           map[string]*JoinSwap
	JoinMap           map[string]*Join
	MapPreloadBrick   map[string]*ToyBrick
	BelongToPreload   map[string]*BelongToPreload
	OneToOnePreload   map[string]*OneToOnePreload
	OneToManyPreload  map[string]*OneToManyPreload
	ManyToManyPreload map[string]*ManyToManyPreload
}

func NewJoinSwap(alias string) *JoinSwap {
	return &JoinSwap{
		Alias:             alias,
		SwapMap:           map[string]*JoinSwap{},
		JoinMap:           map[string]*Join{},
		MapPreloadBrick:   map[string]*ToyBrick{},
		BelongToPreload:   map[string]*BelongToPreload{},
		OneToOnePreload:   map[string]*OneToOnePreload{},
		OneToManyPreload:  map[string]*OneToManyPreload{},
		ManyToManyPreload: map[string]*ManyToManyPreload{},
	}
}

func (m *JoinSwap) Copy() *JoinSwap {
	newMap := *m
	return &newMap
}
