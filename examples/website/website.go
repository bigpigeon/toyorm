package main

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"github.com/bigpigeon/toyorm"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	. "unsafe"

	// when database is mysql
	_ "github.com/go-sql-driver/mysql"
	// when database is sqlite3
	//_ "github.com/mattn/go-sqlite3"
	"database/sql"
	"errors"
	"github.com/google/uuid"
)

type PostUser struct {
	Name     string   `toyorm:"unique index" json:"name" form:"name"`
	Password Password `json:"password" form:"password" toyorm:"type:VARCHAR(40)"`
}

func (p *PostUser) HashPassword(password string) {
	bytePass := sha1.Sum([]byte(password))

	p.Password = Password(base64.StdEncoding.EncodeToString(bytePass[:]))
}

type User struct {
	toyorm.ModelDefault
	PostUser
	Orders []Order
}

type PostOrder struct {
	Name string `form:"name" json:"name"`
	Num  int    `form:"num" json:"num"`
}

type Order struct {
	toyorm.ModelDefault
	UserID uint32 `toyorm:"index"`
	PostOrder
	User *User
}

type Engine struct {
	Toy *toyorm.Toy
	// use to cache sign in status
	Session map[string]uint32
}

func (e Engine) MainPage(ctx *gin.Context) {
	brick := e.Toy.Model(&User{}).Limit(10).Preload(Offsetof(User{}.Orders)).Enter()
	var users []User
	result, err := brick.Find(&users)
	fmt.Printf("find all users:\n")
	if ResultProcess(result, err, ctx) == false {
		return
	}
	ctx.HTML(http.StatusOK, "index.tmpl", gin.H{
		"current": ctx.Keys["current"],
		"users":   users,
	})

}

func (e Engine) LoginPage(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "login.tmpl", gin.H{
		"current": ctx.Keys["current"],
	})
}

func (e Engine) RegisterPage(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "register.tmpl", gin.H{
		"current": ctx.Keys["current"],
	})
}

func (e Engine) UserPage(ctx *gin.Context) {
	userId := ctx.Param("id")
	brick := e.Toy.Model(&User{}).Where("=", Offsetof(User{}.ID), userId).
		Preload(Offsetof(User{}.Orders)).Limit(10).Enter()
	var user User
	result, err := brick.Find(&user)
	fmt.Printf("find user\n")
	if ResultProcess(result, err, ctx) {
		ctx.HTML(http.StatusOK, "user.tmpl", gin.H{
			"user":    user,
			"current": ctx.Keys["current"],
		})
	}
}

func (e Engine) UserEditPage(ctx *gin.Context) {
	if ctx.Keys["current"] == nil {
		ctx.AbortWithError(400, errors.New("current session is nil"))
		return
	}
	brick := e.Toy.Model(&User{}).Where("=", Offsetof(User{}.ID), ctx.Keys["current"].(*User).ID).
		Preload(Offsetof(User{}.Orders)).Limit(10).Enter()
	var user User
	result, err := brick.Find(&user)
	if ResultProcess(result, err, ctx) {
		ctx.HTML(http.StatusOK, "user_edit.tmpl", gin.H{
			"current": &user,
		})
	} else {
		ctx.AbortWithError(400, errors.New("user not found"))
	}
}

func (e Engine) Logout(ctx *gin.Context) {
	session, err := ctx.Cookie("session")
	if err == nil {
		delete(e.Session, session)
	}
	ctx.SetCookie("session", "", -1, "", "", false, false)
	ctx.HTML(http.StatusOK, "redirect.tmpl", gin.H{
		"redirect": "/",
	})
}

func (e Engine) PostNewUser(ctx *gin.Context) {
	var postUser PostUser
	if err := ctx.ShouldBind(&postUser); err == nil {
		user := User{
			PostUser: postUser,
		}
		user.HashPassword(string(user.Password))
		result, err := e.Toy.Model(&User{}).Insert(&user)
		fmt.Printf("create new user:\n")
		if ResultProcess(result, err, ctx) {
			session, err := uuid.NewRandom()
			if err != nil {
				ctx.AbortWithError(500, err)
			}
			e.Session[session.String()] = user.ID
			ctx.SetCookie("session", session.String(), 0, "", "", false, false)

			ctx.HTML(http.StatusOK, "redirect.tmpl", gin.H{
				"redirect": "/",
			})
		}
	} else {
		ctx.AbortWithError(500, err)
	}
}

func (e Engine) UpdateUser(ctx *gin.Context) {
	type UpdateUser struct {
		ID   uint32 `json:"id" form:"id"`
		Name string `json:"name" form:"name"`
	}
	var user UpdateUser
	current := ctx.Keys["current"].(*User)
	if current != nil {
		if err := ctx.ShouldBind(&user); err == nil && current.ID == user.ID {
			current.Name = user.Name
			result, err := e.Toy.Model(&User{}).Save(current)
			fmt.Printf("update user:\n")
			if ResultProcess(result, err, ctx) {
				ctx.HTML(http.StatusOK, "redirect.tmpl", gin.H{
					"redirect": fmt.Sprintf("/edit/user/%d", user.ID),
				})
			}
		} else {
			ctx.AbortWithError(500, err)
		}
	} else {
		ctx.AbortWithError(400, errors.New("session is nil or id not match"))
	}
}

func (e Engine) GetUserWithSession(ctx *gin.Context) {
	ctx.Keys = map[string]interface{}{}
	session, err := ctx.Cookie("session")
	if err == nil && e.Session[session] != 0 {
		var current User
		brick := e.Toy.Model(&User{})
		result, err := brick.Where("=", Offsetof(User{}.ID), e.Session[session]).Find(&current)
		fmt.Printf("find current user:\n")
		if ResultProcess(result, err, ctx) {
			ctx.Keys["current"] = &current
		} else {
			ctx.Abort()
		}
	} else {
		ctx.SetCookie("session", "", -1, "", "", false, false)
	}

}

func (e Engine) PostOrder(ctx *gin.Context) {
	if current := ctx.Keys["current"].(*User); current != nil {
		var postOrder PostOrder
		if err := ctx.ShouldBind(&postOrder); err == nil {
			order := Order{
				PostOrder: postOrder,
				UserID:    current.ID,
			}
			result, err := e.Toy.Model(&Order{}).Insert(&order)
			fmt.Printf("post order user:\n")
			if ResultProcess(result, err, ctx) {
				ctx.HTML(http.StatusOK, "redirect.tmpl", gin.H{
					"redirect": fmt.Sprintf("/edit/user/%d", current.ID),
				})
			}

		} else {
			ctx.AbortWithError(400, err)
		}
	} else {
		ctx.AbortWithError(400, errors.New("session invalid"))
	}

}

func (e Engine) PostSession(ctx *gin.Context) {
	var postUser PostUser
	if err := ctx.ShouldBind(&postUser); err == nil {
		user := User{
			PostUser: postUser,
		}
		user.HashPassword(string(user.Password))
		var findData User
		result, err := e.Toy.Model(&User{}).
			Where("=", Offsetof(User{}.Name), user.Name).And().
			Condition("=", Offsetof(User{}.Password), user.Password).
			Find(&findData)
		fmt.Printf("check user password and name user:\n")
		if err == sql.ErrNoRows {
			ctx.AbortWithError(400, err)
			return
		}
		if ResultProcess(result, err, ctx) {
			session, err := uuid.NewRandom()
			if err != nil {
				ctx.AbortWithError(500, err)
			}
			e.Session[session.String()] = findData.ID
			ctx.SetCookie("session", session.String(), 0, "", "", false, false)
			ctx.HTML(http.StatusOK, "redirect.tmpl", gin.H{
				"redirect": "/",
			})
		}
	} else {
		ctx.AbortWithError(500, err)
	}
}

func createTableAndFillData(e Engine) {
	userBrick := e.Toy.Model(&User{}).Preload(Offsetof(User{}.Orders)).Enter()
	result, err := userBrick.DropTableIfExist()
	if err != nil {
		panic(err)
	}
	if err := result.Err(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Printf("report: \n%s\n", result.Report())
	result, err = userBrick.CreateTable()
	if err != nil {
		panic(err)
	}
	if err := result.Err(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Printf("report: \n%s\n", result.Report())

	// fill admin user
	admin := User{
		PostUser: PostUser{
			Name: "admin",
		},
		Orders: []Order{
			{PostOrder: PostOrder{Name: "coffee", Num: 1}},
		},
	}
	admin.HashPassword("abcd1234")
	result, err = userBrick.Insert(&admin)
	if err != nil {
		panic(err)
	}
	if err := result.Err(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Printf("report: \n%s\n", result.Report())

}

func main() {
	// when database is mysql, make sure your mysql have toyorm_example schema
	toy, err := toyorm.Open("mysql", "root:@tcp(localhost:3306)/toyorm_example?charset=utf8&parseTime=True")
	// when database is sqlite3
	//toy, err := toyorm.Open("sqlite3", ":memory")
	if err != nil {
		panic(err)
	}

	e := Engine{toy, map[string]uint32{}}
	createTableAndFillData(e)

	// init gin engine
	router := gin.Default()
	router.LoadHTMLGlob("templates/*")
	group := router.Group("/", e.GetUserWithSession)

	// init page url
	group.GET("/", e.MainPage)
	group.GET("/login", e.LoginPage)
	group.GET("/register", e.RegisterPage)
	group.GET("/user/:id", e.UserPage)
	group.GET("/edit/user/:id", e.UserEditPage)

	// init api
	group.GET("/logout", e.Logout)
	group.POST("/user", e.PostNewUser)
	group.POST("/order", e.PostOrder)
	group.POST("/update/user", e.UpdateUser)
	router.POST("/session", e.PostSession)

	// start server
	router.Run(":8080")
}
