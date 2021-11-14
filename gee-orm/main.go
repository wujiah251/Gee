package main

import (
	mylog "Gee/gee-orm/log"

	_ "github.com/mattn/go-sqlite3"
)

func init() {
	mylog.SetLevel(mylog.InfoLevel)
}

func main() {
	mylog.Info("123123\n")
	mylog.Error("123123")
}
