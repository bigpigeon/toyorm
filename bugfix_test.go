/*
 * Copyright 2018. bigpigeon. All rights reserved.
 * Use of this source code is governed by a MIT style
 * license that can be found in the LICENSE file.
 */

package toyorm

import (
	"github.com/stretchr/testify/assert"
	"testing"
	. "unsafe"
)

type BugMiddleReportLossTable struct {
	ID   int `toyorm:"primary key"`
	Data string

	Child []BugMiddleReportLossTable
}

func TestBugCreateTableManyToManyReportLoss(t *testing.T) {
	brick := TestDB.Model(&BugMiddleReportLossTable{}).
		Preload(Offsetof(BugMiddleReportLossTable{}.Child)).Enter()
	result, err := brick.DropTableIfExist()
	assert.NoError(t, err)
	assert.NoError(t, result.Err())
	t.Log(result.Report())
	for _, preloadResult := range result.MiddleModelPreload {
		assert.NotEqual(t, preloadResult.Report(), "")
	}
	result, err = brick.CreateTableIfNotExist()
	t.Log(result.Report())
	assert.NoError(t, result.Err())
	for _, preloadResult := range result.MiddleModelPreload {
		assert.NotEqual(t, preloadResult.Report(), "")
	}
}
