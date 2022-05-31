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
	// .envの読み込み(開発環境の時のみ読み込むようにしたい)
	err := godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
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

	http.HandleFunc("/ping", h1)
	http.HandleFunc("/error", h2)
	http.HandleFunc("/youtube", YoutubeHandler)
	http.HandleFunc("/twitter", TwitterHandler)

	// log.Debug().Msgf("listening on port %s", port)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal().Err(err).Msg("start http server failed")
	}
}
