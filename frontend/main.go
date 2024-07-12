package main

import (
	"context"
	"database/sql"
	"embed"
	"encoding/json"
	"io/fs"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"

	nsa "github.com/aopontann/nijisanji-songs-announcement"
)

type ReqBody struct {
	Song bool `json:"song"`
	Info bool `json:"info"`
}

//go:embed dist/*
var dist embed.FS

// test
func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger) // 以降、JSON形式で出力される。

	ctx := context.Background()
	config, err := pgx.ParseConfig(os.Getenv("DSN"))
	if err != nil {
		panic(err)
	}
	sqldb := stdlib.OpenDB(*config)
	db := bun.NewDB(sqldb, pgdialect.New())

	dist, err := fs.Sub(dist, "dist")
	if err != nil {
		panic(err)
	}
	http.Handle("/", http.FileServer(http.FS(dist)))

	http.HandleFunc("/api/subscription", func(w http.ResponseWriter, r *http.Request) {
		if len(r.Header["Authorization"]) == 0 {
			http.Error(w, "NG", http.StatusBadRequest)
			return
		}
		token := strings.Split(r.Header["Authorization"][0], " ")[1]

		if r.Method == http.MethodGet {
			getHandler(w, token, db, ctx)
			return
		}

		if r.Method == http.MethodPost {
			var b ReqBody
			if err = json.NewDecoder(r.Body).Decode(&b); err != nil {
				slog.Error("NewDecoder error",
					slog.String("severity", "ERROR"),
					slog.String("message", err.Error()),
				)
				http.Error(w, "リクエストボディが不正です", http.StatusInternalServerError)
				return
			}

			slog.Info("POST", 
				slog.String("severity", "INFO"),
				slog.String("token", token),
				slog.String("User-Agent", r.Header["User-Agent"][0]),
			)

			_, err := db.NewInsert().
				Model(&nsa.User{Token: token, Song: b.Song, Info: b.Info}).
				On("CONFLICT (token) DO UPDATE").
				Set("song = EXCLUDED.song").
				Set("info = EXCLUDED.info").
				Exec(ctx)
			if err != nil {
				slog.Error("Upsert error",
					slog.String("severity", "ERROR"),
					slog.String("message", err.Error()),
				)
				http.Error(w, "処理に失敗しました", http.StatusInternalServerError)
				return
			}

			w.Write([]byte("OK!!"))
			return
		}

		if r.Method == http.MethodDelete {
			_, err := db.NewDelete().Model((*nsa.User)(nil)).Where("token = ?", token).Exec(ctx)
			if err != nil {
				slog.Error("delete token error",
					slog.String("severity", "ERROR"),
					slog.String("message", err.Error()),
				)
				http.Error(w, "削除に失敗しました", http.StatusInternalServerError)
				return
			}
			w.Write([]byte("OK!!"))
			return
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

func getHandler(w http.ResponseWriter, token string, db *bun.DB, ctx context.Context) {
	user := new(nsa.User)
	err := db.NewSelect().Column("song", "info").Model(user).Where("token = ?", token).Scan(ctx)
	if err == sql.ErrNoRows {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if err != nil {
		slog.Error("select token error",
			slog.String("severity", "ERROR"),
			slog.String("message", err.Error()),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("content-Type", "application/json")

	v, err := json.Marshal(&ReqBody{user.Song, user.Info})
	if err != nil {
		slog.Error("json.Marshal error",
			slog.String("severity", "ERROR"),
			slog.String("message", err.Error()),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(v)
}
