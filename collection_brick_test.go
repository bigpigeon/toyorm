package toyorm

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
	. "unsafe"
)

func dbSelector(key interface{}, n int) int {
	switch val := key.(type) {
	case int:
		return val % n
	case int32:
		return int(val) % n
	case uint:
		return int(val) % n
	case uint32:
		return int(val) % n
	default:
		panic("primary key type not match")
	}
}

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
	brick := TestCollectionDB.Model(&TestInsertTable{}).Debug().Selector(dbSelector)
	// add id generator
	TestCollectionDB.SetModelHandlers("Insert", brick.model, CollectionHandlersChain{CollectionIDGenerate()})
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
	brick := TestCollectionDB.Model(&TestInsertTable{}).Selector(dbSelector).Debug()
	// add id generator
	TestCollectionDB.SetModelHandlers("Insert", brick.model, CollectionHandlersChain{CollectionIDGenerate()})

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
		TestCollectionDB.SetModelHandlers("Insert", brick.model, CollectionHandlersChain{CollectionIDGenerate()})

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
			for j := 1; j < 5; j++ {
				for k := 1; k < 5; k++ {
					for l := 1; l < 5; l++ {
						d := strings.Repeat("d", l)
						t1 := TestSearchTable{
							A: strings.Repeat("a", i),
							B: strings.Repeat("b", j),
							C: strings.Repeat("c", k),
							D: &d,
						}
						result, err := TestCollectionDB.Model(&TestSearchTable{}).Debug().Selector(dbSelector).Insert(&t1)
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
