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
			log.Error().Str("severity", "ERROR").Err(err).Send()
		}
	}
	if taskNum == "1" {
		err := TweetTask()
		if err != nil {
			log.Error().Str("severity", "ERROR").Err(err).Send()
		}
	} else {
		err := CheckNewVideoTask()
		if err != nil {
			log.Error().Str("severity", "ERROR").Err(err).Send()
		}
		err = TweetTask()
		if err != nil {
			log.Error().Str("severity", "ERROR").Err(err).Send()
		}
	}
}
