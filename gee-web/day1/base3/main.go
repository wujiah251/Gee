package main

import (
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"

	"Gee/gee-web/day1/base3/gee"
)

func init() {
	logrus.SetLevel(logrus.TraceLevel)
}

func main() {
	r := gee.New()
	r.GET("/", func(w http.ResponseWriter, req *http.Request) {
		_, _ = fmt.Fprintf(w, "UR:.Path = %q\n", req.URL.Path)
	})
	r.GET("/hello", func(w http.ResponseWriter, req *http.Request) {
		for k, v := range req.Header {
			_, _ = fmt.Fprintf(w, "Header[%q] = %q\n", k, v)
		}
	})
	logrus.Fatal(r.Run(":9999"))
}
