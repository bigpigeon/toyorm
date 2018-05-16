package main

import (
	"github.com/bigpigeon/toyorm"
	_ "github.com/mattn/go-sqlite3"
	. "unsafe"
)

type Product struct {
	toyorm.ModelDefault
	Name  string
	Price int
}

func panicErr(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	toy, err := toyorm.Open("sqlite3", "mydb.db")
	panicErr(err)
	defer func() {
		err := toy.Close()
		panicErr(err)
	}()
	// create ToyBrick
	brick := toy.Model(&Product{}).Debug()
	// create table
	_, err = brick.CreateTable()
	panicErr(err)

	// insert product
	_, err = brick.Insert(&Product{
		Name:  "apple",
		Price: 22,
	})
	panicErr(err)
	// update product
	_, err = brick.Where("=", Offsetof(Product{}.Name), "apple").Update(Product{Price: 23})
	panicErr(err)
	var product Product
	// find
	_, err = brick.Find(&product)
	panicErr(err)
	// delete
	_, err = brick.Delete(&product)
	panicErr(err)
}
