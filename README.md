# Toyorm

this is powerful sql orm library for Golang, have some funny features

[![Build Status](https://travis-ci.org/bigpigeon/toyorm.svg)](https://travis-ci.org/bigpigeon/toyorm)
[![codecov](https://codecov.io/gh/bigpigeon/toyorm/branch/master/graph/badge.svg)](https://codecov.io/gh/bigpigeon/toyorm)
[![Go Report Card](https://goreportcard.com/badge/github.com/bigpigeon/toyorm)](https://goreportcard.com/report/github.com/bigpigeon/toyorm)
[![GoDoc](https://godoc.org/github.com/bigpigeon/toyorm?status.svg)](https://godoc.org/github.com/bigpigeon/toyorm)
[![Join the chat](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/toyorm/toyorm)

---

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->


  - [A Simple Example](#a-simple-example)
  - [Database connection](#database-connection)
  - [Model definition](#model-definition)
      - [example](#example)
      - [type translate](#type-translate)
    - [Bind models](#bind-models)
  - [Sql Operation](#sql-operation)
    - [create table](#create-table)
    - [drop table](#drop-table)
    - [insert/save data](#insertsave-data)
    - [update](#update)
    - [find](#find)
    - [delete](#delete)
  - [ToyBrick](#toybrick)
    - [Where condition](#where-condition)
      - [usage](#usage)
      - [SearchExpr](#searchexpr)
      - [example](#example-1)
    - [Transaction](#transaction)
    - [Debug](#debug)
    - [IgnoreMode](#ignoremode)
    - [BindFields](#bindfields)
    - [Scope](#scope)
    - [Thread safe](#thread-safe)
    - [Preload](#preload)
  - [Result](#result)
    - [Selector](#selector)
  - [Full Feature Example](#full-feature-example)
- [Collection](#collection)
  - [ToyCollection](#toycollection)
  - [CollectionBrick](#collectionbrick)
    - [Selector](#selector-1)
    - [id generator](#id-generator)
    - [sql action](#sql-action)
  - [collection example](#collection-example)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

---



### Support database

[sqlite3](https://www.sqlite.org/)  [mysql](https://www.mysql.com/)

### Go version

version go-1.9

---

### A Simple Example

copy following code to a main.go file and run

```golang
package main

import (
	"encoding/json"
	"fmt"
	"github.com/bigpigeon/toyorm"
	. "unsafe"

	// when database is mysql
	_ "github.com/go-sql-driver/mysql"
	// when database is sqlite3
	//_ "github.com/mattn/go-sqlite3"
	"reflect"
)

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

type ProductGroup struct {
	Tag       string
	KindCount int `toyorm:"column:COUNT(*)"`
}

func (p ProductGroup) TableName() string {
	return toyorm.ModelName(reflect.TypeOf(Product{}))
}

func main() {
	var err error
	var toy *toyorm.Toy
	var result *toyorm.Result
	// when database is mysql, make sure your mysql have toyorm_example schema
	toy, err = toyorm.Open("mysql", "root:@tcp(localhost:3306)/toyorm_example?charset=utf8&parseTime=True")
	// when database is sqlite3
	//toy,err = toyorm.Open("sqlite3", "toyorm_test.db")

	brick := toy.Model(&Product{}).Debug()
	result, err = brick.DropTableIfExist()
	if err != nil {
		panic(err)
	}
	if resultErr := result.Err(); resultErr != nil {
		fmt.Print(resultErr)
	}
	result, err = brick.CreateTableIfNotExist()
	if err != nil {
		panic(err)
	}
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
	// find the first food
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

	// find all food
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
		brick.Rollback()
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
		brick.Rollback()
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
}

```

---

### Database connection

import database driver

```golang
// if database is mysql
_ "github.com/go-sql-driver/mysql"
// if database is sqlite3
_ "github.com/mattn/go-sqlite3"
```

create a toy

```golang
// if database is mysql, make sure your mysql have toyorm_example schema
toy, err = toyorm.Open("mysql", "root:@tcp(localhost:3306)/toyorm_example?charset=utf8&parseTime=True")
// if database is sqlite3
toy,err = toyorm.Open("sqlite3", "toyorm_test.db")
```


### Model definition

##### example

```golang
type Extra map[string]interface{}

func (e Extra) Scan(value interface{}) error {
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

type UserDetail struct {
     ID       int  `toyorm:"primary key;auto_increment"`
     UserID   uint `toyorm:"index"`
     MainPage string
     Extra    Extra `toyorm:"type:VARCHAR(1024)"`
}

type Blog struct {
     toyorm.ModelDefault
     UserID  uint   `toyorm:"index"`
     Title   string `toyorm:"index"`
     Content string
}

type User struct {
     toyorm.ModelDefault
     Name    string `toyorm:"unique index"`
     Age     int
     Sex     string
     Detail  *UserDetail
     Friends []*User
     Blog    []Blog
}
```

##### type translate

if sql type is not match, toyorm will ignore it in Sql Operation


you can use **\<type:sql_type\>** field tag to specified their sql type


the following type will auto translate to sql type


Go Type | sql type
--------|-----------
bool    | BOOLEAN
int8,int16,int32,uint8,uint16,uint32| INTEGER
int64,uint64,int,uint| BIGINT
float32,float64| FLOAT
string  | VARCHAR(255)
time.Time | TIMESTAMP
[]byte    | VARCHAR(255)
sql.NullBool | BOOLEAN
sql.NullInt64 | BIGINT
sql.NullFloat64 | FLOAT
sql.NullString | VARCHAR(255)
sql.RawBytes | VARCHAR(255)

**special fields**

1. special fields have some process in handlers, do not try to change it type or set it value


Field Name| Type      | Description
----------|-----------|------------
CreatedAt |time.Time  | generate when element be create
UpdatedAt |time.Time  | generate when element be update/create
DeletedAt |*time.Time | delete mode is soft

**field tags**

1. tag format can be \<key:value\> or \<key\>

2. the following is special tag

Key           |Value                   |Description
--------------|------------------------|-----------
index         | void or string         | use for optimization when search condition have this field,if you want make a combined,just set same index name with fields
unique index  | void or string         | have unique limit index, other same as index
primary key   | void                   | allow multiple primary key,but some operation not support
\-            | void                    | ignore this field in sql
type          | string                  | sql type
column        | string                  | sql column name
auto_increment| void                    | recommend, if your table primary key have auto_increment attribute must add it
autoincrement | void                    | same as auto_increment
foreign key   | void                   | to add foreign key feature when create table

other custom TAG will append to end of CREATE TABLE field


#### Bind models

1. model kind must be a struct or a point with struct

2. the model is what information that toyorm know about the table

```golang
brick := toy.Model(&User{})
// or
brick := toy.Model(User{})
```

### Sql Operation

---

#### create table

```golang
var err error
_, err = toy.Model(&User{}).Debug().CreateTable()
// CREATE TABLE user (id BIGINT AUTO_INCREMENT,created_at TIMESTAMP NULL,updated_at TIMESTAMP NULL,deleted_at TIMESTAMP NULL,name VARCHAR(255),age BIGINT ,sex VARCHAR(255) , PRIMARY KEY(id))
// CREATE INDEX idx_user_deletedat ON user(deleted_at)
// CREATE UNIQUE INDEX udx_user_name ON user(name)
_, err =toy.Model(&UserDetail{}).Debug().CreateTable()
// CREATE TABLE user_detail (id BIGINT AUTO_INCREMENT,user_id BIGINT,main_page Text,extra VARCHAR(1024), PRIMARY KEY(id))
// CREATE INDEX idx_user_detail_userid ON user_detail(user_id)
_, err =toy.Model(&Blog{}).Debug().CreateTable()
// CREATE TABLE blog (id BIGINT AUTO_INCREMENT,created_at TIMESTAMP NULL,updated_at TIMESTAMP NULL,deleted_at TIMESTAMP NULL,user_id BIGINT,title VARCHAR(255),content VARCHAR(255) , PRIMARY KEY(id))
// CREATE INDEX idx_blog_deletedat ON blog(deleted_at)
// CREATE INDEX idx_blog_userid ON blog(user_id)
// CREATE INDEX idx_blog_title ON blog(title)
```

#### drop table

```golang
var err error
_, err =toy.Model(&User{}).Debug().DropTable()
// DROP TABLE user
_, err =toy.Model(&UserDetail{}).Debug().DropTable()
// DROP TABLE user_detail
_, err =toy.Model(&Blog{}).Debug().DropTable()
// DROP TABLE blog
```

#### insert/save data

// insert with autoincrement will set id to source data

```golang
user := &User{
    Name: "bigpigeon",
    Age:  18,
    Sex:  "male",
}
_, err = toy.Model(&User{}).Debug().Insert(&user)
// INSERT INTO user(created_at,updated_at,name,age,sex) VALUES(?,?,?,?,?) , args:[]interface {}{time.Time{wall:0xbe8df5112e7f07c8, ext:210013499, loc:(*time.Location)(0x141af80)}, time.Time{wall:0xbe8df5112e7f1768, ext:210017044, loc:(*time.Location)(0x141af80)}, "bigpigeon", 18, "male"}
// print user format with json
/* {
  "ID": 1,
  "CreatedAt": "2018-01-11T20:47:00.780077+08:00",
  "UpdatedAt": "2018-01-11T20:47:00.780081+08:00",
  "DeletedAt": null,
  "Name": "bigpigeon",
  "Age": 18,
  "Sex": "male",
  "Detail": null,
  "Friends": null,
  "Blog": null
}*/
```

// save data use "REPLACE INTO" when primary key exist

```golang
users := []User{
    {
        ModelDefault: toyorm.ModelDefault{ID: 1},
        Name:         "bigpigeon",
        Age:          18,
        Sex:          "male",
    },
    {
        Name: "fatpigeon",
        Age:  27,
        Sex:  "male",
    },
}
_, err = toy.Model(&User{}).Debug().Save(&user)
// SELECT id,created_at FROM user WHERE id IN (?), args:[]interface {}{0x1}
// REPLACE INTO user(id,created_at,updated_at,name,age,sex) VALUES(?,?,?,?,?,?) , args:[]interface {}{0x1, time.Time{wall:0x0, ext:63651278036, loc:(*time.Location)(nil)}, time.Time{wall:0xbe8dfb5511465918, ext:302600558, loc:(*time.Location)(0x141af80)}, "bigpigeon", 18, "male"}
// INSERT INTO user(created_at,updated_at,name,age,sex) VALUES(?,?,?,?,?) , args:[]interface {}{time.Time{wall:0xbe8dfb551131b7d8, ext:301251230, loc:(*time.Location)(0x141af80)}, time.Time{wall:0xbe8dfb5511465918, ext:302600558, loc:(*time.Location)(0x141af80)}, "fatpigeon", 27, "male"}

```

#### update

```golang
toy.Model(&User{}).Debug().Update(&User{
    Age: 4,
})
// UPDATE user SET updated_at=?,age=? WHERE deleted_at IS NULL, args:[]interface {}{time.Time{wall:0xbe8df4eb81b6c050, ext:233425327, loc:(*time.Location)(0x141af80)}, 4}
```


#### find

find one

```golang
var user User
_, err = toy.Model(&User{}).Debug().Find(&user}
// SELECT id,created_at,updated_at,deleted_at,name,age,sex FROM user WHERE deleted_at IS NULL LIMIT 1, args:[]interface {}(nil)
// print user format with json
/* {
  "ID": 1,
  "CreatedAt": "2018-01-11T12:47:01Z",
  "UpdatedAt": "2018-01-11T12:47:01Z",
  "DeletedAt": null,
  "Name": "bigpigeon",
  "Age": 4,
  "Sex": "male",
  "Detail": null,
  "Friends": null,
  "Blog": null
}*/
```

find multiple

```golang
var users []User
_, err = brick.Debug().Find(&users)
fmt.Printf("find users %s\n", JsonEncode(&users))

// SELECT id,created_at,updated_at,deleted_at,name,age,sex FROM user WHERE deleted_at IS NULL, args:[]interface {}(nil)
```

#### delete

delete with primary key

```golang
_, err = brick.Debug().Delete(&user)
// UPDATE user SET deleted_at=? WHERE id IN (?), args:[]interface {}{(*time.Time)(0xc4200f0520), 0x1}
```

delete with condition

```golang
_, err = brick.Debug().Where(toyorm.ExprEqual, Offsetof(User{}.Name), "bigpigeon").DeleteWithConditions()
// UPDATE user SET deleted_at=? WHERE name = ?, args:[]interface {}{(*time.Time)(0xc4200dbfa0), "bigpigeon"}
```

### ToyBrick

-----

use **toy.Model** will create a ToyBrick, you need use it to build grammar and operate the database

#### Where condition

affective update/find/delete operation

##### usage

where will clean old conditions and make new one

    brick.Where(<expr>, <Key>, [value])

conditions will copy conditions and clean old conditions

    brick.Conditions(<toyorm.Search>)

or & and condition will use or/and to link new condition when current condition is not nil

    brick.Or().Condition(<expr>, <Key>, [value])
    brick.And().Condition(<expr>, <Key>, [value])

or & and conditions will use or/and to link new conditions

    brick.Or().Conditions(<toyorm.Search>)
    brick.And().Conditions(<toyorm.Search>)

##### SearchExpr

SearchExpr        |  to sql      | example
------------------|--------------|:----------------
ExprAnd           | AND          | brick.Where(ExprAnd, Product{Name:"food one", Count: 4}) // WHERE name = "food one" AND Count = 4
ExprOr            | OR           | brick.Where(ExprOr, Product{Name:"food one", Count: 4}) // WHERE name = "food one" OR Count = "4"
ExprEqual         | =            | brick.Where(ExprEqual, OffsetOf(Product{}.Name), "food one") // WHERE name = "find one"
ExprNotEqual      | <>           | brick.Where(ExprNotEqual, OffsetOf(Product{}.Name), "food one") // WHERE name <> "find one"
ExprGreater       | >            | brick.Where(ExprGreater, OffsetOf(Product{}.Count), 3) // WHERE count > 3
ExprGreaterEqual  | >=           | brick.Where(ExprGreaterEqual, OffsetOf(Product{}.Count), 3) // WHERE count >= 3
ExprLess          | <            | brick.Where(ExprLess, OffsetOf(Product{}.Count), 3) // WHERE count < 3
ExprLessEqual     | <=           | brick.Where(ExprLessEqual, OffsetOf(Product{}.Count), 3) // WHERE count <= 3
ExprBetween       | Between      | brick.Where(ExprBetween, OffsetOf(Product{}.Count), [2]int{2,3}) // WHERE count BETWEEN 2 AND 3
ExprNotBetween    | NOT Between  | brick.Where(ExprNotBetween, OffsetOf(Product{}.Count), [2]int{2,3}) // WHERE count NOT BETWEEN 2 AND 3
ExprIn            | IN           | brick.Where(ExprIn, OffsetOf(Product{}.Count), []int{1, 2, 3}) // WHERE count IN (1,2,3)
ExprNotIn         | NOT IN       | brick.Where(ExprNotIn, OffsetOf(Product{}.Count), []int{1, 2, 3}) // WHERE count NOT IN (1,2,3)
ExprLike          | LIKE         | brick.Where(ExprLike, OffsetOf(Product{}.Name), "one") // WHERE name LIKE "one"
ExprNotLike       | NOT LIKE     | brick.Where(ExprNotLike, OffsetOf(Product{}.Name), "one") // WHERE name NOT LIKE "one"
ExprNull          | IS NULL      | brick.Where(ExprNull, OffsetOf(Product{}.DeletedAt)) // WHERE DeletedAt IS NULL
ExprNotNull       | IS NOT NULL  | brick.Where(ExprNotNull, OffsetOf(Product{}.DeletedAt)) // WHERE DeletedAt IS NOT NULL

##### example

single condition

```golang
brick = brick.Where(toyorm.ExprEqual, Offsetof(Product{}.Tag), "food")
or use string
brick = brick.Where("=", Offsetof(Product{}.Tag), "food")
// WHERE tag = "food"
```

combination condition

```golang
brick = brick.Where(toyorm.ExprEqual, Offsetof(Product{}.Count), 2).And().
    Condition(toyorm.ExprGreater, Offsetof(Product{}.Price), 3).Or().
    Condition(toyorm.ExprEqual, Offsetof(Product{}.Count), 4)
or use string
brick = brick.Where("=", Offsetof(Product{}.Count), 2).And().
    Condition(">", Offsetof(Product{}.Price), 3).Or().
    Condition("=", Offsetof(Product{}.Count), 4)
// WHERE count = 2 and price > 3 or count = 4
```

priority condition

```golang
brick.Where(toyorm.ExprGreater, Offsetof(Product{}.Price), 3).And().Conditions(
    brick.Where(toyorm.ExprEqual, Offsetof(Product{}.Count), 2).Or().
    Condition(toyorm.ExprEqual, Offsetof(Product{}.Count), 1).Search
)
or use string
brick.Where(">", Offsetof(Product{}.Price), 3).And().Conditions(
    brick.Where("=", Offsetof(Product{}.Count), 2).Or().
    Condition("=", Offsetof(Product{}.Count), 1).Search
)
// WHERE price > 3 and (count = 2 or count = 1)
```


```golang
brick.Conditions(
    brick.Where(toyorm.ExprEqual, Offsetof(Product{}.Count), 2).Or().
        Condition(toyorm.ExprEqual, Offsetof(Product{}.Count), 1).Search,
).And().Conditions(
    brick.Where(toyorm.ExprEqual, Offsetof(Product{}.Price), 3).Or().
        Condition(toyorm.ExprEqual, Offsetof(Product{}.Price), 4).Search,
)
or use string
brick.Conditions(
    brick.Where("=", Offsetof(Product{}.Count), 2).Or().
        Condition("=", Offsetof(Product{}.Count), 1).Search,
).And().Conditions(
    brick.Where("=", Offsetof(Product{}.Price), 3).Or().
        Condition("=", Offsetof(Product{}.Price), 4).Search,
)
// WHERE (count = ? OR count = ?) AND (price = ? OR price = ?)
```

limit & offset

```golang
brick := brick.Offset(2).Limit(2)
// LIMIT 2 OFFSET 2
```

order by

```golang
brick = brick.OrderBy(Offsetof(Product{}.Name))
// ORDER BY name
```

order by desc

```golang
brick = brick.OrderBy(brick.ToDesc(Offsetof(Product{}.Name)))
// ORDER BY name DESC
```


#### Transaction

start a transaction

```golang
brick = brick.Begin()
```

rollback all sql action

```golang
err = brick.Rollback()
```

commit all sql action

```golang
err = brick.Commit()
```

#### Debug

if Set debug all sql action will have log

```golang
brick = brick.Debug()
```


#### IgnoreMode

when I Update or Search with struct that have some zero value, did I update it ?


use IgnoreMode to differentiate what zero value should update

```golang
brick = brick.IgnoreMode(toyorm.Mode("Update"), toyorm.IgnoreZero ^ toyorm.IgnoreZeroLen)
// ignore all zeor value but excloud zero len slice
// now value = []int(nil) will ignore when update
// but value = []int{} will update when update
```

In default

Operation | Mode       | affect
----------|------------|--------
Insert    | IgnoreNo   | brick.Insert(<struct>)
Replace   | IgnoreNo   | brick.Replace(<struct>)
Condition | IgnoreZero | brick.Where(ExprAnd/ExprOr, <struct>)
Update    | IgnoreZero | brick.Update(<struct>)

**All of IgnoreMode**

mode              |  effective
------------------|---------------
IgnoreFalse       | ignore field type is bool and value is false
IgnoreZeroInt     | ignore field type is int/uint/uintptr(incloud their 16,32,64 bit type) and value is 0
IgnoreZeroFloat   | ignore field type is float32/float64 and value is 0.0
IgnoreZeroComplex | ignore field type is complex64/complex128 and value is 0 + 0i
IgnoreNilString   | ignore field type is string and value is ""
IgnoreNilPoint    | ignore field type is point/map/slice and value is nil
IgnoreZeroLen     | ignore field type is map/array/slice and len = 0
IgnoreNullStruct  | ignore field type is struct and value is zero value struct e.g type A struct{A string,B int}, A{"", 0} will be ignore
IgnoreNil         | ignore with IgnoreNilPoint and IgnoreZeroLen
IgnoreZero        | ignore all of the above


#### BindFields

if bind a not null fields, the IgnoreMode will failure

```golang
{
    var p Product
    result, err := brick.BindDefaultFields(Offsetof(p.Price), Offsetof(p.UpdatedAt)).Update(&Product{
        Price: 0,
    })
   // process error
   ...
}
var products []Product
result, err = brick.Find(&products)
// process error
...

for _, p := range products {
    fmt.Printf("product name %s, price %v\n", p.Name, p.Price)
}
```

#### Scope

use scope to do some custom operation


```golang
// desc all order by fields
brick.Scope(func(t *ToyBrick) *ToyBrick{

    newOrderBy := make([]*ModelFields, len(t.orderBy))
    for i, f := range t.orderBy {
        newOrderBy = append(newOrderBy, t.ToDesc(f))
    }
    newt := *t
    newt.orderBy = newOrderBy
    return &newt
})
```

#### Thread safe

Thread safe if you comply with the following agreement

1. make sure **ToyBrick** object is read only, if you want to change it, create a new one

2. do not use **append** to change ToyBrick's slice data,use **make** and **copy** to clone new slice


#### Preload

preload need have relation field and container field


relations field is used to link the main record and sub record


container field is used to hold sub record

**one to one**

relation field at sub model


relation field name must be main model type name + main model primary key name

```golang
type User struct {
    toyorm.ModelDefault
    // container field
    Detail  *UserDetail
}

type UserDetail struct {
    ID       int    `toyorm:"primary key;auto_increment"`
    // relation field
    UserID   uint   `toyorm:"index"`
    MainPage string `toyorm:"type:Text"`
}

// load preload
brick = toy.Model(&User{}).Debug().Preload(OffsetOf(User.Detail)).Enter()

```

**belong to**

relation field at main model


relation field name must be container field name + sub model primary key name

```golang
type User struct {
    toyorm.ModelDefault
    // container field
    Detail   *UserDetail
    // relation field
    DetailID int `toyorm:"index"`
}

type UserDetail struct {
    ID       int    `toyorm:"primary key;auto_increment"`
    MainPage string `toyorm:"type:Text"`
}

```

**one to many**

relation field at sub model



relation field name must be main model type name + main model primary key name


```golang
type User struct {
    toyorm.ModelDefault
    // container field
    Blog    []Blog
}

type Blog struct {
    toyorm.ModelDefault
    // relation field
    UserID  uint   `toyorm:"index"`
    Title   string `toyorm:"index"`
    Content string
}

```

**many to many**

many to many not need to specified the relation ship,it relation field at middle model

```
type User struct {
    toyorm.ModelDefault
    // container field
    Friends    []*User
}
```

**load preload**

when you finish model definition, it time to load preload

```golang
// create a main brick
brick = toy.Model(&User{})
// create a sub brick
subBrick := brick.Preload(OffsetOf(User.Blog))
// you can editing any attribute what you want, just like editing it on main model
subBrick = subBrick.Where(ExprEqual, OffsetOf(Blog.Title), "my blog")
// finished change ,use Enter() go back the main brick
brick = subBrick.Enter()
```


if you not like relation field name rule,use custom module to create it

```golang
// one to one custom
brick.CustomOneToOnePreload(<main container>, <sub relation>, [sub model struct])
// belong to custom
brick.CustomBelongToPreload(<main container>, <main relation>, [sub model struct])
// one to many
brick.CustomOneToManyPreload(<main container>, <sub relation>, [sub model struct])
// many to many
brick.CustomManyToManyPreload(<middle model struct>, <main container>, <main relation>, <sub relation>, [sub model struct])
```


### Result

**use Report to view sql action**

report format


insert

```golang
user := User{
    Detail: &UserDetail{
        MainPage: "some html code with you page",
        Extra:    Extra{"title": "my blog"},
    },
    Blog: []Blog{
        {Title: "how to write a blog", Content: "first ..."},
        {Title: "blog introduction", Content: "..."},
    },
    Friends: []*User{
        {
            Detail: &UserDetail{
                MainPage: "some html code with you page",
                Extra:    Extra{},
            },
            Blog: []Blog{
                {Title: "some python tech", Content: "first ..."},
                {Title: "my eleme_union_meal usage", Content: "..."},
            },
            Name: "fatpigeon",
            Age:  18,
            Sex:  "male",
        },
    },
    Name: "bigpigeon",
    Age:  18,
    Sex:  "male",
}
result, err = brick.Save(&user)
// error process ...
fmt.Printf("report:\n%s\n", result.Report())

/*
// [0, ] means affected the 0 element
// [0-0, ] means affected the 0 element the 0 sub element
report:
[0, ] INSERT INTO user(created_at,updated_at,deleted_at,name,age,sex) VALUES(?,?,?,?,?,?)  args:["2018-02-28T17:31:20.012285+08:00","2018-02-28T17:31:20.012285+08:00",null,"bigpigeon",18,"male"]
	preload Detail
	[0-, ] INSERT INTO user_detail(user_id,main_page,extra) VALUES(?,?,?)  args:[1,"some html code with you page",{"title":"my blog"}]
	preload Blog
	[0-0, ] INSERT INTO blog(created_at,updated_at,deleted_at,user_id,title,content) VALUES(?,?,?,?,?,?)  args:["2018-02-28T17:31:20.013968+08:00","2018-02-28T17:31:20.013968+08:00",null,1,"how to write a blog","first ..."]
	[0-1, ] INSERT INTO blog(created_at,updated_at,deleted_at,user_id,title,content) VALUES(?,?,?,?,?,?)  args:["2018-02-28T17:31:20.013968+08:00","2018-02-28T17:31:20.013968+08:00",null,1,"blog introduction","..."]
	preload Friends
	[0-0, ] INSERT INTO user(created_at,updated_at,deleted_at,name,age,sex) VALUES(?,?,?,?,?,?)  args:["2018-02-28T17:31:20.015207+08:00","2018-02-28T17:31:20.015207+08:00",null,"fatpigeon",18,"male"]
		preload Detail
		[0-0-, ] INSERT INTO user_detail(user_id,main_page,extra) VALUES(?,?,?)  args:[2,"some html code with you page",{}]
		preload Blog
		[0-0-0, ] INSERT INTO blog(created_at,updated_at,deleted_at,user_id,title,content) VALUES(?,?,?,?,?,?)  args:["2018-02-28T17:31:20.016389+08:00","2018-02-28T17:31:20.016389+08:00",null,2,"some python tech","first ..."]
		[0-0-1, ] INSERT INTO blog(created_at,updated_at,deleted_at,user_id,title,content) VALUES(?,?,?,?,?,?)  args:["2018-02-28T17:31:20.016389+08:00","2018-02-28T17:31:20.016389+08:00",null,2,"my eleme_union_meal usage","..."]
*/
```

find

```golang
brick := brick.Preload(Offsetof(User{}.Friends)).
    Preload(Offsetof(User{}.Detail)).Enter().
    Preload(Offsetof(User{}.Blog)).Enter().
    Enter()
var users []User
result, err = brick.Find(&users)
// some error process
...
// print the report
fmt.Printf("report:\n%s\n", result.Report())

// report log
/*
report:
[0, 1, ] SELECT id,created_at,updated_at,deleted_at,name,age,sex FROM user WHERE deleted_at IS NULL  args:null
	preload Detail
	[0-, 1-, ] SELECT id,user_id,main_page,extra FROM user_detail WHERE user_id IN (?,?)  args:[2,1]
	preload Blog
	[0-0, 0-1, 1-0, 1-1, ] SELECT id,created_at,updated_at,deleted_at,user_id,title,content FROM blog WHERE deleted_at IS NULL AND user_id IN (?,?)  args:[1,2]
	preload Friends
	[0-0, ] SELECT id,created_at,updated_at,deleted_at,name,age,sex FROM user WHERE deleted_at IS NULL AND id IN (?)  args:[2]
		preload Detail
		[0-0-, ] SELECT id,user_id,main_page,extra FROM user_detail WHERE user_id IN (?)  args:[2]
		preload Blog
		[0-0-0, 0-0-1, ] SELECT id,created_at,updated_at,deleted_at,user_id,title,content FROM blog WHERE deleted_at IS NULL AND user_id IN (?)  args:[2]
*/

```

**use Err to view sql action error**

```golang
var users []struct {
    ID     uint32
    Age    bool
    Detail *UserDetail
    Blog   []Blog
}
result, err = brick.Find(&users)
if err != nil {
    panic(err)
}
if err := result.Err(); err != nil {
    fmt.Printf("error:\n%s\n", err)
}

/*
error:
SELECT id,age FROM user WHERE deleted_at IS NULL  args:null errors(
	[0]sql: Scan error on column index 1: sql/driver: couldn't convert "18" into type bool
	[1]sql: Scan error on column index 1: sql/driver: couldn't convert "18" into type bool
)
*/
```

#### Selector

toyorm support following selector

operation  \\  selector | OffsetOf | Name string | map\[OffsetOf\]interface{} | map\[string\]interface{} | struct |
--------------------|----------|-------------|--------------------------------|--------------------------|--------|
Update              | no       | no          | yes                            | yes                      | yes
Insert              | no       | no          | yes                            | yes                      | yes
Save                | no       | no          | yes                            | yes                      | yes
Where & Conditions  | yes      | yes         | yes                            | yes                      | yes
BindFields          | yes      | yes         | no                             | no                       | no
Preload & Custom Preload | yes | yes         | no                             | no                       | no
OrderBy             | yes      | yes         | no                             | no                       | no
Find                | no       | no          | no                             | no                       | yes





### Full Feature Example

```golang
package main

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bigpigeon/toyorm"
	"runtime"
	. "unsafe"

	// when database is mysql
	_ "github.com/go-sql-driver/mysql"
	// when database is sqlite3
	//_ "github.com/mattn/go-sqlite3"
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

type UserDetail struct {
	ID       int    `toyorm:"primary key;auto_increment"`
	UserID   uint32 `toyorm:"index"`
	MainPage string `toyorm:"type:Text"`
	Extra    Extra  `toyorm:"type:VARCHAR(1024)"`
}

type Blog struct {
	toyorm.ModelDefault
	UserID  uint32 `toyorm:"index"`
	Title   string `toyorm:"index"`
	Content string
}

type User struct {
	toyorm.ModelDefault
	Name    string `toyorm:"unique index"`
	Age     int
	Sex     string
	Detail  *UserDetail
	Friends []*User
	Blog    []Blog
}

func JsonEncode(v interface{}) string {
	jsonData, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		panic(err)
	}
	return string(jsonData)
}

func processError(result *toyorm.Result, err error) {
	if err != nil {
		panic(err)
	}
	if rErr := result.Err(); rErr != nil {
		runtime.Caller(2)
		fmt.Printf("%s\n", rErr)
	}
}

func main() {
	var err error
	var toy *toyorm.Toy
	// when database is mysql, make sure your mysql have toyorm_example schema
	toy, err = toyorm.Open("mysql", "root:@tcp(localhost:3306)/toyorm_example?charset=utf8&parseTime=True")
	// when database is sqlite3
	//toy,err = toyorm.Open("sqlite3", "toyorm_test.db")

	var result *toyorm.Result
	if err != nil {
		panic(err)
	}

	var friendUserId uint32
	brick := toy.Model(&User{}).Debug().
		Preload(Offsetof(User{}.Detail)).Enter().
		Preload(Offsetof(User{}.Blog)).Enter()
	{
		brick := brick.Preload(Offsetof(User{}.Friends)).Enter()
		result, err = brick.DropTableIfExist()
		if err != nil {
			panic(err)
		}
		if err := result.Err(); err != nil {
			fmt.Printf("%s\n", err)
		}

		result, err = brick.CreateTableIfNotExist()
		if err != nil {
			panic(err)
		}
		if err := result.Err(); err != nil {
			fmt.Printf("%s\n", err)
		}
	}

	fmt.Println("insert")
	// insert test
	{
		brick := brick.Preload(Offsetof(User{}.Friends)).
			Preload(Offsetof(User{}.Detail)).Enter().
			Preload(Offsetof(User{}.Blog)).Enter().
			Enter()
		user := User{
			Detail: &UserDetail{
				MainPage: "some html code with you page",
				Extra:    Extra{"title": "my blog"},
			},
			Blog: []Blog{
				{Title: "how to write a blog", Content: "first ..."},
				{Title: "blog introduction", Content: "..."},
			},
			Friends: []*User{
				{
					Detail: &UserDetail{
						MainPage: "some html code with you page",
						Extra:    Extra{},
					},
					Blog: []Blog{
						{Title: "some python tech", Content: "first ..."},
						{Title: "my eleme_union_meal usage", Content: "..."},
					},
					Name: "fatpigeon",
					Age:  18,
					Sex:  "male",
				},
			},
			Name: "bigpigeon",
			Age:  18,
			Sex:  "male",
		}
		result, err = brick.Save(&user)
		if err != nil {
			panic(err)
		}
		if err := result.Err(); err != nil {
			fmt.Printf("%s\n", err)
		}
		fmt.Printf("report:\n%s\n", result.Report())
		fmt.Printf("user %v\n", JsonEncode(user))
		friendUserId = user.Friends[0].ID
	}

	fmt.Println("find one")
	// find one
	{
		brick := brick.Where(toyorm.ExprEqual, Offsetof(User{}.ID), friendUserId).
			Preload(Offsetof(User{}.Detail)).Enter().
			Preload(Offsetof(User{}.Blog)).Where(toyorm.ExprLike, Offsetof(Blog{}.Title), "%tech%").Enter().
			RightValuePreload(Offsetof(User{}.Friends)).
			Preload(Offsetof(User{}.Detail)).Enter().
			Preload(Offsetof(User{}.Blog)).Enter().
			Enter()
		var user User
		result, err = brick.Find(&user)
		if err != nil {
			panic(err)
		}
		if err := result.Err(); err != nil {
			fmt.Printf("%s\n", err)
		}
		fmt.Printf("report:\n%s\n", result.Report())
		fmt.Printf("user %v\n", JsonEncode(user))
	}

	fmt.Println("find")
	var deleteUsers []User
	// find
	{
		brick := brick.Preload(Offsetof(User{}.Friends)).
			Preload(Offsetof(User{}.Detail)).Enter().
			Preload(Offsetof(User{}.Blog)).Enter().
			Enter()
		var users []User
		result, err = brick.Find(&users)
		if err != nil {
			panic(err)
		}
		if err := result.Err(); err != nil {
			fmt.Printf("%s\n", err)
		}
		fmt.Printf("report:\n%s\n", result.Report())
		fmt.Printf("users %v\n", JsonEncode(users))
		deleteUsers = users
	}

	fmt.Println("error find")
	// report error with find
	{
		var users []struct {
			ID     uint32
			Age    bool
			Detail *UserDetail
			Blog   []Blog
		}
		result, err = brick.Find(&users)
		if err != nil {
			panic(err)
		}
		if err := result.Err(); err != nil {
			fmt.Printf("error:\n%s\n", err)
		}
	}

	fmt.Println("delete")
	// delete
	{
		brick := brick.Preload(Offsetof(User{}.Friends)).
			Preload(Offsetof(User{}.Detail)).Enter().
			Preload(Offsetof(User{}.Blog)).Enter().
			Enter()
		result, err = brick.Delete(&deleteUsers)
		if err != nil {
			panic(err)
		}
		if err := result.Err(); err != nil {
			fmt.Printf("%s\n", err)
		}
		fmt.Printf("report:\n%s\n", result.Report())
	}
}


```

## Collection

---

collection provide multiple database operation

### ToyCollection

ToyCollection is basic of the collection , it like Toy

```toyorm
toyCollection, err = toyorm.OpenCollection("sqlite3", []string{"", ""}...)
```

### CollectionBrick

CollectionBrick use to build grammar and operate the database, like ToyBrick

```toyorm
brick := toyCollection.Model(&User{})
```

#### Selector

Selector use to select database when Insert/Save


CollectionBrick has default selector **dbPrimaryKeySelector**


you can custom db selector


```toyorm
func idSelector(n int, keys ...interface{}) int {
	sum := 0
	for _, k := range keys {
		switch val := k.(type) {
		case int:
			sum += val
		case int32:
			sum += int(val)
		case uint:
			sum += int(val)
		case uint32:
			sum += int(val)
		default:
			panic("primary key type not match")
		}
	}
	return sum % n
}

brick = brick.Selector(idSelector)
```

#### id generator

in this mode,Field tag **auto_increment** was invalid


you need create id generator


```golang
// a context id generator
type IDGenerator map[*toyorm.Model]chan int

func (g IDGenerator) CollectionIDGenerate(ctx *toyorm.CollectionContext) error {
	if g[ctx.Brick.Model] == nil {
		idGenerate := make(chan int)
		go func() {
			current := 1
			for {
				// if have redis, use redis-cli
				idGenerate <- current
				current++
			}

		}()
		g[ctx.Brick.Model] = idGenerate
	}
	primaryKey := ctx.Brick.Model.GetOnePrimary()
	for _, record := range ctx.Result.Records.GetRecords() {
		if field := record.Field(primaryKey.Name()); field.IsValid() == false || toyorm.IsZero(field) {
			v := <-g[ctx.Brick.Model]
			record.SetField(primaryKey.Name(), reflect.ValueOf(v))
		}
	}
	return nil

}

// set id genetor to model context
toy.SetModelHandlers("Save", brick.Model, toyorm.CollectionHandlersChain{idGenerate.CollectionIDGenerate})
// set id genetor to all preload models context
for _, pBrick := range brick.MapPreloadBrick {
		toy.SetModelHandlers("Save", pBrick.Model, toyorm.CollectionHandlersChain{idGenerate.CollectionIDGenerate})
}
```


#### sql action

toy collection sql action is same as Toy


insert

```golang
users := []User{
    {Name: "Turing"},
    {Name: "Shannon"},
    {Name: "Ritchie"},
    {Name: "Jobs"},
}
result, err = userBrick.Insert(&users)
// error process

// view report log
fmt.Printf("report:\n%s", result.Report())
```

find

```golang
var jobs User
result, err = userBrick.Where(toyorm.ExprEqual, Offsetof(User{}.Name), "Jobs").Find(&jobs)
// error process

// view report log
fmt.Printf("report:\n%s", result.Report())
```

delete

```golang
result, err = userBrick.Delete(&jobs)
// error process

// view report log
fmt.Printf("delete report:\n%s\n", result.Report())
```

### collection example

```golang
package main

import (
	"encoding/json"
	"fmt"
	"github.com/bigpigeon/toyorm"
	. "unsafe"

	// when database is mysql
	//_ "github.com/go-sql-driver/mysql"
	// when database is sqlite3
	_ "github.com/mattn/go-sqlite3"
	"reflect"
	"time"
)

func JsonEncode(v interface{}) string {
	jsonData, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		panic(err)
	}
	return string(jsonData)
}

func idSelector(n int, keys ...interface{}) int {
	sum := 0
	for _, k := range keys {
		switch val := k.(type) {
		case int:
			sum += val
		case int32:
			sum += int(val)
		case uint:
			sum += int(val)
		case uint32:
			sum += int(val)
		default:
			panic("primary key type not match")
		}
	}
	return sum % n
}

func PostsEncode(p *Post, indent string) string {
	var s string
	s = indent + fmt.Sprintf("%s say: %s \n", p.User.Name, p.Title) + indent + "Text:" + p.Text
	s += "\nreply:"
	for _, relay := range p.Reply {
		s += "\n" + PostsEncode(&relay, indent+"  ")
	}
	return s
}

type User struct {
	toyorm.ModelDefault
	Name  string `toyorm:"unique index"`
	Posts []Post
}

type Post struct {
	ID        uint32    `toyorm:"primary key;auto_increment"`
	CreatedAt time.Time `toyorm:"NULL"`
	Title     string    `toyorm:"index"`
	Text      string
	UserID    uint32 `toyorm:"index"`
	User      *User

	PostID uint32
	Reply  []Post
}

type IDGenerator map[*toyorm.Model]chan int

func (g IDGenerator) CollectionIDGenerate(ctx *toyorm.CollectionContext) error {
	if g[ctx.Brick.Model] == nil {
		idGenerate := make(chan int)
		go func() {
			current := 1
			for {
				// if have redis, use redis-cli
				idGenerate <- current
				current++
			}

		}()
		g[ctx.Brick.Model] = idGenerate
	}
	primaryKey := ctx.Brick.Model.GetOnePrimary()
	for _, record := range ctx.Result.Records.GetRecords() {
		if field := record.Field(primaryKey.Name()); field.IsValid() == false || toyorm.IsZero(field) {
			v := <-g[ctx.Brick.Model]
			record.SetField(primaryKey.Name(), reflect.ValueOf(v))
		}
	}
	return nil

}

func main() {
	var err error
	var toy *toyorm.ToyCollection
	var result *toyorm.Result
	idGenerate := IDGenerator{}
	// when database is mysql, make sure your mysql have toyorm_example schema
	//toy, err = toyorm.OpenCollection("mysql", []string{
	//	"root:@tcp(localhost:3306)/toyorm1?charset=utf8&parseTime=True",
	//	"root:@tcp(localhost:3306)/toyorm2?charset=utf8&parseTime=True",
	//})
	// when database is sqlite3
	toy, err = toyorm.OpenCollection("sqlite3", []string{"", ""}...)

	userBrick := toy.Model(&User{})
	brick := toy.Model(&Post{}).
		Preload(Offsetof(Post{}.User)).Enter().
		Preload(Offsetof(Post{}.Reply)).Preload(Offsetof(Post{}.User)).Enter().Enter()

	// add id generate
	toy.SetModelHandlers("Save", brick.Model, toyorm.CollectionHandlersChain{idGenerate.CollectionIDGenerate})
	toy.SetModelHandlers("Insert", brick.Model, toyorm.CollectionHandlersChain{idGenerate.CollectionIDGenerate})
	for _, pBrick := range brick.MapPreloadBrick {
		toy.SetModelHandlers("Save", pBrick.Model, toyorm.CollectionHandlersChain{idGenerate.CollectionIDGenerate})
		toy.SetModelHandlers("Insert", pBrick.Model, toyorm.CollectionHandlersChain{idGenerate.CollectionIDGenerate})
	}

	result, err = brick.DropTableIfExist()
	if err != nil {
		panic(err)
	}
	if resultErr := result.Err(); resultErr != nil {
		fmt.Print(resultErr)
	}

	result, err = brick.CreateTableIfNotExist()
	if err != nil {
		panic(err)
	}
	if resultErr := result.Err(); resultErr != nil {
		fmt.Print(resultErr)
	}
	fmt.Printf("add users \n")
	users := []User{
		{Name: "Turing"},
		{Name: "Shannon"},
		{Name: "Ritchie"},
		{Name: "Jobs"},
	}
	result, err = userBrick.Insert(&users)
	if err != nil {
		panic(err)
	}
	if resultErr := result.Err(); resultErr != nil {
		fmt.Print(resultErr)
	}
	fmt.Printf("report:\n%s", result.Report())

	fmt.Printf("import posts \n")
	posts := []Post{
		{
			Title: "welcome",
			Text:  "",
			User:  &users[0],
			Reply: []Post{
				{
					Title: "reports",
					User:  &users[1],
				},
				{
					Title: "reports",
					User:  &users[2],
				},
				{
					Title: "reports",
					User:  &users[3],
				},
			},
		},
		{
			Title: "one",
			Text:  "",
			User:  &users[0],
			Reply: []Post{
				{
					Title: "two",
					User:  &users[1],
				},
				{
					Title: "three",
					User:  &users[2],
				},
				{
					Title: "four",
					User:  &users[3],
				},
			},
		},
		{
			Title: "reverse data",
			Text:  "",
			User:  &users[3],
			Reply: []Post{
				{
					Title: "reverse data",
					User:  &users[2],
				},
				{
					Title: "reverse data",
					User:  &users[1],
				},
				{
					Title: "reverse data",
					User:  &users[0],
				},
			},
		},
	}

	result, err = brick.Insert(&posts)
	if err != nil {
		panic(err)
	}
	if resultErr := result.Err(); resultErr != nil {
		fmt.Print(resultErr)
	}
	fmt.Printf("insert report:\n%s", result.Report())
	fmt.Println("posts")
	for _, p := range posts {
		fmt.Println(PostsEncode(&p, ""))
	}

	// find all top post
	fmt.Printf("find posts \n")
	var topPosts []Post
	result, err = brick.Where(toyorm.ExprEqual, Offsetof(Post{}.PostID), 0).Find(&topPosts)
	if err != nil {
		panic(err)
	}
	if resultErr := result.Err(); resultErr != nil {
		fmt.Print(resultErr)
	}
	fmt.Printf("find report:\n%s\n", result.Report())
	fmt.Println("posts")
	for _, p := range topPosts {
		fmt.Println(PostsEncode(&p, ""))
	}

	// delete the one's user and his post data
	userBrick = userBrick.Preload(Offsetof(User{}.Posts)).Enter()
	var jobs User
	result, err = userBrick.Where(toyorm.ExprEqual, Offsetof(User{}.Name), "Jobs").Find(&jobs)
	if err != nil {
		panic(err)
	}
	if resultErr := result.Err(); resultErr != nil {
		fmt.Print(resultErr)
	}
	fmt.Printf("delete report:\n%s\n", result.Report())

	result, err = userBrick.Delete(&jobs)
	if err != nil {
		panic(err)
	}
	if resultErr := result.Err(); resultErr != nil {
		fmt.Print(resultErr)
	}

	fmt.Printf("delete report:\n%s\n", result.Report())
	fmt.Printf("was delete user data:\n%s", JsonEncode(jobs))
	fmt.Printf("try to find posts again\n")
	{
		var topPosts []Post
		result, err = brick.Where(toyorm.ExprEqual, Offsetof(Post{}.PostID), 0).Find(&topPosts)
		if err != nil {
			panic(err)
		}
		if resultErr := result.Err(); resultErr != nil {
			fmt.Print(resultErr)
		}
		fmt.Printf("find report:\n%s\n", result.Report())
		fmt.Println("posts")
		for _, p := range topPosts {
			fmt.Println(PostsEncode(&p, ""))
		}
	}
}
```