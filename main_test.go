package toyorm

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"reflect"
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
	ManyToMany []TestHardDeleteManyToMany

	SoftBelongToID uint32 `toyorm:"unique index"`
	SoftBelongTo   *TestSoftDeleteTableBelongTo
	SoftOneToOne   *TestSoftDeleteTableOneToOne
	SoftOneToMany  []TestSoftDeleteTableOneToMany
	SoftManyToMany []TestSoftDeleteManyToMany
}

type TestSoftDeleteTable struct {
	ModelDefault
	Data       string
	BelongToID uint32 `toyorm:"unique index"`
	BelongTo   *TestHardDeleteTableBelongTo
	OneToOne   *TestHardDeleteTableOneToOne
	OneToMany  []TestHardDeleteTableOneToMany
	ManyToMany []TestHardDeleteManyToMany

	SoftBelongToID uint32 `toyorm:"unique index"`
	SoftBelongTo   *TestSoftDeleteTableBelongTo
	SoftOneToOne   *TestSoftDeleteTableOneToOne
	SoftOneToMany  []TestSoftDeleteTableOneToMany
	SoftManyToMany []TestSoftDeleteManyToMany
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
	OneData  string
	ParentID uint32 `toyorm:"index"`
}

type TestCustomPreloadBelongTo struct {
	ID      uint32 `toyorm:"primary key;auto_increment"`
	TwoData string
}

type TestCustomPreloadOneToMany struct {
	ID        uint32 `toyorm:"primary key;auto_increment"`
	ThreeData string
	ParentID  uint32 `toyorm:"index"`
}

type TestCustomPreloadManyToManyMiddle struct {
	ParentID uint32 `toyorm:"primary key"`
	ChildID  uint32 `toyorm:"primary key"`
}

type TestCustomPreloadManyToMany struct {
	ID       uint32 `toyorm:"primary key;auto_increment"`
	FourData string
}

type TestHardDeleteTableBelongTo struct {
	ID   uint32 `toyorm:"primary key;auto_increment"`
	Data string
}

type TestHardDeleteTableOneToOne struct {
	ID                    uint32 `toyorm:"primary key;auto_increment"`
	Data                  string
	TestHardDeleteTableID *uint32 `toyorm:"index;foreign key"`
	TestSoftDeleteTableID *uint32 `toyorm:"index;foreign key"`
}

type TestHardDeleteTableOneToMany struct {
	ID                    uint32 `toyorm:"primary key;auto_increment"`
	Data                  string
	TestHardDeleteTableID *uint32 `toyorm:"index;foreign key"`
	TestSoftDeleteTableID *uint32 `toyorm:"index;foreign key"`
}

type TestHardDeleteManyToMany struct {
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
	TestHardDeleteTableID *uint32 `toyorm:"index;foreign key"`
	TestSoftDeleteTableID *uint32 `toyorm:"index;foreign key"`
}

type TestSoftDeleteTableOneToMany struct {
	ModelDefault
	Data                  string
	TestHardDeleteTableID *uint32 `toyorm:"index;foreign key"`
	TestSoftDeleteTableID *uint32 `toyorm:"index;foreign key"`
}

type TestSoftDeleteManyToMany struct {
	ModelDefault
	Data string
}

type TestGroupByTable struct {
	ModelDefault
	Name    string `toyorm:"index"`
	Address string `toyorm:"index"`
	Age     int
}

type TestGroupByTableGroup struct {
	Name     string `toyorm:"index"`
	Address  string `toyorm:"index"`
	MaxAge   int    `toyorm:"column:MAX(age)"`
	CountNum int    `toyorm:"column:COUNT(*)"`
}

func (t *TestGroupByTableGroup) TableName() string {
	return ModelName(reflect.TypeOf(TestGroupByTable{}))
}

type TestForeignKeyTable struct {
	ModelDefault
	Data     string
	BelongTo *TestForeignKeyTableBelongTo
	// foreign key cannot set to 0
	BelongToID *uint32 `toyorm:"foreign key"`

	OneToOne   *TestForeignKeyTableOneToOne
	OneToMany  []TestForeignKeyTableOneToMany
	ManyToMany []TestForeignKeyTableManyToMany
}

type TestForeignKeyTableBelongTo struct {
	ID   uint32 `toyorm:"primary key;auto_increment"`
	Data string
}

type TestForeignKeyTableOneToOne struct {
	ID                    uint32 `toyorm:"primary key;auto_increment"`
	TestForeignKeyTableID uint32 `toyorm:"foreign key"`
	Data                  string
}

type TestForeignKeyTableOneToMany struct {
	ID                    uint32 `toyorm:"primary key;auto_increment"`
	TestForeignKeyTableID uint32 `toyorm:"foreign key"`
	Data                  string
}

type TestForeignKeyTableMiddle struct {
	TestForeignKeyTableID           uint32 `toyorm:"primary key;foreign key"`
	TestForeignKeyTableManyToManyID uint32 `toyorm:"primary key;foreign key"`
}

type TestForeignKeyTableManyToMany struct {
	ID   uint32 `toyorm:"primary key;auto_increment"`
	Data string
}

type TestPreloadIgnoreTable struct {
	ID         uint32 `toyorm:"primary key;auto_increment"`
	Data       string
	BelongToID uint32
	BelongTo   TestPreloadIgnoreBelongTo
	OneToOne   TestPreloadIgnoreOneToOne
}

type TestPreloadIgnoreBelongTo struct {
	ID   uint32 `toyorm:"primary key;auto_increment"`
	Data string
}

type TestPreloadIgnoreOneToOne struct {
	ID                       uint32 `toyorm:"primary key;auto_increment"`
	Data                     string
	TestPreloadIgnoreTableID uint32
}

type TestMissTable struct {
	ID         uint32 `toyorm:"primary key;auto_increment"`
	Data       string
	BelongToID uint32
	BelongTo   *TestMissBelongTo
	OneToOne   *TestMissOneToOne
	OneToMany  []TestMissOneToMany
	ManyToMany []TestMissManyToMany
}

type TestMissBelongTo struct {
	ID           uint32 `toyorm:"primary key;auto_increment"`
	BelongToData string
}

type TestMissOneToOne struct {
	ID              uint32 `toyorm:"primary key;auto_increment"`
	OneToOneData    string
	TestMissTableID uint32
}

type TestMissOneToMany struct {
	ID              uint32 `toyorm:"primary key;auto_increment"`
	OneToManyData   string
	TestMissTableID uint32
}

type TestMissManyToMany struct {
	ID             uint32 `toyorm:"primary key;auto_increment"`
	ManyToManyData string
}

type TestSameBelongIdTable struct {
	ID         uint32 `toyorm:"primary key;auto_increment"`
	Data       string
	BelongToID uint32
	BelongTo   TestSameBelongIdBelongTo
}

type TestSameBelongIdBelongTo struct {
	ID   uint32 `toyorm:"primary key;auto_increment"`
	Data string
}

type TestPointContainerTable struct {
	ID         uint32 `toyorm:"primary key;auto_increment"`
	Data       string
	OneToMany  *[]*TestPointContainerOneToMany
	ManyToMany *[]*TestPointContainerManyToMany
}

type TestPointContainerOneToMany struct {
	ID                        uint32 `toyorm:"primary key;auto_increment"`
	OneToManyData             string
	TestPointContainerTableID uint32
}

type TestPointContainerManyToMany struct {
	ID             uint32 `toyorm:"primary key;auto_increment"`
	ManyToManyData string
}

type TestReportTable struct {
	ModelDefault
	Data       string
	BelongToID uint32
	BelongTo   *TestReportSub1
	OneToOne   *TestReportSub2
	OneToMany  []TestReportSub3
	ManyToMany []TestReportSub4
}

type TestReportSub1 struct {
	ID       uint32 `toyorm:"primary key;auto_increment"`
	Sub1Data string

	BelongToID uint32
	BelongTo   *TestReportSub1Sub1
	OneToOne   *TestReportSub1Sub2
	OneToMany  []TestReportSub1Sub3
	ManyToMany []TestReportSub1Sub4
}

type TestReportSub2 struct {
	ID                uint32 `toyorm:"primary key;auto_increment"`
	TestReportTableID uint32
	Sub2Data          string

	BelongToID uint32
	BelongTo   *TestReportSub2Sub1
	OneToOne   *TestReportSub2Sub2
	OneToMany  []TestReportSub2Sub3
	ManyToMany []TestReportSub2Sub4
}

type TestReportSub3 struct {
	ID                uint32 `toyorm:"primary key;auto_increment"`
	TestReportTableID uint32
	Sub3Data          string

	BelongToID uint32
	BelongTo   *TestReportSub3Sub1
	OneToOne   *TestReportSub3Sub2
	OneToMany  []TestReportSub3Sub3
	ManyToMany []TestReportSub3Sub4
}

type TestReportSub4 struct {
	ID       uint32 `toyorm:"primary key;auto_increment"`
	Sub4Data string

	BelongToID uint32
	BelongTo   *TestReportSub4Sub1
	OneToOne   *TestReportSub4Sub2
	OneToMany  []TestReportSub4Sub3
	ManyToMany []TestReportSub4Sub4
}

type TestReportSub1Sub1 struct {
	ID       uint32 `toyorm:"primary key;auto_increment"`
	Sub1Data string
}

type TestReportSub1Sub2 struct {
	ID               uint32 `toyorm:"primary key;auto_increment"`
	TestReportSub1ID uint32
	Sub2Data         string
}

type TestReportSub1Sub3 struct {
	ID               uint32 `toyorm:"primary key;auto_increment"`
	TestReportSub1ID uint32
	Sub3Data         string
}

type TestReportSub1Sub4 struct {
	ID       uint32 `toyorm:"primary key;auto_increment"`
	Sub4Data string
}

type TestReportSub2Sub1 struct {
	ID       uint32 `toyorm:"primary key;auto_increment"`
	Sub1Data string
}

type TestReportSub2Sub2 struct {
	ID               uint32 `toyorm:"primary key;auto_increment"`
	TestReportSub2ID uint32
	Sub2Data         string
}

type TestReportSub2Sub3 struct {
	ID               uint32 `toyorm:"primary key;auto_increment"`
	TestReportSub2ID uint32
	Sub3Data         string
}

type TestReportSub2Sub4 struct {
	ID       uint32 `toyorm:"primary key;auto_increment"`
	Sub4Data string
}

type TestReportSub3Sub1 struct {
	ID       uint32 `toyorm:"primary key;auto_increment"`
	Sub1Data string
}

type TestReportSub3Sub2 struct {
	ID               uint32 `toyorm:"primary key;auto_increment"`
	TestReportSub3ID uint32
	Sub2Data         string
}

type TestReportSub3Sub3 struct {
	ID               uint32 `toyorm:"primary key;auto_increment"`
	TestReportSub3ID uint32
	Sub3Data         string
}

type TestReportSub3Sub4 struct {
	ID       uint32 `toyorm:"primary key;auto_increment"`
	Sub4Data string
}

type TestReportSub4Sub1 struct {
	ID       uint32 `toyorm:"primary key;auto_increment"`
	Sub1Data string
}

type TestReportSub4Sub2 struct {
	ID               uint32 `toyorm:"primary key;auto_increment"`
	TestReportSub4ID uint32
	Sub2Data         string
}

type TestReportSub4Sub3 struct {
	ID               uint32 `toyorm:"primary key;auto_increment"`
	TestReportSub4ID uint32
	Sub3Data         string
}

type TestReportSub4Sub4 struct {
	ID       uint32 `toyorm:"primary key;auto_increment"`
	Sub4Data string
}

type TestRightValuePreloadTable struct {
	ID         uint32 `toyorm:"primary key;auto_increment"`
	Data       string
	ManyToMany []TestRightValuePreloadTable
}

type TestPreloadCheckTable struct {
	ID   uint32 `toyorm:"primary key;auto_increment"`
	Data string

	BelongToID uint32
	BelongTo   *TestPreloadCheckBelongTo
	OneToOne   *TestPreloadCheckOneToOne
	OneToMany  []TestPreloadCheckOneToMany
	ManyToMany []TestPreloadCheckManyToMany
}

type TestPreloadCheckBelongTo struct {
	ID   uint32 `toyorm:"primary key;auto_increment"`
	Data string
}

type TestPreloadCheckOneToOne struct {
	ID                      uint32 `toyorm:"primary key;auto_increment"`
	TestPreloadCheckTableID uint32
	Data                    string
}

type TestPreloadCheckOneToMany struct {
	ID                      uint32 `toyorm:"primary key;auto_increment"`
	TestPreloadCheckTableID uint32
	Data                    string
}

type TestPreloadCheckManyToMany struct {
	ID   uint32 `toyorm:"primary key;auto_increment"`
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

// use to create many to many preload which have foreign key
func foreignKeyManyToManyPreload(v interface{}) func(*ToyBrick) *ToyBrick {
	return func(t *ToyBrick) *ToyBrick {
		field := t.model.fieldSelect(v)
		if subBrick, ok := t.MapPreloadBrick[field.Name()]; ok {
			return subBrick
		}
		subModel := t.Toy.GetModel(LoopTypeIndirectSliceAndPtr(field.StructField().Type))
		newSubt := NewToyBrick(t.Toy, subModel).CopyStatus(t)

		newt := *t
		newt.MapPreloadBrick = t.CopyMapPreloadBrick()
		newt.MapPreloadBrick[field.Name()] = newSubt
		newSubt.preBrick = PreToyBrick{&newt, field}
		if preload := newt.Toy.manyToManyPreloadWithTag(newt.model, field, false, `toyorm:"primary key;foreign key"`); preload != nil {
			newt.ManyToManyPreload = t.CopyManyToManyPreload()
			newt.ManyToManyPreload[field.Name()] = preload
		} else {
			panic(ErrInvalidPreloadField{t.model.ReflectType.Name(), field.Name()})
		}
		return newSubt
	}
}

func createTableUnit(brick *ToyBrick) func(t *testing.T) {
	return func(t *testing.T) {
		result, err := brick.DropTableIfExist()
		assert.Nil(t, err)
		if err := result.Err(); err != nil {
			t.Error(err)
		}

		result, err = brick.CreateTable()
		assert.Nil(t, err)
		if err := result.Err(); err != nil {
			t.Error(err)
		}
	}
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
