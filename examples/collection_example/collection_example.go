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
