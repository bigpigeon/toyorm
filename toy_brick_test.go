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
	for _, tab := range models {
		// start a session
		brick := TestDB.Model(tab).Begin().Debug()
		hastable, err = brick.HasTable()
		assert.Nil(t, err)
		t.Logf("table %s exist:%v\n", brick.model.Name, hastable)
		_, err = brick.DropTableIfExist()
		assert.Nil(t, err)
		_, err = brick.CreateTable()
		assert.Nil(t, err)
		err = brick.Commit()
		assert.Nil(t, err)
	}
}

func TestInsertData(t *testing.T) {
	for i := 1; i < 5; i++ {
		d := strings.Repeat("d", i)
		t1 := TestSearchTable{
			A: strings.Repeat("a", i),
			B: strings.Repeat("b", i),
			C: strings.Repeat("c", i),
			D: &d,
		}
		_, err := TestDB.Model(&TestSearchTable{}).Debug().Insert(&t1)
		assert.Nil(t, err)
		t.Logf("%#v\n", t1)
	}
	t2 := TestSearchTable{
		A: "a",
		B: "b",
		C: "c",
	}
	_, err := TestDB.Model(&TestSearchTable{}).Debug().Insert(&t2)

	assert.Nil(t, err)

	// insert with map[string]interface{}
	t3 := map[string]interface{}{
		"Name":        "one",
		"Value":       0,
		"PtrPtrValue": 20,
	}
	_, err = TestDB.Model(&TestTable4{}).Debug().Insert(&t3)
	assert.Nil(t, err)
	t.Log(t3)

	t4 := map[uintptr]interface{}{
		Offsetof(TestTable4{}.Name):        "two",
		Offsetof(TestTable4{}.Value):       0,
		Offsetof(TestTable4{}.PtrPtrValue): 20,
	}
	_, err = TestDB.Model(&TestTable4{}).Debug().Insert(&t4)
	assert.Nil(t, err)
	t.Logf("%#v\n", t4)

	t5 := TestTable5{
		Name: "one",
		Sub1: &TestTable5Sub1{
			Name: "sub one",
		},
		Sub2: &TestTable5Sub2{
			Name: "sub two",
		},
		Sub3: []TestTable5Sub3{
			TestTable5Sub3{Name: "sub three a"},
			TestTable5Sub3{Name: "sub three two"},
		},
	}
	_, err = TestDB.Model(&TestTable5{}).Debug().
		Preload(Offsetof(t5.Sub1)).Enter().
		Preload(Offsetof(t5.Sub2)).Enter().
		Preload(Offsetof(t5.Sub3)).Enter().Insert(&t5)
	assert.Nil(t, err)
	t.Logf("%#v\n", t5.Sub1)
	t.Logf("%#v\n", t5.Sub2)
	t.Logf("%#v\n", t5)

	t6 := map[string]interface{}{
		"Name": "two",
		"Sub1": map[string]interface{}{
			"Name": "sub 1",
		},
		"Sub2": map[string]interface{}{
			"Name": "sub 2",
		},
		"Sub3": []map[string]interface{}{
			{"Name": "sub 3 a"},
			{"Name": "sub 3 b"},
		},
	}
	_, err = TestDB.Model(&TestTable5{}).Debug().
		Preload(Offsetof(TestTable5{}.Sub1)).Enter().
		Preload(Offsetof(TestTable5{}.Sub2)).Enter().
		Preload(Offsetof(TestTable5{}.Sub3)).Enter().Insert(&t6)
	assert.Nil(t, err)
	t.Logf("%#v\n", t6)

	{
		t2.B = "bbb"
		_, err := TestDB.Model(&TestSearchTable{}).Debug().Save(&t2)
		assert.Nil(t, err)
	}
}

func TestFind(t *testing.T) {
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
	// test find with struct preload
	{
		table := TestTable5{}
		_, err := TestDB.Model(&TestTable5{}).Debug().
			Preload(Offsetof(table.Sub1)).Enter().
			Preload(Offsetof(table.Sub2)).Enter().
			Preload(Offsetof(table.Sub3)).Enter().
			Find(&table)
		assert.Nil(t, err)
		t.Logf("%#v\n", table)
	}
	{
		table := []TestTable5{}
		_, err := TestDB.Model(&TestTable5{}).Debug().
			Preload(Offsetof(TestTable5{}.Sub1)).Enter().
			Preload(Offsetof(TestTable5{}.Sub2)).Enter().
			Preload(Offsetof(TestTable5{}.Sub3)).Enter().
			Find(&table)
		assert.Nil(t, err)

		t.Logf("%#v\n", table)
	}
	//test find map
	//{
	//	table := map[string]interface{}{}
	//	err := TestDB.Model(&TestTable5{}).Debug().
	//		Preload(Offsetof(TestTable5{}.Sub1)).Enter().
	//		Preload(Offsetof(TestTable5{}.Sub2)).Enter().
	//		Preload(Offsetof(TestTable5{}.Sub3)).Enter().
	//		Find(&table)
	//	assert.Nil(t, err)
	//	t.Logf("%#v\n", table)
	//}
}

func TestConditionFind(t *testing.T) {
	base := TestSearchTable{}
	//SELECT id,a,b,c,d FROM test_search_table WHERE a = ? AND b = ?, args:[]interface {}{"a", "b"}
	{
		table := []TestSearchTable{}
		_, err := TestDB.Model(&TestSearchTable{}).Debug().
			Where(ExprAnd, TestSearchTable{A: "a", B: "b"}).Find(&table)
		assert.Nil(t, err)
		t.Logf("%#v\n", table)
	}
	//SELECT id,a,b,c,d FROM test_search_table WHERE (a = ? OR b = ?), args:[]interface {}{"a", "bb"}
	{
		table := []TestSearchTable{}
		_, err := TestDB.Model(&TestSearchTable{}).Debug().
			Where(ExprOr, TestSearchTable{A: "a", B: "bb"}).Find(&table)
		assert.Nil(t, err)
		t.Logf("%#v\n", table)
	}
	//SELECT id,a,b,c,d FROM test_search_table WHERE a = ? AND b = ?, args:[]interface {}{"a", "b"}
	{
		table := []TestSearchTable{}
		_, err := TestDB.Model(&TestSearchTable{}).Debug().
			Where(ExprEqual, Offsetof(base.A), "a").
			And().Condition(ExprEqual, Offsetof(base.B), "b").Find(&table)
		assert.Nil(t, err)
		t.Logf("%#v\n", table)
	}
	//SELECT id,a,b,c,d FROM test_search_table WHERE ((a = ? AND b = ? OR c = ?) OR d = ? AND a = ?), args:[]interface {}{"a", "b", "c", "d", "aa"}
	{
		table := []TestSearchTable{}
		_, err := TestDB.Model(&TestSearchTable{}).Debug().
			Where(ExprEqual, Offsetof(base.A), "a").And().
			Condition(ExprEqual, Offsetof(base.B), "b").Or().
			Condition(ExprEqual, Offsetof(base.C), "c").Or().
			Condition(ExprEqual, Offsetof(base.D), "d").And().
			Condition(ExprEqual, Offsetof(base.A), "aa").
			Find(&table)
		assert.Nil(t, err)
		t.Logf("%#v\n", table)
	}
	//SELECT id,a,b,c,d FROM test_search_table WHERE a = ? AND ((a = ? OR b = ?) OR c = ?), args:[]interface {}{"aa", "a", "b", "c"}
	{
		table := []TestSearchTable{}
		brick := TestDB.Model(&TestSearchTable{}).Debug()
		inlineSearch := brick.Where(ExprEqual, Offsetof(base.A), "a").Or().
			Condition(ExprEqual, Offsetof(base.B), "b").Or().
			Condition(ExprEqual, Offsetof(base.C), "c").Search
		_, err := brick.Where(ExprEqual, Offsetof(base.A), "aa").And().
			Conditions(inlineSearch).
			Find(&table)
		assert.Nil(t, err)
		t.Logf("%#v\n", table)
	}
}

func TestUpdate(t *testing.T) {
	table := TestSearchTable{A: "aaaaa", B: "bbbbb"}
	brick := TestDB.Model(&TestSearchTable{}).Debug()
	_, err := brick.Where(ExprEqual, Offsetof(TestSearchTable{}.A), "a").Update(&table)
	assert.Nil(t, err)
	var tableList []TestSearchTable
	brick.Find(&tableList)
	t.Logf("%#v\n", tableList)
}

func TestSave(t *testing.T) {
	table := []TestTable5{
		{
			Name: "test save 1",
			Sub1: &TestTable5Sub1{
				Name: "test save sub1",
			},
			Sub2: &TestTable5Sub2{
				Name: "test save sub2",
			},
			Sub3: []TestTable5Sub3{
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
			Sub1: &TestTable5Sub1{
				Name: "test save 2 sub1",
			},
			Sub2: &TestTable5Sub2{
				Name: "test save 2 sub2",
			},
			Sub3: []TestTable5Sub3{
				{
					Name: "test save 2 sub3 sub1",
				},
				{
					Name: "test save 2 sub3 sub2",
				},
			},
		},
	}
	brick := TestDB.Model(&TestTable5{}).Debug()
	brick = brick.Preload(Offsetof(TestTable5{}.Sub1)).Enter()
	brick = brick.Preload(Offsetof(TestTable5{}.Sub2)).Enter()
	brick = brick.Preload(Offsetof(TestTable5{}.Sub3)).Enter()
	_, err := brick.Save(&table)
	assert.Nil(t, err)
	t.Logf("%#v", table)
	table[0].Sub3[1].ID = 0
	table[0].Sub1.Name = "test save sub1 try save"
	_, err = brick.Save(&table)
	assert.Nil(t, err)
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
	_, err = brick.DropTableIfExist()
	assert.Nil(t, err)
	_, err = brick.CreateTable()
	assert.Nil(t, err)

	_, err = brick.Save([]TestHardDeleteTable{
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
	var hardDeleteData []TestHardDeleteTable
	_, err = brick.Find(&hardDeleteData)
	assert.Nil(t, err)
	_, err = brick.Delete(&hardDeleteData)
	assert.Nil(t, err)
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
	_, err := brick.DropTableIfExist()
	assert.Nil(t, err)
	_, err = brick.CreateTableIfNotExist()
	assert.Nil(t, err)

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
	_, err = brick.Save(product)
	assert.Nil(t, err)

	// add a new tag
	tagBrick := TestDB.Model(&Tag{}).Debug()
	_, err = tagBrick.Insert(&Tag{Code: "nice"})
	assert.Nil(t, err)
	//bind new tag to the one's product
	middleBrick := TestDB.MiddleModel(&Product{}, &Tag{}).Debug()
	_, err = middleBrick.Save(&struct {
		ProductID uint
		TagCode   string
	}{product[0].ID, "nice"})
	assert.Nil(t, err)

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
	_, err = brick.Find(&newProducts)
	assert.Nil(t, err)
	jsonBytes, err := json.MarshalIndent(newProducts, "", "  ")
	assert.Nil(t, err)
	t.Logf("\n%v", string(jsonBytes))
	brick.Delete(&newProducts)
}
