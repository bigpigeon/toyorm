package main

import (
	"encoding/json"
	"fmt"
	"github.com/bigpigeon/toyorm"
	"reflect"
	. "unsafe"

	// when database is mysql
	_ "github.com/go-sql-driver/mysql"
	// when database is sqlite3
	_ "github.com/mattn/go-sqlite3"
	// when database is postgres
	//_ "github.com/lib/pq"
)

/* ----------------------
		data definition
 ------------------------- */
func JsonEncode(v interface{}) string {
	jsonData, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		panic(err)
	}
	return string(jsonData)
}

type Product struct {
	toyorm.ModelDefault
	Name  string  `toyorm:"index"`
	Price float64 `toyorm:"index"`
	Count int
	Tag   string `toyorm:"index"`
}

// use by group by
type ProductGroup struct {
	Tag       string
	KindCount int `toyorm:"column:COUNT(*)"`
}

// map to sql table name
func (p ProductGroup) TableName() string {
	return toyorm.ModelName(reflect.ValueOf(Product{}))
}

func main() {
	var err error
	var toy *toyorm.Toy
	var result *toyorm.Result
	// when database is mysql, make sure your mysql have toyorm_example schema
	toy, err = toyorm.Open("mysql", "root:@tcp(localhost:3306)/toyorm_example?charset=utf8&parseTime=True")
	// when database is sqlite3
	//toy, err = toyorm.Open("sqlite3", "toyorm_test.db")
	// when database is postgres
	//toy, err = toyorm.Open("postgres", "user=postgres dbname=toyorm sslmode=disable")

	// brick is basic "sql builder"
	brick := toy.Model(&Product{}).Debug()
	// use drop table operation to clear old data
	result, err = brick.DropTableIfExist()
	if err != nil {
		panic(err)
	}
	// print sql error if exist
	if resultErr := result.Err(); resultErr != nil {
		fmt.Print(resultErr)
	}
	// create table
	result, err = brick.CreateTableIfNotExist()
	if err != nil {
		panic(err)
	}
	// print sql error if exist
	if resultErr := result.Err(); resultErr != nil {
		fmt.Print(resultErr)
	}
	// Insert will set id to source data when primary key is auto_increment
	result, err = brick.Insert(&[]Product{
		{Name: "food one", Price: 1, Count: 4, Tag: "food"},
		{Name: "food two", Price: 2, Count: 3, Tag: "food"},
		{Name: "food three", Price: 3, Count: 2, Tag: "food"},
		{Name: "food four", Price: 4, Count: 1, Tag: "food"},
		{Name: "toolkit one", Price: 1, Count: 8, Tag: "toolkit"},
		{Name: "toolkit two", Price: 2, Count: 6, Tag: "toolkit"},
		{Name: "toolkit one", Price: 3, Count: 4, Tag: "toolkit"},
		{Name: "toolkit two", Price: 4, Count: 2, Tag: "toolkit"},
	})
	if err != nil {
		panic(err)
	}
	if resultErr := result.Err(); resultErr != nil {
		fmt.Print(resultErr)
	}
	// find one with the tag=food condition
	{
		var product Product
		result, err := brick.Where(toyorm.ExprEqual, Offsetof(Product{}.Tag), "food").Find(&product)
		if err != nil {
			panic(err)
		}
		if resultErr := result.Err(); resultErr != nil {
			fmt.Print(resultErr)
		}
		fmt.Printf("food %s\n", JsonEncode(product))
	}

	// find all tag=food
	{
		var products []Product
		result, err := brick.Where(toyorm.ExprEqual, Offsetof(Product{}.Tag), "food").Find(&products)
		if err != nil {
			panic(err)
		}
		if resultErr := result.Err(); resultErr != nil {
			fmt.Print(resultErr)
		}
		fmt.Printf("foods %s\n", JsonEncode(products))
	}

	// find count = 2
	{
		var products []Product
		result, err := brick.Where(toyorm.ExprEqual, Offsetof(Product{}.Count), 2).Find(&products)
		if err != nil {
			panic(err)
		}
		if resultErr := result.Err(); resultErr != nil {
			fmt.Print(resultErr)
		}
		fmt.Printf("count > 2 products %s\n", JsonEncode(products))
	}
	// find count = 2 and price > 3
	{
		var products []Product
		result, err := brick.Where(toyorm.ExprEqual, Offsetof(Product{}.Count), 2).And().
			Condition(toyorm.ExprGreater, Offsetof(Product{}.Price), 3).Find(&products)
		if err != nil {
			panic(err)
		}
		if resultErr := result.Err(); resultErr != nil {
			fmt.Print(resultErr)
		}
		fmt.Printf("count = 2 and price > 3 products %s\n", JsonEncode(products))
	}
	// find count = 2 and price > 3 or count = 4
	{
		var products []Product
		result, err := brick.Where(toyorm.ExprEqual, Offsetof(Product{}.Count), 2).And().
			Condition(toyorm.ExprGreater, Offsetof(Product{}.Price), 3).Or().
			Condition(toyorm.ExprEqual, Offsetof(Product{}.Count), 4).Find(&products)
		if err != nil {
			panic(err)
		}
		if resultErr := result.Err(); resultErr != nil {
			fmt.Print(resultErr)
		}
		fmt.Printf("count = 2 and price > 3 or count = 4  products %s\n", JsonEncode(products))
	}
	// find price > 3 and (count = 2 or count = 1)
	{
		var products []Product
		result, err := brick.Where(toyorm.ExprGreater, Offsetof(Product{}.Price), 3).And().Conditions(
			brick.Where(toyorm.ExprEqual, Offsetof(Product{}.Count), 2).Or().
				Condition(toyorm.ExprEqual, Offsetof(Product{}.Count), 1).Search,
		).Find(&products)
		if err != nil {
			panic(err)
		}
		if resultErr := result.Err(); resultErr != nil {
			fmt.Print(resultErr)
		}
		fmt.Printf("price > 3 and (count = 2 or count = 1)  products %s\n", JsonEncode(products))
	}

	// find (count = 2 or count = 1) and (price = 3 or price = 4)
	{
		var products []Product
		result, err := brick.Conditions(
			brick.Where(toyorm.ExprEqual, Offsetof(Product{}.Count), 2).Or().
				Condition(toyorm.ExprEqual, Offsetof(Product{}.Count), 1).Search,
		).And().Conditions(
			brick.Where(toyorm.ExprEqual, Offsetof(Product{}.Price), 3).Or().
				Condition(toyorm.ExprEqual, Offsetof(Product{}.Price), 4).Search,
		).Find(&products)
		if err != nil {
			panic(err)
		}
		if resultErr := result.Err(); resultErr != nil {
			fmt.Print(resultErr)
		}
		fmt.Printf("(count = 2 or count = 1) and (price = 3 or price = 4)  products %s\n", JsonEncode(products))
	}
	// find offset 2 limit 2
	{
		var products []Product
		result, err := brick.Offset(2).Limit(2).Find(&products)
		if err != nil {
			panic(err)
		}
		if resultErr := result.Err(); resultErr != nil {
			fmt.Print(resultErr)
		}
		fmt.Printf("offset 1 limit 2 products %s\n", JsonEncode(products))
	}
	// order by
	{
		var products []Product
		result, err := brick.OrderBy(Offsetof(Product{}.Name)).Find(&products)
		if err != nil {
			panic(err)
		}
		if resultErr := result.Err(); resultErr != nil {
			fmt.Print(resultErr)
		}
		fmt.Printf("order by name products %s\n", JsonEncode(products))
	}
	{
		var products []Product
		result, err := brick.OrderBy(brick.ToDesc(Offsetof(Product{}.Name))).Find(&products)
		if err != nil {
			panic(err)
		}
		if resultErr := result.Err(); resultErr != nil {
			fmt.Print(resultErr)
		}
		fmt.Printf("order by name desc products %s\n", JsonEncode(products))
	}
	// update to count = 4
	{
		result, err := brick.Update(&Product{
			Count: 4,
		})
		if err != nil {
			panic(err)
		}
		if resultErr := result.Err(); resultErr != nil {
			fmt.Print(resultErr)
		}
		var Counters []struct {
			Name  string
			Count int
		}
		result, err = brick.Find(&Counters)
		if err != nil {
			panic(err)
		}
		if resultErr := result.Err(); resultErr != nil {
			fmt.Print(resultErr)
		}
		for _, counter := range Counters {
			fmt.Printf("product name %s, count %d\n", counter.Name, counter.Count)
		}
	}
	// use bind fields to update a zero value
	{
		{
			var p Product
			result, err := brick.BindDefaultFields(Offsetof(p.Price), Offsetof(p.UpdatedAt)).Update(&Product{
				Price: 0,
			})
			if err != nil {
				panic(err)
			}
			if resultErr := result.Err(); resultErr != nil {
				fmt.Print(resultErr)
			}
		}
		var products []Product
		result, err = brick.Find(&products)
		if err != nil {
			panic(err)
		}
		if resultErr := result.Err(); resultErr != nil {
			fmt.Print(resultErr)
		}
		for _, p := range products {
			fmt.Printf("product name %s, price %v\n", p.Name, p.Price)
		}
	}
	// delete with element
	{
		// make a transaction, because I do not really delete a data
		brick := brick.Begin()
		var product Product
		result, err := brick.Where(toyorm.ExprEqual, Offsetof(Product{}.Name), "food four").Find(&product)
		if err != nil {
			panic(err)
		}
		if resultErr := result.Err(); resultErr != nil {
			fmt.Print(resultErr)
		}

		result, err = brick.Delete(&product)
		if err != nil {
			panic(err)
		}
		if resultErr := result.Err(); resultErr != nil {
			fmt.Print(resultErr)
		}

		var disappearProduct Product
		result, err = brick.Where(toyorm.ExprEqual, Offsetof(Product{}.Name), "food four").Find(&disappearProduct)
		if resultErr := result.Err(); resultErr != nil {
			fmt.Print(resultErr)
		}
		fmt.Printf("error(%s)\n", err)
		err = brick.Rollback()
		if err != nil {
			panic(err)
		}
	}
	// delete with condition
	{
		// make a transaction, because I am not really delete a data
		brick := brick.Begin()
		result, err := brick.Where(toyorm.ExprEqual, Offsetof(Product{}.Name), "food four").DeleteWithConditions()
		if err != nil {
			panic(err)
		}
		if resultErr := result.Err(); resultErr != nil {
			fmt.Print(resultErr)
		}
		var product Product
		_, err = brick.Where(toyorm.ExprEqual, Offsetof(Product{}.Name), "food four").Find(&product)
		fmt.Printf("error(%s)\n", err)
		err = brick.Rollback()
		if err != nil {
			panic(err)
		}
	}
	// group by
	{
		var tab ProductGroup
		brick := toy.Model(&tab).Debug().GroupBy(Offsetof(tab.Tag))
		var groups []ProductGroup
		result, err := brick.Find(&groups)
		if err != nil {
			panic(err)
		}
		if resultErr := result.Err(); resultErr != nil {
			fmt.Print(resultErr)
		}
		for _, g := range groups {
			fmt.Printf("group %#v\n", g)
		}
	}
	// custom insert
	var saveData Product
	{
		data := Product{
			Name:  "bag",
			Price: 9999,
			Count: 2,
			Tag:   "container",
		}
		result, err := brick.Template("INSERT INTO $ModelName($Columns) Values($Values)").Insert(&data)
		if err != nil {
			panic(err)
		}
		if resultErr := result.Err(); resultErr != nil {
			fmt.Print(resultErr)
		}
		saveData = data
	}
	// custom save
	{
		saveData.Price = 9988
		result, err := brick.Template("INSERT INTO $ModelName($Columns) VALUES($Values) ON DUPLICATE KEY UPDATE $Cas $UpdateValues").Save(&saveData)
		if err != nil {
			panic(err)
		}
		if resultErr := result.Err(); resultErr != nil {
			fmt.Print(resultErr)
		}
	}
	// custom usave
	{
		saveData.Price = 9987
		result, err := brick.Template("UPDATE $ModelName SET $UpdateValues $Conditions").USave(&saveData)
		if err != nil {
			panic(err)
		}
		if resultErr := result.Err(); resultErr != nil {
			fmt.Print(resultErr)
		}
	}
	// custom find
	{
		var data Product
		// if driver is mysql use "USE INDEX" replace "INDEXED BY"
		result, err := brick.Template("SELECT $Columns FROM $ModelName idx_product_name $Conditions OFFSET 0").
			Where("=", Offsetof(Product{}.Name), "bag").Find(&data)
		if err != nil {
			panic(err)
		}
		if resultErr := result.Err(); resultErr != nil {
			fmt.Print(resultErr)
		}
		fmt.Printf("find where name = bag %v\n", data)
	}
	// custom update
	{
		fmt.Printf("the template source %s\n", fmt.Sprintf("UPDATE $ModelName SET $UpdateValues,$FN-Count = $FN-Count + ? $Conditions"))
		result, err := brick.Template(fmt.Sprintf("UPDATE $ModelName SET $UpdateValues,$FN-Count = $FN-Count + ? $Conditions"), 2).
			Where("=", Offsetof(Product{}.Name), "bag").Update(&Product{Price: 200})
		if err != nil {
			panic(err)
		}
		if resultErr := result.Err(); resultErr != nil {
			fmt.Print(resultErr)
		}
		var data Product
		result, err = brick.Where("=", Offsetof(Product{}.Name), "bag").Find(&data)
		if err != nil {
			panic(err)
		}
		if resultErr := result.Err(); resultErr != nil {
			fmt.Print(resultErr)
		}

		fmt.Printf("now bag product count is %d\n", data.Count)
	}
}
