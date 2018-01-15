package toyorm

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"path/filepath"
	"testing"
	"time"
)

var (
	TestDB *Toy
)

type TestCreateTable1 struct {
	ID       int    `toyorm:"primary key;AUTOINCREMENT"`
	Name     string `toyorm:"not null"`
	Category string `toyorm:"index"`
}

type TestCreateTable2 struct {
	ID   string `toyorm:"primary key"`
	Name string `toyorm:"not null;index:idx_name_code"`
	Code string `toyorm:"not null;index:idx_name_code"`
}

type TestCreateTable3 struct {
	ID       int      `toyorm:"primary key;auto_increment"`
	Name     string   `toyorm:"not null"`
	Category string   `toyorm:"index"`
	Value    *float64 `toyorm:""`
}

type TestCreateTable4 struct {
	ID          int    `toyorm:"primary key;auto_increment"`
	Name        string `toyorm:"not null"`
	Value       int
	PtrPtrValue **int
}

type TestCreateTable5Sub1 struct {
	ID   int `toyorm:"primary key;auto_increment"`
	Name string
}

type TestCreateTable5Sub2 struct {
	ID                 int `toyorm:"primary key;auto_increment"`
	Name               string
	TestCreateTable5ID uint `toyorm:"index"`
}

type TestCreateTable5Sub3 struct {
	ID                 int `toyorm:"primary key;auto_increment"`
	Name               string
	TestCreateTable5ID uint `toyorm:"index"`
}

type TestCreateTable5Sub4 struct {
	ID   int `toyorm:"primary key;auto_increment"`
	Name string
}

type TestCreateTable5 struct {
	ModelDefault
	Name   string `toyorm:"not null"`
	Sub1ID int
	Sub1   *TestCreateTable5Sub1
	Sub2   *TestCreateTable5Sub2
	Sub3   []TestCreateTable5Sub3
	Sub4   []TestCreateTable5Sub4
}

type TestInsertTable struct {
	ModelDefault
	DataStr     string
	DataInt     int
	DataFloat   float64
	DataComplex complex64
	PtrStr      *string
	PtrInt      *int
	PtrFloat    *float64
	PtrComplex  *complex128
}

type TestSearchTable struct {
	ModelDefault
	A string
	B string
	C string
	D *string
}

type TestHardDeleteTable struct {
	ID         uint `toyorm:"primary key;auto_increment"`
	Data       string
	BelongToID uint `toyorm:"unique index"`
	BelongTo   *TestHardDeleteTableBelongTo
	OneToOne   *TestHardDeleteTableOneToOne
	OneToMany  []TestHardDeleteTableOneToMany
	ManyToMany []TestHardDeleteTableManyToMany

	SoftBelongToID uint `toyorm:"unique index"`
	SoftBelongTo   *TestSoftDeleteTableBelongTo
	SoftOneToOne   *TestSoftDeleteTableOneToOne
	SoftOneToMany  []TestSoftDeleteTableOneToMany
	SoftManyToMany []TestSoftDeleteTableManyToMany
}

type TestSoftDeleteTable struct {
	ModelDefault
	Data       string
	BelongToID uint `toyorm:"unique index"`
	BelongTo   *TestHardDeleteTableBelongTo
	OneToOne   *TestHardDeleteTableOneToOne
	OneToMany  []TestHardDeleteTableOneToMany
	ManyToMany []TestHardDeleteTableManyToMany

	SoftBelongToID uint `toyorm:"unique index"`
	SoftBelongTo   *TestSoftDeleteTableBelongTo
	SoftOneToOne   *TestSoftDeleteTableOneToOne
	SoftOneToMany  []TestSoftDeleteTableOneToMany
	SoftManyToMany []TestSoftDeleteTableManyToMany
}

type TestCustomPreloadTable struct {
	ID            uint `toyorm:"primary key;auto_increment"`
	Data          string
	BelongToID    uint `toyorm:"index"`
	ChildOne      *TestCustomPreloadOneToOne
	ChildTwo      *TestCustomPreloadBelongTo
	Children      []TestCustomPreloadOneToMany
	OtherChildren []TestCustomPreloadManyToMany
}

type TestCustomPreloadOneToOne struct {
	ID       uint `toyorm:"primary key;auto_increment"`
	Data     string
	ParentID uint `toyorm:"index"`
}

type TestCustomPreloadBelongTo struct {
	ID   uint `toyorm:"primary key;auto_increment"`
	Data string
}

type TestCustomPreloadOneToMany struct {
	ID       uint `toyorm:"primary key;auto_increment"`
	Data     string
	ParentID uint `toyorm:"index"`
}

type TestCustomPreloadManyToManyMiddle struct {
	ParentID uint `toyorm:"primary key"`
	ChildID  int  `toyorm:"primary key"`
}

type TestCustomPreloadManyToMany struct {
	ID   uint `toyorm:"primary key;auto_increment"`
	Data string
}

type TestHardDeleteTableBelongTo struct {
	ID   uint `toyorm:"primary key;auto_increment"`
	Data string
}

type TestHardDeleteTableOneToOne struct {
	ID                    uint `toyorm:"primary key;auto_increment"`
	Data                  string
	TestHardDeleteTableID uint `toyorm:"index"`
	TestSoftDeleteTableID uint `toyorm:"index"`
}

type TestHardDeleteTableOneToMany struct {
	ID                    uint `toyorm:"primary key;auto_increment"`
	Data                  string
	TestHardDeleteTableID uint `toyorm:"index"`
	TestSoftDeleteTableID uint `toyorm:"index"`
}

type TestHardDeleteTableManyToMany struct {
	ID   uint `toyorm:"primary key"`
	Data string
}

type TestSoftDeleteTableBelongTo struct {
	ModelDefault
	Data string
}

type TestSoftDeleteTableOneToOne struct {
	ModelDefault
	Data                  string
	TestHardDeleteTableID uint `toyorm:"index"`
	TestSoftDeleteTableID uint `toyorm:"index"`
}

type TestSoftDeleteTableOneToMany struct {
	ModelDefault
	Data                  string
	TestHardDeleteTableID uint `toyorm:"index"`
	TestSoftDeleteTableID uint `toyorm:"index"`
}

type TestSoftDeleteTableManyToMany struct {
	ModelDefault
	Data string
}

type Product struct {
	ModelDefault
	Detail    *ProductDetail
	Price     float64
	Discount  float32
	Amount    int
	Address   []Address
	Contact   string
	Favorites int
	Version   string
	Tag       []Tag
	Friend    []Product
}

type ProductDetail struct {
	ID         uint `toyorm:"primary key;auto_increment"`
	ProductID  uint `toyorm:"index"`
	Page       string
	Parameters map[string]interface{}
}

type Tag struct {
	Code        string `toyorm:"primary key"`
	Description string `toyorm:"not null;default ''"`
}

type Address struct {
	ID        int    `toyorm:"primary key;auto_increment"`
	Address1  string `toyorm:"index"`
	Address2  string
	ProductID uint `toyorm:"index"`
}

type User struct {
	ModelDefault
	Birthday time.Time `toyorm:"index;NULL"`
	Name     string    `toyorm:"unique index"`
	Age      int
	Height   int
	Sex      bool
}

type SqlTypeTable struct {
	ID    int `toyorm:"primary key;auto_increment"`
	Name  sql.NullString
	Age   sql.NullInt64
	Sex   sql.NullBool
	Money sql.NullFloat64
}

func TestMain(m *testing.M) {
	for _, sqldata := range []struct {
		Driver string
		Source string
	}{
		{"mysql", "root:@tcp(localhost:3306)/toyorm?charset=utf8&parseTime=True"},
		{"sqlite3", filepath.Join(os.TempDir() + "toyorm_test.db")},
	} {
		var err error
		TestDB, err = Open(sqldata.Driver, sqldata.Source)
		fmt.Printf("=========== %s ===========\n", sqldata.Driver)
		fmt.Printf("connect to %s \n\n", sqldata.Source)
		if err == nil {
			err = TestDB.db.Ping()
			if err == nil {
				m.Run()
				TestDB.db.Close()
				continue
			}
		}
		fmt.Printf("Error: cannot test sql %s because (%s)\n", sqldata.Driver, err)
	}
}
