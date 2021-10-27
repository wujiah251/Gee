package main

import (
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
)

type Engine struct{}

func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch req.URL.Path {
	case "/":
		fmt.Fprintf(w, "URL.Path = %q\n", req.URL.Path)
	case "/hello":
		for k, v := range req.Header {
			fmt.Fprintf(w, "Header[%q] = %q\n", k, v)
		}
	default:
		fmt.Fprintf(w, "404 NOT FOUND: %s\n", req.URL)
	}
}

func init() {
	logrus.SetLevel(logrus.TraceLevel)
}

func main() {
	engine := new(Engine)
	logrus.Fatal(http.ListenAndServe(":9999", engine))
}
