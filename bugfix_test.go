/*
 * Copyright 2018. bigpigeon. All rights reserved.
 * Use of this source code is governed by a MIT style
 * license that can be found in the LICENSE file.
 */

package toyorm

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestBugCreateCustomNameTableDirtyCache(t *testing.T) {
	tab := TestCustomTableNameTable{
		FragNum:  1,
		BelongTo: &TestCustomTableNameBelongTo{FragNum: 1}}
	brick := TestDB.Model(&tab).
		Preload(Offsetof(tab.BelongTo)).Enter()
	t.Log(brick.BelongToPreload["BelongTo"].SubModel.Name)
	tab2 := TestCustomTableNameTable{
		FragNum:  1,
		BelongTo: &TestCustomTableNameBelongTo{FragNum: 2}}
	brick = TestDB.Model(&tab2).
		Preload(Offsetof(tab.BelongTo)).Enter()
	require.Equal(t, brick.BelongToPreload["BelongTo"].SubModel.Name, "test_custom_table_name_belong_to_2")
}

func TestBugOrderByConflictWithLimit(t *testing.T) {
	type TestOrderByLimitTable struct {
		ModelDefault
		Data string
	}
	var tab TestOrderByLimitTable
	brick := TestDB.Model(&tab)
	createTableUnit(brick)(t)
	var data []TestOrderByLimitTable
	result, err := brick.OrderBy(Offsetof(tab.CreatedAt)).Limit(2).Find(&data)
	require.NoError(t, err)
	require.NoError(t, result.Err())
}
