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

	h2 := func(w http.ResponseWriter, _ *http.Request) {
		log.Info().Str("severity", "ERROR").Msg("error!!!")
		io.WriteString(w, "error-demo\n")
	}

	send := func(w http.ResponseWriter, _ *http.Request) {
		sendMail("id001", "test-subject", "test2-message")
		io.WriteString(w, "send-demo\n")
	}

	http.HandleFunc("/ping", h1)
	http.HandleFunc("/error", h2)
	http.HandleFunc("/mail", send)
	http.HandleFunc("/youtube", YoutubeHandler)
	http.HandleFunc("/youtube/updateVideoCount", UpdateVideoCountHandler)
	http.HandleFunc("/youtube/checkNewVideo", CheckNewUploadHandler)
	http.HandleFunc("/twitter", TwitterHandler)

	// log.Debug().Msgf("listening on port %s", port)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal().Err(err).Msg("start http server failed")
	}
}
