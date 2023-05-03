package main

import (
	"database/sql"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/rs/zerolog/log"
)

var DB *sql.DB

func DBInit() {
	var err error
	DB, err = sql.Open("mysql", os.Getenv("DSN"))
	if err != nil {
		log.Fatal().Err(err).Msg("db connection failed")
	}

	err = DB.Ping()
	if err != nil {
		log.Fatal().Err(err).Msg("db ping failed")
	}
}
