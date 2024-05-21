package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"

	nsa "github.com/aopontann/nijisanji-songs-announcement"
)

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

	http.Handle("/", http.FileServer(http.Dir("frontend/public/")))

	http.HandleFunc("/is-subscription", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("/is-subscription")
		if r.Method == http.MethodPost {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				slog.Error("is-subscription",
					slog.String("severity", "ERROR"),
					slog.String("message", err.Error()),
				)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			exists, err := db.NewSelect().Model((*nsa.User)(nil)).Where("token = ?", string(body)).Exists(ctx)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}
			if !exists {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte("NotFound"))
				return
			}
			w.Write([]byte("OK!!"))
			return
		}
	})

	http.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("access")
		body, err := io.ReadAll(r.Body)
		if err != nil {
			slog.Error("insert token error",
				slog.String("severity", "ERROR"),
				slog.String("message", err.Error()),
			)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if r.Method == http.MethodPost {
			_, err := db.NewInsert().Model(&nsa.User{Token: string(body)}).Ignore().Exec(ctx)
			if err != nil {
				slog.Error("insert token error",
					slog.String("severity", "ERROR"),
					slog.String("message", err.Error()),
				)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Write([]byte("OK!!"))
			return
		}

		if r.Method == http.MethodDelete {
			_, err := db.NewDelete().Model((*nsa.User)(nil)).Where("token = ?", string(body)).Exec(ctx)
			if err != nil {
				slog.Error("delete token error",
					slog.String("severity", "ERROR"),
					slog.String("message", err.Error()),
				)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Write([]byte("OK!!"))
			return
		}
	})

	http.HandleFunc("/key", func(w http.ResponseWriter, r *http.Request) {
		publicKey := os.Getenv("WEBPUSH_PUBLIC_KEY")
		w.Write([]byte(publicKey))
	})

	// Determine port for HTTP service.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	// Start HTTP server.
	log.Printf("Listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
