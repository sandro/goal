package main

import (
	"net/http"
	"os"

	"github.com/sandro/go-sqlite-lite/sqx"
	"github.com/sandro/imigrate"
)

func main() {
	myDB := sqx.NewConn("file:db.sqlite3", false)
	fs := http.Dir("")
	migrator := imigrate.NewIMigrator(myDB, fs)
	imigrate.CLI(migrator)
	os.Exit(0)
}
