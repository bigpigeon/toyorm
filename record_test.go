/*
 * Copyright 2018. bigpigeon. All rights reserved.
 * Use of this source code is governed by a MIT style
 * license that can be found in the LICENSE file.
 */

package toyorm

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
	. "unsafe"
)

func TestStructRecord(t *testing.T) {
	model := NewModel(reflect.ValueOf(TestCreateTable3{}).Type())
	data := []struct {
		Name     string
		Category string
		Value    float64
	}{{
		"bigpigeon",
		"user",
		20,
	}}
	records := NewRecords(model, reflect.ValueOf(&data).Elem())

	assert.Equal(t, records.GetRecord(0).Field(model.GetPosField(0).Name()), reflect.Value{})
	records.GetRecord(0).SetField(model.GetPosField(0).Name(), reflect.ValueOf(1))
	assert.Equal(t, records.GetRecord(0).Field(model.GetPosField(0).Name()).Interface(), int32(1))
	records.GetRecord(0).SetField(model.GetPosField(1).Name(), reflect.ValueOf("verybigpigeon"))
	assert.Equal(t, data[0].Name, "verybigpigeon")
	t.Log(data)
	// test add
	elem := struct {
		Name     string
		Category string
		Value    float64
	}{
		"whatever",
		"stuff",
		30,
	}
	records.Add(reflect.ValueOf(elem))
	t.Log(data)
	assert.Equal(t, data[1], elem)
}

func TestNameMapRecord(t *testing.T) {
	model := NewModel(reflect.ValueOf(TestCreateTable3{}).Type())
	data := []map[string]interface{}{{
		"Name":     "bigpigeon",
		"Category": "user",
		"Value":    20,
	}}
	records := NewRecords(model, reflect.ValueOf(&data).Elem())

	assert.Equal(t, records.GetRecord(0).Field(model.GetPosField(0).Name()), reflect.Value{})
	records.GetRecord(0).SetField(model.GetPosField(0).Name(), reflect.ValueOf(1))
	t.Log(records.GetRecord(0).Source())
	assert.Equal(t, records.GetRecord(0).Field(model.GetPosField(0).Name()).Interface(), int32(1))
	records.GetRecord(0).SetField(model.GetPosField(1).Name(), reflect.ValueOf("verybigpigeon"))
	t.Log(data)
	assert.Equal(t, data[0]["Name"], "verybigpigeon")
	// test add
	elem := map[string]interface{}{
		"Name":     "whatever",
		"Category": "stuff",
		"Value":    30,
	}
	records.Add(reflect.ValueOf(elem))
	t.Log(data)
	assert.Equal(t, data[1], elem)
}

func TestOffsetMapRecord(t *testing.T) {
	model := NewModel(reflect.ValueOf(TestCreateTable3{}).Type())
	data := []map[uintptr]interface{}{{
		Offsetof(TestCreateTable3{}.Name):     "bigpigeon",
		Offsetof(TestCreateTable3{}.Category): "user",
		Offsetof(TestCreateTable3{}.Value):    20,
	}}
	records := NewRecords(model, reflect.ValueOf(&data).Elem())

	assert.Equal(t, records.GetRecord(0).Field(model.GetPosField(0).Name()), reflect.Value{})
	records.GetRecord(0).SetField(model.GetPosField(0).Name(), reflect.ValueOf(1))
	assert.Equal(t, records.GetRecord(0).Field(model.GetPosField(0).Name()).Interface(), int32(1))
	records.GetRecord(0).SetField(model.GetPosField(1).Name(), reflect.ValueOf("verybigpigeon"))
	assert.Equal(t, data[0][Offsetof(TestCreateTable3{}.Name)], "verybigpigeon")
	t.Log(data)
	// test add
	elem := map[uintptr]interface{}{
		Offsetof(TestCreateTable3{}.Name):     "whatever",
		Offsetof(TestCreateTable3{}.Category): "stuff",
		Offsetof(TestCreateTable3{}.Value):    30,
	}
	records.Add(reflect.ValueOf(elem))
	t.Log(data)
	assert.Equal(t, data[1], elem)
}

func TestStructRecordGroupBy(t *testing.T) {
	model := NewModel(reflect.ValueOf(TestCreateTable3{}).Type())
	data := []struct {
		Name     string
		Category string
		Value    float64
	}{{
		"bigpigeon",
		"user",
		20,
	}, {
		"fatpigeon",
		"user",
		21,
	}, {
		"pigeon",
		"user",
		21,
	}}
	records := NewRecords(model, reflect.ValueOf(&data).Elem())
	group1 := records.GroupBy("Name")
	for key, recordList := range group1 {
		var source []interface{}
		for _, record := range recordList {
			source = append(source, record.Source().Interface())
			assert.Equal(t, key, record.Field("Name").Interface())
		}
		t.Logf("group with name %s %v", key, source)

	}

	group2 := records.GroupBy("Category")
	for key, recordList := range group2 {
		var source []interface{}
		for _, record := range recordList {
			source = append(source, record.Source().Interface())
			assert.Equal(t, key, record.Field("Category").Interface())
		}
		t.Logf("group with catetory %s %v", key, source)
	}

	group3 := records.GroupBy("Value")
	for key, recordList := range group3 {
		var source []interface{}
		for _, record := range recordList {
			source = append(source, record.Source().Interface())
			assert.Equal(t, key, record.Field("Value").Interface())
		}
		t.Logf("group with value %v %v", key, source)
	}
}

func TestNameMapRecordGroupBy(t *testing.T) {
	model := NewModel(reflect.ValueOf(TestCreateTable3{}).Type())
	data := []map[string]interface{}{{
		"Name":     "bigpigeon",
		"Category": "user",
		"Value":    20,
	}, {
		"Name":     "fatpigeon",
		"Category": "user",
		"Value":    21,
	}, {
		"Name":     "pigeon",
		"Category": "user",
		"Value":    21,
	}}
	records := NewRecords(model, reflect.ValueOf(&data).Elem())
	group1 := records.GroupBy("Name")
	for key, recordList := range group1 {
		var source []interface{}
		for _, record := range recordList {
			source = append(source, record.Source().Interface())
			assert.Equal(t, key, record.Field("Name").Interface())
		}
		t.Logf("group with name %s %v", key, source)

	}

	group2 := records.GroupBy("Category")
	for key, recordList := range group2 {
		var source []interface{}
		for _, record := range recordList {
			source = append(source, record.Source().Interface())
			assert.Equal(t, key, record.Field("Category").Interface())
		}
		t.Logf("group with catetory %s %v", key, source)
	}

	group3 := records.GroupBy("Value")
	for key, recordList := range group3 {
		var source []interface{}
		for _, record := range recordList {
			source = append(source, record.Source().Interface())
			assert.Equal(t, key, record.Field("Value").Interface())
		}
		t.Logf("group with value %v %v", key, source)
	}
}

func TestOffsetMapRecordGroupBy(t *testing.T) {
	model := NewModel(reflect.ValueOf(TestCreateTable3{}).Type())
	data := []map[uintptr]interface{}{{
		Offsetof(TestCreateTable3{}.Name):     "bigpigeon",
		Offsetof(TestCreateTable3{}.Category): "user",
		Offsetof(TestCreateTable3{}.Value):    20,
	}, {
		Offsetof(TestCreateTable3{}.Name):     "fatpigeon",
		Offsetof(TestCreateTable3{}.Category): "user",
		Offsetof(TestCreateTable3{}.Value):    21,
	}, {
		Offsetof(TestCreateTable3{}.Name):     "pigeon",
		Offsetof(TestCreateTable3{}.Category): "user",
		Offsetof(TestCreateTable3{}.Value):    21,
	}}
	records := NewRecords(model, reflect.ValueOf(&data).Elem())
	group1 := records.GroupBy("Name")
	for key, recordList := range group1 {
		var source []interface{}
		for _, record := range recordList {
			source = append(source, record.Source().Interface())
			assert.Equal(t, key, record.Field("Name").Interface())
		}
		t.Logf("group with name %s %v", key, source)

	}

	group2 := records.GroupBy("Category")
	for key, recordList := range group2 {
		var source []interface{}
		for _, record := range recordList {
			source = append(source, record.Source().Interface())
			assert.Equal(t, key, record.Field("Category").Interface())
		}
		t.Logf("group with catetory %s %v", key, source)
	}

	group3 := records.GroupBy("Value")
	for key, recordList := range group3 {
		var source []interface{}
		for _, record := range recordList {
			source = append(source, record.Source().Interface())
			assert.Equal(t, key, record.Field("Value").Interface())
		}
		t.Logf("group with value %v %v", key, source)
	}
}

func TestRecordFieldType(t *testing.T) {
	type Extra struct {
		Name string
		Val  string
	}
	type User struct {
		ModelDefault
		Tags     []int `toyorm:"type:VARCHAR(1024)"`
		JsonData Extra `toyorm:"type:VARCHAR(1024)"`
	}
	model := NewModel(reflect.ValueOf(User{}).Type())
	for _, data := range []interface{}{
		[]User{
			{
				ModelDefault: ModelDefault{
					ID: 1,
				},
				JsonData: Extra{"key", "222"},
			},
		},
		[]map[string]interface{}{
			{
				"ID":       1,
				"JsonData": Extra{"key", "222"},
			},
		},
		[]map[uintptr]interface{}{
			{
				Offsetof(User{}.ID):       1,
				Offsetof(User{}.JsonData): Extra{"key", "222"},
			},
		},
	} {
		records := NewRecords(model, reflect.ValueOf(&data).Elem().Elem())
		assert.Equal(t, records.GetFieldType("JsonData"), reflect.TypeOf(Extra{}))
		assert.Equal(t, records.GetFieldType("Tags"), reflect.TypeOf([]int{}))
		assert.Equal(t, records.GetRecord(0).GetFieldType("JsonData"), reflect.TypeOf(Extra{}))
		assert.Equal(t, records.GetRecord(0).GetFieldType("Tags"), reflect.TypeOf([]int{}))
	}

}
