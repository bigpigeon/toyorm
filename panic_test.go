package toyorm

import (
	"testing"
)

func TestErrorModelStruct(t *testing.T) {
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

func TestErrorModelName(t *testing.T) {
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

//func TestErrorPreloadField(t *testing.T) {
//	NonNameStruct := struct {
//		ID    int `toyorm:"primary key"`
//		Name  string
//		Value string
//	}{}
//	defer func() {
//		err := recover()
//		t.Log(err)
//		if _, ok := err.(ErrInvalidModelName); ok == false {
//			t.Log("non panic when model name is nil")
//			t.Fail()
//		}
//	}()
//	TestDB.Model(&NonNameStruct)
//}
