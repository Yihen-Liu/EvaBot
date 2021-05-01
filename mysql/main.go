/**
 * Description:
 * Author: Yihen.Liu
 * Create: 2021-05-01
 */
package main

import (
	"fmt"

	"github.com/gofiber/template/html"

	"github.com/gofiber/fiber/v2"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	DB *gorm.DB
)

type Todo struct {
	ID     int    `json:"id"`
	Tiltle string `json:"title"`
	Status bool   `json:"status"`
}

func initMysql() (err error) {
	//初始化数据库连接
	dns := "root:123456@(localhost:3306)/todo?charset=utf8mb4&parseTime=True&loc=Local"
	DB, err = gorm.Open(mysql.Open(dns), &gorm.Config{})
	if err != nil {
		fmt.Println("mysql content faild, err:", err)
		return
	}
	return
}

func main() {
	err := initMysql()
	if err != nil {
		panic(err)
	}
	DB.AutoMigrate(&Todo{})
	/*
	   fiber默认使用html/template加载模板文件，可自定义使用其他模板引擎加载。
	   支持amber，handlebars，mustache，pug等等...
	*/
	engine := html.New("./templattes", ".html")
	app := fiber.New(fiber.Config{
		Views: engine,
	})
	//加载静态文件
	app.Static("/static", "./static")
	/*
	   生成首页
	   注意新版的fiber中要求匿名函数后必须使用一个error的返回值，
	   fiber框架下很多的函数都是默认定义了error返回值，
	   所以我们都可以直接return一个执行函数即可。
	*/
	app.Get("/", func(c *fiber.Ctx) error {
		//fiber中定义使用了类似Gin框架的gin.H{}做了一个fiber.Map{}，返回任意内容
		return c.Render("index", fiber.Map{
			"code":2000,
			"msg":"Todo list sussce!",
		})
	})
	// 注册一个路由组
	v1 := app.Group("/v1")
	// 添加一个todo
	v1.Post("/todo", func(c *fiber.Ctx) error {
		var todo Todo
		c.BodyParser(&todo)
		if err = DB.Create(&todo).Error; err != nil {
			return c.JSON(fiber.Map{
				"code": 2001,
				"msg":  "add a todo message faild",
			})
		} else {
			return c.JSON(todo)
		}

	})
	// 查看todo列表
	v1.Get("/todo", func(c *fiber.Ctx) error {
		var todolist []Todo
		if err = DB.Find(&todolist).Error; err != nil {
			return c.JSON(fiber.Map{
				"code": 2002,
				"msg":  "don't get todo list",
			})
		} else {
			return c.JSON(todolist)
		}

	})
	// 根据id修改todo
	v1.Put("/todo/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		var todo Todo
		if err = DB.Where("id=?", id).First(&todo).Error; err != nil {
			return c.JSON(fiber.Map{
				"code": 2003,
				"msg":  "don't search todo message by id ",
			})
		}
		c.BodyParser(&todo)
		if err = DB.Save(&todo).Error; err != nil {
			return c.JSON(fiber.Map{
				"code": 2004,
				"msg":  "don't update todo message by id",
			})
		} else {
			return c.JSON(todo)
		}
	})

	//根据id删除todo
	v1.Delete("/todo/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		if err = DB.Where("id=?", id).Delete(Todo{}).Error; err == nil {
			return c.JSON(fiber.Map{
				"code": 2000,
				"msg":  "delete todo massage success ",
			})
		}
		return err
	})
	app.Listen(":3000")
}
