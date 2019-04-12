/*
 * Copyright 2018. bigpigeon. All rights reserved.
 * Use of this source code is governed by a MIT style
 * license that can be found in the LICENSE file.
 */

package toyorm

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
	"time"
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
	assert.NoError(t, err)
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

func TestBugPrimaryKeyWithZero(t *testing.T) {
	skillTestDB(t, "sqlite3")
	type TestPrimaryKeyWithZero struct {
		ID   uint32 `toyorm:"primary key"`
		Data string
	}
	var tab TestPrimaryKeyWithZero
	brick := TestDB.Model(&tab)
	createTableUnit(brick)(t)
	data := TestPrimaryKeyWithZero{
		Data: "some meta data",
	}
	result, err := brick.Debug().Save(&data)
	require.NoError(t, err)
	require.Error(t, result.Err())
	//{
	//	result, err := brick.Debug().Save(&data)
	//	require.NoError(t, err)
	//	require.NoError(t, result.Err())
	//}
}

func TestBugCustomDefaultPrimaryKey(t *testing.T) {
	type TestCustomDefaultPrimaryKey struct {
		ID   uint32 `toyorm:"primary key"`
		Data string
	}
	brick := TestDB.Model(&TestCustomDefaultPrimaryKey{})
	//createTableUnit(brick)(t)
	result, err := brick.DropTableIfExist()
	require.NoError(t, err)
	require.NoError(t, result.Err())
	if TestDriver == "postgres" {
		_, err := brick.Exec(DefaultExec{query: `CREATE TABLE "test_custom_default_primary_key" (id SERIAL,data VARCHAR(255) ,PRIMARY KEY(id))`})
		require.NoError(t, err)
	} else if TestDriver == "mysql" {
		_, err := brick.Exec(DefaultExec{query: "CREATE TABLE `test_custom_default_primary_key` (id INTEGER AUTO_INCREMENT,data VARCHAR(255) ,PRIMARY KEY(id))"})
		require.NoError(t, err)
	} else if TestDriver == "sqlite3" {
		_, err := brick.Exec(DefaultExec{query: "CREATE TABLE `test_custom_default_primary_key` (id INTEGER PRIMARY KEY AUTOINCREMENT,data VARCHAR(255) )"})
		require.NoError(t, err)
	} else {
		return
	}
	data := TestCustomDefaultPrimaryKey{
		Data: "some thing",
	}
	result, err = brick.Insert(&data)
	require.NoError(t, err)
	require.NoError(t, result.Err())
	require.Equal(t, uint32(1), data.ID)
}

func TestBugUSaveCreatedAt(t *testing.T) {
	type TestUSaveCreatedAt struct {
		ID        uint32 `toyorm:"primary key;auto_increment"`
		Data      string
		CreatedAt time.Time
	}
	brick := TestDB.Model(&TestUSaveCreatedAt{})
	createTableUnit(brick)(t)
	data := TestUSaveCreatedAt{
		Data:      "test data",
		CreatedAt: time.Now().Add(-1 * time.Second), // mysql only save millisecond level datetime,need ensure the created at will change on next save
	}
	result, err := brick.Insert(&data)
	require.NoError(t, err)
	require.NoError(t, result.Err())
	t.Log(data.CreatedAt)
	var oldCreatedAt time.Time
	{
		var data TestUSaveCreatedAt
		result, err = brick.Find(&data)
		require.NoError(t, err)
		require.NoError(t, result.Err())
		oldCreatedAt = data.CreatedAt
	}
	data.CreatedAt = time.Time{}
	result, err = brick.USave(&data)
	require.NoError(t, err)
	require.NoError(t, result.Err())
	{
		var data TestUSaveCreatedAt
		result, err = brick.Find(&data)
		require.NoError(t, err)
		require.NoError(t, result.Err())
		require.Equal(t, oldCreatedAt, data.CreatedAt)
	}
}

func TestBugZeroWithNotStruct(t *testing.T) {
	type TestDataSub struct {
		Data string
	}
	var d []TestDataSub
	v := IsZero(reflect.ValueOf(d))
	require.True(t, v)
}

func TestBugNotZeroCreatedAt(t *testing.T) {
	type TestNotZeroCreatedAtTable struct {
		ID        uint32 `toyorm:"primary key;auto_increment"`
		CreatedAt time.Time
	}
	var tab TestNotZeroCreatedAtTable
	brick := TestDB.Model(&tab)
	createTableUnit(brick)(t)
	date, err := time.Parse(time.RFC3339, "2006-01-02T15:04:05Z")
	require.NoError(t, err)
	data := TestNotZeroCreatedAtTable{
		CreatedAt: date,
	}
	result, err := brick.Insert(&data)
	require.NoError(t, err)
	require.NoError(t, result.Err())
	var fData TestNotZeroCreatedAtTable
	result, err = brick.Find(&fData)
	require.NoError(t, err)
	require.NoError(t, result.Err())
	t.Logf("old date %s new date %s", date, fData.CreatedAt)
	require.True(t, date.Equal(fData.CreatedAt))
}

func TestBugOrderByConflictWithGroupBy(t *testing.T) {
	{
		brick := TestDB.Model(&TestGroupByTable{})
		createTableUnit(brick)(t)
	}

	var tab TestGroupByTableGroup
	brick := TestDB.Model(&tab)

	brick = brick.GroupBy(Offsetof(tab.Name), Offsetof(tab.Address)).OrderBy(Offsetof(tab.Name))
	var data []TestGroupByTableGroup
	result, err := brick.Find(&data)
	require.NoError(t, err)
	require.NoError(t, result.Err())
}

func TestBugSaveNilCreatedAt(t *testing.T) {
	skillTestDB(t, "sqlite3")
	type TestSaveCreatedAt struct {
		ID        uint32 `toyorm:"primary key;auto_increment"`
		Data      string
		CreatedAt time.Time `toyorm:"NULL"`
	}
	brick := TestDB.Model(&TestSaveCreatedAt{})
	createTableUnit(brick)(t)
	//brick = brick.Debug()
	data := TestSaveCreatedAt{
		Data:      "test data",
		CreatedAt: time.Now().Add(-10 * time.Second),
	}
	result, err := brick.Insert(&data)
	require.NoError(t, err)
	require.NoError(t, result.Err())
	var oldCreatedAt time.Time
	{
		var data TestSaveCreatedAt
		result, err = brick.Find(&data)
		require.NoError(t, err)
		require.NoError(t, result.Err())
		oldCreatedAt = data.CreatedAt
	}
	//time.Sleep(100 * time.Millisecond)
	// FIXME mysql will update created at field
	t.Log(oldCreatedAt)
	data.CreatedAt = time.Time{}
	data.Data += " save again"
	result, err = brick.Save(&data)
	require.NoError(t, err)
	require.NoError(t, result.Err())
	t.Log(result.Report())
	{
		var data TestSaveCreatedAt
		result, err = brick.Find(&data)
		require.NoError(t, err)
		require.NoError(t, result.Err())
		assert.Equal(t, oldCreatedAt, data.CreatedAt)
		t.Log(data.CreatedAt)
		assert.Equal(t, data.Data, "test data save again")
	}
}
