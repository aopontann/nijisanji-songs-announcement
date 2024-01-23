package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"

	nsa "github.com/aopontann/nijisanji-songs-announcement"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger) // 以降、JSON形式で出力される。
	http.HandleFunc("/hello", nsa.Hello)
	http.HandleFunc("/check", nsa.CheckNewVideo)
	http.HandleFunc("/video", nsa.SaveVideos)
	http.HandleFunc("/song", nsa.SongVideoAnnounce)
	http.HandleFunc("/keyword", nsa.KeywordAnnounce)
	// Determine port for HTTP service.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	// Start HTTP server.
	log.Printf("Listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
