package main

import (
	"context"
	"io"
	"net/http"
	"os"

	"cloud.google.com/go/firestore"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

var YoutubeService *youtube.Service
var FireStoreService *firestore.Client

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
	FireStoreService, err = firestore.NewClient(ctx, "aerial-rush-345708")
	if err != nil {
		log.Fatal().Err(err).Msg("firestore.NewService create failed")
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

	h3 := func(w http.ResponseWriter, _ *http.Request) {
		ctx := context.Background()
		client, err := firestore.NewClient(ctx, "aerial-rush-345708")
		if err != nil {
			io.WriteString(w, "error NewClient firestore-demo\n")
			return
		}

		_, _, err = client.Collection("demo").Add(ctx, map[string]interface{}{
			"first":  "Alan2",
			"middle": "Mathison",
			"last":   "Turing",
			"born":   1912,
		})
		if err != nil {
			log.Printf("Failed adding aturing: %v", err)
			io.WriteString(w, "error firestore-demo\n")
		}

		// docsnap, err := client.Collection("ライバー").Doc("UCeGendL8CO5RkffB6IFwHow").Get(ctx)
		// if err != nil {
		// 	log.Print(err.Error())
		// 	io.WriteString(w, "error firestore-demo\n")
		// 	return
		// }
		// dataMap := docsnap.Data()
		// fmt.Println(dataMap)
		io.WriteString(w, "firestore-demo\n")
	}

	http.HandleFunc("/ping", h1)
	http.HandleFunc("/error", h2)
	http.HandleFunc("/firestore", h3)
	http.HandleFunc("/youtube", YoutubeHandler)
	http.HandleFunc("/twitter", TwitterHandler)
	http.HandleFunc("/twitter/search", TwitterSearchHandler)

	// log.Debug().Msgf("listening on port %s", port)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal().Err(err).Msg("start http server failed")
	}
}
