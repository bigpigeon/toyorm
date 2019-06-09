/*
 * Copyright 2018. bigpigeon. All rights reserved.
 * Use of this source code is governed by a MIT style
 * license that can be found in the LICENSE file.
 */

package toyorm

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sort"
	"strings"
	"testing"
	"time"
	. "unsafe"
)

func TestCreateTable(t *testing.T) {
	var hastable bool
	var err error
	for _, tab := range []interface{}{
		TestCreateTable1{},
		TestCreateTable2{},
		TestCreateTable3{},
		TestCreateTable4{},
		SqlTypeTable{},
	} {
		// start a session
		brick := TestDB.Model(tab).Begin()
		hastable, err = brick.HasTable()
		assert.NoError(t, err)
		t.Logf("table %s exist:%v\n", brick.Model.Name, hastable)
		result := brick.DropTableIfExist()
		assert.Nil(t, result.Err())
		if err := result.Err(); err != nil {
			t.Error(err)
			t.Fail()
		}
		result = brick.CreateTable()
		assert.Nil(t, result.Err())
		if err := result.Err(); err != nil {
			t.Error(err)
			t.Fail()
		}
		err = brick.Commit()
		assert.NoError(t, err)
	}
}

func TestInsertData(t *testing.T) {
	brick := TestDB.Model(&TestInsertTable{})
	//create table
	{
		result := brick.DropTableIfExist()
		assert.Nil(t, result.Err())

		result = brick.CreateTable()
		assert.Nil(t, result.Err())
		if err := result.Err(); err != nil {
			t.Error(err)
			t.Fail()
		}
	}
	// insert with struct
	{
		dstr := "some str data"
		dint := 100
		dfloat := 100.0
		dcomplex := 100 + 1i
		tab := TestInsertTable{
			DataStr:     "some str data",
			DataInt:     100,
			DataFloat:   100.0,
			DataComplex: 100 + 1i,
			PtrStr:      &dstr,
			PtrInt:      &dint,
			PtrFloat:    &dfloat,
			PtrComplex:  &dcomplex,
		}
		result := brick.Insert(tab)
		assert.Nil(t, result.Err())

		t.Logf("id %v created at: %v updated at: %v", tab.ID, tab.CreatedAt, tab.UpdatedAt)
	}
	// test insert map[string]interface{}
	{
		dstr := "some str data"
		dint := 100
		dfloat := 100.0
		dcomplex := 100 + 1i
		tab := map[string]interface{}{
			"DataStr":     "some str data",
			"DataInt":     100,
			"DataFloat":   100.0,
			"DataComplex": 100 + 1i,
			"PtrStr":      &dstr,
			"PtrInt":      &dint,
			"PtrFloat":    &dfloat,
			"PtrComplex":  &dcomplex,
		}
		result := brick.Insert(tab)
		assert.Nil(t, result.Err())

		t.Logf("id %v created at: %v updated at: %v", tab["ID"], tab["CreatedAt"], tab["UpdatedAt"])
	}
	// test insert map[OffsetOf]interface{}
	{
		dstr := "some str data"
		dint := 100
		dfloat := 100.0
		dcomplex := 100 + 1i
		var tab TestInsertTable
		tabMap := map[uintptr]interface{}{
			Offsetof(tab.DataStr):     "some str data",
			Offsetof(tab.DataInt):     100,
			Offsetof(tab.DataFloat):   100.0,
			Offsetof(tab.DataComplex): 100 + 1i,
			Offsetof(tab.PtrStr):      &dstr,
			Offsetof(tab.PtrInt):      &dint,
			Offsetof(tab.PtrFloat):    &dfloat,
			Offsetof(tab.PtrComplex):  &dcomplex,
		}
		result := brick.Insert(tabMap)
		assert.Nil(t, result.Err())
		t.Logf("id %v created at: %v updated at: %v", tabMap[Offsetof(tab.ID)], tabMap[Offsetof(tab.CreatedAt)], tabMap[Offsetof(tab.UpdatedAt)])
	}
	// insert list
	// insert with struct
	{
		dstr := "some str data"
		dint := 100
		dfloat := 100.0
		dcomplex := 100 + 1i
		tab := []TestInsertTable{
			{
				DataStr:     "some str data",
				DataInt:     100,
				DataFloat:   100.0,
				DataComplex: 100 + 1i,
				PtrStr:      &dstr,
				PtrInt:      &dint,
				PtrFloat:    &dfloat,
				PtrComplex:  &dcomplex,
			},
			{
				DataStr:     "some str data.",
				DataInt:     101,
				DataFloat:   101.0,
				DataComplex: 101 + 1i,
				PtrStr:      &dstr,
				PtrInt:      &dint,
				PtrFloat:    &dfloat,
				PtrComplex:  &dcomplex,
			},
		}
		result := brick.Insert(tab)
		assert.Nil(t, result.Err())

		t.Logf("id %v %v", tab[0].ID, tab[1].ID)
	}
	// test insert map[string]interface{}
	{
		dstr := "some str data"
		dint := 100
		dfloat := 100.0
		dcomplex := 100 + 1i
		tab := []map[string]interface{}{
			{
				"DataStr":     "some str data",
				"DataInt":     100,
				"DataFloat":   100.0,
				"DataComplex": 100 + 1i,
				"PtrStr":      &dstr,
				"PtrInt":      &dint,
				"PtrFloat":    &dfloat,
				"PtrComplex":  &dcomplex,
			},
			{
				"DataStr":     "some str data.",
				"DataInt":     101,
				"DataFloat":   101.0,
				"DataComplex": 101 + 1i,
				"PtrStr":      &dstr,
				"PtrInt":      &dint,
				"PtrFloat":    &dfloat,
				"PtrComplex":  &dcomplex,
			},
		}
		result := brick.Insert(tab)
		assert.Nil(t, result.Err())
		t.Logf("id %v %v", tab[0]["ID"], tab[1]["ID"])
	}
	// test insert map[OffsetOf]interface{}
	{
		dstr := "some str data"
		dint := 100
		dfloat := 100.0
		dcomplex := 100 + 1i
		var tab TestInsertTable
		tabMap := []map[uintptr]interface{}{
			{
				Offsetof(tab.DataStr):     "some str data",
				Offsetof(tab.DataInt):     100,
				Offsetof(tab.DataFloat):   100.0,
				Offsetof(tab.DataComplex): 100 + 1i,
				Offsetof(tab.PtrStr):      &dstr,
				Offsetof(tab.PtrInt):      &dint,
				Offsetof(tab.PtrFloat):    &dfloat,
				Offsetof(tab.PtrComplex):  &dcomplex,
			},
			{
				Offsetof(tab.DataStr):     "some str data.",
				Offsetof(tab.DataInt):     101,
				Offsetof(tab.DataFloat):   100.1,
				Offsetof(tab.DataComplex): 101 + 2i,
				Offsetof(tab.PtrStr):      &dstr,
				Offsetof(tab.PtrInt):      &dint,
				Offsetof(tab.PtrFloat):    &dfloat,
				Offsetof(tab.PtrComplex):  &dcomplex,
			},
		}
		result := brick.Insert(tabMap)
		assert.Nil(t, result.Err())
		if err := result.Err(); err != nil {
			t.Error(err)
			t.Fail()
		}
		t.Logf("id %v %v", tabMap[0][Offsetof(tab.ID)], tabMap[1][Offsetof(tab.ID)])

	}
}

func TestInsertPointData(t *testing.T) {
	brick := TestDB.Model(&TestInsertTable{})
	// insert with struct
	{
		dstr := "some str data"
		dint := 100
		dfloat := 100.0
		dcomplex := 100 + 1i
		tab := TestInsertTable{
			DataStr:     "some str data",
			DataInt:     100,
			DataFloat:   100.0,
			DataComplex: 100 + 1i,
			PtrStr:      &dstr,
			PtrInt:      &dint,
			PtrFloat:    &dfloat,
			PtrComplex:  &dcomplex,
		}
		result := brick.Insert(&tab)

		if err := result.Err(); err != nil {
			t.Error(err)
			t.Fail()
		}
		t.Logf("id %v created at: %v updated at: %v", tab.ID, tab.CreatedAt, tab.UpdatedAt)
		assert.NotZero(t, tab.ID)
		assert.NotZero(t, tab.CreatedAt)
		assert.NotZero(t, tab.UpdatedAt)
	}
	// test insert map[string]interface{}
	{
		dstr := "some str data"
		dint := 100
		dfloat := 100.0
		dcomplex := 100 + 1i
		tab := map[string]interface{}{
			"DataStr":     "some str data",
			"DataInt":     100,
			"DataFloat":   100.0,
			"DataComplex": 100 + 1i,
			"PtrStr":      &dstr,
			"PtrInt":      &dint,
			"PtrFloat":    &dfloat,
			"PtrComplex":  &dcomplex,
		}
		result := brick.Insert(&tab)

		if err := result.Err(); err != nil {
			t.Error(err)
			t.Fail()
		}
		t.Logf("id %v created at: %v updated at: %v", tab["ID"], tab["CreatedAt"], tab["UpdatedAt"])
		assert.NotZero(t, tab["ID"])
		assert.NotZero(t, tab["CreatedAt"])
		assert.NotZero(t, tab["UpdatedAt"])
	}
	// test insert map[OffsetOf]interface{}
	{
		dstr := "some str data"
		dint := 100
		dfloat := 100.0
		dcomplex := 100 + 1i
		var tab TestInsertTable
		tabMap := map[uintptr]interface{}{
			Offsetof(tab.DataStr):     "some str data",
			Offsetof(tab.DataInt):     100,
			Offsetof(tab.DataFloat):   100.0,
			Offsetof(tab.DataComplex): 100 + 1i,
			Offsetof(tab.PtrStr):      &dstr,
			Offsetof(tab.PtrInt):      &dint,
			Offsetof(tab.PtrFloat):    &dfloat,
			Offsetof(tab.PtrComplex):  &dcomplex,
		}
		result := brick.Insert(&tabMap)

		if err := result.Err(); err != nil {
			t.Error(err)
			t.Fail()
		}
		t.Logf("id %v created at: %v updated at: %v", tabMap[Offsetof(tab.ID)], tabMap[Offsetof(tab.CreatedAt)], tabMap[Offsetof(tab.UpdatedAt)])
		assert.NotZero(t, tabMap[Offsetof(tab.ID)])
		assert.NotZero(t, tabMap[Offsetof(tab.CreatedAt)])
		assert.NotZero(t, tabMap[Offsetof(tab.UpdatedAt)])
	}
}

func TestFind(t *testing.T) {
	{
		brick := TestDB.Model(&TestSearchTable{})
		result := brick.DropTableIfExist()

		if err := result.Err(); err != nil {
			t.Error(err)
			t.Fail()
		}
		result = brick.CreateTable()

		if err := result.Err(); err != nil {
			t.Error(err)
			t.Fail()
		}
	}
	// fill data
	{
		for i := 1; i < 4; i++ {
			for j := 1; j < 4; j++ {
				for k := 1; k < 4; k++ {
					for l := 1; l < 4; l++ {
						d := strings.Repeat("d", l)
						t1 := TestSearchTable{
							A: strings.Repeat("a", i),
							B: strings.Repeat("b", j),
							C: strings.Repeat("c", k),
							D: &d,
						}
						result := TestDB.Model(&TestSearchTable{}).Insert(&t1)

						if err := result.Err(); err != nil {
							t.Error(err)
							t.Fail()
						}
						t.Logf("%#v\n", t1)
					}
				}
			}

		}
	}
	CheckNotZero := func(t *testing.T, tab *TestSearchTable) {
		assert.NotZero(t, tab.ID)
		assert.NotZero(t, tab.CreatedAt)
		assert.NotZero(t, tab.UpdatedAt)
		assert.NotZero(t, tab.A)
		assert.NotZero(t, tab.B)
		assert.NotZero(t, tab.C)
		assert.NotZero(t, tab.D)
	}
	// test find with struct
	{
		table := TestSearchTable{}
		result := TestDB.Model(&TestSearchTable{}).Find(&table)

		if err := result.Err(); err != nil {
			t.Error(err)
			t.Fail()
		}
		t.Logf("%#v\n", table)
		CheckNotZero(t, &table)
	}
	// test find with struct list
	{
		var tables []TestSearchTable
		result := TestDB.Model(&TestSearchTable{}).Find(&tables)

		if err := result.Err(); err != nil {
			t.Error(err)
			t.Fail()
		}
		t.Logf("%#v\n", tables)
		for _, tab := range tables {
			CheckNotZero(t, &tab)
		}
	}
}

func TestConditionFind(t *testing.T) {
	base := TestSearchTable{}
	//SELECT id,a,b,c,d FROM test_search_table WHERE a = ? AND b = ?, args:[]interface {}{"a", "b"}
	{
		var tabs []TestSearchTable
		result := TestDB.Model(&TestSearchTable{}).
			WhereGroup(ExprAnd, TestSearchTable{A: "a", B: "b"}).Find(&tabs)

		if err := result.Err(); err != nil {
			t.Error(err)
			t.Fail()
		}
		t.Logf("%#v\n", tabs)
		for _, tab := range tabs {
			assert.Equal(t, "a", tab.A)
			assert.Equal(t, "b", tab.B)
		}
	}
	//SELECT id,a,b,c,d FROM test_search_table WHERE (a = ? OR b = ?), args:[]interface {}{"a", "bb"}
	{
		var tabs []TestSearchTable
		result := TestDB.Model(&TestSearchTable{}).
			WhereGroup(ExprOr, TestSearchTable{A: "a", B: "bb"}).Find(&tabs)

		if err := result.Err(); err != nil {
			t.Error(err)
			t.Fail()
		}
		t.Logf("%#v\n", tabs)
		for _, tab := range tabs {
			assert.True(t, tab.A == "a" || tab.B == "bb")
		}
	}
	//SELECT id,a,b,c,d FROM test_search_table WHERE a = ? AND b = ?, args:[]interface {}{"a", "b"}
	{
		var tabs []TestSearchTable
		result := TestDB.Model(&TestSearchTable{}).
			Where(ExprEqual, Offsetof(base.A), "a").
			And().Condition(ExprEqual, Offsetof(base.B), "b").Find(&tabs)

		if err := result.Err(); err != nil {
			t.Error(err)
			t.Fail()
		}
		t.Logf("%#v\n", tabs)
		for _, tab := range tabs {
			assert.Equal(t, "a", tab.A)
			assert.Equal(t, "b", tab.B)
		}
	}
	//SELECT id,a,b,c,d FROM test_search_table WHERE ((a = ? AND b = ? OR c = ?) OR d = ? AND a = ?), args:[]interface {}{"a", "b", "c", "d", "aa"}
	{
		var tabs []TestSearchTable
		result := TestDB.Model(&TestSearchTable{}).
			Where(ExprEqual, Offsetof(base.A), "a").And().
			Condition(ExprEqual, Offsetof(base.B), "b").Or().
			Condition(ExprEqual, Offsetof(base.C), "c").Or().
			Condition(ExprEqual, Offsetof(base.D), "d").And().
			Condition(ExprEqual, Offsetof(base.A), "aa").
			Find(&tabs)

		if err := result.Err(); err != nil {
			t.Error(err)
			t.Fail()
		}
		t.Logf("%#v\n", tabs)
		for _, tab := range tabs {
			assert.True(t, (tab.A == "a" && tab.B == "b" || tab.C == "c") || *tab.D == "d" && tab.A == "aa")
		}
	}
	//SELECT id,a,b,c,d FROM test_search_table WHERE a = ? AND ((a = ? OR b = ?) OR c = ?), args:[]interface {}{"aa", "a", "b", "c"}
	{
		var tabs []TestSearchTable
		brick := TestDB.Model(&TestSearchTable{})

		result := brick.Where(ExprEqual, Offsetof(base.A), "aa").And().
			Conditions(
				brick.Where(ExprEqual, Offsetof(base.A), "a").Or().
					Condition(ExprEqual, Offsetof(base.B), "b").Or().
					Condition(ExprEqual, Offsetof(base.C), "c").Search,
			).
			Find(&tabs)

		if err := result.Err(); err != nil {
			t.Error(err)
			t.Fail()
		}
		t.Logf("%#v\n", tabs)
		for _, tab := range tabs {
			assert.True(t, tab.A == "aa" && ((tab.A == "a" || tab.B == "b") || tab.C == "c"))
		}
	}
}

func TestCombinationConditionFind(t *testing.T) {
	brick := TestDB.Model(&TestSearchTable{})
	{
		var tabs []TestSearchTable
		result := brick.WhereGroup(ExprAnd, TestSearchTable{A: "a", B: "b"}).Find(&tabs)

		if err := result.Err(); err != nil {
			t.Error(err)
			t.Fail()
		}
		t.Logf("%#v\n", tabs)
		for _, tab := range tabs {
			assert.Equal(t, "a", tab.A)
			assert.Equal(t, "b", tab.B)
		}
	}
	{
		var tabs []TestSearchTable
		result := brick.WhereGroup(ExprAnd, map[string]interface{}{"A": "a", "B": "b"}).Find(&tabs)

		if err := result.Err(); err != nil {
			t.Error(err)
			t.Fail()
		}
		t.Logf("%#v\n", tabs)
		for _, tab := range tabs {
			assert.Equal(t, "a", tab.A)
			assert.Equal(t, "b", tab.B)
		}
	}
	{
		var tabs []TestSearchTable
		result := brick.WhereGroup(ExprAnd, map[uintptr]interface{}{
			Offsetof(TestSearchTable{}.A): "a",
			Offsetof(TestSearchTable{}.B): "b",
		}).Find(&tabs)

		if err := result.Err(); err != nil {
			t.Error(err)
			t.Fail()
		}
		t.Logf("%#v\n", tabs)
		for _, tab := range tabs {
			assert.Equal(t, "a", tab.A)
			assert.Equal(t, "b", tab.B)
		}
	}

	{
		var tabs []TestSearchTable
		result := brick.WhereGroup(ExprOr, map[string]interface{}{"A": "a", "B": "b"}).And().
			Condition(ExprEqual, "C", "c").Find(&tabs)

		if err := result.Err(); err != nil {
			t.Error(err)
			t.Fail()
		}
		t.Logf("%#v\n", tabs)
		for _, tab := range tabs {
			assert.True(t, tab.A == "a" || tab.B == "b")
		}
	}
}

func TestOrderByFind(t *testing.T) {
	brick := TestDB.Model(&TestSearchTable{})
	{
		brick := brick.OrderBy(Offsetof(TestSearchTable{}.C))
		var data []TestSearchTable
		result := brick.Find(&data)

		if err := result.Err(); err != nil {
			t.Error(err)
			t.Fail()
		}
		var cList []string
		for _, tab := range data {
			cList = append(cList, tab.C)
		}
		fmt.Printf("c: %s\n", cList)
		assert.True(t, sort.StringsAreSorted(cList))
	}
	{
		brick := brick.OrderBy(brick.ToDesc(Offsetof(TestSearchTable{}.C)))
		var data []TestSearchTable
		result := brick.Find(&data)

		if err := result.Err(); err != nil {
			t.Error(err)
			t.Fail()
		}
		var cList []string
		for _, tab := range data {
			cList = append(cList, tab.C)
		}
		fmt.Printf("c: %s\n", cList)
		assert.True(t, sort.SliceIsSorted(cList, func(i, j int) bool {
			return cList[i] > cList[j]
		}))
	}
}

func TestUpdate(t *testing.T) {
	table := TestSearchTable{A: "aaaaa", B: "bbbbb"}
	brick := TestDB.Model(&TestSearchTable{})
	result := brick.Where(ExprEqual, Offsetof(TestSearchTable{}.A), "a").Update(&table)

	if err := result.Err(); err != nil {
		t.Error(err)
	}
	{
		var tableList []TestSearchTable
		brick.Where(ExprEqual, Offsetof(TestSearchTable{}.A), "a").Find(&tableList)
		t.Logf("%#v\n", tableList)
		assert.Equal(t, len(tableList), 0)
	}
	{
		var tableList []TestSearchTable
		brick.Where(ExprEqual, Offsetof(TestSearchTable{}.A), "aaaaa").And().
			Condition(ExprEqual, Offsetof(TestSearchTable{}.B), "bbbbb").
			Find(&tableList)
		t.Logf("%#v\n", tableList)
		assert.True(t, len(tableList) > 0)
	}
}

func TestPreloadCreateTable(t *testing.T) {
	brick := TestDB.Model(TestPreloadTable{}).
		Preload(Offsetof(TestPreloadTable{}.BelongTo)).Enter().
		Preload(Offsetof(TestPreloadTable{}.OneToOne)).Enter().
		Preload(Offsetof(TestPreloadTable{}.OneToMany)).Enter().
		Preload(Offsetof(TestPreloadTable{}.ManyToMany)).Enter()
	brick.CreateTable()
	hastable, err := brick.HasTable()
	assert.NoError(t, err)
	t.Logf("table %s exist:%v\n", brick.Model.Name, hastable)
	result := brick.DropTableIfExist()

	if err := result.Err(); err != nil {
		t.Error(err)
		t.Fail()
	}
	result = brick.CreateTable()

	if err := result.Err(); err != nil {
		t.Error(err)
		t.Fail()
	}
}

func TestPreloadInsertData(t *testing.T) {
	brick := TestDB.Model(TestPreloadTable{}).
		Preload(Offsetof(TestPreloadTable{}.BelongTo)).Enter().
		Preload(Offsetof(TestPreloadTable{}.OneToOne)).Enter().
		Preload(Offsetof(TestPreloadTable{}.OneToMany)).Enter().
		Preload(Offsetof(TestPreloadTable{}.ManyToMany)).Enter()
	{
		tab := TestPreloadTable{
			Name: "test insert data",
			BelongTo: &TestPreloadTableBelongTo{
				Name: "test insert data belong to",
			},
			OneToOne: &TestPreloadTableOneToOne{
				Name: "test insert data one to one",
			},
			OneToMany: []TestPreloadTableOneToMany{
				{Name: "test insert data one to many"},
				{Name: "test insert data one to many."},
			},
			ManyToMany: []TestPreloadTableManyToMany{
				{Name: "test insert data many to many"},
				{Name: "test insert data many to many."},
			},
		}
		result := brick.Insert(&tab)

		if err := result.Err(); err != nil {
			t.Error(err)
			t.Fail()
		}
		t.Logf("main id %v, belong id %v, one to one id %v, one to many id [%v,%v], many to many id [%v, %v]", tab.ID, tab.BelongTo.ID, tab.OneToOne.ID, tab.OneToMany[0].ID, tab.OneToMany[1].ID, tab.ManyToMany[0].ID, tab.ManyToMany[1].ID)
		assert.NotZero(t, tab.ID)
		assert.NotZero(t, tab.CreatedAt)
		assert.NotZero(t, tab.UpdatedAt)
		assert.NotZero(t, tab.BelongTo.ID)
		assert.NotZero(t, tab.OneToOne.ID)
		assert.NotZero(t, tab.OneToMany[0].ID)
		assert.NotZero(t, tab.OneToMany[1].ID)
		assert.NotZero(t, tab.ManyToMany[0].ID)
		assert.NotZero(t, tab.ManyToMany[1].ID)

		assert.Equal(t, tab.BelongToID, tab.BelongTo.ID)
		assert.Equal(t, tab.ID, tab.OneToOne.TestPreloadTableID)
		assert.Equal(t, tab.ID, tab.OneToMany[0].TestPreloadTableID)
		assert.Equal(t, tab.ID, tab.OneToMany[1].TestPreloadTableID)
	}
	{
		tab := map[string]interface{}{
			"Name": "test insert data",
			"BelongTo": map[string]interface{}{
				"Name": "test insert data belong to",
			},
			"OneToOne": map[string]interface{}{
				"Name": "test insert data one to one",
			},
			"OneToMany": []map[string]interface{}{
				{"Name": "test insert data one to many"},
				{"Name": "test insert data one to many."},
			},
			"ManyToMany": []map[string]interface{}{
				{"Name": "test insert data many to many"},
				{"Name": "test insert data many to many."},
			},
		}
		result := brick.Insert(tab)

		if err := result.Err(); err != nil {
			t.Error(err)
			t.Fail()
		}
		t.Logf("main id %v, belong id %v, one to one id %v, one to many id [%v,%v], many to many id [%v, %v]", tab["ID"], tab["BelongTo"].(map[string]interface{})["ID"], tab["OneToOne"].(map[string]interface{})["ID"], tab["OneToMany"].([]map[string]interface{})[0]["ID"], tab["OneToMany"].([]map[string]interface{})[1]["ID"], tab["ManyToMany"].([]map[string]interface{})[0]["ID"], tab["ManyToMany"].([]map[string]interface{})[1]["ID"])
		assert.NotZero(t, tab["ID"])
		assert.NotZero(t, tab["CreatedAt"])
		assert.NotZero(t, tab["UpdatedAt"])
		assert.NotZero(t, tab["BelongTo"].(map[string]interface{})["ID"])
		assert.NotZero(t, tab["OneToOne"].(map[string]interface{})["ID"])
		assert.NotZero(t, tab["OneToMany"].([]map[string]interface{})[0]["ID"])
		assert.NotZero(t, tab["OneToMany"].([]map[string]interface{})[1]["ID"])
		assert.NotZero(t, tab["ManyToMany"].([]map[string]interface{})[0]["ID"])
		assert.NotZero(t, tab["ManyToMany"].([]map[string]interface{})[1]["ID"])

		assert.Equal(t, tab["BelongToID"], tab["BelongTo"].(map[string]interface{})["ID"])
		assert.Equal(t, tab["ID"], tab["OneToOne"].(map[string]interface{})["TestPreloadTableID"])
		assert.Equal(t, tab["ID"], tab["OneToMany"].([]map[string]interface{})[0]["TestPreloadTableID"])
		assert.Equal(t, tab["ID"], tab["OneToMany"].([]map[string]interface{})[1]["TestPreloadTableID"])
	}
	{
		var tPreload TestPreloadTable
		var tBelong TestPreloadTableBelongTo
		var tOneToOne TestPreloadTableOneToOne
		var tOneToMany TestPreloadTableOneToMany
		var tManyToMany TestPreloadTableManyToMany
		tab := map[uintptr]interface{}{
			Offsetof(tPreload.Name): "test insert data",
			Offsetof(tPreload.BelongTo): map[uintptr]interface{}{
				Offsetof(tBelong.Name): "test insert data belong to",
			},
			Offsetof(tPreload.OneToOne): map[uintptr]interface{}{
				Offsetof(tOneToOne.Name): "test insert data one to one",
			},
			Offsetof(tPreload.OneToMany): []map[uintptr]interface{}{
				{Offsetof(tOneToMany.Name): "test insert data one to many"},
				{Offsetof(tOneToMany.Name): "test insert data one to many."},
			},
			Offsetof(tPreload.ManyToMany): []map[uintptr]interface{}{
				{Offsetof(tManyToMany.Name): "test insert data many to many"},
				{Offsetof(tManyToMany.Name): "test insert data many to many."},
			},
		}
		result := brick.Insert(tab)

		if err := result.Err(); err != nil {
			t.Error(err)
			t.Fail()
		}
		t.Logf(
			"main id %v, belong id %v, one to one id %v, one to many id [%v,%v], many to many id [%v, %v]",
			tab[Offsetof(tPreload.ID)],
			tab[Offsetof(tPreload.BelongTo)].(map[uintptr]interface{})[Offsetof(tBelong.ID)],
			tab[Offsetof(tPreload.OneToOne)].(map[uintptr]interface{})[Offsetof(tOneToOne.ID)],
			tab[Offsetof(tPreload.OneToMany)].([]map[uintptr]interface{})[0][Offsetof(tOneToMany.ID)],
			tab[Offsetof(tPreload.OneToMany)].([]map[uintptr]interface{})[1][Offsetof(tOneToMany.ID)],
			tab[Offsetof(tPreload.ManyToMany)].([]map[uintptr]interface{})[0][Offsetof(tManyToMany.ID)],
			tab[Offsetof(tPreload.ManyToMany)].([]map[uintptr]interface{})[1][Offsetof(tManyToMany.ID)],
		)
		assert.NotZero(t, tab[Offsetof(tPreload.ID)])
		assert.NotZero(t, tab[Offsetof(tPreload.CreatedAt)])
		assert.NotZero(t, tab[Offsetof(tPreload.UpdatedAt)])
		assert.NotZero(t, tab[Offsetof(tPreload.BelongTo)].(map[uintptr]interface{})[Offsetof(tBelong.ID)])
		assert.NotZero(t, tab[Offsetof(tPreload.OneToOne)].(map[uintptr]interface{})[Offsetof(tOneToOne.ID)])
		assert.NotZero(t, tab[Offsetof(tPreload.OneToMany)].([]map[uintptr]interface{})[0][Offsetof(tOneToMany.ID)])
		assert.NotZero(t, tab[Offsetof(tPreload.OneToMany)].([]map[uintptr]interface{})[1][Offsetof(tOneToMany.ID)])
		assert.NotZero(t, tab[Offsetof(tPreload.ManyToMany)].([]map[uintptr]interface{})[0][Offsetof(tManyToMany.ID)])
		assert.NotZero(t, tab[Offsetof(tPreload.ManyToMany)].([]map[uintptr]interface{})[1][Offsetof(tManyToMany.ID)])

		assert.Equal(t, tab[Offsetof(tPreload.BelongToID)], tab[Offsetof(tPreload.BelongTo)].(map[uintptr]interface{})[Offsetof(tBelong.ID)])
		assert.Equal(t, tab[Offsetof(tPreload.ID)], tab[Offsetof(tPreload.OneToOne)].(map[uintptr]interface{})[Offsetof(tOneToOne.TestPreloadTableID)])
		assert.Equal(t, tab[Offsetof(tPreload.ID)], tab[Offsetof(tPreload.OneToMany)].([]map[uintptr]interface{})[0][Offsetof(tOneToMany.TestPreloadTableID)])
		assert.Equal(t, tab[Offsetof(tPreload.ID)], tab[Offsetof(tPreload.OneToMany)].([]map[uintptr]interface{})[1][Offsetof(tOneToMany.TestPreloadTableID)])
	}
}

func TestPreloadSave(t *testing.T) {
	brick := TestDB.Model(&TestPreloadTable{})
	brick = brick.Preload(Offsetof(TestPreloadTable{}.BelongTo)).Enter()
	brick = brick.Preload(Offsetof(TestPreloadTable{}.OneToOne)).Enter()
	brick = brick.Preload(Offsetof(TestPreloadTable{}.OneToMany)).Enter()
	manyToManyPreload := brick.Preload(Offsetof(TestPreloadTable{}.ManyToMany))
	brick = manyToManyPreload.Enter()

	{
		manyToMany := []TestPreloadTableManyToMany{
			{
				Name: "test save 1 many_to_many 1",
			},
			{
				Name: "test save 1 many_to_many 2",
			},
		}
		tables := []TestPreloadTable{
			{
				Name: "test save 1",
				BelongTo: &TestPreloadTableBelongTo{
					Name: "test save sub1",
				},
				OneToOne: &TestPreloadTableOneToOne{
					Name: "test save sub2",
				},
				OneToMany: []TestPreloadTableOneToMany{
					{
						Name: "test save sub3 sub1",
					},
					{
						Name: "test save sub3 sub2",
					},
				},
			},
			{
				Name: "test save 2",
				BelongTo: &TestPreloadTableBelongTo{
					Name: "test save 2 sub1",
				},
				OneToOne: &TestPreloadTableOneToOne{
					Name: "test save 2 sub2",
				},
				OneToMany: []TestPreloadTableOneToMany{
					{
						Name: "test save 2 sub3 sub1",
					},
					{
						Name: "test save 2 sub3 sub2",
					},
				},
			},
		}
		// insert many to many
		result := manyToManyPreload.Insert(&manyToMany)

		if err := result.Err(); err != nil {
			t.Error(err)
			t.Failed()
		}
		assert.NotZero(t, manyToMany[0].ID)
		assert.NotZero(t, manyToMany[1].ID)
		// now many to many object have id information
		tables[0].ManyToMany = manyToMany
		tables[1].ManyToMany = manyToMany

		result = brick.Save(&tables)

		if err := result.Err(); err != nil {
			t.Error(err)
			t.Failed()
		}
		for _, tab := range tables {
			t.Logf("main id %v, belong id %v, one to one id %v, one to many id [%v,%v], many to many id [%v, %v]", tab.ID, tab.BelongTo.ID, tab.OneToOne.ID, tab.OneToMany[0].ID, tab.OneToMany[1].ID, tab.ManyToMany[0].ID, tab.ManyToMany[1].ID)
			assert.NotZero(t, tab.ID)
			assert.NotZero(t, tab.CreatedAt)
			assert.NotZero(t, tab.UpdatedAt)
			assert.NotZero(t, tab.BelongTo.ID)
			assert.NotZero(t, tab.OneToOne.ID)
			assert.NotZero(t, tab.OneToMany[0].ID)
			assert.NotZero(t, tab.OneToMany[1].ID)
			assert.NotZero(t, tab.ManyToMany[0].ID)
			assert.NotZero(t, tab.ManyToMany[1].ID)

			assert.Equal(t, tab.BelongToID, tab.BelongTo.ID)
			assert.Equal(t, tab.ID, tab.OneToOne.TestPreloadTableID)
			assert.Equal(t, tab.ID, tab.OneToMany[0].TestPreloadTableID)
			assert.Equal(t, tab.ID, tab.OneToMany[1].TestPreloadTableID)
		}
		// try to update soft delete
		now := time.Now()
		tables[0].DeletedAt = &now
		result = brick.Save(&tables)

		if err := result.Err(); err != nil {
			t.Error(err)
			t.Failed()
		}
	}

}

func TestPreloadFind(t *testing.T) {
	brick := TestDB.Model(TestPreloadTable{}).
		Preload(Offsetof(TestPreloadTable{}.BelongTo)).Enter().
		Preload(Offsetof(TestPreloadTable{}.OneToOne)).Enter().
		Preload(Offsetof(TestPreloadTable{}.OneToMany)).Enter().
		Preload(Offsetof(TestPreloadTable{}.ManyToMany)).Enter()
	{
		var tabs []TestPreloadTable
		brick.Find(&tabs)
		for _, tab := range tabs {
			var oneToManyIds []int32
			for _, sub := range tab.OneToMany {
				oneToManyIds = append(oneToManyIds, sub.ID)
			}
			var manyToManyIds []int32
			for _, sub := range tab.ManyToMany {
				manyToManyIds = append(manyToManyIds, sub.ID)
			}
			t.Logf("%#v\n", tab)
			t.Logf("main id %v, belong id %v, one to one id %v, one to many id list %v, many to many id list %v", tab.ID, tab.BelongTo.ID, tab.OneToOne.ID, oneToManyIds, manyToManyIds)
			assert.NotZero(t, tab.ID)
			assert.NotZero(t, tab.CreatedAt)
			assert.NotZero(t, tab.UpdatedAt)
			assert.NotZero(t, tab.BelongTo.ID)
			assert.NotZero(t, tab.OneToOne.ID)
			assert.Equal(t, len(tab.OneToMany), 2)
			for _, sub := range tab.OneToMany {
				assert.NotZero(t, sub.ID)
			}
			assert.Equal(t, len(tab.ManyToMany), 2)
			for _, sub := range tab.ManyToMany {
				assert.NotZero(t, sub.ID)
			}

			assert.Equal(t, tab.BelongToID, tab.BelongTo.ID)
			assert.Equal(t, tab.ID, tab.OneToOne.TestPreloadTableID)
			for _, sub := range tab.OneToMany {
				assert.Equal(t, tab.ID, sub.TestPreloadTableID)
			}
		}
	}
}

func TestPreloadDelete(t *testing.T) {
	var hardTab TestHardDeleteTable
	var softTab TestSoftDeleteTable
	var result *Result

	// delete middle first
	{
		hardHardMiddleBrick := TestDB.MiddleModel(&hardTab, &TestHardDeleteManyToMany{})
		result = hardHardMiddleBrick.DropTableIfExist()

		if err := result.Err(); err != nil {
			t.Error(err)
		}

		hardSoftMiddleBrick := TestDB.MiddleModel(&hardTab, &TestSoftDeleteManyToMany{})
		result = hardSoftMiddleBrick.DropTableIfExist()

		if err := result.Err(); err != nil {
			t.Error(err)
		}

		softHardMiddleBrick := TestDB.MiddleModel(&softTab, &TestHardDeleteManyToMany{})
		result = softHardMiddleBrick.DropTableIfExist()

		if err := result.Err(); err != nil {
			t.Error(err)
		}

		softSoftMiddleBrick := TestDB.MiddleModel(&softTab, &TestSoftDeleteManyToMany{})
		result = softSoftMiddleBrick.DropTableIfExist()

		if err := result.Err(); err != nil {
			t.Error(err)
		}
	}
	// delete hard table
	{
		hardBrick := TestDB.Model(&hardTab).
			Preload(Offsetof(hardTab.BelongTo)).Enter().
			Preload(Offsetof(hardTab.OneToOne)).Enter().
			Preload(Offsetof(hardTab.OneToMany)).Enter().
			Scope(foreignKeyManyToManyPreload(Offsetof(hardTab.ManyToMany))).Enter().
			Preload(Offsetof(hardTab.SoftBelongTo)).Enter().
			Preload(Offsetof(hardTab.SoftOneToOne)).Enter().
			Preload(Offsetof(hardTab.SoftOneToMany)).Enter().
			Scope(foreignKeyManyToManyPreload(Offsetof(hardTab.SoftManyToMany))).Enter()

		result = hardBrick.DropTableIfExist()

		if err := result.Err(); err != nil {
			t.Error(err)
		}
	}
	// delete soft table
	{
		brick := TestDB.Model(&softTab).
			Preload(Offsetof(softTab.BelongTo)).Enter().
			Preload(Offsetof(softTab.OneToOne)).Enter().
			Preload(Offsetof(softTab.OneToMany)).Enter().
			Scope(foreignKeyManyToManyPreload(Offsetof(softTab.ManyToMany))).Enter().
			Preload(Offsetof(softTab.SoftBelongTo)).Enter().
			Preload(Offsetof(softTab.SoftOneToOne)).Enter().
			Preload(Offsetof(softTab.SoftOneToMany)).Enter().
			Scope(foreignKeyManyToManyPreload(Offsetof(softTab.SoftManyToMany))).Enter()

		result = brick.DropTableIfExist()

		if err := result.Err(); err != nil {
			t.Error(err)
		}
	}
	// have same target foreign key ,need create table following table first
	{
		result := TestDB.Model(&TestHardDeleteTableBelongTo{}).CreateTable()

		if err := result.Err(); err != nil {
			t.Error(err)
		}

		result = TestDB.Model(&TestSoftDeleteTableBelongTo{}).CreateTable()

		if err := result.Err(); err != nil {
			t.Error(err)
		}

		result = TestDB.Model(&hardTab).CreateTable()

		if err := result.Err(); err != nil {
			t.Error(err)
		}

		result = TestDB.Model(&softTab).CreateTable()

		if err := result.Err(); err != nil {
			t.Error(err)
		}
	}
	t.Run("Hard Delete", func(t *testing.T) {
		brick := TestDB.Model(&hardTab).
			Preload(Offsetof(hardTab.BelongTo)).Enter().
			Preload(Offsetof(hardTab.OneToOne)).Enter().
			Preload(Offsetof(hardTab.OneToMany)).Enter().
			Scope(foreignKeyManyToManyPreload(Offsetof(hardTab.ManyToMany))).Enter().
			Preload(Offsetof(hardTab.SoftBelongTo)).Enter().
			Preload(Offsetof(hardTab.SoftOneToOne)).Enter().
			Preload(Offsetof(hardTab.SoftOneToMany)).Enter().
			Scope(foreignKeyManyToManyPreload(Offsetof(hardTab.SoftManyToMany))).Enter()

		result = brick.CreateTableIfNotExist()

		if err := result.Err(); err != nil {
			t.Error(err)
		}

		result = brick.Save([]TestHardDeleteTable{
			{
				Data:     "hard delete main model",
				BelongTo: &TestHardDeleteTableBelongTo{Data: "belong to data"},
				OneToOne: &TestHardDeleteTableOneToOne{Data: "one to one data"},
				OneToMany: []TestHardDeleteTableOneToMany{
					{Data: "one to many data"},
					{Data: "one to many data"},
				},
				ManyToMany: []TestHardDeleteManyToMany{
					{ID: 1, Data: "many to many data"},
					{ID: 2, Data: "many to many data"},
				},
				SoftBelongTo: &TestSoftDeleteTableBelongTo{Data: "belong to data"},
				SoftOneToOne: &TestSoftDeleteTableOneToOne{Data: "one to one data"},
				SoftOneToMany: []TestSoftDeleteTableOneToMany{
					{Data: "one to many data"},
					{Data: "one to many data"},
				},
				SoftManyToMany: []TestSoftDeleteManyToMany{
					{ModelDefault: ModelDefault{ID: 1}, Data: "many to many data"},
					{ModelDefault: ModelDefault{ID: 2}, Data: "many to many data"},
				},
			},
			{
				Data:     "hard delete main model",
				BelongTo: &TestHardDeleteTableBelongTo{Data: "belong to data"},
				OneToOne: &TestHardDeleteTableOneToOne{Data: "one to one data"},
				OneToMany: []TestHardDeleteTableOneToMany{
					{Data: "one to many data"},
					{Data: "one to many data"},
				},
				ManyToMany: []TestHardDeleteManyToMany{
					{ID: 1, Data: "many to many data"},
					{ID: 2, Data: "many to many data"},
				},
				SoftBelongTo: &TestSoftDeleteTableBelongTo{Data: "belong to data"},
				SoftOneToOne: &TestSoftDeleteTableOneToOne{Data: "one to one data"},
				SoftOneToMany: []TestSoftDeleteTableOneToMany{
					{Data: "one to many data"},
					{Data: "one to many data"},
				},
				SoftManyToMany: []TestSoftDeleteManyToMany{
					{ModelDefault: ModelDefault{ID: 1}, Data: "many to many data"},
					{ModelDefault: ModelDefault{ID: 2}, Data: "many to many data"},
				},
			},
		})

		if err := result.Err(); err != nil {
			t.Error(err)
		}

		var hardDeleteData []TestHardDeleteTable
		result = brick.Find(&hardDeleteData)

		if err := result.Err(); err != nil {
			t.Error(err)
		}

		result = brick.Delete(&hardDeleteData)
		if err := result.Err(); err != nil {
			t.Error(err)
		}
	})
	t.Run("SoftDelete", func(t *testing.T) {
		brick := TestDB.Model(&softTab).
			Preload(Offsetof(softTab.BelongTo)).Enter().
			Preload(Offsetof(softTab.OneToOne)).Enter().
			Preload(Offsetof(softTab.OneToMany)).Enter().
			Scope(foreignKeyManyToManyPreload(Offsetof(softTab.ManyToMany))).Enter().
			Preload(Offsetof(softTab.SoftBelongTo)).Enter().
			Preload(Offsetof(softTab.SoftOneToOne)).Enter().
			Preload(Offsetof(softTab.SoftOneToMany)).Enter().
			Scope(foreignKeyManyToManyPreload(Offsetof(softTab.SoftManyToMany))).Enter()
		result = brick.CreateTableIfNotExist()

		if err := result.Err(); err != nil {
			t.Error(err)
		}

		result = brick.Save([]TestSoftDeleteTable{
			{
				Data:     "hard delete main model",
				BelongTo: &TestHardDeleteTableBelongTo{Data: "belong to data"},
				OneToOne: &TestHardDeleteTableOneToOne{Data: "one to one data"},
				OneToMany: []TestHardDeleteTableOneToMany{
					{Data: "one to many data"},
					{Data: "one to many data"},
				},
				ManyToMany: []TestHardDeleteManyToMany{
					{ID: 1, Data: "many to many data"},
					{ID: 2, Data: "many to many data"},
				},
				SoftBelongTo: &TestSoftDeleteTableBelongTo{Data: "belong to data"},
				SoftOneToOne: &TestSoftDeleteTableOneToOne{Data: "one to one data"},
				SoftOneToMany: []TestSoftDeleteTableOneToMany{
					{Data: "one to many data"},
					{Data: "one to many data"},
				},
				SoftManyToMany: []TestSoftDeleteManyToMany{
					{ModelDefault: ModelDefault{ID: 1}, Data: "many to many data"},
					{ModelDefault: ModelDefault{ID: 2}, Data: "many to many data"},
				},
			},
			{
				Data:     "hard delete main model",
				BelongTo: &TestHardDeleteTableBelongTo{Data: "belong to data"},
				OneToOne: &TestHardDeleteTableOneToOne{Data: "one to one data"},
				OneToMany: []TestHardDeleteTableOneToMany{
					{Data: "one to many data"},
					{Data: "one to many data"},
				},
				ManyToMany: []TestHardDeleteManyToMany{
					{ID: 1, Data: "many to many data"},
					{ID: 2, Data: "many to many data"},
				},
				SoftBelongTo: &TestSoftDeleteTableBelongTo{Data: "belong to data"},
				SoftOneToOne: &TestSoftDeleteTableOneToOne{Data: "one to one data"},
				SoftOneToMany: []TestSoftDeleteTableOneToMany{
					{Data: "one to many data"},
					{Data: "one to many data"},
				},
				SoftManyToMany: []TestSoftDeleteManyToMany{
					{ModelDefault: ModelDefault{ID: 1}, Data: "many to many data"},
					{ModelDefault: ModelDefault{ID: 2}, Data: "many to many data"},
				},
			},
		})

		if err := result.Err(); err != nil {
			t.Error(err)
		}

		var softDeleteData []TestSoftDeleteTable
		result = brick.Find(&softDeleteData)

		if err := result.Err(); err != nil {
			t.Error(err)
		}

		result = brick.Delete(&softDeleteData)

		if err := result.Err(); err != nil {
			t.Error(err)
		}
	})

}

func TestCustomPreload(t *testing.T) {
	table := TestCustomPreloadTable{}
	tableOne := TestCustomPreloadOneToOne{}
	tableThree := TestCustomPreloadOneToMany{}
	middleTable := TestCustomPreloadManyToManyMiddle{}
	brick := TestDB.Model(&table).
		CustomOneToOnePreload(Offsetof(table.ChildOne), Offsetof(tableOne.ParentID)).Enter().
		CustomBelongToPreload(Offsetof(table.ChildTwo), Offsetof(table.BelongToID)).Enter().
		CustomOneToManyPreload(Offsetof(table.Children), Offsetof(tableThree.ParentID)).Enter().
		CustomManyToManyPreload(middleTable, Offsetof(table.OtherChildren), Offsetof(middleTable.ParentID), Offsetof(middleTable.ChildID)).Enter()
	result := brick.DropTableIfExist()

	if err := result.Err(); err != nil {
		t.Error(err)
	}

	result = brick.CreateTable()

	if err := result.Err(); err != nil {
		t.Error(err)
	}

	data := TestCustomPreloadTable{
		Data:     "custom main data",
		ChildOne: &TestCustomPreloadOneToOne{OneData: "custom one to one data"},
		ChildTwo: &TestCustomPreloadBelongTo{TwoData: "custom belong to data"},
		Children: []TestCustomPreloadOneToMany{
			{ThreeData: "custom one to many data 1"},
			{ThreeData: "custom one to many data 2"},
		},
		OtherChildren: []TestCustomPreloadManyToMany{
			{FourData: "custom many to many data 1"},
			{FourData: "custom many to many data 2"},
		},
	}
	result = brick.Insert(&data)

	if err := result.Err(); err != nil {
		t.Error(err)
	}

	assert.NotZero(t, data.ID)
	assert.NotZero(t, data.ChildOne.ID)
	assert.NotZero(t, data.ChildTwo.ID)
	assert.NotZero(t, data.Children[0].ID)
	assert.NotZero(t, data.Children[1].ID)
	assert.NotZero(t, data.OtherChildren[0].ID)
	assert.NotZero(t, data.OtherChildren[1].ID)

	assert.Equal(t, data.ChildTwo.ID, data.BelongToID)
	assert.Equal(t, data.ChildOne.ParentID, data.ID)
	assert.Equal(t, data.ID, data.Children[0].ParentID)
	assert.Equal(t, data.ID, data.Children[1].ParentID)

	result = brick.Delete(&data)

	if err := result.Err(); err != nil {
		t.Error(err)
	}
}

func TestFlow(t *testing.T) {
	// create a brick with product
	brick := TestDB.Model(&Product{}).
		Preload(Offsetof(Product{}.Detail)).Enter().
		Preload(Offsetof(Product{}.Address)).Enter().
		Preload(Offsetof(Product{}.Tag)).Enter().
		Preload(Offsetof(Product{}.Friend)).
		Preload(Offsetof(Product{}.Detail)).Enter().
		Preload(Offsetof(Product{}.Address)).Enter().
		Preload(Offsetof(Product{}.Tag)).Enter().
		Enter()
	// drow table if exist
	result := brick.DropTableIfExist()

	assert.Nil(t, result.Err())

	result = brick.CreateTableIfNotExist()

	assert.Nil(t, result.Err())

	product := []Product{
		{
			Detail: &ProductDetail{
				Page: "html code at here",
				Parameters: map[string]interface{}{
					"value":  "1",
					"value2": 1.0,
					"value3": 1.0,
				},
			},
			Price:    10.0,
			Discount: 1.0,
			Amount:   100000,
			Address: []Address{
				{Address1: "what ever", Address2: "what ever"},
				{Address1: "product origin", Address2: "origin 2"},
			},
			Contact: "phone number or email",
			//Favorites: 50,
			Version: "1.0",
			Tag: []Tag{
				{"food", ""},
				{"recommend", ""},
			},
			Friend: []Product{
				{
					Detail: &ProductDetail{
						Page: "html code at here",
						Parameters: map[string]interface{}{
							"value":  "1",
							"value2": 1.0,
							"value3": 1.0,
						},
					},
					Price:    11.0,
					Discount: 1.0,
					Amount:   100000,
					Address: []Address{
						{Address1: "what ever", Address2: "what ever"},
						{Address1: "product origin", Address2: "origin 2"},
					},
					Contact:   "phone number or email",
					Favorites: 50,
					Version:   "1.0",
					Tag: []Tag{
						{"food", ""},
						{"recommend", ""},
					},
				},
			},
		},
		{
			Detail: &ProductDetail{
				Page: "html code at here",
				Parameters: map[string]interface{}{
					"value":  "2",
					"value2": 2.0,
					"value3": 2.0,
				},
			},
			Price:    20.0,
			Discount: 1.0,
			Amount:   100000,
			Address: []Address{
				{Address1: "what ever", Address2: "what ever"},
				{Address1: "product origin", Address2: "origin 2"},
			},
			Contact:   "phone number or email",
			Favorites: 50,
			Version:   "1.0",
			Tag: []Tag{
				{"food", ""},
				{"bad", ""},
			},
		},
	}
	result = brick.Save(product)

	if err := result.Err(); err != nil {
		t.Error(err)
	}

	// add a new tag
	tagBrick := TestDB.Model(&Tag{})
	result = tagBrick.Insert(&Tag{Code: "nice"})

	if err := result.Err(); err != nil {
		t.Error(err)
	}

	//bind new tag to the one's product
	middleBrick := TestDB.MiddleModel(&Product{}, &Tag{})
	result = middleBrick.Save(&struct {
		ProductID uint32
		TagCode   string
	}{product[0].ID, "nice"})

	if err := result.Err(); err != nil {
		t.Error(err)
	}

	// try to find
	var newProducts []Product
	result = brick.Find(&newProducts)

	if err := result.Err(); err != nil {
		t.Error(err)
	}

	jsonBytes, err := json.MarshalIndent(newProducts, "", "  ")
	assert.NoError(t, err)
	t.Logf("\n%v", string(jsonBytes))
	result = brick.Debug().Delete(&product)
	assert.NoError(t, result.Err())
	t.Log(result.Report())
}

func TestGroupBy(t *testing.T) {
	//create table and insert data
	{
		brick := TestDB.Model(&TestGroupByTable{})
		result := brick.DropTableIfExist()

		if err := result.Err(); err != nil {
			t.Error(err)
		}

		result = brick.CreateTable()

		if err := result.Err(); err != nil {
			t.Error(err)
		}

		result = brick.Insert([]TestGroupByTable{
			{Name: "pigeon", Address: "aaa", Age: 1},
			{Name: "pigeon", Address: "aaa", Age: 2},
			{Name: "pigeon", Address: "aaa", Age: 3},
			{Name: "pigeon", Address: "aaa", Age: 4},
			{Name: "pigeon", Address: "bbb", Age: 2},
			{Name: "pigeon", Address: "bbb", Age: 4},
			{Name: "pigeon", Address: "bbb", Age: 1},
			{Name: "bigpigeon", Address: "aaa", Age: 1},
			{Name: "bigpigeon", Address: "bbb", Age: 1},
			{Name: "bigpigeon", Address: "bbb", Age: 2},
		})

		if err := result.Err(); err != nil {
			t.Error(err)
		}
	}
	{
		var tab TestGroupByTableGroup
		brick := TestDB.Model(&tab)

		brick = brick.GroupBy(Offsetof(tab.Name), Offsetof(tab.Address))
		var data []TestGroupByTableGroup
		result := brick.Find(&data)

		if err := result.Err(); err != nil {
			t.Error(err)
		}
		for _, d := range data {
			t.Logf("%#v\n", d)
		}

	}
}

func TestForeignKey(t *testing.T) {
	var tab TestForeignKeyTable
	var middleTab TestForeignKeyTableMiddle

	brick := TestDB.Model(&tab).
		Preload(Offsetof(tab.BelongTo)).Enter().
		Preload(Offsetof(tab.OneToOne)).Enter().
		Preload(Offsetof(tab.OneToMany)).Enter()
	brick = brick.CustomManyToManyPreload(
		&middleTab, Offsetof(tab.ManyToMany),
		Offsetof(middleTab.TestForeignKeyTableID),
		Offsetof(middleTab.TestForeignKeyTableManyToManyID),
	).Enter()

	result := brick.DropTableIfExist()

	if err := result.Err(); err != nil {
		t.Error(err)
	}

	result = brick.CreateTable()

	if err := result.Err(); err != nil {
		t.Error(err)
	}

	data := TestForeignKeyTable{
		Data: "test foreign key",
		BelongTo: &TestForeignKeyTableBelongTo{
			Data: "belong to data",
		},
		OneToOne: &TestForeignKeyTableOneToOne{
			Data: "one to one data",
		},
		OneToMany: []TestForeignKeyTableOneToMany{
			{Data: "one to many 1"},
			{Data: "one to mant 2"},
		},
		ManyToMany: []TestForeignKeyTableManyToMany{
			{Data: "many to many 1"},
			{Data: "many to many 2"},
		},
	}
	result = brick.Insert(&data)

	if err := result.Err(); err != nil {
		t.Error(err)
	}

	assert.NotZero(t, data.ID)
	assert.NotZero(t, data.BelongTo.ID)
	assert.NotZero(t, data.OneToOne.ID)
	assert.NotZero(t, data.OneToMany[0].ID)
	assert.NotZero(t, data.OneToMany[1].ID)

	assert.Equal(t, *data.BelongToID, data.BelongTo.ID)
	assert.Equal(t, data.ID, data.OneToOne.TestForeignKeyTableID)
	assert.Equal(t, data.ID, data.OneToMany[0].TestForeignKeyTableID)
	assert.Equal(t, data.ID, data.OneToMany[1].TestForeignKeyTableID)

	result = brick.Delete(&data)

	if err := result.Err(); err != nil {
		t.Error(err)
	}

}

func TestIgnorePreloadInsert(t *testing.T) {
	var tab TestPreloadIgnoreTable
	brick := TestDB.Model(&tab).
		Preload(Offsetof(tab.BelongTo)).Enter().
		Preload(Offsetof(tab.OneToOne)).Enter()

	result := brick.DropTableIfExist()

	if err := result.Err(); err != nil {
		t.Error(err)
	}

	result = brick.CreateTable()

	if err := result.Err(); err != nil {
		t.Error(err)
	}

	data := []TestPreloadIgnoreTable{
		{Data: "ignore preload 1"},
		{
			Data:     "ignore preload 2",
			BelongTo: TestPreloadIgnoreBelongTo{Data: "ignore preload 2 belong to"},
			OneToOne: TestPreloadIgnoreOneToOne{Data: "ignore preload 2 one to one"},
		},
	}
	result = brick.Insert(&data)

	if err := result.Err(); err != nil {
		t.Error(err)
	}
	assert.Zero(t, data[0].BelongToID)
	assert.Equal(t, data[1].OneToOne.TestPreloadIgnoreTableID, data[1].ID)
}

func TestMissPreloadFind(t *testing.T) {
	var tab TestMissTable
	var belongTab TestMissBelongTo
	var manyToManyTab TestMissManyToMany
	brick := TestDB.Model(&tab).
		Preload(Offsetof(tab.BelongTo)).Enter().
		Preload(Offsetof(tab.OneToOne)).Enter().
		Preload(Offsetof(tab.OneToMany)).Enter().
		Preload(Offsetof(tab.ManyToMany)).Enter()
	result := brick.DropTableIfExist()

	if err := result.Err(); err != nil {
		t.Error(err)
	}
	result = brick.CreateTable()

	if err := result.Err(); err != nil {
		t.Error(err)
	}

	missData := []TestMissTable{
		{
			Data: " miss data 1",
			BelongTo: &TestMissBelongTo{
				BelongToData: "miss data 1 belong to",
			},
			ManyToMany: []TestMissManyToMany{
				{ManyToManyData: "miss data 1 many to many 1"},
				{ManyToManyData: "miss data 1 many to many 2"},
			},
		},
		{
			Data: "miss data 2",
			BelongTo: &TestMissBelongTo{
				BelongToData: "miss data 2 belong to",
			},
			ManyToMany: []TestMissManyToMany{
				{ManyToManyData: "miss data 2 many to many 1"},
				{ManyToManyData: "miss data 2 many to many 2"},
			},
		},
	}
	// insert some data
	result = brick.Insert(&missData)

	if err := result.Err(); err != nil {
		t.Error(err)
	}
	// remove belong to data and many to many data
	result = TestDB.Model(&belongTab).
		Delete([]TestMissBelongTo{*missData[0].BelongTo, *missData[1].BelongTo})

	if err := result.Err(); err != nil {
		t.Error(err)
	}
	result = TestDB.Model(&manyToManyTab).
		Delete([]TestMissManyToMany{missData[0].ManyToMany[0], missData[1].ManyToMany[0]})

	if err := result.Err(); err != nil {
		t.Error(err)
	}
	// find again
	var scanMissData []TestMissTable

	result = brick.Find(&scanMissData)

	if err := result.Err(); err != nil {
		t.Error(err)
	}

	t.Logf("%#v\n", scanMissData)
	// TODO miss data need warning or error
}

func TestSameBelongId(t *testing.T) {
	var tab TestSameBelongIdTable
	brick := TestDB.Model(&tab).
		Preload(Offsetof(tab.BelongTo)).Enter()

	result := brick.DropTableIfExist()

	if err := result.Err(); err != nil {
		t.Error(err)
	}
	result = brick.CreateTable()

	if err := result.Err(); err != nil {
		t.Error(err)
	}
	data := []TestSameBelongIdTable{
		{Data: "test same belong id 1", BelongTo: TestSameBelongIdBelongTo{ID: 1, Data: "belong data"}},
		{Data: "test same belong id 2", BelongTo: TestSameBelongIdBelongTo{ID: 1, Data: "belong data"}},
	}
	result = brick.Save(&data)

	if err := result.Err(); err != nil {
		t.Error(err)
	}

	var findData []TestSameBelongIdTable
	result = brick.Find(&findData)
	t.Logf("%#v", findData)
}

func TestPointContainerField(t *testing.T) {
	var tab TestPointContainerTable
	brick := TestDB.Model(&tab).
		Preload(Offsetof(tab.OneToMany)).Enter().
		Preload(Offsetof(tab.ManyToMany)).Enter()
	createTableUnit(brick)(t)
	data := []*TestPointContainerTable{
		{
			Data: "point container table 1",
			OneToMany: &[]*TestPointContainerOneToMany{
				{OneToManyData: "point container table 1 one to many 1"},
				{OneToManyData: "point container table 1 one to many 2"},
			},
			ManyToMany: &[]*TestPointContainerManyToMany{
				{ManyToManyData: "point container table many to many 1"},
				{ManyToManyData: "point container table many to many 2"},
			},
		},
		{
			Data: "point container table 1",
			OneToMany: &[]*TestPointContainerOneToMany{
				{OneToManyData: "point container table 2 one to many 1"},
				{OneToManyData: "point container table 2 one to many 2"},
			},
			ManyToMany: &[]*TestPointContainerManyToMany{
				{ManyToManyData: "point container table many to many 3"},
				{ManyToManyData: "point container table many to many 4"},
			},
		},
	}
	result := brick.Insert(&data)

	if err := result.Err(); err != nil {
		t.Error(err)
	}

	var findData []*TestPointContainerTable
	result = brick.Find(&findData)

	if err := result.Err(); err != nil {
		t.Error(err)
	}
	jsonBytes, err := json.MarshalIndent(findData, "", "  ")
	assert.NoError(t, err)
	t.Logf("\n%v", string(jsonBytes))

	assert.Equal(t, data, findData)
}

func TestReport(t *testing.T) {
	var tab TestReportTable
	var tabSub1 TestReportSub1
	var tabSub2 TestReportSub2
	var tabSub3 TestReportSub3
	var tabSub4 TestReportSub4
	brick := TestDB.Model(&tab).
		Preload(Offsetof(tab.BelongTo)).
		Preload(Offsetof(tabSub1.BelongTo)).Enter().
		Preload(Offsetof(tabSub1.OneToOne)).Enter().
		Preload(Offsetof(tabSub1.OneToMany)).Enter().
		Preload(Offsetof(tabSub1.ManyToMany)).Enter().
		Enter().
		Preload(Offsetof(tab.OneToOne)).
		Preload(Offsetof(tabSub2.BelongTo)).Enter().
		Preload(Offsetof(tabSub2.OneToOne)).Enter().
		Preload(Offsetof(tabSub2.OneToMany)).Enter().
		Preload(Offsetof(tabSub2.ManyToMany)).Enter().
		Enter().
		Preload(Offsetof(tab.OneToMany)).
		Preload(Offsetof(tabSub3.BelongTo)).Enter().
		Preload(Offsetof(tabSub3.OneToOne)).Enter().
		Preload(Offsetof(tabSub3.OneToMany)).Enter().
		Preload(Offsetof(tabSub3.ManyToMany)).Enter().
		Enter().
		Preload(Offsetof(tab.ManyToMany)).
		Preload(Offsetof(tabSub4.BelongTo)).Enter().
		Preload(Offsetof(tabSub4.OneToOne)).Enter().
		Preload(Offsetof(tabSub4.OneToMany)).Enter().
		Preload(Offsetof(tabSub4.ManyToMany)).Enter().
		Enter()

	result := brick.DropTableIfExist()

	if err := result.Err(); err != nil {
		t.Error(err)
	}
	t.Log("\n", result.Report())

	result = brick.CreateTable()

	if err := result.Err(); err != nil {
		t.Error(err)
	}
	t.Log("\n", result.Report())

	var data []TestReportTable
	for i := 0; i < 2; i++ {
		reportStr := fmt.Sprintf("report data(%d)", i)
		tab := TestReportTable{
			Data: reportStr,
			BelongTo: &TestReportSub1{
				Sub1Data: reportStr + " sub 1",
				BelongTo: &TestReportSub1Sub1{
					Sub1Data: reportStr + " sub 1 sub 1",
				},
				OneToOne: &TestReportSub1Sub2{
					Sub2Data: reportStr + " sub 1 sub 2",
				},
			},
			OneToOne: &TestReportSub2{
				Sub2Data: reportStr + " sub 2",
				BelongTo: &TestReportSub2Sub1{
					Sub1Data: reportStr + " sub 2 sub 1",
				},
				OneToOne: &TestReportSub2Sub2{
					Sub2Data: reportStr + " sub 2 sub 2",
				},
			},
		}
		for j := 0; j < 2; j++ {
			tab.BelongTo.OneToMany = append(tab.BelongTo.OneToMany, TestReportSub1Sub3{
				Sub3Data: reportStr + " sub 1 " + fmt.Sprintf("sub 3(%d)", j),
			})
			tab.BelongTo.ManyToMany = append(tab.BelongTo.ManyToMany, TestReportSub1Sub4{
				Sub4Data: reportStr + " sub 1 " + fmt.Sprintf("sub 4(%d)", j),
			})
			tab.OneToOne.OneToMany = append(tab.OneToOne.OneToMany, TestReportSub2Sub3{
				Sub3Data: reportStr + " sub 2 " + fmt.Sprintf("sub 3(%d)", j),
			})
			tab.OneToOne.ManyToMany = append(tab.OneToOne.ManyToMany, TestReportSub2Sub4{
				Sub4Data: reportStr + " sub 2 " + fmt.Sprintf("sub 4(%d)", j),
			})
			sub3Str := fmt.Sprintf(" sub 3(%d)", j)
			tab.OneToMany = append(tab.OneToMany, TestReportSub3{
				Sub3Data: reportStr + sub3Str,
				BelongTo: &TestReportSub3Sub1{
					Sub1Data: reportStr + sub3Str + " sub 1",
				},
				OneToOne: &TestReportSub3Sub2{
					Sub2Data: reportStr + sub3Str + " sub 2",
				},
			})
			sub4Str := fmt.Sprintf(" sub 4(%d)", j)
			tab.ManyToMany = append(tab.ManyToMany, TestReportSub4{
				Sub4Data: reportStr + fmt.Sprintf(" sub 4(%d)", j),
				BelongTo: &TestReportSub4Sub1{
					Sub1Data: reportStr + sub4Str + " sub 1",
				},
				OneToOne: &TestReportSub4Sub2{
					Sub2Data: reportStr + sub4Str + " sub 2",
				},
			})
			for k := 0; k < 2; k++ {
				tab.OneToMany[j].OneToMany = append(tab.OneToMany[j].OneToMany, TestReportSub3Sub3{
					Sub3Data: reportStr + sub3Str + fmt.Sprintf(" sub 3(%d)", k),
				})

				tab.OneToMany[j].ManyToMany = append(tab.OneToMany[j].ManyToMany, TestReportSub3Sub4{
					Sub4Data: reportStr + sub3Str + fmt.Sprintf(" sub 4(%d)", k),
				})

				tab.ManyToMany[j].OneToMany = append(tab.ManyToMany[j].OneToMany, TestReportSub4Sub3{
					Sub3Data: reportStr + sub4Str + fmt.Sprintf(" sub 3(%d)", k),
				})

				tab.ManyToMany[j].ManyToMany = append(tab.ManyToMany[j].ManyToMany, TestReportSub4Sub4{
					Sub4Data: reportStr + sub4Str + fmt.Sprintf(" sub 4(%d)", k),
				})
			}
		}
		data = append(data, tab)
	}
	data[0].ID = 2

	result = brick.Save(&data)

	if err := result.Err(); err != nil {
		t.Error(err)
	}
	t.Log("\n", result.Report())

	var scanData []TestReportTable
	result = brick.Find(&scanData)

	if err := result.Err(); err != nil {
		t.Error(err)
	}
	t.Log("\n", result.Report())

	jsonBytes, err := json.MarshalIndent(scanData, "", "  ")
	assert.NoError(t, err)
	t.Logf("\n%v", string(jsonBytes))

	result = brick.Delete(&scanData)

	if err := result.Err(); err != nil {
		t.Error(err)
	}
	t.Log("\n", result.Report())
}

func TestRightValuePreload(t *testing.T) {
	var tab TestRightValuePreloadTable
	baseBrick := TestDB.Model(&tab)
	brick := baseBrick.Preload(Offsetof(tab.ManyToMany)).Enter()

	result := brick.DropTableIfExist()

	if err := result.Err(); err != nil {
		t.Error(err)
	}

	result = brick.CreateTableIfNotExist()

	if err := result.Err(); err != nil {
		t.Error(err)
	}

	data := TestRightValuePreloadTable{
		Data: "test right value preload",
		ManyToMany: []TestRightValuePreloadTable{
			{Data: "test right value preload sub 1"},
			{Data: "test right value preload sub 2"},
		},
	}

	result = brick.Insert(&data)

	if err := result.Err(); err != nil {
		t.Error(err)
	}

	var scanData TestRightValuePreloadTable
	findBrick := baseBrick.RightValuePreload(Offsetof(tab.ManyToMany)).Enter()
	findBrick = findBrick.Where(ExprEqual, Offsetof(tab.ID), data.ManyToMany[0].ID)
	result = findBrick.Find(&scanData)

	if err := result.Err(); err != nil {
		t.Error(err)
	}
	assert.Equal(t, scanData.ManyToMany[0].ID, data.ID)
	t.Logf("result:\n%s\n", result.Report())
}

func TestPreloadCheck(t *testing.T) {
	var tab TestPreloadCheckTable
	brick := TestDB.Model(&tab).
		Preload(Offsetof(tab.BelongTo)).Enter().
		Preload(Offsetof(tab.OneToOne)).Enter().
		Preload(Offsetof(tab.OneToMany)).Enter().
		Preload(Offsetof(tab.ManyToMany)).Enter()
	createTableUnit(brick)(t)
	type missingID struct {
		Data string

		BelongToID uint32
		BelongTo   *TestPreloadCheckBelongTo
		OneToOne   *TestPreloadCheckOneToOne
		OneToMany  []TestPreloadCheckOneToMany
		ManyToMany []TestPreloadCheckManyToMany
	}
	type missingBelongToID struct {
		ID   uint32
		Data string

		BelongTo   *TestPreloadCheckBelongTo
		OneToOne   *TestPreloadCheckOneToOne
		OneToMany  []TestPreloadCheckOneToMany
		ManyToMany []TestPreloadCheckManyToMany
	}
	type missingBelongTo struct {
		ID   uint32
		Data string

		BelongToID uint32
		OneToOne   *TestPreloadCheckOneToOne
		OneToMany  []TestPreloadCheckOneToMany
		ManyToMany []TestPreloadCheckManyToMany
	}

	type missingOneToOne struct {
		ID   uint32
		Data string

		BelongToID uint32
		BelongTo   *TestPreloadCheckBelongTo
		//OneToOne   *TestPreloadCheckOneToOne
		OneToMany  []TestPreloadCheckOneToMany
		ManyToMany []TestPreloadCheckManyToMany
	}

	type missingOneToOneRelationship struct {
		ID   uint32
		Data string

		BelongToID uint32
		BelongTo   *TestPreloadCheckBelongTo
		OneToOne   *struct {
			ID   uint32
			Data string
		}
		OneToMany  []TestPreloadCheckOneToMany
		ManyToMany []TestPreloadCheckManyToMany
	}

	type missingOneToMany struct {
		ID   uint32
		Data string

		BelongToID uint32
		BelongTo   *TestPreloadCheckBelongTo
		OneToOne   *TestPreloadCheckOneToOne
		//OneToMany  []TestPreloadCheckOneToMany
		ManyToMany []TestPreloadCheckManyToMany
	}

	type missingOneToManyRelationship struct {
		ID   uint32
		Data string

		BelongToID uint32
		BelongTo   *TestPreloadCheckBelongTo
		OneToOne   *TestPreloadCheckOneToOne
		OneToMany  []struct {
			ID   uint32
			Data string
		}
		ManyToMany []TestPreloadCheckManyToMany
	}

	type missingManyToMany struct {
		ID   uint32
		Data string

		BelongToID uint32
		BelongTo   *TestPreloadCheckBelongTo
		OneToOne   *TestPreloadCheckOneToOne
		OneToMany  []TestPreloadCheckOneToMany
		//ManyToMany []TestPreloadCheckManyToMany
	}

	type missingManyToManyID struct {
		ID   uint32
		Data string

		BelongToID uint32
		BelongTo   *TestPreloadCheckBelongTo
		OneToOne   *TestPreloadCheckOneToOne
		OneToMany  []TestPreloadCheckOneToMany
		ManyToMany []struct {
			//ID   uint32 `toyorm:"primary key;auto_increment"`
			Data string
		}
	}

	result := brick.Find(&missingID{})
	assert.Equal(t, result.Err().Error(), "struct missing ID field")

	result = brick.Find(&missingBelongToID{})
	assert.Equal(t, result.Err().Error(), "struct missing BelongToID field")

	result = brick.Find(&missingBelongTo{})
	assert.Equal(t, result.Err().Error(), "struct missing BelongTo field")

	result = brick.Find(&missingOneToOne{})
	assert.Equal(t, result.Err().Error(), "struct missing OneToOne field")

	result = brick.Find(&missingOneToOneRelationship{})
	assert.Equal(t, result.Err().Error(), "struct of the OneToOne field missing TestPreloadCheckTableID field")

	result = brick.Find(&missingOneToMany{})
	assert.Equal(t, result.Err().Error(), "struct missing OneToMany field")

	result = brick.Find(&missingOneToManyRelationship{})
	assert.Equal(t, result.Err().Error(), "struct of the OneToMany field missing TestPreloadCheckTableID field")

	result = brick.Find(&missingManyToMany{})
	assert.Equal(t, result.Err().Error(), "struct missing ManyToMany field")

	result = brick.Find(&missingManyToManyID{})
	assert.Equal(t, result.Err().Error(), "struct of the ManyToMany field missing ID field")
}

// some database cannot use table name like "order, group"
func TestTableNameProtect(t *testing.T) {

	brick := TestDB.Model(&User{}).
		Preload(Offsetof(User{}.Orders)).Enter()
	result := brick.DropTableIfExist()

	if err := result.Err(); err != nil {
		t.Error(err)
	}

	result = brick.CreateTable()

	if err := result.Err(); err != nil {
		t.Error(err)
	}

	data := User{
		Name:     "admin",
		Password: "12345",
		Orders: []Order{
			{Name: "MacBook", Num: 1},
		},
	}
	result = brick.Insert(&data)

	if err := result.Err(); err != nil {
		t.Error(err)
	}

	data.Orders[0].Num += 1
	result = brick.Save(&data)

	if err := result.Err(); err != nil {
		t.Error(err)
	}

	orderBrick := brick.Preload(Offsetof(User{}.Orders))
	result = orderBrick.Update(Order{
		Name: "Surface",
	})

	if err := result.Err(); err != nil {
		t.Error(err)
	}

	count, err := orderBrick.Count()
	assert.NoError(t, err)
	assert.Equal(t, count, 1)

	var scanData []User
	result = brick.Find(&scanData)

	if err := result.Err(); err != nil {
		t.Error(err)
	}

	result = brick.Delete(&data)

	if err := result.Err(); err != nil {
		t.Error(err)
	}
}

func TestCount(t *testing.T) {
	var tab TestCountTable
	brick := TestDB.Model(&tab)

	createTableUnit(brick)(t)
	// insert data
	var data []TestCountTable
	for i := 0; i < 21; i++ {
		data = append(data, TestCountTable{Data: fmt.Sprintf("test count %d", i)})
	}
	result := brick.Insert(data)

	if err := result.Err(); err != nil {
		t.Error(err)
	}

	count, err := brick.Count()
	assert.NoError(t, err)
	assert.Equal(t, count, 21)
}

func TestInsertFailure(t *testing.T) {
	type NotExistTable struct {
		ModelDefault
		Data string
	}
	brick := TestDB.Model(&NotExistTable{})
	data := NotExistTable{
		Data: "not exist table 1",
	}
	result := brick.Insert(&data)

	if err := result.Err(); err != nil {
		t.Log(err)
	} else {
		t.Error("insert failure need return error")
	}
}

func TestCustomExec(t *testing.T) {
	brick := TestDB.Model(&TestCustomExecTable{})
	createTableUnit(brick)(t)

	data := []TestCustomExecTable{
		{Data: "test custom exec table 1", Sync: 1},
		{ModelDefault: ModelDefault{ID: 55}, Data: "test custom exec table 2", Sync: 2},
	}

	var result *Result

	if TestDriver == "postgres" {
		result = brick.Template("INSERT INTO $ModelName($Columns) Values($Values) RETURNING $FN-ID").Insert(&data)
	} else {
		result = brick.Template("INSERT INTO $ModelName($Columns) Values($Values)").Insert(&data)
	}

	if err := result.Err(); err != nil {
		t.Error(err)
	}
	t.Logf("report:\n%s\n", result.Report())
	exec0 := result.ActionFlow[0].(ExecAction).Exec
	exec1 := result.ActionFlow[1].(ExecAction).Exec
	if TestDriver == "postgres" {
		assert.Equal(t, exec0.Query(), "INSERT INTO test_custom_exec_table(created_at,updated_at,deleted_at,data,sync,cas) Values($1,$2,$3,$4,$5,$6) RETURNING id")
		assert.Equal(t, exec1.Query(), "INSERT INTO test_custom_exec_table(id,created_at,updated_at,deleted_at,data,sync,cas) Values($1,$2,$3,$4,$5,$6,$7) RETURNING id")
	} else {
		assert.Equal(t, result.ActionFlow[0].(ExecAction).Exec.Query(), "INSERT INTO test_custom_exec_table(created_at,updated_at,deleted_at,data,sync,cas) Values(?,?,?,?,?,?)")
		assert.Equal(t, result.ActionFlow[1].(ExecAction).Exec.Query(), "INSERT INTO test_custom_exec_table(id,created_at,updated_at,deleted_at,data,sync,cas) Values(?,?,?,?,?,?,?)")
	}
	assert.NotZero(t, exec0.Args()[0])
	assert.Equal(t, exec0.Args()[0], data[0].CreatedAt)
	assert.NotZero(t, exec0.Args()[1])
	assert.Equal(t, exec0.Args()[1], data[0].UpdatedAt)
	assert.Nil(t, exec0.Args()[2])
	assert.Equal(t, exec0.Args()[3], data[0].Data)
	assert.Equal(t, exec0.Args()[4], data[0].Sync)
	assert.Equal(t, exec0.Args()[5], 1)

	assert.Equal(t, exec1.Args()[0], uint32(55))
	assert.NotZero(t, exec1.Args()[1])
	assert.Equal(t, exec1.Args()[1], data[1].CreatedAt)
	assert.NotZero(t, exec1.Args()[2])
	assert.Equal(t, exec1.Args()[2], data[1].UpdatedAt)
	assert.Nil(t, exec1.Args()[3])
	assert.Equal(t, exec1.Args()[4], data[1].Data)
	assert.Equal(t, exec1.Args()[5], data[1].Sync)
	assert.Equal(t, exec1.Args()[6], 1)

	switch TestDriver {
	case "mysql":
		result = brick.Template("INSERT  INTO $ModelDef($Columns) VALUES($Values) ON DUPLICATE KEY UPDATE $Cas $UpdateValues").Save(&data)

		if err := result.Err(); err != nil {
			t.Error(err)
		}
		tempStr := "INSERT  INTO `test_custom_exec_table`(id,created_at,updated_at,deleted_at,data,sync,cas) VALUES(?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE cas = IF(cas = VALUES(cas) - 1, VALUES(cas) , \"update failure\"), id = VALUES(id),updated_at = VALUES(updated_at),deleted_at = VALUES(deleted_at),data = VALUES(data),sync = VALUES(sync)"
		assert.Equal(t, result.ActionFlow[0].(ExecAction).Exec.Query(), tempStr)
		assert.Equal(t, result.ActionFlow[1].(ExecAction).Exec.Query(), tempStr)
	case "sqlite3":
		result = brick.Template("REPLACE  INTO $ModelDef($Columns) Values($Values)").Save(&data)

		if err := result.Err(); err != nil {
			t.Error(err)
		}
		tempStr := "REPLACE  INTO `test_custom_exec_table`(id,created_at,updated_at,deleted_at,data,sync,cas) Values(?,?,?,?,?,?,?)"
		assert.Equal(t, result.ActionFlow[0].(ExecAction).Exec.Query(), tempStr)
		assert.Equal(t, result.ActionFlow[1].(ExecAction).Exec.Query(), tempStr)
	case "postgres":
		result = brick.Template("INSERT  INTO $ModelDef($Columns) VALUES($Values) ON CONFLICT($PrimaryColumns) DO UPDATE SET $UpdateValues $Cas").Save(&data)

		if err := result.Err(); err != nil {
			t.Error(err)
		}
		tempStr := `INSERT  INTO "test_custom_exec_table"(id,created_at,updated_at,deleted_at,data,sync,cas) VALUES($1,$2,$3,$4,$5,$6,$7) ON CONFLICT(id) DO UPDATE SET id = Excluded.id,updated_at = Excluded.updated_at,deleted_at = Excluded.deleted_at,data = Excluded.data,sync = Excluded.sync,cas = Excluded.cas  WHERE test_custom_exec_table.cas = $8`
		assert.Equal(t, result.ActionFlow[0].(ExecAction).Exec.Query(), tempStr)
		assert.Equal(t, result.ActionFlow[1].(ExecAction).Exec.Query(), tempStr)
	}
	assert.Equal(t, data[0].Cas, 2)
	assert.Equal(t, data[1].Cas, 2)
	assert.True(t, data[0].UpdatedAt.After(data[0].CreatedAt))
	assert.True(t, data[1].UpdatedAt.After(data[1].CreatedAt))

	var scanData []TestCustomExecTable
	result = brick.Template("SELECT $Columns FROM $ModelName $Conditions").
		Limit(5).Offset(0).OrderBy(Offsetof(TestCustomExecTable{}.Data)).Find(&scanData)

	if err := result.Err(); err != nil {
		t.Error(err)
	}
	t.Logf("report:\n%s\n", result.Report())
	assert.Equal(t, result.ActionFlow[0].(QueryAction).Exec.Query(), "SELECT id,created_at,updated_at,deleted_at,data,sync,cas FROM test_custom_exec_table  WHERE deleted_at IS NULL ORDER BY data LIMIT 5")
	assert.Equal(t, len(scanData), len(data))
	for i := range data {
		assert.Equal(t, data[i].ID, scanData[i].ID)
		assert.Equal(t, data[i].Data, scanData[i].Data)
		assert.NotZero(t, scanData[i].CreatedAt)
		assert.NotZero(t, scanData[i].UpdatedAt)
	}

	result = brick.Template("UPDATE $ModelName SET $UpdateValues WHERE id = ?", 2).Update(&TestCustomExecTable{Sync: 5})

	if err := result.Err(); err != nil {
		t.Error(err)
	}
	t.Logf("report:\n%s\n", result.Report())
	if TestDriver == "postgres" {
		assert.Equal(t, result.ActionFlow[0].(ExecAction).Exec.Query(), "UPDATE test_custom_exec_table SET updated_at = $1,sync = $2 WHERE id = $3")
	} else {
		assert.Equal(t, result.ActionFlow[0].(ExecAction).Exec.Query(), "UPDATE test_custom_exec_table SET updated_at = ?,sync = ? WHERE id = ?")
	}
}

func TestToyBrickCopyOnWrite(t *testing.T) {
	var tab TestPreloadTable
	brick := TestDB.Model(&tab)

	{
		searchBrick := brick.Where(ExprEqual, Offsetof(tab.Name), "22").Offset(2).Limit(3)

		assert.Equal(t, brick.Search, SearchList(nil))
		assert.Equal(t, brick.offset, 0)
		assert.Equal(t, brick.limit, 0)

		searchBrick2 := searchBrick.And().Condition(ExprEqual, Offsetof(tab.Name), "33")
		sExec := DefaultDialect{}.SearchExec(searchBrick.Search)
		assert.Equal(t, sExec.Source(), "name = ?")
		assert.Equal(t, sExec.Args(), []interface{}{"22"})
		assert.Equal(t, searchBrick.offset, 2)
		assert.Equal(t, searchBrick.limit, 3)

		s2Exec := DefaultDialect{}.SearchExec(searchBrick2.Search)
		assert.Equal(t, s2Exec.Source(), "name = ? AND name = ?")
		assert.Equal(t, s2Exec.Args(), []interface{}{"22", "33"})
	}
	{
		preloadBrick := brick.Preload(Offsetof(tab.BelongTo)).Enter()
		assert.Zero(t, len(brick.MapPreloadBrick))
		assert.Zero(t, len(brick.BelongToPreload))

		preloadBrick2 := preloadBrick.Preload(Offsetof(tab.OneToOne)).Enter()
		assert.Equal(t, len(preloadBrick.MapPreloadBrick), 1)
		assert.Zero(t, len(preloadBrick.OneToOnePreload))

		preloadBrick3 := preloadBrick2.Preload(Offsetof(tab.OneToMany)).Enter()
		assert.Equal(t, len(preloadBrick2.MapPreloadBrick), 2)
		assert.Zero(t, len(preloadBrick2.OneToManyPreload))

		preloadBrick4 := preloadBrick3.Preload(Offsetof(tab.ManyToMany)).Enter()
		assert.Equal(t, len(preloadBrick3.MapPreloadBrick), 3)
		assert.Zero(t, len(preloadBrick3.ManyToManyPreload))

		assert.Equal(t, len(preloadBrick4.MapPreloadBrick), 4)
		assert.Equal(t, len(preloadBrick4.BelongToPreload), 1)
		assert.Equal(t, len(preloadBrick4.OneToOnePreload), 1)
		assert.Equal(t, len(preloadBrick4.OneToManyPreload), 1)
		assert.Equal(t, len(preloadBrick4.ManyToManyPreload), 1)

	}
	{
		var tab TestJoinTable
		var subTab TestJoinNameTable
		joinBrick := TestDB.Model(&tab)
		joinBrick2 := joinBrick.Join(Offsetof(tab.NameJoin))
		joinBrick2 = joinBrick2.OrderBy(Offsetof(subTab.SubData), Offsetof(subTab.Name))
		joinBrick2 = joinBrick2.GroupBy(Offsetof(subTab.Name))
		joinBrick2 = joinBrick2.Where(ExprEqual, Offsetof(subTab.Name), "sub")
		joinBrick2 = joinBrick2.Alias("sub")
		mainBrick := joinBrick2.Swap()
		assert.Equal(t, joinBrick.alias, "")
		assert.Equal(t, mainBrick.alias, "m")
		assert.Equal(t, mainBrick.Model, joinBrick.Model)

		assert.Equal(t, len(joinBrick2.OwnSearch), 1)
		assert.Equal(t, len(joinBrick2.OwnGroupBy), 1)
		assert.Equal(t, len(joinBrick2.OwnOrderBy), 2)
		assert.Equal(t, len(mainBrick.OwnSearch), 0)
		assert.Equal(t, len(mainBrick.OwnGroupBy), 0)
		assert.Equal(t, len(mainBrick.OwnOrderBy), 0)
	}
	{
		var tab TestJoinTable
		var nameTab TestJoinNameTable
		var priceTab TestJoinPriceTable
		brick := TestDB.Model(&tab).OrderBy(Offsetof(tab.Name)).GroupBy(Offsetof(tab.Data)).
			Where(ExprEqual, Offsetof(tab.Data), "test join 1").
			Join(Offsetof(tab.NameJoin)).OrderBy(Offsetof(nameTab.SubData)).GroupBy(Offsetof(nameTab.Name)).
			Or().Condition(ExprEqual, Offsetof(nameTab.SubData), "test join name 3").Swap().
			Join(Offsetof(tab.PriceJoin)).Join(Offsetof(priceTab.StarJoin)).Swap().Swap()
		brick2 := brick.OrderBy()
		assert.Equal(t, len(brick.OwnOrderBy), 1)
		assert.Equal(t, len(brick.Join(Offsetof(tab.NameJoin)).OwnOrderBy), 1)
		assert.Equal(t, len(brick2.OwnOrderBy), 0)
		assert.Equal(t, len(brick2.Join(Offsetof(tab.NameJoin)).OwnOrderBy), 0)

		brick3 := brick.GroupBy()
		assert.Equal(t, len(brick.OwnGroupBy), 1)
		assert.Equal(t, len(brick.Join(Offsetof(tab.NameJoin)).OwnGroupBy), 1)
		assert.Equal(t, len(brick3.OwnGroupBy), 0)
		assert.Equal(t, len(brick3.Join(Offsetof(tab.NameJoin)).OwnGroupBy), 0)

		brick4 := brick.Conditions(nil)
		assert.Equal(t, len(brick.OwnSearch), 1)
		assert.Equal(t, len(brick.Join(Offsetof(tab.NameJoin)).OwnSearch), 1)
		assert.Equal(t, len(brick4.OwnSearch), 0)
		assert.Equal(t, len(brick4.Join(Offsetof(tab.NameJoin)).OwnSearch), 0)
	}

}

func TestJoin(t *testing.T) {
	var tab TestJoinTable
	var nameTab TestJoinNameTable
	var priceTab TestJoinPriceTable
	var starTab TestJoinPriceSubStarTable
	// create table
	tabBrick := TestDB.Model(&tab)
	nameTabBrick := TestDB.Model(&nameTab).Preload(Offsetof(nameTab.OneToMany)).Enter()
	priceTabBrick := TestDB.Model(&priceTab)
	starTabBrick := TestDB.Model(&starTab)
	createTableUnit(tabBrick)(t)
	createTableUnit(nameTabBrick)(t)
	createTableUnit(priceTabBrick)(t)
	createTableUnit(starTabBrick)(t)
	// import data
	tabBrick.Insert(TestJoinTable{Name: "name 1", Data: "test join 1", Price: 1})
	tabBrick.Insert(TestJoinTable{Name: "name 2", Data: "test join 2", Price: 2})
	tabBrick.Insert(TestJoinTable{Name: "name 3", Data: "test join 3", Price: 3})

	nameTabBrick.Insert(TestJoinNameTable{Name: "name 1", SubData: "test join name 1", OneToMany: []TestJoinNameOneToManyTable{
		{PreloadData: "test name 1 one to many 1"},
		{PreloadData: "test name 1 one to many 2"},
	}})
	nameTabBrick.Insert(TestJoinNameTable{Name: "name 2", SubData: "test join name 2", OneToMany: []TestJoinNameOneToManyTable{
		{PreloadData: "test name 2 one to many 1"},
		{PreloadData: "test name 2 one to many 2"},
	}})
	nameTabBrick.Insert(TestJoinNameTable{Name: "name 3", SubData: "test join name 3", OneToMany: []TestJoinNameOneToManyTable{
		{PreloadData: "test name 3 one to many 1"},
		{PreloadData: "test name 3 one to many 2"},
	}})

	priceTabBrick.Insert(TestJoinPriceTable{Price: 1, SubData: "test join name 1", Star: 4})
	priceTabBrick.Insert(TestJoinPriceTable{Price: 2, SubData: "test join name 2", Star: 5})
	priceTabBrick.Insert(TestJoinPriceTable{Price: 3, SubData: "test join name 3", Star: 6})

	starTabBrick.Insert(TestJoinPriceSubStarTable{Star: 4, SubData: "test join name 1"})
	starTabBrick.Insert(TestJoinPriceSubStarTable{Star: 5, SubData: "test join name 2"})
	starTabBrick.Insert(TestJoinPriceSubStarTable{Star: 6, SubData: "test join name 3"})

	// join test
	{
		brick := tabBrick.
			Join(Offsetof(tab.NameJoin)).Swap().
			Join(Offsetof(tab.PriceJoin)).Join(Offsetof(priceTab.StarJoin)).Swap().Swap()
		var scanData []TestJoinTable
		result := brick.Find(&scanData)

		if err := result.Err(); err != nil {
			t.Error(err)
		}
		t.Logf("report:\n%s\n", result.Report())

		assert.Equal(t, len(scanData), 3)
		for _, elem := range scanData {
			assert.NotNil(t, elem.NameJoin)
			assert.Equal(t, elem.Name, elem.NameJoin.Name)
			assert.Equal(t, elem.Price, elem.PriceJoin.Price)
			assert.Equal(t, elem.PriceJoin.Star, elem.PriceJoin.StarJoin.Star)
		}
	}
	// join count
	{
		brick := tabBrick.Where(ExprEqual, Offsetof(tab.Name), "test join 1").
			Join(Offsetof(tab.NameJoin)).Where(ExprEqual, Offsetof(tab.NameJoin.SubData), "test join name 1").Swap().
			Join(Offsetof(tab.PriceJoin)).Join(Offsetof(priceTab.StarJoin)).Swap().Swap()
		count, err := brick.Count()
		require.NoError(t, err)
		require.Equal(t, count, 1)
		t.Logf("count %d", count)
	}
	// condition join test
	{
		brick := tabBrick.Where(ExprEqual, Offsetof(tab.Data), "test join 1").
			Join(Offsetof(tab.NameJoin)).Or().Condition(ExprEqual, Offsetof(nameTab.SubData), "test join name 3").Swap().
			Join(Offsetof(tab.PriceJoin)).Join(Offsetof(priceTab.StarJoin)).Swap().Swap()
		var scanData []TestJoinTable
		result := brick.Find(&scanData)

		if err := result.Err(); err != nil {
			t.Error(err)
		}
		t.Logf("report:\n%s\n", result.Report())

		assert.Equal(t, len(scanData), 2)

		for _, elem := range scanData {
			assert.Equal(t, elem.Name, elem.NameJoin.Name)
			assert.Equal(t, elem.Price, elem.PriceJoin.Price)
			assert.Equal(t, elem.PriceJoin.Star, elem.PriceJoin.StarJoin.Star)
		}
	}
	// preload on join
	{
		brick := tabBrick.
			Join(Offsetof(tab.NameJoin)).
			Preload(Offsetof(nameTab.OneToMany)).Enter().Swap()

		var scanData []TestJoinTable
		result := brick.Find(&scanData)

		if err := result.Err(); err != nil {
			t.Error(err)
		}
		t.Logf("report:\n%s\n", result.Report())

		assert.Equal(t, len(scanData), 3)
		for _, elem := range scanData {
			assert.NotNil(t, elem.NameJoin)
			assert.Equal(t, elem.Name, elem.NameJoin.Name)
			assert.Equal(t, len(elem.NameJoin.OneToMany), 2)
			for _, subElem := range elem.NameJoin.OneToMany {
				assert.Equal(t, subElem.TestJoinNameTableID, elem.NameJoin.ID)
			}
		}

	}
}

// bug: when call brick.Conditions(brick.Search) will lose all ownSearch information
func TestJoinAlias(t *testing.T) {
	var tab TestJoinTable
	var nameTab TestJoinNameTable
	var priceTab TestJoinPriceTable
	brick := TestDB.Model(&tab).OrderBy(Offsetof(tab.Name)).
		Where(ExprEqual, Offsetof(tab.Data), "test join 1").
		Join(Offsetof(tab.NameJoin)).
		Or().Condition(ExprEqual, Offsetof(nameTab.SubData), "test join name 3").Swap().
		Join(Offsetof(tab.PriceJoin)).Join(Offsetof(priceTab.StarJoin)).Swap().Swap()

	for _, i := range brick.OwnOrderBy {
		assert.Equal(t, brick.orderBy[i].Column(), brick.alias+"."+brick.orderBy[i].Source().Column())
	}
	for _, i := range brick.OwnGroupBy {
		assert.Equal(t, brick.groupBy[i].Column(), brick.alias+"."+brick.groupBy[i].Source().Column())
	}
	for _, i := range brick.OwnSearch {
		assert.Equal(t, brick.Search[i].Val.Column(), brick.alias+"."+brick.Search[i].Val.Source().Column())
	}
	for name := range brick.JoinMap {
		brick := brick.Join(name)
		for _, i := range brick.OwnOrderBy {
			assert.Equal(t, brick.orderBy[i].Column(), brick.alias+"."+brick.orderBy[i].Source().Column())
		}
		for _, i := range brick.OwnGroupBy {
			assert.Equal(t, brick.groupBy[i].Column(), brick.alias+"."+brick.groupBy[i].Source().Column())
		}
		for _, i := range brick.OwnSearch {
			assert.Equal(t, brick.Search[i].Val.Column(), brick.alias+"."+brick.Search[i].Val.Source().Column())
		}
	}

	brick = brick.Alias("m1")
	brick = brick.Join(Offsetof(tab.NameJoin)).Alias("n1").Swap()
	for _, i := range brick.OwnOrderBy {
		assert.Equal(t, brick.orderBy[i].Column(), brick.alias+"."+brick.orderBy[i].Source().Column())
	}
	for _, i := range brick.OwnGroupBy {
		assert.Equal(t, brick.groupBy[i].Column(), brick.alias+"."+brick.groupBy[i].Source().Column())
	}
	for _, i := range brick.OwnSearch {
		assert.Equal(t, brick.Search[i].Val.Column(), brick.alias+"."+brick.Search[i].Val.Source().Column())
	}
	for name := range brick.JoinMap {
		brick := brick.Join(name)
		for _, i := range brick.OwnOrderBy {
			assert.Equal(t, brick.orderBy[i].Column(), brick.alias+"."+brick.orderBy[i].Source().Column())
		}
		for _, i := range brick.OwnGroupBy {
			assert.Equal(t, brick.groupBy[i].Column(), brick.alias+"."+brick.groupBy[i].Source().Column())
		}
		for _, i := range brick.OwnSearch {
			assert.Equal(t, brick.Search[i].Val.Column(), brick.alias+"."+brick.Search[i].Val.Source().Column())
		}
	}
}

func TestAliasRelatedPreloadName(t *testing.T) {
	type BelongToSub struct {
		ID   uint32 `toyorm:"primary key;auto_increment"`
		Name string
	}
	type OneToOneSub struct {
		ID     uint32 `toyorm:"primary key;auto_increment"`
		MainID uint32 `toyorm:"one to one:OneToOneData"`
		Name   string
	}
	type OneToManySub struct {
		ID     uint32 `toyorm:"primary key;auto_increment"`
		MainID uint32 `toyorm:"one to many:OneToManyData"`
		Name   string
	}
	type MainTable struct {
		ModelDefault
		Name            string
		AliasBelongToID uint32 `toyorm:"belong to:BelongToData"`
		BelongToData    BelongToSub
		OneToOneData    OneToOneSub
		OneToManyData   []OneToManySub
	}

	brick := TestDB.Model(MainTable{}).
		Preload(Offsetof(MainTable{}.BelongToData)).Enter().
		Preload(Offsetof(MainTable{}.OneToOneData)).Enter().
		Preload(Offsetof(MainTable{}.OneToManyData)).Enter()

	belongToPreload := brick.BelongToPreload["BelongToData"]
	assert.Equal(t, belongToPreload.RelationField.Name(), "AliasBelongToID")
	oneToOnePreload := brick.OneToOnePreload["OneToOneData"]
	assert.Equal(t, oneToOnePreload.RelationField.Name(), "MainID")
	oneToManyPreload := brick.OneToManyPreload["OneToManyData"]
	assert.Equal(t, oneToManyPreload.RelationField.Name(), "MainID")

}

func TestSave(t *testing.T) {
	brick := TestDB.Model(&TestSaveTable{}).Debug()
	createTableUnit(brick)(t)

	data := TestSaveTable{
		Data: "test save",
	}

	result := brick.Save(&data)
	assert.NoError(t, result.Err())
	assert.NotZero(t, data.CreatedAt)
	assert.NotZero(t, data.UpdatedAt)

	oldCreatedAt := data.CreatedAt
	oldUpdateAt := data.UpdatedAt

	result = brick.Save(&data)
	assert.NoError(t, result.Err())
	assert.Equal(t, oldCreatedAt, data.CreatedAt)
	assert.True(t, data.UpdatedAt.After(oldUpdateAt))
}

func TestSaveNoneCreatedAt(t *testing.T) {
	skillTestDB(t, "sqlite3")
	brick := TestDB.Model(&TestSaveTable{}).Debug()
	createTableUnit(brick)(t)

	data := TestSaveTable{
		Data: "test save",
	}

	result := brick.Save(&data)
	assert.NoError(t, result.Err())
	assert.NotZero(t, data.CreatedAt)
	assert.NotZero(t, data.UpdatedAt)
	// reset created at & save & find it again
	var oldFData TestSaveTable
	result = brick.Find(&oldFData)
	require.NoError(t, result.Err())

	data.CreatedAt = time.Time{}
	result = brick.Save(&data)
	require.NoError(t, result.Err())

	var fData TestSaveTable
	result = brick.Find(&fData)

	require.NoError(t, result.Err())

	t.Logf("old createdAt %s new createdAt %s", oldFData.CreatedAt, fData.CreatedAt)
	assert.True(t, oldFData.CreatedAt.Equal(fData.CreatedAt))
}

func TestSaveCas(t *testing.T) {
	skillTestDB(t, "sqlite3")
	brick := TestDB.Model(&TestCasTable{}).Debug()
	createTableUnit(brick)(t)
	data := TestCasTable{
		Name:       "test cas data",
		UniqueData: "unique data",
	}
	result := brick.Insert(&data)
	assert.NoError(t, result.Err())
	assert.Equal(t, data.Cas, 1)

	data.Name += " 2"
	result = brick.Save(&data)
	assert.NoError(t, result.Err())
	assert.Equal(t, data.Cas, 2)

	data.Name += " 2"
	data.Cas--
	result = brick.Save(&data)

	resultErr := result.Err()
	assert.NotNil(t, resultErr)
	t.Log("error:\n", resultErr)

	t.Log("report:\n", result.Report())
}

func TestSaveWithUniqueIndex(t *testing.T) {
	brick := TestDB.Model(&TestUniqueIndexSaveTable{})
	createTableUnit(brick)(t)
	oldData := TestUniqueIndexSaveTable{
		ID:   1,
		Name: "unique",
		Data: "some data",
	}
	result := brick.Insert(&oldData)
	assert.NoError(t, result.Err())
	newData := TestUniqueIndexSaveTable{
		ID:   2,
		Name: "unique",
		Data: "some data 2",
	}
	// if here use save, will replace first record data in sqlite3
	//result = brick.Save(&newData)
	//resultProcessor(result, err)(t)

	// change name and insert
	newData.Name = "unique other"
	result = brick.Insert(&newData)
	assert.NoError(t, result.Err())

	// use USave(Save with Update) will get error
	newData.Name = "unique"
	result = brick.USave(&newData)
	assert.NoError(t, result.Err())
	resultErr := result.Err()
	assert.NotNil(t, resultErr)
	t.Log("error:\n", resultErr)
}

func TestRelateFieldTypeConvert(t *testing.T) {
	type TestBelongToSub struct {
		ID   int32 `toyorm:"primary key;auto_increment"`
		Data string
	}
	type TestOneToOneSub struct {
		ID                      int32 `toyorm:"primary key;auto_increment"`
		TestFieldConvertTableID int32
		Data                    string
	}
	type TestOneToManySub struct {
		ID                      int32 `toyorm:"primary key;auto_increment"`
		TestFieldConvertTableID int32
		Data                    string
	}
	type TestManyToManySub struct {
		ID   int32 `toyorm:"primary key;auto_increment"`
		Data string
	}
	type TestFieldConvertTable struct {
		ID   int32 `toyorm:"primary key;auto_increment"`
		Data string

		BelongToID int32
		BelongTo   *TestBelongToSub
		OneToOne   *TestOneToOneSub
		OneToMany  []TestOneToManySub
		ManyToMany []TestManyToManySub
	}

	// Follow data use to Insert or find
	type DtoID int32
	type DtoBelongTo struct {
		ID   DtoID
		Data string
	}
	type DtoOneToOne struct {
		ID                      DtoID
		TestFieldConvertTableID DtoID
		Data                    string
	}
	type DtoOneToMany struct {
		ID                      DtoID
		TestFieldConvertTableID DtoID
		Data                    string
	}
	type DtoManyToMany struct {
		ID   DtoID
		Data string
	}
	type DtoTable struct {
		ID   DtoID
		Data string

		BelongToID DtoID
		BelongTo   *DtoBelongTo
		OneToOne   *DtoOneToOne
		OneToMany  []DtoOneToMany
		ManyToMany []DtoManyToMany
	}

	var tab TestFieldConvertTable
	brick := TestDB.Model(&tab).
		Preload(Offsetof(tab.BelongTo)).Enter().
		Preload(Offsetof(tab.OneToOne)).Enter().
		Preload(Offsetof(tab.OneToMany)).Enter().
		Preload(Offsetof(tab.ManyToMany)).Enter()
	createTableUnit(brick)(t)
	data := []DtoTable{
		{Data: "main data 1", BelongTo: &DtoBelongTo{
			Data: "belong to 1 sub ",
		}, OneToOne: &DtoOneToOne{
			Data: "one to one 1 sub ",
		}, OneToMany: []DtoOneToMany{
			{Data: "one to many 1 data 1"},
			{Data: "one to many 1 data 2"},
		}, ManyToMany: []DtoManyToMany{
			{Data: "many to many 1 data 1"},
			{Data: "many to many 1 data 2"},
		}},
		{Data: "main data 2", BelongTo: &DtoBelongTo{
			Data: "belong to 2 sub ",
		}, OneToOne: &DtoOneToOne{
			Data: "one to one 2 sub ",
		}, OneToMany: []DtoOneToMany{
			{Data: "one to many 2 data 1"},
			{Data: "one to many 2 data 2"},
		}, ManyToMany: []DtoManyToMany{
			{Data: "many to many 2 data 1"},
			{Data: "many to many 2 data 2"},
		}},
	}
	result := brick.Insert(&data)
	assert.NoError(t, result.Err())

	for _, d := range data {
		assert.NotZero(t, d.ID)
		assert.NotZero(t, d.BelongToID)
		assert.NotZero(t, d.BelongTo)
		assert.NotZero(t, d.BelongTo.ID)
		assert.NotZero(t, d.OneToOne)
		assert.NotZero(t, d.OneToOne.ID)
		assert.Equal(t, len(d.OneToMany), 2)
		for _, s := range d.OneToMany {
			assert.NotZero(t, s.ID)
			assert.NotZero(t, s.TestFieldConvertTableID)
		}
		assert.Equal(t, len(d.ManyToMany), 2)
		for _, s := range d.ManyToMany {
			assert.NotZero(t, s.ID)
		}
	}
	t.Log(result.Report())
	{
		var data []DtoTable
		result := brick.Find(&data)

		assert.NoError(t, result.Err())
		for _, d := range data {
			assert.NotZero(t, d.ID)
			assert.NotZero(t, d.BelongToID)
			assert.NotZero(t, d.BelongTo)
			assert.NotZero(t, d.BelongTo.ID)
			assert.NotZero(t, d.OneToOne)
			assert.NotZero(t, d.OneToOne.ID)
			assert.Equal(t, len(d.OneToMany), 2)
			for _, s := range d.OneToMany {
				assert.NotZero(t, s.ID)
				assert.NotZero(t, s.TestFieldConvertTableID)
			}
			assert.Equal(t, len(d.ManyToMany), 2)
			for _, s := range d.ManyToMany {
				assert.NotZero(t, s.ID)
			}
		}
		t.Log(result.Report())
	}
}

func TestSaveWithOther(t *testing.T) {
	skillTestDB(t, "sqlite3")
	brick := TestDB.Model(&TestSaveWithOtherTable{})
	createTableUnit(brick)(t)
	data := TestSaveWithOtherTable{
		Name: "pigeon",
	}
	result := brick.Save(&data)

	assert.NoError(t, result.Err())

	type OtherTable struct {
		ID  uint32
		Age int
	}
	otherData := OtherTable{ID: data.ID, Age: 22}
	result = brick.Save(&otherData)

	assert.NoError(t, result.Err())

	var fData TestSaveWithOtherTable
	result = brick.Find(&fData)

	assert.NoError(t, result.Err())
	assert.Equal(t, fData.Name, "pigeon")

	// test insert only id
	type OtherTable2 struct {
		ID uint32
	}
	otherData2 := OtherTable2{}
	result = brick.Save(&otherData2)

	assert.NoError(t, result.Err())

	var fData2 TestSaveWithOtherTable
	result = brick.Where(ExprEqual, Offsetof(TestSaveWithOtherTable{}.ID), otherData2.ID).Find(&fData2)

	assert.NoError(t, result.Err())
	assert.Equal(t, fData2.Name, "")
	assert.Equal(t, fData2.Age, 0)
}

func TestCustomTableName(t *testing.T) {
	for _, i := range []int{1, 2, 10} {
		tab := TestCustomTableNameTable{
			FragNum:    i,
			BelongTo:   &TestCustomTableNameBelongTo{FragNum: i},
			OneToOne:   &TestCustomTableNameOneToOne{FragNum: i},
			OneToMany:  []TestCustomTableNameOneToMany{{FragNum: i}},
			ManyToMany: []TestCustomTableNameManyToMany{{FragNum: i}},
			Join:       &TestCustomTableNameJoin{FragNum: i},
		}
		brick := TestDB.Model(&tab).
			Preload(Offsetof(tab.BelongTo)).Enter().
			Preload(Offsetof(tab.OneToOne)).Enter().
			Preload(Offsetof(tab.OneToMany)).Enter().
			Preload(Offsetof(tab.ManyToMany)).Enter().
			Join(Offsetof(tab.Join)).Swap()
		require.Equal(t, brick.Model.Name, "test_custom_table_name_table_"+fmt.Sprint(i))
		require.Equal(t, brick.BelongToPreload["BelongTo"].SubModel.Name, "test_custom_table_name_belong_to_"+fmt.Sprint(i))
		require.Equal(t, brick.OneToOnePreload["OneToOne"].SubModel.Name, "test_custom_table_name_one_to_one_"+fmt.Sprint(i))
		require.Equal(t, brick.OneToManyPreload["OneToMany"].SubModel.Name, "test_custom_table_name_one_to_many_"+fmt.Sprint(i))
		require.Equal(t, brick.ManyToManyPreload["ManyToMany"].SubModel.Name, "test_custom_table_name_many_to_many_"+fmt.Sprint(i))
		require.Equal(t, brick.ManyToManyPreload["ManyToMany"].MiddleModel.Name, fmt.Sprintf("test_custom_table_name_many_to_many_%[1]d_test_custom_table_name_table_%[1]d", i))
		require.Equal(t, brick.JoinMap["Join"].SubModel.Name, "test_custom_table_name_join_"+fmt.Sprint(i))

	}
}

func TestDefaultValue(t *testing.T) {
	type TestDefaultTable struct {
		ID           uint32 `toyorm:"primary key;auto_increment"`
		Data         string
		DefaultStr   string  `toyorm:"default:'test';NOT NULL"`
		DefaultInt   int     `toyorm:"default:200;NOT NULL"`
		DefaultFloat float64 `toyorm:"default:52.5;NOT NULL"`
		DefaultBool  bool    `toyorm:"default:true;NOT NULL"`
	}
	brick := TestDB.Model(&TestDefaultTable{}).Debug()
	createTableUnit(brick)(t)
	data := TestDefaultTable{}
	result := brick.Insert(&data)
	require.NoError(t, result.Err())

	fData := TestDefaultTable{Data: "test default value"}
	result = brick.Find(&fData)
	require.NoError(t, result.Err())

	t.Log(fData)
	assert.Equal(t, fData.DefaultStr, "test")
	assert.Equal(t, fData.DefaultInt, 200)
	assert.Equal(t, fData.DefaultFloat, 52.5)
	assert.Equal(t, fData.DefaultBool, true)
}

func TestTempField(t *testing.T) {
	type TestTempFieldTable struct {
		ID    uint32 `toyorm:"primary key;auto_increment"`
		Data  string
		Tag   string
		Score int32
	}

	brick := TestDB.Model(&TestTempFieldTable{})
	createTableUnit(brick)(t)

	for i := 0; i < 10; i++ {
		result := brick.Insert(&TestTempFieldTable{
			Data:  "TEST",
			Tag:   fmt.Sprint(i % 2),
			Score: int32(i),
		})
		assert.NoError(t, result.Err())
	}
	brick = brick.Alias("m")
	var data []TestTempFieldTable
	result := brick.BindDefaultFields(
		Offsetof(TestTempFieldTable{}.Tag),
		brick.TempField(Offsetof(TestTempFieldTable{}.Score), "MAX(%s)"),
	).GroupBy(Offsetof(TestTempFieldTable{}.Tag)).
		Where(ExprEqual, brick.TempField(Offsetof(TestTempFieldTable{}.Data), "LOWER(%s)"), "test").
		OrderBy(brick.TempField(Offsetof(TestTempFieldTable{}.Tag), "%s DESC")).
		Find(&data)
	assert.NoError(t, result.Err())
	t.Logf("%#v\n", data)
	for _, d := range data {
		if d.Tag == "1" {
			require.Equal(t, d.Score, int32(9))
		} else if d.Tag == "0" {
			require.Equal(t, d.Score, int32(8))
		}
	}
}

func TestDelete(t *testing.T) {
	{
		type DeleteTable struct {
			ID   uint32 `toyorm:"primary key"`
			Data string
			Num  int
		}

		brick := TestDB.Model(&DeleteTable{})
		createTableUnit(brick)(t)
		data := []DeleteTable{
			{ID: 1, Data: "test data 1", Num: 1},
			{ID: 2, Data: "test data 2", Num: 1},
			{ID: 3, Data: "test data 3", Num: 2},
			{ID: 4, Data: "test data 4", Num: 2},
		}
		result := brick.Insert(&data)

		require.NoError(t, result.Err())
		result = brick.Delete(data)

		require.NoError(t, result.Err())
		t.Log(result.Report())
		result = brick.Where(ExprEqual, Offsetof(DeleteTable{}.Num), 2).DeleteWithConditions()

		require.NoError(t, result.Err())
		count, err := brick.Count()
		require.NoError(t, err)
		require.Equal(t, count, 0)
	}
	{

		type DeleteSoftTable struct {
			ID        uint32 `toyorm:"primary key"`
			Data      string
			Num       int
			DeletedAt *time.Time
		}
		brick := TestDB.Model(&DeleteSoftTable{})
		createTableUnit(brick)(t)
		data := []DeleteSoftTable{
			{ID: 1, Data: "test data 1", Num: 1},
			{ID: 2, Data: "test data 2", Num: 1},
			{ID: 3, Data: "test data 3", Num: 2},
			{ID: 4, Data: "test data 4", Num: 2},
		}
		result := brick.Insert(&data)

		require.NoError(t, result.Err())
		result = brick.Delete(data[:2])

		require.NoError(t, result.Err())
		t.Log(result.Report())
		result = brick.Where(ExprEqual, Offsetof(DeleteSoftTable{}.Num), 2).DeleteWithConditions()

		require.NoError(t, result.Err())
		count, err := brick.Count()
		require.NoError(t, err)
		require.Equal(t, count, 0)
	}
}
