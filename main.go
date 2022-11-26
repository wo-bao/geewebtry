package main

import (
	"fmt"
	"geeTry/gee"
	"html/template"
	"net/http"
	"time"
)

type student struct {
	Name string
	Age int8
}

func FormatAsDate(t time.Time) string {
	year, month, day := t.Date()
	return fmt.Sprintf("%d-%02d-%02d", year, month, day)
}

func main() {
	r := gee.Default()
	//r.Use(gee.Logger()) // global middleWare
	r.SetFuncMap(template.FuncMap{"FormatAsDate": FormatAsDate})
	r.LoadHTMLGlob("src/templates/*")
	r.Static("/assets", "src/geeTry/static")  // root需要绝对路径 D:/Programs/Go/projects/src/geeTry
	stu1 := &student{Name: "wobao", Age: 27}
	stu2 := &student{Name: "huanhuan", Age: 18}
	r.GET("/", func(c *gee.Context) {
		c.HTML(http.StatusOK, "css.tmpl", nil)
	})
	r.GET("/students", func(c *gee.Context) {
		c.HTML(http.StatusOK, "arr.tmpl", gee.H{
			"title" : "gee",
			"stuArr": [2]*student{stu1, stu2},
		})
	})
	r.GET("/date", func(c *gee.Context) {
		c.HTML(http.StatusOK, "custom_func.tmpl", gee.H{
			"title": "gee",
			"now": time.Date(2022, 6, 29,0,0,0,0, time.UTC),
		})
	})
	_ = r.Run(":9999")
}