# Introduction

this is powerful sql orm library for Golang, it has few dependencies and high performance

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

func main() {
	var err error
	var toy *toyorm.Toy
	// when database is mysql, make sure your mysql have toyorm_example schema
	toy, err = toyorm.Open("mysql", "root:@tcp(localhost:3306)/toyorm_example?charset=utf8&parseTime=True")
	// when database is sqlite3
	//toy,err = toyorm.Open("sqlite3", "toyorm_test.db")

	brick := toy.Model(&Product{}).Debug()
	_, err = brick.DropTableIfExist()
	if err != nil {
		panic(err)
	}
	_, err = brick.CreateTableIfNotExist()
	if err != nil {
		panic(err)
	}
	// Insert will set id to source data when primary key is auto_increment
	brick.Insert(&[]Product{
		{Name: "food one", Price: 1, Count: 4, Tag: "food"},
		{Name: "food two", Price: 2, Count: 3, Tag: "food"},
		{Name: "food three", Price: 3, Count: 2, Tag: "food"},
		{Name: "food four", Price: 4, Count: 1, Tag: "food"},
		{Name: "toolkit one", Price: 1, Count: 8, Tag: "toolkit"},
		{Name: "toolkit two", Price: 2, Count: 6, Tag: "toolkit"},
		{Name: "toolkit one", Price: 3, Count: 4, Tag: "toolkit"},
		{Name: "toolkit two", Price: 4, Count: 2, Tag: "toolkit"},
	})
	// find the first food
	{
		var product Product
		_, err := brick.Where(toyorm.ExprEqual, Offsetof(Product{}.Tag), "food").Find(&product)
		if err != nil {
			panic(err)
		}
		fmt.Printf("food %s\n", JsonEncode(product))
	}

	// find all food
	{
		var products []Product
		_, err := brick.Where(toyorm.ExprEqual, Offsetof(Product{}.Tag), "food").Find(&products)
		if err != nil {
			panic(err)
		}
		fmt.Printf("foods %s\n", JsonEncode(products))
	}

	// find count = 2
	{
		var products []Product
		_, err := brick.Where(toyorm.ExprEqual, Offsetof(Product{}.Count), 2).Find(&products)
		if err != nil {
			panic(err)
		}
		fmt.Printf("count > 2 products %s\n", JsonEncode(products))
	}
	// find count = 2 and price > 3
	{
		var products []Product
		_, err := brick.Where(toyorm.ExprEqual, Offsetof(Product{}.Count), 2).And().
			Condition(toyorm.ExprGreater, Offsetof(Product{}.Price), 3).Find(&products)
		if err != nil {
			panic(err)
		}
		fmt.Printf("count = 2 and price > 3 products %s\n", JsonEncode(products))
	}
	// find count = 2 and price > 3 or count = 4
	{
		var products []Product
		_, err := brick.Where(toyorm.ExprEqual, Offsetof(Product{}.Count), 2).And().
			Condition(toyorm.ExprGreater, Offsetof(Product{}.Price), 3).Or().
			Condition(toyorm.ExprEqual, Offsetof(Product{}.Count), 4).Find(&products)
		if err != nil {
			panic(err)
		}
		fmt.Printf("count = 2 and price > 3 or count = 4  products %s\n", JsonEncode(products))
	}
	// find price > 3 and (count = 2 or count = 1)
	{
		var products []Product
		_, err := brick.Where(toyorm.ExprGreater, Offsetof(Product{}.Price), 3).And().Conditions(
			brick.Where(toyorm.ExprEqual, Offsetof(Product{}.Count), 2).Or().
				Condition(toyorm.ExprEqual, Offsetof(Product{}.Count), 1).Search,
		).Find(&products)
		if err != nil {
			panic(err)
		}
		fmt.Printf("price > 3 and (count = 2 or count = 1)  products %s\n", JsonEncode(products))
	}
	// update to count = 4
	{
		_, err := brick.Update(&Product{
			Count: 4,
		})
		if err != nil {
			panic(err)
		}
		var Counters []struct {
			Name  string
			Count int
		}
		_, err = brick.Find(&Counters)
		if err != nil {
			panic(err)
		}
		for _, counter := range Counters {
			fmt.Printf("product name %s, count %d\n", counter.Name, counter.Count)
		}
	}
	// delete with element
	{
		// make a transaction, because I do not really delete a data
		brick := brick.Begin()
		var product Product
		_, err := brick.Where(toyorm.ExprEqual, Offsetof(Product{}.Name), "food four").Find(&product)
		if err != nil {
			panic(err)
		}

		_, err = brick.Delete(&product)
		if err != nil {
			panic(err)
		}

		var disappearProduct Product
		_, err = brick.Where(toyorm.ExprEqual, Offsetof(Product{}.Name), "food four").Find(&disappearProduct)
		fmt.Printf("when I delete name = food four product , I try to find it will have error(%s)\n", err)
		brick.Rollback()
	}
	// delete with condition
	{
		// make a transaction, because I am not really delete a data
		brick := brick.Begin()
		_, err := brick.Where(toyorm.ExprEqual, Offsetof(Product{}.Name), "food four").DeleteWithConditions()
		if err != nil {
			panic(err)
		}
		var product Product
		_, err = brick.Where(toyorm.ExprEqual, Offsetof(Product{}.Name), "food four").Find(&product)
		fmt.Printf("when I delete name = food four product , I try to find it will have error(%s)\n", err)
		brick.Rollback()
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

some model definition example

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

**special fields**

1. all special fields have some process in handlers, do not try to change it type or set it value


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

other custom TAG will append to end of CREATE TABLE field

### bind models

1. model kind must be a struct or a point with struct

2. the model is what information that toyorm know about the table

```golang
toy.Model(&User{})
// or
toy.Model(User{})
```


### ToyBrick

-----

use **toy.Model** will create a ToyBrick, you need use it to build grammar and operate the database

#### Where condition

where will clean old conditions and make new condition

    brick.Where(<expr>, <Key>, [value])

conditions will copy conditions and clean old conditions

    brick.Conditions(<toyorm.Search>)

or & and condition will use or/and to link new condition when current condition is not nil

    brick.Or().Condition(<expr>, <Key>, [value])
    brick.And().Condition(<expr>, <Key>, [value])

or & and conditions will use or/and to link new conditions

    brick.Or().Conditions(<toyorm.Search>)
    brick.And().Conditions(<toyorm.Search>)

**SearchExpr**

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

**some example**

single condition

```golang
brick = brick.Where(toyorm.ExprEqual, Offsetof(Product{}.Tag), "food")
// WHERE tag = "food"
```

combination condition

```golang
brick = brick.Where(toyorm.ExprEqual, Offsetof(Product{}.Count), 2).And().
    Condition(toyorm.ExprGreater, Offsetof(Product{}.Price), 3).Or().
    Condition(toyorm.ExprEqual, Offsetof(Product{}.Count), 4)
// WHERE count = 2 and price > 3 or count = 4
```

priority condition

```golang
brick.Where(toyorm.ExprGreater, Offsetof(Product{}.Price), 3).And().Conditions(
    brick.Where(toyorm.ExprEqual, Offsetof(Product{}.Count), 2).Or().
    Condition(toyorm.ExprEqual, Offsetof(Product{}.Count), 1).Search
)
// WHERE price > 3 and (count = 2 or count = 1)
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



#### Thread safe

Thread safe if you comply with the following agreement

1. all **ToyBrick** object is read only, if you want to change it please create a new one

2. do not use **append** to change ToyBrick's slice data,use **make** and **copy** to clone new slice


#### Create table

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

### Drop Table

```golang
var err error
_, err =toy.Model(&User{}).Debug().DropTable()
// DROP TABLE user
_, err =toy.Model(&UserDetail{}).Debug().DropTable()
// DROP TABLE user_detail
_, err =toy.Model(&Blog{}).Debug().DropTable()
// DROP TABLE blog
```

### Insert/Save Data

// insert with autoincrement will set id to source data

```golang
user := &User{
    Name: "bigpigeon",
    Age:  18,
    Sex:  "male",
}
_, err = toy.Model(&User{}).Debug().Insert(&user)
// INSERT INTO user(created_at,updated_at,name,age,sex) VALUES(?,?,?,?,?) , args:[]interface {}{time.Time{wall:0xbe8d7d8a2b9dbdb0, ext:284219817, loc:(*time.Location)(0x141cf80)}, time.Time{wall:0xbe8d7d8a2b9dd138, ext:284224274, loc:(*time.Location)(0x141cf80)}, "fatpigeon", 18, "male"}
// print user format with json
/* user {
    "ID": 1,
    "CreatedAt": "2018-01-10T10:47:04.722534+08:00",
    "UpdatedAt": "2018-01-10T10:47:04.722543+08:00",
    "DeletedAt": null,
    "Name": "bigpigeon",
    "Age": 18,
    "Sex": "male",
    ...
}*/
```

// save data use "REPLACE INTO" when primary key exist

```golang
user := &User{
    ModelDefault: toyorm.ModelDefault{ID: 1},
    Name: "bigpigeon",
    Age:  18,
    Sex:  "male",
}
_, err =toy.Model(&User{}).Debug().Save(&user)
// SELECT id,created_at FROM user WHERE id IN (?), args:[]interface {}{0x1}
// REPLACE INTO user(id,created_at,updated_at,name,age,sex) VALUES(?,?,?,?,?,?) , args:[]interface {}{0x1, time.Time{wall:0xbe8d861c14f91460, ext:329878361, loc:(*time.Location)(0x141cf80)}, time.Time{wall:0xbe8d861c150009a0, ext:330334354, loc:(*time.Location)(0x141cf80)}, "bigpigeon", 18, "male"}

```

### Full Feature Example

TODO