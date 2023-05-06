package main

import (
	"os"
)

func main() {
	// YouTube Data API 初期化
	YTNew()

	// DB接続初期化
	DBInit()
	defer DB.Close()

	taskNum := os.Getenv("CLOUD_RUN_TASK_INDEX")
	if taskNum == "0" {
		CheckNewVideoTask()
	}
	if taskNum == "1" {
		TweetTask()
	} else {
		CheckNewVideoTask()
		TweetTask()
	}
}
