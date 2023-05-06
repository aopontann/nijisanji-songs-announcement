package main

import (
	"io"
	"net/http"
	"os"

	"github.com/rs/zerolog/log"
)

func main() {
	port := os.Getenv("PORT")
	log.Debug().Str("severity", "DEBUG").Str("PORT", port).Send()
	if port == "" {
		port = "8080"
	}

	// YouTube Data API 初期化
	YTNew()

	// DB接続初期化
	DBInit()
	defer DB.Close()

	h1 := func(w http.ResponseWriter, _ *http.Request) {
		log.Info().Str("severity", "INFO").Msg("pong!!!")
		io.WriteString(w, "pong\n")
	}

	send := func(w http.ResponseWriter, _ *http.Request) {
		SendMail("test-subject", "test2-message")
		io.WriteString(w, "send-demo\n")
	}

	tweet := func(w http.ResponseWriter, _ *http.Request) {
		video := GetVideoInfo{ID: "test", Title: "test"}
		err := video.Tweets()
		if err != nil {
			io.WriteString(w, "failed-tweet\n")
			return
		}
		io.WriteString(w, "success-tweet\n")
	}

	http.HandleFunc("/tweet", TweetHandler)
	http.HandleFunc("/item-count", UpdateItemCountHandler) // PUT
	http.HandleFunc("/check-new-video", CheckNewVideoHAndler) // POST

	// 検証用
	http.HandleFunc("/ping", h1)
	http.HandleFunc("/mail", send)
	http.HandleFunc("/test/tweet", tweet)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal().Err(err).Msg("start http server failed")
	}
}
