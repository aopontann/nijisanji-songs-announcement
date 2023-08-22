package main

import (
	"database/sql"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/rs/zerolog/log"
)

func main() {

	db, err := sql.Open("mysql", os.Getenv("DSN"))
	if err != nil {
		log.Fatal().Str("severity", "ERROR").Str("service", "sql.Open").Msg(err.Error())
	}

	err = CheckNewVideoTask(db)
	if err != nil {
		log.Fatal().Str("severity", "ERROR").Msg(err.Error())
	}
	err = MisskeyPostTask(db)
	if err != nil {
		log.Fatal().Str("severity", "ERROR").Msg(err.Error())
	}
	err = TweetTask(db)
	if err != nil {
		log.Fatal().Str("severity", "ERROR").Msg(err.Error())
	}
}
