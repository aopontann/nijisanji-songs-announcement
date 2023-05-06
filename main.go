package main

import (
	"os"

	"github.com/rs/zerolog/log"
)

func main() {
	// YouTube Data API 初期化
	YTNew()

	// DB接続初期化
	DBInit()
	defer DB.Close()

	taskNum := os.Getenv("CLOUD_RUN_TASK_INDEX")
	if taskNum == "0" {
		err := CheckNewVideoTask()
		if err != nil {
			log.Fatal().Str("severity", "ERROR").Msg(err.Error())
		}
	}
	if taskNum == "1" {
		err := TweetTask()
		if err != nil {
			log.Fatal().Str("severity", "ERROR").Msg(err.Error())
		}
	} else {
		err := CheckNewVideoTask()
		if err != nil {
			log.Fatal().Str("severity", "ERROR").Msg(err.Error())
		}
		err = TweetTask()
		if err != nil {
			log.Fatal().Str("severity", "ERROR").Msg(err.Error())
		}
	}
}
