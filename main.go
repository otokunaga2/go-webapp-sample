package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo"
	_ "github.com/mattn/go-sqlite3"
	"gopkg.in/go-playground/validator.v8"
	"gopkg.in/gorp.v1"
)

type Validator struct {
	validator *validator.Validate
}

func (v *Validator) Validate(i interface{}) error {
	return v.validator.Struct(i)
}

type Comment struct {
	Id      int64     `json:"id" db:"id,primarykey,autoincrement"`
	Name    string    `json:"nane" db:"name,notnull,default:'noname',size:200"`
	Text    string    `json:"text" db:"text,notnull,size:399"`
	Created time.Time `json:"created" db:"created,notnull"`
	Updated time.Time `json:"updated" db:"updated,notnull"`
}

func main() {
	fmt.Println("vim-go")
	e := echo.New()
	dbmap := initDb()
	defer dbmap.Db.Close()
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "hello world")
	})
	e.GET("/api/comments", func(c echo.Context) error {
		var comments []Comment
		_, err := dbmap.Select(&comments, "SELECT * FROM comments ORDER by created desc LIMIT 10")
		if err != nil {
			c.Logger().Error("Select: ", err)
			return c.String(http.StatusBadRequest, "Select: "+err.Error())
		}
		return c.JSON(http.StatusOK, comments)
	})
	e.POST("/api/comments", func(c echo.Context) error {
		var comment Comment
		if err := c.Bind(&comment); err != nil {
			c.Logger().Error("Bind", err)
			return c.String(http.StatusBadRequest, "Bind:"+err.Error())
		}
		if err := c.Validate(&comment); err != nil {
			c.Logger().Error("Validate: ", err)
			return c.String(http.StatusBadRequest, "Validate"+err.Error())

		}
		c.Logger().Info("Added: %v", comment.Id)
		return c.JSON(http.StatusCreated, "")

	})
	e.Static("/", "static/")
	e.Logger.Fatal(e.Start(":8080"))
}

func initDb() *gorp.DbMap {
	// connect to db using standard Go database/sql API
	// use whatever database/sql driver you wish
	db, err := sql.Open("sqlite3", "/tmp/post_db.bin")
	checkErr(err, "sql.Open failed")

	// construct a gorp DbMap
	dbmap := &gorp.DbMap{Db: db, Dialect: gorp.SqliteDialect{}}

	// add a table, setting the table name to 'posts' and
	// specifying that the Id property is an auto incrementing PK
	dbmap.AddTableWithName(Comment{}, "comments").SetKeys(true, "Id")

	// create the table. in a production system you'd generally
	// use a migration tool, or create the tables via scripts
	err = dbmap.CreateTablesIfNotExists()
	checkErr(err, "Create tables failed")

	return dbmap
}

func checkErr(err error, msg string) {
	if err != nil {
		log.Fatalln(msg, err)
	}
}
