package toyorm

import (
	"testing"
	. "unsafe"
)

func TestPanicRepeatField(t *testing.T) {
	type SomeEmbedFields struct {
		ID   uint
		Data string
	}
	type RepeatStruct struct {
		SomeEmbedFields
		Data string
	}
	defer func() {
		err := recover()
		t.Log(err)
		if _, ok := err.(ErrRepeatField); ok == false {
			t.Log("non panic when model struct is error")
			t.Fail()
		}
	}()
	TestDB.Model(&RepeatStruct{})
}

func TestPanicSameColumnName(t *testing.T) {
	type SomeEmbedFields struct {
		ID     uint
		DataSS string
	}
	type SameColumnStruct struct {
		SomeEmbedFields
		Data_ss string
	}
	defer func() {
		err := recover()
		t.Log(err)
		if _, ok := err.(ErrSameColumnName); ok == false {
			t.Log("non panic when model struct is error")
			t.Fail()
		}
	}()
	TestDB.Model(&SameColumnStruct{})
}

func TestPanicModelStruct(t *testing.T) {
	ErrorModelStruct := 2
	defer func() {
		err := recover()
		t.Log(err)
		if _, ok := err.(ErrInvalidModelType); ok == false {
			t.Log("non panic when model struct is error")
			t.Fail()
		}
	}()
	TestDB.Model(&ErrorModelStruct)
}

func TestPanicModelName(t *testing.T) {
	NonNameStruct := struct {
		ID    int `toyorm:"primary key"`
		Name  string
		Value string
	}{}
	defer func() {
		err := recover()
		t.Log(err)
		if _, ok := err.(ErrInvalidModelName); ok == false {
			t.Log("non panic when model name is nil")
			t.Fail()
		}
	}()
	TestDB.Model(&NonNameStruct)
}

func TestPanicPreloadField(t *testing.T) {
	type TestPanicPreloadTableSub struct {
		ID uint `toyorm:"primary key"`
	}
	type TestPanicPreloadTable struct {
		ModelDefault
		Name string
		Sub  *TestPanicPreloadTableSub
	}
	defer func() {
		err := recover()
		t.Log(err)
		if _, ok := err.(ErrInvalidPreloadField); ok == false {
			t.Log("non panic when preload is error")
			t.Fail()
		}
	}()
	TestDB.Model(&TestPanicPreloadTable{}).Preload(Offsetof(TestPanicPreloadTable{}.Sub))
}

func TestPanicConditionKey(t *testing.T) {
	// that's right
	TestDB.Model(&TestSearchTable{}).Where(ExprAnd, TestSearchTable{A: "a", B: "b"})
	defer func() {
		err := recover()
		t.Log(err)
		if _, ok := err.(ErrInvalidRecordType); ok == false {
			t.Log("non panic when preload is error")
			t.Fail()
		}
	}()
	// error example
	TestDB.Model(&TestSearchTable{}).Where(ExprAnd, Offsetof(TestSearchTable{}.A))
}
