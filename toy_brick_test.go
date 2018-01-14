package toyorm

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
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

func TestPreloadCreateTable(t *testing.T) {
	brick := TestDB.Model(TestCreateTable5{}).Debug().
		Preload(Offsetof(TestCreateTable5{}.Sub1)).Enter().
		Preload(Offsetof(TestCreateTable5{}.Sub2)).Enter().
		Preload(Offsetof(TestCreateTable5{}.Sub3)).Enter().
		Preload(Offsetof(TestCreateTable5{}.Sub4)).Enter()
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
		t.Logf("id %v", tab.ID)
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
		t.Logf("id %v", tab["ID"])
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
		t.Logf("id %v", tabMap[Offsetof(tab.ID)])
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
		t.Logf("id %v", tab.ID)
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
		t.Logf("id %v", tab["ID"])
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
		t.Logf("id %v", tabMap[Offsetof(tab.ID)])
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
		_, err := TestDB.Model(&TestSearchTable{}).Debug().Find(&table)
		assert.Nil(t, err)
		t.Logf("%#v\n", table)
	}
	// test find with map
	//{
	//	table := map[string]interface{}{}
	//	err := TestDB.Model(&TestSearchTable{}).Debug().Find(&table)
	//	assert.Nil(t, err)
	//	t.Logf("%#v\n", table)
	//}
	// test find with struct list
	{
		tables := []TestSearchTable{}
		_, err := TestDB.Model(&TestSearchTable{}).Debug().Find(&tables)
		assert.Nil(t, err)
		t.Logf("%#v\n", tables)
	}
	// test find with map list
	//{
	//	tables := []map[string]interface{}{}
	//	err := TestDB.Model(&TestSearchTable{}).Debug().Find(&tables)
	//	assert.Nil(t, err)
	//	t.Logf("%#v\n", tables)
	//}
}

func TestConditionFind(t *testing.T) {
	base := TestSearchTable{}
	//SELECT id,a,b,c,d FROM test_search_table WHERE a = ? AND b = ?, args:[]interface {}{"a", "b"}
	{
		table := []TestSearchTable{}
		result, err := TestDB.Model(&TestSearchTable{}).Debug().
			Where(ExprAnd, TestSearchTable{A: "a", B: "b"}).Find(&table)
		assert.Nil(t, err)
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
		assert.Nil(t, err)
		if err := result.Err(); err != nil {
			t.Error(err)
			t.Fail()
		}
		t.Logf("%#v\n", table)
	}
}

func TestUpdate(t *testing.T) {
	table := TestSearchTable{A: "aaaaa", B: "bbbbb"}
	brick := TestDB.Model(&TestSearchTable{}).Debug()
	result, err := brick.Where(ExprEqual, Offsetof(TestSearchTable{}.A), "a").Update(&table)
	assert.Nil(t, err)
	assert.Nil(t, err)
	if err := result.Err(); err != nil {
		t.Error(err)
	}
	var tableList []TestSearchTable
	brick.Find(&tableList)
	t.Logf("%#v\n", tableList)
}

func TestSave(t *testing.T) {
	table := []TestCreateTable5{
		{
			Name: "test save 1",
			Sub1: &TestCreateTable5Sub1{
				Name: "test save sub1",
			},
			Sub2: &TestCreateTable5Sub2{
				Name: "test save sub2",
			},
			Sub3: []TestCreateTable5Sub3{
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
			Sub1: &TestCreateTable5Sub1{
				Name: "test save 2 sub1",
			},
			Sub2: &TestCreateTable5Sub2{
				Name: "test save 2 sub2",
			},
			Sub3: []TestCreateTable5Sub3{
				{
					Name: "test save 2 sub3 sub1",
				},
				{
					Name: "test save 2 sub3 sub2",
				},
			},
		},
	}
	brick := TestDB.Model(&TestCreateTable5{}).Debug()
	brick = brick.Preload(Offsetof(TestCreateTable5{}.Sub1)).Enter()
	brick = brick.Preload(Offsetof(TestCreateTable5{}.Sub2)).Enter()
	brick = brick.Preload(Offsetof(TestCreateTable5{}.Sub3)).Enter()
	result, err := brick.Save(&table)
	assert.Nil(t, err)
	assert.Nil(t, err)
	if err := result.Err(); err != nil {
		t.Error(err)
	}
	t.Logf("%#v", table)
	table[0].Sub3[1].ID = 0
	table[0].Sub1.Name = "test save sub1 try save"
	result, err = brick.Save(&table)
	assert.Nil(t, err)
	assert.Nil(t, err)
	if err := result.Err(); err != nil {
		t.Error(err)
	}
	t.Logf("%#v", table)
}

func TestDelete(t *testing.T) {
	brick := TestDB.Model(&TestHardDeleteTable{}).Debug()
	brick = brick.Preload(Offsetof(TestHardDeleteTable{}.BelongTo)).Enter()
	brick = brick.Preload(Offsetof(TestHardDeleteTable{}.OneToOne)).Enter()
	brick = brick.Preload(Offsetof(TestHardDeleteTable{}.OneToMany)).Enter()
	brick = brick.Preload(Offsetof(TestHardDeleteTable{}.ManyToMany)).Enter()
	brick = brick.Preload(Offsetof(TestHardDeleteTable{}.SoftBelongTo)).Enter()
	brick = brick.Preload(Offsetof(TestHardDeleteTable{}.SoftOneToOne)).Enter()
	brick = brick.Preload(Offsetof(TestHardDeleteTable{}.SoftOneToMany)).Enter()
	brick = brick.Preload(Offsetof(TestHardDeleteTable{}.SoftManyToMany)).Enter()
	var err error
	var result *Result
	result, err = brick.DropTableIfExist()
	assert.Nil(t, err)
	assert.Nil(t, err)
	if err := result.Err(); err != nil {
		t.Error(err)
	}

	result, err = brick.CreateTable()
	assert.Nil(t, err)
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
			ManyToMany: []TestHardDeleteTableManyToMany{
				{ID: 1, Data: "many to many data"},
				{ID: 2, Data: "many to many data"},
			},
			SoftBelongTo: &TestSoftDeleteTableBelongTo{Data: "belong to data"},
			SoftOneToOne: &TestSoftDeleteTableOneToOne{Data: "one to one data"},
			SoftOneToMany: []TestSoftDeleteTableOneToMany{
				{Data: "one to many data"},
				{Data: "one to many data"},
			},
			SoftManyToMany: []TestSoftDeleteTableManyToMany{
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
			ManyToMany: []TestHardDeleteTableManyToMany{
				{ID: 1, Data: "many to many data"},
				{ID: 2, Data: "many to many data"},
			},
			SoftBelongTo: &TestSoftDeleteTableBelongTo{Data: "belong to data"},
			SoftOneToOne: &TestSoftDeleteTableOneToOne{Data: "one to one data"},
			SoftOneToMany: []TestSoftDeleteTableOneToMany{
				{Data: "one to many data"},
				{Data: "one to many data"},
			},
			SoftManyToMany: []TestSoftDeleteTableManyToMany{
				{ModelDefault: ModelDefault{ID: 1}, Data: "many to many data"},
				{ModelDefault: ModelDefault{ID: 2}, Data: "many to many data"},
			},
		},
	})
	assert.Nil(t, err)
	assert.Nil(t, err)
	if err := result.Err(); err != nil {
		t.Error(err)
	}

	var hardDeleteData []TestHardDeleteTable
	result, err = brick.Find(&hardDeleteData)
	assert.Nil(t, err)
	assert.Nil(t, err)
	if err := result.Err(); err != nil {
		t.Error(err)
	}

	_, err = brick.Delete(&hardDeleteData)
	assert.Nil(t, err)
	assert.Nil(t, err)
	if err := result.Err(); err != nil {
		t.Error(err)
	}
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
		CustomManyToManyPreload(Offsetof(table.OtherChildren), middleTable, Offsetof(middleTable.ParentID), Offsetof(middleTable.ChildID)).Enter()
	result, err := brick.DropTableIfExist()
	assert.Nil(t, err)
	assert.Nil(t, err)
	if err := result.Err(); err != nil {
		t.Error(err)
	}

	result, err = brick.CreateTable()
	assert.Nil(t, err)
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
		ProductID uint
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
