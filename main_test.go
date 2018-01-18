package toyorm

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
	"testing"
	"time"
)

var (
	TestDB *Toy
)

type TestCreateTable1 struct {
	ID       int32  `toyorm:"primary key;AUTOINCREMENT"`
	Name     string `toyorm:"not null"`
	Category string `toyorm:"index"`
}

type TestCreateTable2 struct {
	ID   string `toyorm:"primary key"`
	Name string `toyorm:"not null;index:idx_name_code"`
	Code string `toyorm:"not null;index:idx_name_code"`
}

type TestCreateTable3 struct {
	ID       int32    `toyorm:"primary key;auto_increment"`
	Name     string   `toyorm:"not null"`
	Category string   `toyorm:"index"`
	Value    *float64 `toyorm:""`
}

type TestCreateTable4 struct {
	ID          int32  `toyorm:"primary key;auto_increment"`
	Name        string `toyorm:"not null"`
	Value       int
	PtrPtrValue **int
}

type TestPreloadTableBelongTo struct {
	ID   int32 `toyorm:"primary key;auto_increment"`
	Name string
}

type TestPreloadTableOneToOne struct {
	ID                 int32 `toyorm:"primary key;auto_increment"`
	Name               string
	TestPreloadTableID uint32 `toyorm:"index"`
}

type TestPreloadTableOneToMany struct {
	ID                 int32 `toyorm:"primary key;auto_increment"`
	Name               string
	TestPreloadTableID uint32 `toyorm:"index"`
}

type TestPreloadTableManyToMany struct {
	ID   int32 `toyorm:"primary key;auto_increment"`
	Name string
}

type TestPreloadTable struct {
	ModelDefault
	Name       string `toyorm:"not null"`
	BelongToID int32
	BelongTo   *TestPreloadTableBelongTo
	OneToOne   *TestPreloadTableOneToOne
	OneToMany  []TestPreloadTableOneToMany
	ManyToMany []TestPreloadTableManyToMany
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
	ID         uint32 `toyorm:"primary key;auto_increment"`
	Data       string
	BelongToID uint32 `toyorm:"unique index"`
	BelongTo   *TestHardDeleteTableBelongTo
	OneToOne   *TestHardDeleteTableOneToOne
	OneToMany  []TestHardDeleteTableOneToMany
	ManyToMany []TestHardDeleteTableManyToMany

	SoftBelongToID uint32 `toyorm:"unique index"`
	SoftBelongTo   *TestSoftDeleteTableBelongTo
	SoftOneToOne   *TestSoftDeleteTableOneToOne
	SoftOneToMany  []TestSoftDeleteTableOneToMany
	SoftManyToMany []TestSoftDeleteTableManyToMany
}

type TestSoftDeleteTable struct {
	ModelDefault
	Data       string
	BelongToID uint32 `toyorm:"unique index"`
	BelongTo   *TestHardDeleteTableBelongTo
	OneToOne   *TestHardDeleteTableOneToOne
	OneToMany  []TestHardDeleteTableOneToMany
	ManyToMany []TestHardDeleteTableManyToMany

	SoftBelongToID uint32 `toyorm:"unique index"`
	SoftBelongTo   *TestSoftDeleteTableBelongTo
	SoftOneToOne   *TestSoftDeleteTableOneToOne
	SoftOneToMany  []TestSoftDeleteTableOneToMany
	SoftManyToMany []TestSoftDeleteTableManyToMany
}

type TestCustomPreloadTable struct {
	ID            uint32 `toyorm:"primary key;auto_increment"`
	Data          string
	BelongToID    uint32 `toyorm:"index"`
	ChildOne      *TestCustomPreloadOneToOne
	ChildTwo      *TestCustomPreloadBelongTo
	Children      []TestCustomPreloadOneToMany
	OtherChildren []TestCustomPreloadManyToMany
}

type TestCustomPreloadOneToOne struct {
	ID       uint32 `toyorm:"primary key;auto_increment"`
	Data     string
	ParentID uint32 `toyorm:"index"`
}

type TestCustomPreloadBelongTo struct {
	ID   uint32 `toyorm:"primary key;auto_increment"`
	Data string
}

type TestCustomPreloadOneToMany struct {
	ID       uint32 `toyorm:"primary key;auto_increment"`
	Data     string
	ParentID uint32 `toyorm:"index"`
}

type TestCustomPreloadManyToManyMiddle struct {
	ParentID uint32 `toyorm:"primary key"`
	ChildID  int32  `toyorm:"primary key"`
}

type TestCustomPreloadManyToMany struct {
	ID   uint32 `toyorm:"primary key;auto_increment"`
	Data string
}

type TestHardDeleteTableBelongTo struct {
	ID   uint32 `toyorm:"primary key;auto_increment"`
	Data string
}

type TestHardDeleteTableOneToOne struct {
	ID                    uint32 `toyorm:"primary key;auto_increment"`
	Data                  string
	TestHardDeleteTableID uint32 `toyorm:"index"`
	TestSoftDeleteTableID uint32 `toyorm:"index"`
}

type TestHardDeleteTableOneToMany struct {
	ID                    uint32 `toyorm:"primary key;auto_increment"`
	Data                  string
	TestHardDeleteTableID uint32 `toyorm:"index"`
	TestSoftDeleteTableID uint32 `toyorm:"index"`
}

type TestHardDeleteTableManyToMany struct {
	ID   uint32 `toyorm:"primary key"`
	Data string
}

type TestSoftDeleteTableBelongTo struct {
	ModelDefault
	Data string
}

type TestSoftDeleteTableOneToOne struct {
	ModelDefault
	Data                  string
	TestHardDeleteTableID uint32 `toyorm:"index"`
	TestSoftDeleteTableID uint32 `toyorm:"index"`
}

type TestSoftDeleteTableOneToMany struct {
	ModelDefault
	Data                  string
	TestHardDeleteTableID uint32 `toyorm:"index"`
	TestSoftDeleteTableID uint32 `toyorm:"index"`
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
	ID         uint32 `toyorm:"primary key;auto_increment"`
	ProductID  uint32 `toyorm:"index"`
	Page       string
	Parameters map[string]interface{}
}

type Tag struct {
	Code        string `toyorm:"primary key"`
	Description string `toyorm:"not null;default ''"`
}

type Address struct {
	ID        int32  `toyorm:"primary key;auto_increment"`
	Address1  string `toyorm:"index"`
	Address2  string
	ProductID uint32 `toyorm:"index"`
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
	ID    int32 `toyorm:"primary key;auto_increment"`
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
		{"sqlite3", ":memory:"},
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
