package main

import (
	"context"
	"io"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

var YoutubeService *youtube.Service

func main() {
	var err error

	port := os.Getenv("PORT")
	// log.Debug().Str("severity", "DEBUG").Str("PORT", port).Send()
	if port == "" {
		err := godotenv.Load(".env.local")
		if err != nil {
			log.Fatal().Err(err).Msg("godotenv.Load() error")
		}
		port = os.Getenv("PORT")
	}

	ctx := context.Background()
	YoutubeService, err = youtube.NewService(ctx, option.WithAPIKey(os.Getenv("YOUTUBE_API_KEY")))
	if err != nil {
		log.Fatal().Err(err).Msg("youtube.NewService create failed")
	}

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
		sendMail("test", "test2-message")
		io.WriteString(w, "send-demo\n")
	}

	http.HandleFunc("/ping", h1)
	http.HandleFunc("/error", h2)
	http.HandleFunc("/mail", send)
	http.HandleFunc("/youtube", YoutubeHandler)
	http.HandleFunc("/twitter", TwitterHandler)
	http.HandleFunc("/twitter/search", TwitterSearchHandler)

	// log.Debug().Msgf("listening on port %s", port)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal().Err(err).Msg("start http server failed")
	}
}
