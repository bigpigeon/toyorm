package toyorm

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
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
		brick := TestDB.Model(tab).Begin().Debug()
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
		err = brick.Commit()
		assert.Nil(t, err)
	}
}

func TestInsertData(t *testing.T) {
	brick := TestDB.Model(&TestInsertTable{}).Debug()
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
		result, err := brick.Insert(tab)
		assert.Nil(t, err)
		if err := result.Err(); err != nil {
			t.Error(err)
			t.Fail()
		}
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
		result, err := brick.Insert(tab)
		assert.Nil(t, err)
		if err := result.Err(); err != nil {
			t.Error(err)
			t.Fail()
		}
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
		result, err := brick.Insert(tabMap)
		assert.Nil(t, err)
		if err := result.Err(); err != nil {
			t.Error(err)
			t.Fail()
		}
		t.Logf("id %v %v", tabMap[0][Offsetof(tab.ID)], tabMap[1][Offsetof(tab.ID)])

	}
}

func TestInsertPointData(t *testing.T) {
	brick := TestDB.Model(&TestInsertTable{}).Debug()
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

func TestFind(t *testing.T) {
	{
		brick := TestDB.Model(&TestSearchTable{}).Debug()
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
		for i := 1; i < 5; i++ {
			d := strings.Repeat("d", i)
			t1 := TestSearchTable{
				A: strings.Repeat("a", i),
				B: strings.Repeat("b", i),
				C: strings.Repeat("c", i),
				D: &d,
			}
			result, err := TestDB.Model(&TestSearchTable{}).Debug().Insert(&t1)
			assert.Nil(t, err)
			if err := result.Err(); err != nil {
				t.Error(err)
				t.Fail()
			}
			t.Logf("%#v\n", t1)
		}
	}
	// test find with struct
	{
		table := TestSearchTable{}
		result, err := TestDB.Model(&TestSearchTable{}).Debug().Find(&table)
		assert.Nil(t, err)
		if err := result.Err(); err != nil {
			t.Error(err)
			t.Fail()
		}
		t.Logf("%#v\n", table)
	}
	// test find with struct list
	{
		var tables []TestSearchTable
		result, err := TestDB.Model(&TestSearchTable{}).Debug().Find(&tables)
		assert.Nil(t, err)
		if err := result.Err(); err != nil {
			t.Error(err)
			t.Fail()
		}
		t.Logf("%#v\n", tables)
	}
}

func TestConditionFind(t *testing.T) {
	base := TestSearchTable{}
	//SELECT id,a,b,c,d FROM test_search_table WHERE a = ? AND b = ?, args:[]interface {}{"a", "b"}
	{
		table := []TestSearchTable{}
		result, err := TestDB.Model(&TestSearchTable{}).Debug().
			Where(ExprAnd, TestSearchTable{A: "a", B: "b"}).Find(&table)
		assert.Nil(t, err)
		if err := result.Err(); err != nil {
			t.Error(err)
			t.Fail()
		}
		t.Logf("%#v\n", table)
	}
	//SELECT id,a,b,c,d FROM test_search_table WHERE (a = ? OR b = ?), args:[]interface {}{"a", "bb"}
	{
		table := []TestSearchTable{}
		result, err := TestDB.Model(&TestSearchTable{}).Debug().
			Where(ExprOr, TestSearchTable{A: "a", B: "bb"}).Find(&table)
		assert.Nil(t, err)
		if err := result.Err(); err != nil {
			t.Error(err)
			t.Fail()
		}
		t.Logf("%#v\n", table)
	}
	//SELECT id,a,b,c,d FROM test_search_table WHERE a = ? AND b = ?, args:[]interface {}{"a", "b"}
	{
		table := []TestSearchTable{}
		result, err := TestDB.Model(&TestSearchTable{}).Debug().
			Where(ExprEqual, Offsetof(base.A), "a").
			And().Condition(ExprEqual, Offsetof(base.B), "b").Find(&table)
		assert.Nil(t, err)
		if err := result.Err(); err != nil {
			t.Error(err)
			t.Fail()
		}
		t.Logf("%#v\n", table)
	}
	//SELECT id,a,b,c,d FROM test_search_table WHERE ((a = ? AND b = ? OR c = ?) OR d = ? AND a = ?), args:[]interface {}{"a", "b", "c", "d", "aa"}
	{
		table := []TestSearchTable{}
		result, err := TestDB.Model(&TestSearchTable{}).Debug().
			Where(ExprEqual, Offsetof(base.A), "a").And().
			Condition(ExprEqual, Offsetof(base.B), "b").Or().
			Condition(ExprEqual, Offsetof(base.C), "c").Or().
			Condition(ExprEqual, Offsetof(base.D), "d").And().
			Condition(ExprEqual, Offsetof(base.A), "aa").
			Find(&table)
		assert.Nil(t, err)
		if err := result.Err(); err != nil {
			t.Error(err)
			t.Fail()
		}
		t.Logf("%#v\n", table)
	}
	//SELECT id,a,b,c,d FROM test_search_table WHERE a = ? AND ((a = ? OR b = ?) OR c = ?), args:[]interface {}{"aa", "a", "b", "c"}
	{
		table := []TestSearchTable{}
		brick := TestDB.Model(&TestSearchTable{}).Debug()
		inlineSearch := brick.Where(ExprEqual, Offsetof(base.A), "a").Or().
			Condition(ExprEqual, Offsetof(base.B), "b").Or().
			Condition(ExprEqual, Offsetof(base.C), "c").Search
		result, err := brick.Where(ExprEqual, Offsetof(base.A), "aa").And().
			Conditions(inlineSearch).
			Find(&table)
		assert.Nil(t, err)
		if err := result.Err(); err != nil {
			t.Error(err)
			t.Fail()
		}
		t.Logf("%#v\n", table)
	}
}

func TestCombinationConditionFind(t *testing.T) {
	brick := TestDB.Model(&TestSearchTable{}).Debug()
	{
		var tabs []TestSearchTable
		result, err := brick.Where(ExprAnd, TestSearchTable{A: "a", B: "b"}).Find(&tabs)
		assert.Nil(t, err)
		if err := result.Err(); err != nil {
			t.Error(err)
			t.Fail()
		}
		t.Logf("%#v\n", tabs)
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
	}
}

func TestUpdate(t *testing.T) {
	table := TestSearchTable{A: "aaaaa", B: "bbbbb"}
	brick := TestDB.Model(&TestSearchTable{}).Debug()
	result, err := brick.Where(ExprEqual, Offsetof(TestSearchTable{}.A), "a").Update(&table)
	assert.Nil(t, err)
	if err := result.Err(); err != nil {
		t.Error(err)
	}
	var tableList []TestSearchTable
	brick.Find(&tableList)
	t.Logf("%#v\n", tableList)
}

func TestPreloadCreateTable(t *testing.T) {
	brick := TestDB.Model(TestPreloadTable{}).Debug().
		Preload(Offsetof(TestPreloadTable{}.BelongTo)).Enter().
		Preload(Offsetof(TestPreloadTable{}.OneToOne)).Enter().
		Preload(Offsetof(TestPreloadTable{}.OneToMany)).Enter().
		Preload(Offsetof(TestPreloadTable{}.ManyToMany)).Enter()
	brick.CreateTable()
	hastable, err := brick.HasTable()
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

func TestPreloadInsertData(t *testing.T) {
	brick := TestDB.Model(TestPreloadTable{}).Debug().
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
				{"Name": "test insert data one to many"},
				{"Name": "test insert data one to many."},
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

func TestPreloadSave(t *testing.T) {
	brick := TestDB.Model(&TestPreloadTable{}).Debug()
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
		table := []TestPreloadTable{
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
		table[0].ManyToMany = manyToMany
		table[1].ManyToMany = manyToMany

		result, err = brick.Save(&table)
		assert.Nil(t, err)
		if err := result.Err(); err != nil {
			t.Error(err)
			t.Failed()
		}
		// try to update soft delete
		now := time.Now()
		table[0].DeletedAt = &now
		result, err = brick.Save(&table)
		assert.Nil(t, err)
		if err := result.Err(); err != nil {
			t.Error(err)
			t.Failed()
		}
	}

}

func TestPreloadFind(t *testing.T) {
	brick := TestDB.Model(TestPreloadTable{}).Debug().
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
	var err error
	var result *Result

	// delete middle first
	{
		hardHardMiddleBrick := TestDB.MiddleModel(&hardTab, &TestHardDeleteManyToMany{}).Debug()
		result, err = hardHardMiddleBrick.DropTableIfExist()
		assert.Nil(t, err)
		if err := result.Err(); err != nil {
			t.Error(err)
		}

		hardSoftMiddleBrick := TestDB.MiddleModel(&hardTab, &TestSoftDeleteManyToMany{}).Debug()
		result, err = hardSoftMiddleBrick.DropTableIfExist()
		assert.Nil(t, err)
		if err := result.Err(); err != nil {
			t.Error(err)
		}

		softHardMiddleBrick := TestDB.MiddleModel(&softTab, &TestHardDeleteManyToMany{}).Debug()
		result, err = softHardMiddleBrick.DropTableIfExist()
		assert.Nil(t, err)
		if err := result.Err(); err != nil {
			t.Error(err)
		}

		softSoftMiddleBrick := TestDB.MiddleModel(&softTab, &TestSoftDeleteManyToMany{}).Debug()
		result, err = softSoftMiddleBrick.DropTableIfExist()
		assert.Nil(t, err)
		if err := result.Err(); err != nil {
			t.Error(err)
		}
	}
	// delete hard table
	{
		hardBrick := TestDB.Model(&hardTab).Debug().
			Preload(Offsetof(hardTab.BelongTo)).Enter().
			Preload(Offsetof(hardTab.OneToOne)).Enter().
			Preload(Offsetof(hardTab.OneToMany)).Enter().
			Scope(foreignKeyManyToManyPreload(Offsetof(hardTab.ManyToMany))).Enter().
			Preload(Offsetof(hardTab.SoftBelongTo)).Enter().
			Preload(Offsetof(hardTab.SoftOneToOne)).Enter().
			Preload(Offsetof(hardTab.SoftOneToMany)).Enter().
			Scope(foreignKeyManyToManyPreload(Offsetof(hardTab.SoftManyToMany))).Enter()

		result, err = hardBrick.DropTableIfExist()
		assert.Nil(t, err)
		if err := result.Err(); err != nil {
			t.Error(err)
		}
	}
	// delete soft table
	{
		brick := TestDB.Model(&softTab).Debug().
			Preload(Offsetof(softTab.BelongTo)).Enter().
			Preload(Offsetof(softTab.OneToOne)).Enter().
			Preload(Offsetof(softTab.OneToMany)).Enter().
			Scope(foreignKeyManyToManyPreload(Offsetof(softTab.ManyToMany))).Enter().
			Preload(Offsetof(softTab.SoftBelongTo)).Enter().
			Preload(Offsetof(softTab.SoftOneToOne)).Enter().
			Preload(Offsetof(softTab.SoftOneToMany)).Enter().
			Scope(foreignKeyManyToManyPreload(Offsetof(softTab.SoftManyToMany))).Enter()

		result, err = brick.DropTableIfExist()
		assert.Nil(t, err)
		if err := result.Err(); err != nil {
			t.Error(err)
		}
	}
	// have same target foreign key ,need create table following table first
	{
		result, err = TestDB.Model(&TestHardDeleteTableBelongTo{}).Debug().CreateTable()
		assert.Nil(t, err)
		if err := result.Err(); err != nil {
			t.Error(err)
		}

		result, err := TestDB.Model(&TestSoftDeleteTableBelongTo{}).Debug().CreateTable()
		assert.Nil(t, err)
		if err := result.Err(); err != nil {
			t.Error(err)
		}

		result, err = TestDB.Model(&hardTab).Debug().CreateTable()
		assert.Nil(t, err)
		if err := result.Err(); err != nil {
			t.Error(err)
		}

		result, err = TestDB.Model(&softTab).Debug().CreateTable()
		assert.Nil(t, err)
		if err := result.Err(); err != nil {
			t.Error(err)
		}
	}
	t.Run("Hard Delete", func(t *testing.T) {
		brick := TestDB.Model(&hardTab).Debug().
			Preload(Offsetof(hardTab.BelongTo)).Enter().
			Preload(Offsetof(hardTab.OneToOne)).Enter().
			Preload(Offsetof(hardTab.OneToMany)).Enter().
			Scope(foreignKeyManyToManyPreload(Offsetof(hardTab.ManyToMany))).Enter().
			Preload(Offsetof(hardTab.SoftBelongTo)).Enter().
			Preload(Offsetof(hardTab.SoftOneToOne)).Enter().
			Preload(Offsetof(hardTab.SoftOneToMany)).Enter().
			Scope(foreignKeyManyToManyPreload(Offsetof(hardTab.SoftManyToMany))).Enter()

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
		brick := TestDB.Model(&softTab).Debug().
			Preload(Offsetof(softTab.BelongTo)).Enter().
			Preload(Offsetof(softTab.OneToOne)).Enter().
			Preload(Offsetof(softTab.OneToMany)).Enter().
			Scope(foreignKeyManyToManyPreload(Offsetof(softTab.ManyToMany))).Enter().
			Preload(Offsetof(softTab.SoftBelongTo)).Enter().
			Preload(Offsetof(softTab.SoftOneToOne)).Enter().
			Preload(Offsetof(softTab.SoftOneToMany)).Enter().
			Scope(foreignKeyManyToManyPreload(Offsetof(softTab.SoftManyToMany))).Enter()
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

func TestCustomPreload(t *testing.T) {
	table := TestCustomPreloadTable{}
	tableOne := TestCustomPreloadOneToOne{}
	tableThree := TestCustomPreloadOneToMany{}
	middleTable := TestCustomPreloadManyToManyMiddle{}
	brick := TestDB.Model(&table).Debug().
		CustomOneToOnePreload(Offsetof(table.ChildOne), Offsetof(tableOne.ParentID)).Enter().
		CustomBelongToPreload(Offsetof(table.ChildTwo), Offsetof(table.BelongToID)).Enter().
		CustomOneToManyPreload(Offsetof(table.Children), Offsetof(tableThree.ParentID)).Enter().
		CustomManyToManyPreload(middleTable, Offsetof(table.OtherChildren), Offsetof(middleTable.ParentID), Offsetof(middleTable.ChildID)).Enter()
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

func TestFlow(t *testing.T) {
	// create a brick with product
	brick := TestDB.Model(&Product{}).Debug().
		Preload(Offsetof(Product{}.Detail)).Enter().
		Preload(Offsetof(Product{}.Address)).Enter().
		Preload(Offsetof(Product{}.Tag)).Enter().
		Preload(Offsetof(Product{}.Friend)).
		Preload(Offsetof(Product{}.Detail)).Enter().
		Preload(Offsetof(Product{}.Address)).Enter().
		Preload(Offsetof(Product{}.Tag)).Enter().
		Enter()
	// drow table if exist
	result, err := brick.DropTableIfExist()
	assert.Nil(t, err)
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
	assert.Nil(t, err)
	if err := result.Err(); err != nil {
		t.Error(err)
	}

	// add a new tag
	tagBrick := TestDB.Model(&Tag{}).Debug()
	result, err = tagBrick.Insert(&Tag{Code: "nice"})
	assert.Nil(t, err)
	assert.Nil(t, err)
	if err := result.Err(); err != nil {
		t.Error(err)
	}

	//bind new tag to the one's product
	middleBrick := TestDB.MiddleModel(&Product{}, &Tag{}).Debug()
	result, err = middleBrick.Save(&struct {
		ProductID uint32
		TagCode   string
	}{product[0].ID, "nice"})
	assert.Nil(t, err)
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
	assert.Nil(t, err)
	if err := result.Err(); err != nil {
		t.Error(err)
	}

	jsonBytes, err := json.MarshalIndent(newProducts, "", "  ")
	assert.Nil(t, err)
	t.Logf("\n%v", string(jsonBytes))
	brick.Delete(&newProducts)
}

func TestGroupBy(t *testing.T) {
	//create table and insert data
	{
		brick := TestDB.Model(&TestGroupByTable{}).Debug()
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

		result, err = brick.Insert([]TestGroupByTable{
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
		assert.Nil(t, err)
		if err := result.Err(); err != nil {
			t.Error(err)
		}
	}
	{
		var tab TestGroupByTableGroup
		brick := TestDB.Model(&tab).Debug()

		brick = brick.GroupBy(Offsetof(tab.Name), Offsetof(tab.Address))
		var data []TestGroupByTableGroup
		result, err := brick.Find(&data)
		assert.Nil(t, err)
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

	brick := TestDB.Model(&tab).Debug().
		Preload(Offsetof(tab.BelongTo)).Enter().
		Preload(Offsetof(tab.OneToOne)).Enter().
		Preload(Offsetof(tab.OneToMany)).Enter()
	brick = brick.CustomManyToManyPreload(
		&middleTab, Offsetof(tab.ManyToMany),
		Offsetof(middleTab.TestForeignKeyTableID),
		Offsetof(middleTab.TestForeignKeyTableManyToManyID),
	).Enter()

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
	result, err = brick.Insert(&data)
	assert.Nil(t, err)
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

	result, err = brick.Delete(&data)
	assert.Nil(t, err)
	if err := result.Err(); err != nil {
		t.Error(err)
	}

}

func TestMissPreload(t *testing.T) {
	var tab TestMissTable
	var belongTab TestMissBelongTo
	var manyToManyTab TestMissManyToMany
	brick := TestDB.Model(&tab).Debug().
		Preload(Offsetof(tab.BelongTo)).Enter().
		Preload(Offsetof(tab.OneToOne)).Enter().
		Preload(Offsetof(tab.OneToMany)).Enter().
		Preload(Offsetof(tab.ManyToMany)).Enter()
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
	result, err = TestDB.Model(&belongTab).Debug().
		Delete([]TestMissBelongTo{*missData[0].BelongTo, *missData[1].BelongTo})
	assert.Nil(t, err)
	if err := result.Err(); err != nil {
		t.Error(err)
	}
	result, err = TestDB.Model(&manyToManyTab).Debug().
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
