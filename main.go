package main

import (
	"database/sql"
	"os"

	"github.com/rs/zerolog/log"
	_ "github.com/go-sql-driver/mysql"
)

func main() {

	db, err := sql.Open("mysql", os.Getenv("DSN"))
	log.Debug().Str("severity", "DEBUG").Str("test",os.Getenv("DSN")[110:]).Send()
	if err != nil {
		log.Fatal().Str("severity", "ERROR").Str("service", "sql.Open").Msg(err.Error())
	}

	taskNum := os.Getenv("CLOUD_RUN_TASK_INDEX")
	if taskNum == "0" {
		err := CheckNewVideoTask(db)
		if err != nil {
			log.Fatal().Str("severity", "ERROR").Msg(err.Error())
		}
	}
	if taskNum == "1" {
		// 開発環境ではツイートを行わない
		if os.Getenv("ENV") != "dev" {
			err := TweetTask(db)
			if err != nil {
				log.Fatal().Str("severity", "ERROR").Msg(err.Error())
			}
		}
	} else {
		err := CheckNewVideoTask(db)
		if err != nil {
			log.Fatal().Str("severity", "ERROR").Msg(err.Error())
		}
		err = TweetTask(db)
		if err != nil {
			log.Fatal().Str("severity", "ERROR").Msg(err.Error())
		}
	}
}
