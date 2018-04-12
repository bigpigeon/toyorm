package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bigpigeon/toyorm"
	"time"
	. "unsafe"
	// when database is mysql
	_ "github.com/go-sql-driver/mysql"
	// when database is sqlite3
	//_ "github.com/mattn/go-sqlite3"
	// when database is postgres
	//_ "github.com/lib/pq"
	"database/sql/driver"
)

type Extra map[string]interface{}

func (e *Extra) Scan(value interface{}) error {
	switch v := value.(type) {
	case string:
		return json.Unmarshal([]byte(v), e)
	case []byte:
		return json.Unmarshal(v, e)
	default:
		return errors.New("not support type")
	}
}

func (e Extra) Value() (driver.Value, error) {
	return json.Marshal(e)
}

type Color struct {
	Name string `toyorm:"primary key;join:ColorDetail"`
	Code int32
}

type Comment struct {
	toyorm.ModelDefault
	ProductDetailProductID uint32 `toyorm:"index"`
	Data                   string `toyorm:"type:VARCHAR(1024)"`
}

type ProductDetail struct {
	ProductID  uint32 `toyorm:"primary key;join:Detail"`
	Title      string
	CustomPage string `toyorm:"type:text"`
	Extra      Extra  `toyorm:"type:VARCHAR(2048)"`
	Color      string `toyorm:"join:ColorDetail"`
	// use alias to rename field
	ColorJoin Color `toyorm:"alias:ColorDetail"`
	Comment   []Comment
}

type Product struct {
	// join tag value must same as container field name
	ID        uint32     `toyorm:"primary key;auto_increment;join:Detail"`
	CreatedAt time.Time  `toyorm:"NULL"`
	DeletedAt *time.Time `toyorm:"NULL"`
	Name      string
	Count     int
	Price     float64
	Detail    *ProductDetail
}

func main() {
	var err error
	var toy *toyorm.Toy
	// when database is mysql, make sure your mysql have toyorm_example schema
	toy, err = toyorm.Open("mysql", "root:@tcp(localhost:3306)/toyorm_example?charset=utf8&parseTime=True")
	// when database is sqlite3
	//toy,err = toyorm.Open("sqlite3", "toyorm_test.db")
	// when database is postgres
	//toy, err = toyorm.Open("postgres", "user=postgres dbname=toyorm sslmode=disable")
	if err != nil {
		panic(err)
	}

	var tab Product
	var detailTab ProductDetail
	var colorTab Color

	// create table & import data
	dataInit(toy)

	// now use join to query data
	{
		brick := toy.Model(&tab).Debug().
			Join(Offsetof(tab.Detail)).
			Join(Offsetof(detailTab.ColorJoin)).Swap().Swap()
		var scanData []Product
		result, err := brick.Find(&scanData)
		if err != nil {
			panic(err)
		}
		if err := result.Err(); err != nil {
			fmt.Printf("%s\n", err)
		}
		for _, product := range scanData {
			fmt.Printf("product %s\n", JsonEncode(product))
		}
	}
	// add condition
	{
		// where Product.Name = "clean stick" or Color.Name = "black"
		brick := toy.Model(&tab).Debug().Where("=", Offsetof(tab.Name), "clean stick").
			Join(Offsetof(tab.Detail)).
			Join(Offsetof(detailTab.ColorJoin)).Or().Condition("=", Offsetof(colorTab.Name), "black").
			Swap().Swap()
		var scanData []Product
		result, err := brick.Find(&scanData)
		if err != nil {
			panic(err)
		}
		if err := result.Err(); err != nil {
			fmt.Printf("%s\n", err)
		}
		for _, product := range scanData {
			fmt.Printf("product %s\n", JsonEncode(product))
		}
	}
	// order by
	{
		// where Product.Name = "clean stick" or Color.Name = "black"
		brick := toy.Model(&tab).Debug().
			Join(Offsetof(tab.Detail)).
			Join(Offsetof(detailTab.ColorJoin)).OrderBy(Offsetof(colorTab.Name)).
			Swap().Swap()
		var scanData []Product
		result, err := brick.Find(&scanData)
		if err != nil {
			panic(err)
		}
		if err := result.Err(); err != nil {
			fmt.Printf("%s\n", err)
		}
		for _, product := range scanData {
			fmt.Printf("product %s\n", JsonEncode(product))
		}
	}
	// group by also work in join here not example

	// preload on join
	{
		brick := toy.Model(&tab).Debug().
			Join(Offsetof(tab.Detail)).Preload(Offsetof(detailTab.Comment)).Enter().
			Join(Offsetof(detailTab.ColorJoin)).Swap().Swap()
		var scanData []Product
		result, err := brick.Find(&scanData)
		if err != nil {
			panic(err)
		}
		if err := result.Err(); err != nil {
			fmt.Printf("%s\n", err)
		}
		for _, product := range scanData {
			fmt.Printf("product %s\n", JsonEncode(product))
		}
	}
}

func dataInit(toy *toyorm.Toy) {
	var tab Product
	var detailTab ProductDetail
	var colorTab Color
	brick := toy.Model(&tab).Debug()
	detailBrick := toy.Model(&detailTab).Debug().Preload(Offsetof(detailTab.Comment)).Enter()
	colorBrick := toy.Model(&colorTab).Debug()

	// create table
	for _, b := range []*toyorm.ToyBrick{brick, detailBrick, colorBrick} {
		result, err := b.DropTableIfExist()
		if err != nil {
			panic(err)
		}
		if err := result.Err(); err != nil {
			fmt.Printf("%s\n", err)
		}
		result, err = b.CreateTable()
		if err != nil {
			panic(err)
		}
		if err := result.Err(); err != nil {
			fmt.Printf("%s\n", err)
		}
	}

	products := []Product{
		{Name: "clean stick", Price: 2, Count: 1000},
		{Name: "water pipe", Price: 3, Count: 1000},
		{Name: "cable", Price: 1, Count: 1000},
	}

	result, err := brick.Insert(&products)
	if err != nil {
		panic(err)
	}
	if err := result.Err(); err != nil {
		fmt.Printf("%s\n", err)
	}
	details := []ProductDetail{
		{ProductID: products[0].ID, Color: "white", Title: "cheap and quality clean stick",
			CustomPage: "<p>pre {{ .Unit }}/{{ .Price }}</p>", Extra: Extra{"Unit": "meter"},
			Comment: []Comment{{Data: "good quality but I like black color more"}, {Data: "why only has white color"}},
		},
		{ProductID: products[1].ID, Color: "orange", Title: "pipe with non-toxic material",
			CustomPage: "<p>pre {{ .Unit }}/{{ .Price }}</p>", Extra: Extra{"Unit": "meter"},
			Comment: []Comment{{Data: "very good"}, {Data: "I think iron material is better"}},
		},
		{ProductID: products[2].ID, Color: "black", Title: "black cable",
			CustomPage: "<p>pre {{ .Unit }}/{{ .Price }}</p>", Extra: Extra{"Unit": "meter"},
			Comment: []Comment{{Data: "good signal"}, {Data: "emmmm my cat like bite it"}},
		},
	}
	result, err = detailBrick.Insert(&details)
	if err != nil {
		panic(err)
	}
	if err := result.Err(); err != nil {
		fmt.Printf("%s\n", err)
	}

	colors := []Color{
		{"white", 0xffffff},
		{"orange", 0xffa500},
		{"black", 0x000000},
	}
	result, err = colorBrick.Insert(&colors)
	if err != nil {
		panic(err)
	}
	if err := result.Err(); err != nil {
		fmt.Printf("%s\n", err)
	}

}

func JsonEncode(v interface{}) string {
	jsonData, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		panic(err)
	}
	return string(jsonData)
}
