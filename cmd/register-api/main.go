package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"log/slog"
	"net/http"
	"os"

	nsa "github.com/aopontann/nijisanji-songs-announcement"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

type CheckReqBody struct {
	Token string `json:"token"`
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger) // 以降、JSON形式で出力される。

	ctx := context.Background()
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(os.Getenv("DSN"))))
	db := bun.NewDB(sqldb, pgdialect.New())

	http.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		var b CheckReqBody
		if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
			slog.Error("json.NewDecoder",
				slog.String("severity", "ERROR"),
				slog.String("message", err.Error()),
			)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if r.Method == http.MethodPost {
			_, err := db.NewInsert().Model(&nsa.User{Token: b.Token}).Exec(ctx)
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
			_, err := db.NewDelete().Model(&nsa.User{Token: b.Token}).Exec(ctx)
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
