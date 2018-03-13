package toyorm

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
	"time"
	. "unsafe"
)

func TestCollectionCreateTable(t *testing.T) {
	var hastable []bool
	var err error
	for _, tab := range []interface{}{
		TestCreateTable1{},
		TestCreateTable2{},
		TestCreateTable3{},
		TestCreateTable4{},
		SqlTypeTable{},
	} {
		// start a session

		brick := TestCollectionDB.Model(tab).Debug()
		hastable, err = brick.HasTable()
		assert.Nil(t, err)
		t.Logf("table %s exist:%v\n", brick.model.Name, hastable)
		result, err := brick.DropTableIfExist()
		assert.Nil(t, err)
		if err := result.Err(); err != nil {
			t.Error(err)
			t.Fail()
		}
		result, err = brick.CreateTable()
		assert.Nil(t, err)
		assert.Nil(t, err)
		if err := result.Err(); err != nil {
			t.Error(err)
			t.Fail()
		}
	}
}

func TestCollectionInsertData(t *testing.T) {
	brick := TestCollectionDB.Model(&TestInsertTable{}).Debug()
	// add id generator
	TestCollectionDB.SetModelHandlers("Insert", brick.model, CollectionHandlersChain{CollectionIDGenerate})
	//create table
	{
		result, err := brick.DropTableIfExist()
		assert.Nil(t, err)
		if err := result.Err(); err != nil {
			t.Error(err)
			t.Fail()
		}
		result, err = brick.CreateTable()
		assert.Nil(t, err)
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
		result, err := brick.Insert(tab)
		assert.Nil(t, err)
		if err := result.Err(); err != nil {
			t.Error(err)
			t.Fail()
		}
		t.Logf("id %v created at: %v updated at: %v", tab.ID, tab.CreatedAt, tab.UpdatedAt)
		assert.Zero(t, tab.ID)
		assert.Zero(t, tab.CreatedAt)
		assert.Zero(t, tab.UpdatedAt)
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
		result, err := brick.Insert(tab)
		assert.Nil(t, err)
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
		result, err := brick.Insert(tabMap)
		assert.Nil(t, err)
		if err := result.Err(); err != nil {
			t.Error(err)
			t.Fail()
		}
		t.Logf("id %v created at: %v updated at: %v", tabMap[Offsetof(tab.ID)], tabMap[Offsetof(tab.CreatedAt)], tabMap[Offsetof(tab.UpdatedAt)])
		assert.NotZero(t, tabMap[Offsetof(tab.ID)])
		assert.NotZero(t, tabMap[Offsetof(tab.CreatedAt)])
		assert.NotZero(t, tabMap[Offsetof(tab.UpdatedAt)])
	}
	// insert list
	// insert with struct
	{
		dstr := "some str data"
		dint := 100
		dfloat := 100.0
		dcomplex := 100 + 1i
		tabs := []TestInsertTable{
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
		result, err := brick.Insert(tabs)
		assert.Nil(t, err)
		if err := result.Err(); err != nil {
			t.Error(err)
			t.Fail()
		}
		for _, tab := range tabs {
			t.Logf("id %v created at: %v updated at: %v", tab.ID, tab.CreatedAt, tab.UpdatedAt)
			assert.NotZero(t, tab.ID)
			assert.NotZero(t, tab.CreatedAt)
			assert.NotZero(t, tab.UpdatedAt)
		}
	}
	// test insert map[string]interface{}
	{
		dstr := "some str data"
		dint := 100
		dfloat := 100.0
		dcomplex := 100 + 1i
		tabs := []map[string]interface{}{
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
		result, err := brick.Insert(tabs)
		assert.Nil(t, err)
		if err := result.Err(); err != nil {
			t.Error(err)
			t.Fail()
		}
		for _, tab := range tabs {
			t.Logf("id %v created at: %v updated at: %v", tab["ID"], tab["CreatedAt"], tab["UpdatedAt"])
			assert.NotZero(t, tab["ID"])
			assert.NotZero(t, tab["CreatedAt"])
			assert.NotZero(t, tab["UpdatedAt"])
		}
	}
	// test insert map[OffsetOf]interface{}
	{
		dstr := "some str data"
		dint := 100
		dfloat := 100.0
		dcomplex := 100 + 1i
		var tab TestInsertTable
		data := []map[uintptr]interface{}{
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
		result, err := brick.Insert(data)
		assert.Nil(t, err)
		if err := result.Err(); err != nil {
			t.Error(err)
			t.Fail()
		}
		for _, obj := range data {
			t.Logf("id %v created at: %v updated at: %v", obj[Offsetof(tab.ID)], obj[Offsetof(tab.CreatedAt)], obj[Offsetof(tab.UpdatedAt)])
			assert.NotZero(t, obj[Offsetof(tab.ID)])
			assert.NotZero(t, obj[Offsetof(tab.CreatedAt)])
			assert.NotZero(t, obj[Offsetof(tab.UpdatedAt)])
		}

	}
	// have selector data
	{
		brick := TestCollectionDB.Model(&TestInsertTable{}).Debug()
		dstr := "some str data"
		dint := 100
		dfloat := 100.0
		dcomplex := 100 + 1i
		data := []TestInsertSelector{
			{TestInsertTable{
				DataStr:     "some str data",
				DataInt:     100,
				DataFloat:   100.0,
				DataComplex: 100 + 1i,
				PtrStr:      &dstr,
				PtrInt:      &dint,
				PtrFloat:    &dfloat,
				PtrComplex:  &dcomplex,
			}},
			{TestInsertTable{
				DataStr:     "some str data",
				DataInt:     100,
				DataFloat:   100.0,
				DataComplex: 100 + 1i,
				PtrStr:      &dstr,
				PtrInt:      &dint,
				PtrFloat:    &dfloat,
				PtrComplex:  &dcomplex,
			}},
			{TestInsertTable{
				DataStr:     "some str data",
				DataInt:     100,
				DataFloat:   100.0,
				DataComplex: 100 + 1i,
				PtrStr:      &dstr,
				PtrInt:      &dint,
				PtrFloat:    &dfloat,
				PtrComplex:  &dcomplex,
			}},
			{TestInsertTable{
				DataStr:     "some str data",
				DataInt:     100,
				DataFloat:   100.0,
				DataComplex: 100 + 1i,
				PtrStr:      &dstr,
				PtrInt:      &dint,
				PtrFloat:    &dfloat,
				PtrComplex:  &dcomplex,
			}},
		}
		result, err := brick.Insert(data)
		assert.Nil(t, err)
		if err := result.Err(); err != nil {
			t.Error(err)
			t.Fail()
		}
		t.Log(data)
	}
}

func TestCollectionInsertPointData(t *testing.T) {
	brick := TestCollectionDB.Model(&TestInsertTable{}).Debug()
	// add id generator
	TestCollectionDB.SetModelHandlers("Insert", brick.model, CollectionHandlersChain{CollectionIDGenerate})

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
		result, err := brick.Insert(&tab)
		assert.Nil(t, err)
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
		result, err := brick.Insert(&tab)
		assert.Nil(t, err)
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
		result, err := brick.Insert(&tabMap)
		assert.Nil(t, err)
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

func TestCollectionFind(t *testing.T) {
	{
		brick := TestCollectionDB.Model(&TestSearchTable{}).Debug()
		// add id generator
		TestCollectionDB.SetModelHandlers("Insert", brick.model, CollectionHandlersChain{CollectionIDGenerate})

		result, err := brick.DropTableIfExist()
		assert.Nil(t, err)
		if err := result.Err(); err != nil {
			t.Error(err)
			t.Fail()
		}
		result, err = brick.CreateTable()
		assert.Nil(t, err)
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
						result, err := TestCollectionDB.Model(&TestSearchTable{}).Debug().Insert(&t1)
						assert.Nil(t, err)
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
		result, err := TestCollectionDB.Model(&TestSearchTable{}).Debug().Find(&table)
		assert.Nil(t, err)
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
		result, err := TestCollectionDB.Model(&TestSearchTable{}).Debug().Find(&tables)
		assert.Nil(t, err)
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

func TestCollectionConditionFind(t *testing.T) {
	base := TestSearchTable{}
	//SELECT id,a,b,c,d FROM test_search_table WHERE a = ? AND b = ?, args:[]interface {}{"a", "b"}
	{
		var tabs []TestSearchTable
		result, err := TestCollectionDB.Model(&TestSearchTable{}).Debug().
			Where(ExprAnd, TestSearchTable{A: "a", B: "b"}).Find(&tabs)
		assert.Nil(t, err)
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
		result, err := TestCollectionDB.Model(&TestSearchTable{}).Debug().
			Where(ExprOr, TestSearchTable{A: "a", B: "bb"}).Find(&tabs)
		assert.Nil(t, err)
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
		result, err := TestCollectionDB.Model(&TestSearchTable{}).Debug().
			Where(ExprEqual, Offsetof(base.A), "a").
			And().Condition(ExprEqual, Offsetof(base.B), "b").Find(&tabs)
		assert.Nil(t, err)
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
		result, err := TestCollectionDB.Model(&TestSearchTable{}).Debug().
			Where(ExprEqual, Offsetof(base.A), "a").And().
			Condition(ExprEqual, Offsetof(base.B), "b").Or().
			Condition(ExprEqual, Offsetof(base.C), "c").Or().
			Condition(ExprEqual, Offsetof(base.D), "d").And().
			Condition(ExprEqual, Offsetof(base.A), "aa").
			Find(&tabs)
		assert.Nil(t, err)
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
		brick := TestCollectionDB.Model(&TestSearchTable{}).Debug()

		result, err := brick.Where(ExprEqual, Offsetof(base.A), "aa").And().
			Conditions(
				brick.Where(ExprEqual, Offsetof(base.A), "a").Or().
					Condition(ExprEqual, Offsetof(base.B), "b").Or().
					Condition(ExprEqual, Offsetof(base.C), "c").Search,
			).
			Find(&tabs)
		assert.Nil(t, err)
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

func TestCollectionCombinationConditionFind(t *testing.T) {
	brick := TestCollectionDB.Model(&TestSearchTable{}).Debug()
	{
		var tabs []TestSearchTable
		result, err := brick.Where(ExprAnd, TestSearchTable{A: "a", B: "b"}).Find(&tabs)
		assert.Nil(t, err)
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
		result, err := brick.Where(ExprAnd, map[string]interface{}{"A": "a", "B": "b"}).Find(&tabs)
		assert.Nil(t, err)
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
		result, err := brick.Where(ExprAnd, map[uintptr]interface{}{
			Offsetof(TestSearchTable{}.A): "a",
			Offsetof(TestSearchTable{}.B): "b",
		}).Find(&tabs)
		assert.Nil(t, err)
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
		result, err := brick.Where(ExprOr, map[string]interface{}{"A": "a", "B": "b"}).And().
			Condition(ExprEqual, "C", "c").Find(&tabs)
		assert.Nil(t, err)
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

func TestCollectionUpdate(t *testing.T) {
	table := TestSearchTable{A: "aaaaa", B: "bbbbb"}
	brick := TestCollectionDB.Model(&TestSearchTable{}).Debug()
	result, err := brick.Where(ExprEqual, Offsetof(TestSearchTable{}.A), "a").Update(&table)
	assert.Nil(t, err)
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

func TestCollectionPreloadCreateTable(t *testing.T) {
	brick := TestCollectionDB.Model(TestPreloadTable{}).Debug().
		Preload(Offsetof(TestPreloadTable{}.BelongTo)).Enter().
		Preload(Offsetof(TestPreloadTable{}.OneToOne)).Enter().
		Preload(Offsetof(TestPreloadTable{}.OneToMany)).Enter().
		Preload(Offsetof(TestPreloadTable{}.ManyToMany)).Enter()
	brick.CreateTable()
	hasTable, err := brick.HasTable()
	assert.Nil(t, err)
	t.Logf("table %s exist:%v\n", brick.model.Name, hasTable)
	result, err := brick.DropTableIfExist()
	assert.Nil(t, err)
	if err := result.Err(); err != nil {
		t.Error(err)
		t.Fail()
	}
	result, err = brick.CreateTable()
	assert.Nil(t, err)
	if err := result.Err(); err != nil {
		t.Error(err)
		t.Fail()
	}
}

func TestCollectionPreloadInsertData(t *testing.T) {
	brick := TestCollectionDB.Model(TestPreloadTable{}).Debug().
		Preload(Offsetof(TestPreloadTable{}.BelongTo)).Debug().Enter().
		Preload(Offsetof(TestPreloadTable{}.OneToOne)).Debug().Enter().
		Preload(Offsetof(TestPreloadTable{}.OneToMany)).Debug().Enter().
		Preload(Offsetof(TestPreloadTable{}.ManyToMany)).Debug().Enter()
	// add id generator
	TestCollectionDB.SetModelHandlers("Save", brick.model, CollectionHandlersChain{CollectionIDGenerate})
	TestCollectionDB.SetModelHandlers("Insert", brick.model, CollectionHandlersChain{CollectionIDGenerate})

	for _, pBrick := range brick.MapPreloadBrick {
		TestCollectionDB.SetModelHandlers("Save", pBrick.model, CollectionHandlersChain{CollectionIDGenerate})
		TestCollectionDB.SetModelHandlers("Insert", pBrick.model, CollectionHandlersChain{CollectionIDGenerate})
	}
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
				{Name: "test insert data one to many 1"},
				{Name: "test insert data one to many 2"},
			},
			ManyToMany: []TestPreloadTableManyToMany{
				{Name: "test insert data many to many"},
				{Name: "test insert data many to many."},
			},
		}
		result, err := brick.Insert(&tab)
		assert.Nil(t, err)
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
				{"Name": "test insert data one to many 1"},
				{"Name": "test insert data one to many 2"},
			},
			"ManyToMany": []map[string]interface{}{
				{"Name": "test insert data many to many"},
				{"Name": "test insert data many to many."},
			},
		}
		result, err := brick.Insert(tab)
		assert.Nil(t, err)
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
		result, err := brick.Insert(tab)
		assert.Nil(t, err)
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

func TestCollectionPreloadSave(t *testing.T) {
	brick := TestCollectionDB.Model(&TestPreloadTable{}).Debug()
	brick = brick.Preload(Offsetof(TestPreloadTable{}.BelongTo)).Enter()
	brick = brick.Preload(Offsetof(TestPreloadTable{}.OneToOne)).Enter()
	brick = brick.Preload(Offsetof(TestPreloadTable{}.OneToMany)).Enter()
	manyToManyPreload := brick.Preload(Offsetof(TestPreloadTable{}.ManyToMany))
	// add id generator
	TestCollectionDB.SetModelHandlers("Save", brick.model, CollectionHandlersChain{CollectionIDGenerate})
	TestCollectionDB.SetModelHandlers("Insert", brick.model, CollectionHandlersChain{CollectionIDGenerate})
	for _, pBrick := range brick.MapPreloadBrick {
		TestCollectionDB.SetModelHandlers("Save", pBrick.model, CollectionHandlersChain{CollectionIDGenerate})
		TestCollectionDB.SetModelHandlers("Insert", pBrick.model, CollectionHandlersChain{CollectionIDGenerate})
	}
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
		result, err := manyToManyPreload.Insert(&manyToMany)
		assert.Nil(t, err)
		if err := result.Err(); err != nil {
			t.Error(err)
			t.Failed()
		}
		assert.NotZero(t, manyToMany[0].ID)
		assert.NotZero(t, manyToMany[1].ID)
		// now many to many object have id information
		tables[0].ManyToMany = manyToMany
		tables[1].ManyToMany = manyToMany
		t.Logf("1id %v 2id %v\n", tables[0].ID, tables[1].ID)
		result, err = brick.Save(&tables)
		assert.Nil(t, err)
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
		result, err = brick.Save(&tables)
		assert.Nil(t, err)
		if err := result.Err(); err != nil {
			t.Error(err)
			t.Failed()
		}
	}

}

func TestCollectionPreloadFind(t *testing.T) {
	brick := TestCollectionDB.Model(TestPreloadTable{}).Debug().
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

func TestCollectionPreloadDelete(t *testing.T) {
	var hardTab TestHardDeleteTable
	var softTab TestSoftDeleteTable
	var err error
	var result *Result

	t.Run("Hard Delete", func(t *testing.T) {
		brick := TestCollectionDB.Model(&hardTab).Debug().
			Preload(Offsetof(hardTab.BelongTo)).Enter().
			Preload(Offsetof(hardTab.OneToOne)).Enter().
			Preload(Offsetof(hardTab.OneToMany)).Enter().
			Preload(Offsetof(hardTab.ManyToMany)).Enter().
			Preload(Offsetof(hardTab.SoftBelongTo)).Enter().
			Preload(Offsetof(hardTab.SoftOneToOne)).Enter().
			Preload(Offsetof(hardTab.SoftOneToMany)).Enter().
			Preload(Offsetof(hardTab.SoftManyToMany)).Enter()

		TestCollectionDB.SetModelHandlers("Save", brick.model, CollectionHandlersChain{CollectionIDGenerate})
		TestCollectionDB.SetModelHandlers("Insert", brick.model, CollectionHandlersChain{CollectionIDGenerate})

		for _, pBrick := range brick.MapPreloadBrick {
			TestCollectionDB.SetModelHandlers("Save", pBrick.model, CollectionHandlersChain{CollectionIDGenerate})
			TestCollectionDB.SetModelHandlers("Insert", pBrick.model, CollectionHandlersChain{CollectionIDGenerate})
		}
		result, err = brick.DropTableIfExist()
		assert.Nil(t, err)
		if err := result.Err(); err != nil {
			t.Error(err)
		}
		result, err = brick.CreateTableIfNotExist()
		assert.Nil(t, err)
		if err := result.Err(); err != nil {
			t.Error(err)
		}

		result, err = brick.Save([]TestHardDeleteTable{
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
		assert.Nil(t, err)
		if err := result.Err(); err != nil {
			t.Error(err)
		}

		var hardDeleteData []TestHardDeleteTable
		result, err = brick.Find(&hardDeleteData)
		assert.Nil(t, err)
		if err := result.Err(); err != nil {
			t.Error(err)
		}

		_, err = brick.Delete(&hardDeleteData)
		assert.Nil(t, err)
		if err := result.Err(); err != nil {
			t.Error(err)
		}
	})
	t.Run("SoftDelete", func(t *testing.T) {
		brick := TestCollectionDB.Model(&softTab).Debug().
			Preload(Offsetof(softTab.BelongTo)).Enter().
			Preload(Offsetof(softTab.OneToOne)).Enter().
			Preload(Offsetof(softTab.OneToMany)).Enter().
			Preload(Offsetof(softTab.ManyToMany)).Enter().
			Preload(Offsetof(softTab.SoftBelongTo)).Enter().
			Preload(Offsetof(softTab.SoftOneToOne)).Enter().
			Preload(Offsetof(softTab.SoftOneToMany)).Enter().
			Preload(Offsetof(softTab.SoftManyToMany)).Enter()

		TestCollectionDB.SetModelHandlers("Save", brick.model, CollectionHandlersChain{CollectionIDGenerate})
		TestCollectionDB.SetModelHandlers("Insert", brick.model, CollectionHandlersChain{CollectionIDGenerate})
		for _, pBrick := range brick.MapPreloadBrick {
			TestCollectionDB.SetModelHandlers("Save", pBrick.model, CollectionHandlersChain{CollectionIDGenerate})
			TestCollectionDB.SetModelHandlers("Insert", pBrick.model, CollectionHandlersChain{CollectionIDGenerate})
		}
		result, err = brick.DropTableIfExist()
		assert.Nil(t, err)
		if err := result.Err(); err != nil {
			t.Error(err)
		}
		result, err = brick.CreateTableIfNotExist()
		assert.Nil(t, err)
		if err := result.Err(); err != nil {
			t.Error(err)
		}

		result, err = brick.Save([]TestSoftDeleteTable{
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
		assert.Nil(t, err)
		if err := result.Err(); err != nil {
			t.Error(err)
		}

		var softDeleteData []TestSoftDeleteTable
		result, err = brick.Find(&softDeleteData)
		assert.Nil(t, err)
		if err := result.Err(); err != nil {
			t.Error(err)
		}

		_, err = brick.Delete(&softDeleteData)
		assert.Nil(t, err)
		if err := result.Err(); err != nil {
			t.Error(err)
		}
	})
}

func TestCollectionCustomPreload(t *testing.T) {
	table := TestCustomPreloadTable{}
	tableOne := TestCustomPreloadOneToOne{}
	tableThree := TestCustomPreloadOneToMany{}
	middleTable := TestCustomPreloadManyToManyMiddle{}
	brick := TestCollectionDB.Model(&table).Debug().
		CustomOneToOnePreload(Offsetof(table.ChildOne), Offsetof(tableOne.ParentID)).Enter().
		CustomBelongToPreload(Offsetof(table.ChildTwo), Offsetof(table.BelongToID)).Enter().
		CustomOneToManyPreload(Offsetof(table.Children), Offsetof(tableThree.ParentID)).Enter().
		CustomManyToManyPreload(middleTable, Offsetof(table.OtherChildren), Offsetof(middleTable.ParentID), Offsetof(middleTable.ChildID)).Enter()
	TestCollectionDB.SetModelHandlers("Insert", brick.model, CollectionHandlersChain{CollectionIDGenerate})
	for _, pBrick := range brick.MapPreloadBrick {
		TestCollectionDB.SetModelHandlers("Insert", pBrick.model, CollectionHandlersChain{CollectionIDGenerate})
	}

	result, err := brick.DropTableIfExist()
	assert.Nil(t, err)
	if err := result.Err(); err != nil {
		t.Error(err)
	}

	result, err = brick.CreateTable()
	assert.Nil(t, err)
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
	result, err = brick.Insert(&data)
	assert.Nil(t, err)
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

	result, err = brick.Delete(&data)
	assert.Nil(t, err)
	if err := result.Err(); err != nil {
		t.Error(err)
	}
}

func TestCollectionFlow(t *testing.T) {
	// create a brick with product
	brick := TestCollectionDB.Model(&Product{}).Debug().
		Preload(Offsetof(Product{}.Detail)).Enter().
		Preload(Offsetof(Product{}.Address)).Enter().
		Preload(Offsetof(Product{}.Tag)).Enter().
		Preload(Offsetof(Product{}.Friend)).
		Preload(Offsetof(Product{}.Detail)).Enter().
		Preload(Offsetof(Product{}.Address)).Enter().
		Preload(Offsetof(Product{}.Tag)).Enter().
		Enter()

	TestCollectionDB.SetModelHandlers("Save", brick.model, CollectionHandlersChain{CollectionIDGenerate})
	TestCollectionDB.SetModelHandlers("Insert", brick.model, CollectionHandlersChain{CollectionIDGenerate})
	for _, pBrick := range brick.MapPreloadBrick {
		TestCollectionDB.SetModelHandlers("Save", pBrick.model, CollectionHandlersChain{CollectionIDGenerate})
		TestCollectionDB.SetModelHandlers("Insert", pBrick.model, CollectionHandlersChain{CollectionIDGenerate})
	}

	result, err := brick.DropTableIfExist()
	assert.Nil(t, err)
	if err := result.Err(); err != nil {
		t.Error(err)
	}

	result, err = brick.CreateTableIfNotExist()
	assert.Nil(t, err)
	assert.Nil(t, err)
	if err := result.Err(); err != nil {
		t.Error(err)
	}

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
	result, err = brick.Save(product)
	assert.Nil(t, err)
	if err := result.Err(); err != nil {
		t.Error(err)
	}

	// add a new tag
	tagBrick := TestCollectionDB.Model(&Tag{}).Debug()
	result, err = tagBrick.Insert(&Tag{Code: "nice"})
	assert.Nil(t, err)
	if err := result.Err(); err != nil {
		t.Error(err)
	}

	//bind new tag to the one's product
	middleBrick := TestCollectionDB.MiddleModel(&Product{}, &Tag{}).Debug()
	result, err = middleBrick.Save(&struct {
		ProductID uint32
		TagCode   string
	}{product[0].ID, "nice"})
	assert.Nil(t, err)
	if err := result.Err(); err != nil {
		t.Error(err)
	}

	// I try use transaction but sqlite will lock database when have second query
	//if _, err = brick.Insert(&product); err != nil {
	//	t.Logf("error %s", err)
	//	err = brick.Rollback()
	//	assert.Nil(t, err)
	//	t.Log("try to save")
	//	brick = brick.Begin()
	//	if err = brick.Save(&product); err != nil {
	//		t.Logf("error %s", err)
	//		err = brick.Rollback()
	//		assert.Nil(t, err)
	//	} else {
	//		err = brick.Commit()
	//		assert.Nil(t, err)
	//	}
	//} else {
	//	err = brick.Commit()
	//	assert.Nil(t, err)
	//}
	// try to find
	var newProducts []Product
	result, err = brick.Find(&newProducts)
	assert.Nil(t, err)
	if err := result.Err(); err != nil {
		t.Error(err)
	}

	jsonBytes, err := json.MarshalIndent(newProducts, "", "  ")
	assert.Nil(t, err)
	t.Logf("\n%v", string(jsonBytes))
	brick.Delete(&newProducts)
}

func TestCollectionIgnorePreloadInsert(t *testing.T) {
	var tab TestPreloadIgnoreTable
	brick := TestCollectionDB.Model(&tab).Debug().
		Preload(Offsetof(tab.BelongTo)).Enter().
		Preload(Offsetof(tab.OneToOne)).Enter()

	TestCollectionDB.SetModelHandlers("Save", brick.model, CollectionHandlersChain{CollectionIDGenerate})
	TestCollectionDB.SetModelHandlers("Insert", brick.model, CollectionHandlersChain{CollectionIDGenerate})
	for _, pBrick := range brick.MapPreloadBrick {
		TestCollectionDB.SetModelHandlers("Save", pBrick.model, CollectionHandlersChain{CollectionIDGenerate})
		TestCollectionDB.SetModelHandlers("Insert", pBrick.model, CollectionHandlersChain{CollectionIDGenerate})
	}

	result, err := brick.DropTableIfExist()
	assert.Nil(t, err)
	if err := result.Err(); err != nil {
		t.Error(err)
	}

	result, err = brick.CreateTable()
	assert.Nil(t, err)
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
	result, err = brick.Insert(&data)
	assert.Nil(t, err)
	if err := result.Err(); err != nil {
		t.Error(err)
	}
	assert.Zero(t, data[0].BelongToID)
	assert.Equal(t, data[1].OneToOne.TestPreloadIgnoreTableID, data[1].ID)
}

func TestCollectionMissPreloadFind(t *testing.T) {
	var tab TestMissTable
	var belongTab TestMissBelongTo
	var manyToManyTab TestMissManyToMany
	brick := TestCollectionDB.Model(&tab).Debug().
		Preload(Offsetof(tab.BelongTo)).Enter().
		Preload(Offsetof(tab.OneToOne)).Enter().
		Preload(Offsetof(tab.OneToMany)).Enter().
		Preload(Offsetof(tab.ManyToMany)).Enter()

	TestCollectionDB.SetModelHandlers("Save", brick.model, CollectionHandlersChain{CollectionIDGenerate})
	TestCollectionDB.SetModelHandlers("Insert", brick.model, CollectionHandlersChain{CollectionIDGenerate})
	for _, pBrick := range brick.MapPreloadBrick {
		TestCollectionDB.SetModelHandlers("Save", pBrick.model, CollectionHandlersChain{CollectionIDGenerate})
		TestCollectionDB.SetModelHandlers("Insert", pBrick.model, CollectionHandlersChain{CollectionIDGenerate})
	}

	result, err := brick.DropTableIfExist()
	assert.Nil(t, err)
	if err := result.Err(); err != nil {
		t.Error(err)
	}
	result, err = brick.CreateTable()
	assert.Nil(t, err)
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
	result, err = brick.Insert(&missData)
	assert.Nil(t, err)
	if err := result.Err(); err != nil {
		t.Error(err)
	}
	// remove belong to data and many to many data
	result, err = TestCollectionDB.Model(&belongTab).Debug().
		Delete([]TestMissBelongTo{*missData[0].BelongTo, *missData[1].BelongTo})
	assert.Nil(t, err)
	if err := result.Err(); err != nil {
		t.Error(err)
	}
	result, err = TestCollectionDB.Model(&manyToManyTab).Debug().
		Delete([]TestMissManyToMany{missData[0].ManyToMany[0], missData[1].ManyToMany[0]})
	assert.Nil(t, err)
	if err := result.Err(); err != nil {
		t.Error(err)
	}
	// find again
	var scanMissData []TestMissTable

	result, err = brick.Find(&scanMissData)
	assert.Nil(t, err)
	if err := result.Err(); err != nil {
		t.Error(err)
	}

	t.Logf("%#v\n", scanMissData)
	// TODO miss data need warning or error
}

func TestCollectionSameBelongId(t *testing.T) {
	var tab TestSameBelongIdTable
	brick := TestCollectionDB.Model(&tab).Debug().
		Preload(Offsetof(tab.BelongTo)).Enter()

	TestCollectionDB.SetModelHandlers("Save", brick.model, CollectionHandlersChain{CollectionIDGenerate})
	TestCollectionDB.SetModelHandlers("Insert", brick.model, CollectionHandlersChain{CollectionIDGenerate})
	for _, pBrick := range brick.MapPreloadBrick {
		TestCollectionDB.SetModelHandlers("Save", pBrick.model, CollectionHandlersChain{CollectionIDGenerate})
		TestCollectionDB.SetModelHandlers("Insert", pBrick.model, CollectionHandlersChain{CollectionIDGenerate})
	}

	result, err := brick.DropTableIfExist()
	assert.Nil(t, err)
	if err := result.Err(); err != nil {
		t.Error(err)
	}
	result, err = brick.CreateTable()
	assert.Nil(t, err)
	if err := result.Err(); err != nil {
		t.Error(err)
	}
	data := []TestSameBelongIdTable{
		{Data: "test same belong id 1", BelongTo: TestSameBelongIdBelongTo{ID: 1, Data: "belong data"}},
		{Data: "test same belong id 2", BelongTo: TestSameBelongIdBelongTo{ID: 1, Data: "belong data"}},
	}
	result, err = brick.Save(&data)
	assert.Nil(t, err)
	if err := result.Err(); err != nil {
		t.Error(err)
	}

	var findData []TestSameBelongIdTable
	result, err = brick.Find(&findData)
	t.Logf("%#v", findData)
}
