package main

import (
	"database/sql"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/rs/zerolog/log"
	nsa "github.com/aopontann/nijisanji-songs-announcement"
)

func main() {

	db, err := sql.Open("mysql", os.Getenv("DSN"))
	if err != nil {
		log.Fatal().Str("severity", "ERROR").Str("service", "sql.Open").Msg(err.Error())
	}

	err = nsa.CheckNewVideoTask(db)
	if err != nil {
		log.Fatal().Str("severity", "ERROR").Msg(err.Error())
	}
	err = nsa.MisskeyPostTask(db)
	if err != nil {
		log.Fatal().Str("severity", "ERROR").Msg(err.Error())
	}
	err = nsa.TweetTask(db)
	if err != nil {
		log.Fatal().Str("severity", "ERROR").Msg(err.Error())
	}
}
