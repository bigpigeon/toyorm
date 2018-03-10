package toyorm

import (
	"github.com/stretchr/testify/assert"
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
