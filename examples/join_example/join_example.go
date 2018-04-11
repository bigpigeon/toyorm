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

func JsonEncode(v interface{}) string {
	jsonData, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		panic(err)
	}
	return string(jsonData)
}

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

type ProductDetail struct {
	ProductID  uint32 `toyorm:"primary key;join:Detail"`
	Title      string
	CustomPage string `toyorm:"type:text"`
	Extra      Extra  `toyorm:"type:VARCHAR(2048)"`
	Color      string `toyorm:"join:ColorDetail"`
	ColorJoin  Color  `toyorm:"container:ColorDetail"`
	//TODO add preload
}

type Product struct {
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
	var tab Product
	var detailTab ProductDetail
	var colorTab Color
	brick := toy.Model(&tab).Debug()
	detailBrick := toy.Model(&detailTab).Debug()
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
	// import data
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
		{ProductID: products[0].ID, Color: "white", Title: "cheap and quality clean stick", CustomPage: "<p>pre {{ .Unit }}/{{ .Price}}</p>", Extra: Extra{"Unit": "meter"}},
		{ProductID: products[1].ID, Color: "orange", Title: "pipe with non-toxic material", CustomPage: "<p>pre {{ .Unit }}/{{ .Price}}</p>", Extra: Extra{"Unit": "meter"}},
		{ProductID: products[2].ID, Color: "black", Title: "black cable", CustomPage: "<p>pre {{ .Unit }}/{{ .Price}}</p>", Extra: Extra{"Unit": "meter"}},
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

	// now use join to query data
	{
		brick := toy.Model(&tab).Debug().
			Join(Offsetof(tab.Detail)).
			Join(Offsetof(detailTab.ColorJoin)).Swap().Swap()
		var scanData []Product
		result, err = brick.Find(&scanData)
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
		result, err = brick.Find(&scanData)
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
		result, err = brick.Find(&scanData)
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
}
