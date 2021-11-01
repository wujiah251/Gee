package main

import (
	"Gee/gee-web/day6/gee"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

type student struct {
	Name string
	Age  int8
}

func FormatAsDate(t time.Time) string {
	year, month, day := t.Date()
	return fmt.Sprintf("%d-%02d-%02d", year, month, day)
}

func main() {
	r := gee.New()
	r.Use(gee.Logger())
	r.SetFuncMap(template.FuncMap{
		"FormatAsDate": FormatAsDate,
	})
	r.LoadHTMLGlob("templates/*")
	r.Static("/assets", "./static")

	stu1 := &student{Name: "Geektutu", Age: 20}
	stu2 := &student{Name: "Jack", Age: 22}
	r.GET("/", func(c *gee.Context) {
		c.HTML(http.StatusOK, "css.tmpl", nil)
	})
	r.GET("/students", func(c *gee.Context) {
		c.HTML(http.StatusOK, "arr.tmpl", gee.H{
			"title":  "geecache",
			"stuArr": [2]*student{stu1, stu2},
		})
	})
	r.GET("/date", func(c *gee.Context) {
		c.HTML(http.StatusOK, "custom_func.tmpl", gee.H{
			"title": "geecache",
			"now":   time.Date(2019, 8, 17, 0, 0, 0, 0, time.UTC),
		})
	})
	logrus.Fatalln(r.Run(":9999"))
}

/*
(1) render array
$ curl http://localhost:9999/date
<html>
<body>
    <p>hello, geecache</p>
    <p>Date: 2019-08-17</p>
</body>
</html>
*/

/*
(2) custom render function
$ curl http://localhost:9999/students
<html>
<body>
    <p>hello, geecache</p>
    <p>0: Geektutu is 20 years old</p>
    <p>1: Jack is 22 years old</p>
</body>
</html>
*/

/*
(3) serve static files
$ curl http://localhost:9999/assets/css/geektutu.css
p {
    color: orange;
    font-weight: 700;
    font-size: 20px;
}
*/
