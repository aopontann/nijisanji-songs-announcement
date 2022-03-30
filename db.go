package main

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

// Opening a database and save the reference to `Database` struct.
func DBInit() {
	var err error
	// DB接続
	DB, err = sql.Open("mysql", os.Getenv("DSN"))
	if err != nil {
		log.Fatal(err)
	}
	err = DB.Ping()
	if err != nil {
		log.Fatal(err)
	}
}
