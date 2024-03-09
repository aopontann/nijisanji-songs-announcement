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

	http.HandleFunc("/v2/check", func(w http.ResponseWriter, r *http.Request) {
		err := nsa.CheckNewVideoJob()
		if err != nil {
			slog.Error("CheckNewVideoJob",
				slog.String("severity", "ERROR"),
				slog.String("message", err.Error()),
			)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
	http.HandleFunc("/v2/keyword", func(w http.ResponseWriter, r *http.Request) {
		err := nsa.KeywordAnnounceJob()
		if err != nil {
			slog.Error("KeywordAnnounceJob",
				slog.String("severity", "ERROR"),
				slog.String("message", err.Error()),
			)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
	http.HandleFunc("/v2/song", func(w http.ResponseWriter, r *http.Request) {
		err := nsa.SongVideoAnnounceJob()
		if err != nil {
			slog.Error("SongVideoAnnounceJob",
				slog.String("severity", "ERROR"),
				slog.String("message", err.Error()),
			)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
	http.HandleFunc("/v2/delete", func(w http.ResponseWriter, r *http.Request) {
		err := nsa.DeleteVideoJob()
		if err != nil {
			slog.Error("DeleteVideoJob",
				slog.String("severity", "ERROR"),
				slog.String("message", err.Error()),
			)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	http.HandleFunc("/v2/playlistItems/update", func(w http.ResponseWriter, r *http.Request) {
		err := nsa.UpdatePlaylistItemJob()
		if err != nil {
			slog.Error("UpdatePlaylistItemJob",
				slog.String("severity", "ERROR"),
				slog.String("message", err.Error()),
			)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

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
