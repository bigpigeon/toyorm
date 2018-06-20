/*
 * Copyright 2018. bigpigeon. All rights reserved.
 * Use of this source code is governed by a MIT style
 * license that can be found in the LICENSE file.
 */

package toyorm

type Mode int8

const (
	ModeDefault Mode = iota
	ModeInsert
	ModeSave
	ModeUpdate
	ModeScan
	ModeSelect
	ModeCondition
	ModePreload
	ModeEnd
)
