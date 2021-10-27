package main

import (
	"Gee/gee-web/day2/gee"
	"log"
	"net/http"
)

func init() {

}

func main() {
	r := gee.New()
	r.GET("/", func(c *gee.Context) {
		log.Printf("%+v", c.Req)
		c.HTML(http.StatusOK, "<h1>Hello Gee</h1>")
	})
	r.GET("/hello", func(c *gee.Context) {
		log.Printf("%+v", c.Req)
		c.String(http.StatusOK, "hello %s, you're at %s\n", c.Query("name"), c.Path)
	})
	r.POST("/login", func(c *gee.Context) {
		log.Printf("%+v", c.Req)
		c.JSON(http.StatusOK, gee.H{
			"username": c.PostForm("username"),
			"password": c.PostForm("password"),
		})
	})
	r.Run(":9999")
}
