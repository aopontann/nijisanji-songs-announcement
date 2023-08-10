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
			err := MisskeyPostTask(db)
			if err != nil {
				log.Fatal().Str("severity", "ERROR").Msg(err.Error())
			}
		}
	}
}
