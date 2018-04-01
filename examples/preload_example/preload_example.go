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
	// when database is postgres
	//_ "github.com/lib/pq"
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
	// when database is postgres
	//toy, err = toyorm.Open("postgres", "user=postgres dbname=toyorm sslmode=disable")

	var result *toyorm.Result
	if err != nil {
		panic(err)
	}

	var friendUserId uint32
	// bind model
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
