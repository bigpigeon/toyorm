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
