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
  - [Website Example](#website-example)
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
    - [Template](#template)
      - [Custom insert](#custom-insert)
      - [Custom find](#custom-find)
      - [Custom update](#custom-update)
      - [Placeholder](#placeholder)
    - [Thread safe](#thread-safe)
    - [Preload](#preload)
      - [Preload example](#preload-example)
      - [One to one](#one-to-one)
      - [Belong to](#belong-to)
      - [One to many](#one-to-many)
      - [Many to many](#many-to-many)
      - [Load preload](#load-preload)
    - [Join](#join)
      - [Join Example](#join-example)
      - [Model Define](#model-define)
      - [Join in Find](#join-in-find)
      - [Preload On Join](#preload-on-join)
  - [Result](#result)
    - [Selector](#selector)
- [Collection](#collection)
  - [ToyCollection](#toycollection)
  - [CollectionBrick](#collectionbrick)
    - [Selector](#selector-1)
    - [id generator](#id-generator)
    - [sql action](#sql-action)
  - [collection example](#collection-example)
- [toy-doctor](#toy-doctor)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

---



### Support database

[sqlite3](https://www.sqlite.org/)  [mysql](https://www.mysql.com/) [postgresql](https://www.postgresql.org/)


### Go version

version go-1.9

---

### A Simple Example

[here](examples/simple_example)

---

### Website Example

[here](examples/website)

---

### Database connection

import database driver

```golang
// if database is mysql
_ "github.com/go-sql-driver/mysql"
// if database is sqlite3
_ "github.com/mattn/go-sqlite3"
// when database is postgres
_ "github.com/lib/pq"
```

create a toy

```golang
// if database is mysql, make sure your mysql have toyorm_example schema
toy, err = toyorm.Open("mysql", "root:@tcp(localhost:3306)/toyorm_example?charset=utf8&parseTime=True")
// if database is sqlite3
toy,err = toyorm.Open("sqlite3", "toyorm_test.db")
// when database is postgres
toy, err = toyorm.Open("postgres", "user=postgres dbname=toyorm sslmode=disable")
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
Cas       |int        | (not support for sqlite3)save operation will failure when it's value modified by third party e.g in postgres:cas=1 insert xxx conflict(id) update cas = 2 where cas = 1

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
alias         | string                 | change field name with toyorm
join          | string                 | to select related field when call brick.Join
belong to     | string                 | to select related field when call brick.Preload with BelongTo container
one to one    | string                 | to select related field when call brick.Preload with OneToOne container
one to many   | string                 | to select related field when call brick.Preload with OneToMany container
default       | srring                 | default value in database

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

whereGroup add multiple condition with same expr

    brick.WhereGroup(<expr>, <group>)

conditions will copy conditions and clean old conditions

    brick.Conditions(<toyorm.Search>)

or & and condition will use or/and to link new condition when current condition is not nil

    brick.Or().Condition(<expr>, <Key>, [value])
    brick.Or().ConditionGroup(<expr>, <group>)
    brick.And().Condition(<expr>, <Key>, [value])

or & and conditions will use or/and to link new conditions

    brick.Or().Conditions(<toyorm.Search>)
    brick.And().Conditions(<toyorm.Search>)

##### SearchExpr

SearchExpr        |  to sql      | example
------------------|--------------|:----------------
ExprAnd           | AND          | brick.WhereGroup(ExprAnd, Product{Name:"food one", Count: 4}) // WHERE name = "food one" AND Count = 4
ExprOr            | OR           | brick.WhereGroup(ExprOr, Product{Name:"food one", Count: 4}) // WHERE name = "food one" OR Count = "4"
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

#### Template Field

sometimes the condition of sql is not a normal field, use TempField to wrapper normal field

```golang
brick = brick.Where("=", brick.TempField(Offsetof(Product{}.Name), "LOWER(%s)"), "name")
// WHERE LOWER(name) = ?
```

#### Transaction

---

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
// now field = []int(nil) will ignore when update
// but field = []int{} will update when update
// now field = map[int]int(nil) will ignore when update
// but field = map[int]int{} will update when update
```

In default

Operation | Mode       | affect
----------|------------|--------
Insert    | IgnoreNo   | brick.Insert(<struct>)
Replace   | IgnoreNo   | brick.Replace(<struct>)
Condition | IgnoreZero | brick.WhereGroup(ExprAnd/ExprOr, <struct>)
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


#### Template

use template exec to replace default exec


now template only support Insert/Save/Update/Find

##### Custom insert

```golang
data := Product{
    Name:  "bag",
    Price: 9999,
    Count: 2,
    Tag:   "container",
}
result, err := brick.Template("INSERT INTO $ModelName($Columns) Values($Values)").Insert(&data)
// INSERT INTO product(created_at,updated_at,deleted_at,name,price,count,tag) Values(?,?,?,?,?,?,?) args:["2018-04-01T17:05:48.927499+08:00","2018-04-01T17:05:48.927499+08:00",null,"bag",9999,2,"container"]
```

##### Custom save

data := Product{
    Name:  "bag",
    Price: 9999,
    Count: 2,
    Tag:   "container",
}
result, err := brick.Template("INSERT INTO $ModelName($Columns) Values($Values) ON DUPLICATE KEY UPDATE $Cas $UpdateValues").Save(&data)
// INSERT INTO product(id,created_at,updated_at,deleted_at,name,price,count,tag) Values(?,?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE  id = VALUES(id),updated_at = VALUES(updated_at),deleted_at = VALUES(deleted_at),name = VALUES(name),price = VALUES(price),count = VALUES(count),tag = VALUES(tag)  args:[9,"2019-04-01T00:15:13.57342+08:00","2019-04-01T00:15:13.574443+08:00",null,"bag",9988,2,"container"]
```


##### Custom find

```golang
var data Product
// if driver is mysql use "USE INDEX" replace "INDEXED BY"
result, err := brick.Template("SELECT $Columns FROM $ModelName INDEXED BY idx_product_name $Conditions").
    Where("=", Offsetof(Product{}.Name), "bag").Find(&data)
// SELECT id,created_at,updated_at,deleted_at,name,price,count,tag FROM product INDEXED BY idx_product_name  WHERE deleted_at IS NULL AND name = ? LIMIT 1  args:["bag"]
```


##### Custom update

set count = count + 2

```golang
result, err := brick.Template(fmt.Sprintf("UPDATE $ModelName SET $Values,$FN-Count = $0x%x + ? $Conditions", Offsetof(Product{}.Count)), 2).
	Where("=", Offsetof(Product{}.Name), "bag").Update(&Product{Price: 200})
// UPDATE product SET updated_at = ?,price = ?,count = count + ?  WHERE deleted_at IS NULL AND name = ?  args:["2018-04-01T17:50:35.205377+08:00",200,2,"bag"]
```

##### Placeholder

follower placeholder use in template example


two special placeholder

1. $FN- will convert struct field name to table field name e.g $FN-Name => name
2. $0x will convert struct field offset to table field name e.g $0x58 => Count

action \\  placeholder | $ModelName | $Columns    | $Values                | $Conditions
-----------------------|------------|-------------|------------------| -----------|
Find                   | product    | id,data,... | -               |  WHERE ... ORDER BY ... GROUP BY ... LIMIT ... OFFSET ...
Insert                 | product    | id,data,... | ?,?,...         | WHERE ... ORDER BY ... GROUP BY ... LIMIT ... OFFSET ...
Save                   | product    | id,data,... | ?,?,...           | WHERE ... ORDER BY ... GROUP BY ... LIMIT ... OFFSET ...
Update                 | product    | id,data,... | id = ?,data = ?,...    | WHERE ... ORDER BY ... GROUP BY ... LIMIT ... OFFSET ...

#### Thread safe

Thread safe if you comply with the following agreement

1. make sure **ToyBrick** object is read only, if you want to change it, create a new one

2. do not use **append** to change ToyBrick's slice data,use **make** and **copy** to clone new slice


#### Preload

---

preload need have relation field and container field


relations field is used to link the main record and sub record


container field is used to hold sub record

##### Preload example

[here](examples/preload_example)

##### One to one

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

##### Belong to

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

##### One to many

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

##### Many to many

many to many not need to specified the relation ship,it relation field at middle model

```
type User struct {
    toyorm.ModelDefault
    // container field
    Friends    []*User
}
```

##### Load preload

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

or use tag declaration

```golang
type UserDetail struct {
    ID       int    `toyorm:"primary key;auto_increment"`
    MainID   uint32 `toyorm:"index;one to one:Detail"` // declaration the container field
    MainPage string `toyorm:"type:Text"`
    Extra    Extra  `toyorm:"type:VARCHAR(1024)"`
}

type Blog struct {
    toyorm.ModelDefault
    MainID  uint32 `toyorm:"index;one to many:Blog"` // declaration the container field
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

// now custom relation field is MainID
brick = brick.Preload(Offsetof(User{}.Blog)).Enter()
brick = brick.Preload(Offsetof(User{}.Detail)).Enter()
```

#### Join

---

different association query is join query

##### Join Example

[here](examples/join_example)

##### Model Define

use join tag to association related field/ join tag value must same as container field name

```golang
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
	ColorJoin  Color  `toyorm:"alias:ColorDetail"`
	Comment    []Comment
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
```

##### Join in Find

Swap on Join just like Enter on Preload, can back to previous data

```golang
brick := toy.Model(&tab).Debug().
    Join(Offsetof(tab.Detail)).
    Join(Offsetof(detailTab.ColorJoin)).Swap().Swap()
var scanData []Product
result, err = brick.Find(&scanData)
// SELECT m.id,m.created_at,m.deleted_at,m.name,m.count,m.price,m_0.product_id,m_0.title,m_0.custom_page,m_0.extra,m_0.color,m_0_0.name,m_0_0.code FROM `product` as `m` JOIN `product_detail` AS `m_0` ON m.id = m_0.product_id JOIN `color` AS `m_0_0` ON m_0.color = m_0_0.name   WHERE m.deleted_at IS NULL
```

use Join to switch current model and add conditions

```golang
// where Product.Name = "clean stick" or Color.Name = "black"
brick := toy.Model(&tab).Debug().Where("=", Offsetof(tab.Name), "clean stick").
    Join(Offsetof(tab.Detail)).
    Join(Offsetof(detailTab.ColorJoin)).Or().Condition("=", Offsetof(colorTab.Name), "black").
    Swap().Swap()
var scanData []Product
result, err = brick.Find(&scanData)
// SELECT m.id,m.created_at,m.deleted_at,m.name,m.count,m.price,m_0.product_id,m_0.title,m_0.custom_page,m_0.extra,m_0.color,m_0_0.name,m_0_0.code FROM `product` as `m` JOIN `product_detail` AS `m_0` ON m.id = m_0.product_id JOIN `color` AS `m_0_0` ON m_0.color = m_0_0.name   WHERE m.deleted_at IS NULL AND (m.name = ? OR m_0_0.name = ?)  args:["clean stick","black"]
```

set order just like add conditions

```golang
// where Product.Name = "clean stick" or Color.Name = "black"
brick := toy.Model(&tab).Debug().
    Join(Offsetof(tab.Detail)).
    Join(Offsetof(detailTab.ColorJoin)).OrderBy(Offsetof(colorTab.Name)).
    Swap().Swap()
var scanData []Product
result, err = brick.Find(&scanData)
// SELECT m.id,m.created_at,m.deleted_at,m.name,m.count,m.price,m_0.product_id,m_0.title,m_0.custom_page,m_0.extra,m_0.color,m_0_0.name,m_0_0.code FROM `product` as `m` JOIN `product_detail` AS `m_0` ON m.id = m_0.product_id JOIN `color` AS `m_0_0` ON m_0.color = m_0_0.name   WHERE m.deleted_at IS NULL ORDER BY m_0_0.name
```

also can set GroupBy but here not example

##### Preload On Join

Preload method also work on Join mode

```golang
brick := toy.Model(&tab).Debug().
    Join(Offsetof(tab.Detail)).Preload(Offsetof(detailTab.Comment)).Enter().
    Join(Offsetof(detailTab.ColorJoin)).Swap().Swap()
var scanData []Product
result, err = brick.Find(&scanData)
// SELECT m.id,m.created_at,m.deleted_at,m.name,m.count,m.price,m_0.product_id,m_0.title,m_0.custom_page,m_0.extra,m_0.color,m_0_0.name,m_0_0.code FROM `product` as `m` JOIN `product_detail` AS `m_0` ON m.id = m_0.product_id JOIN `color` AS `m_0_0` ON m_0.color = m_0_0.name   WHERE m.deleted_at IS NULL
// SELECT id,created_at,updated_at,deleted_at,product_detail_product_id,data FROM `comment`   WHERE deleted_at IS NULL AND product_detail_product_id IN (?,?,?)  args:[1,2,3]
```

### Custom Table Name

custom your table name with different platform

```golang
type User struct {
    ID uint32 `toyorm:"primary key"`
    Name string `toyorm:"index"`
    Platform string `toyorm:"-"`
}
func (u *User) TableName() string {
    return "user_" + u.Platform
}

brick := toy.Model(&User{Platform:"p1"}).Debug()
brick.CreateTable()
// CREATE TABLE user_p1 (id BIGINT AUTO_INCREMENT,name VARCHAR(255), PRIMARY KEY(id))
brick := toy.Model(&User{Platform:"p2"}).Debug()
brick.CreateTable()
// CREATE TABLE user_p2 (id BIGINT AUTO_INCREMENT,name VARCHAR(255), PRIMARY KEY(id))
```

table name method also work on Preload and Join

```golang
type UserDetail struct {
    ID uint32 `toyorm:"primary key"`
    UserID uint32
    Data string
}
func (u *UserDetail) TableName() string {
    return "user_detail_" + u.Platform
}

type User struct {
    ID uint32 `toyorm:"primary key"`
    Name string `toyorm:"index"`
    Detail *UserDetail
    Platform string `toyorm:"-"`
}
func (u *User) TableName() string {
    return "user_" + u.Platform
}

brick := toy.Model(&User{Platform:"p1", Detail:&UserDetail{Platform:"p1"}}).Debug().
    Preload(Offsetof(User{}.UserDetail)).Enter()
brick.CreateTable()
// CREATE TABLE user_p1 (id BIGINT AUTO_INCREMENT,name VARCHAR(255), PRIMARY KEY(id))
// CREATE TABLE user_detail_p1 (id BIGINT AUTO_INCREMENT, user_id BIGINT, data VARCHAR(255), PRIMARY KEY(id))
```


in one to many or many to many mode , need to set it's first element val

```golang
type Address struct {
    ID uint32 `toyorm:"primary key"`
    UserID uint32
    Addr string
    Platform string `toyorm:"-"`
}
func (a *Address) TableName() string {
    return "address_" + a.Platform
}
type User struct {
    ID uint32 `toyorm:"primary key"`
    Name string `toyorm:"index"`
    Addresses []Address
    Platform string `toyorm:"-"`
}
func (u *User) TableName() string {
    return "user_" + u.Platform
}

brick := toy.Model(&User{Platform:"p1", Addresses:[]Address{{Platform:"p1"}}}).Debug().
    Preload(Offsetof(User{}.Addresses)).Enter()
brick.CreateTable()
// CREATE TABLE user_p1 (id BIGINT AUTO_INCREMENT,name VARCHAR(255), PRIMARY KEY(id))
// CREATE TABLE address_p1 (id BIGINT AUTO_INCREMENT, user_id BIGINT, addr VARCHAR(255), PRIMARY KEY(id))
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
--------------------|----------|-----------------|--------------------------------|--------------------------|--------|
Update              | no       | no              | yes                            | yes                      | yes
Insert              | no       | no              | yes                            | yes                      | yes
Save                | no       | no              | yes                            | yes                      | yes
Where & Conditions  | yes      | yes             | no                             | no                       | no
WhereGroup & ConditionGroup | no |  no           | yes                            | yes                      | yes
BindFields          | yes      | yes             | no                             | no                       | no
Preload & Custom Preload | yes | yes             | no                             | no                       | no
OrderBy             | yes      | yes             | no                             | no                       | no
Find                | no       | no              | no                             | no                       | yes





## Collection


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

[here](examples/collection_example)

## toy-doctor

parameter check within ToyBrick method call

[toy-doctor](https://github.com/bigpigeon/toy-doctor)