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
	"unicode/utf8"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"

	nsa "github.com/aopontann/nijisanji-songs-announcement"
)

type ReqBody struct {
	Song        bool   `json:"song"`
	Keyword     bool   `json:"keyword"`
	KeywordText string `json:"keyword_text"`
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

	fcm := nsa.NewFCM()

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

			// バリデーション処理
			if utf8.RuneCountInString(b.KeywordText) > 20 {
				http.Error(w, "21文字以上のキーワードは登録できません", http.StatusBadRequest)
				return
			}

			user := new(nsa.User)
			err := db.NewSelect().Column("song", "keyword", "keyword_text").Model(user).Where("token = ?", token).Scan(ctx)
			if err == sql.ErrNoRows {
				_, err = db.NewInsert().
					Model(&nsa.User{Token: token, Song: b.Song, Keyword: b.Keyword, KeywordText: b.KeywordText}).
					Exec(ctx)
				if err != nil {
					slog.Error("NewInsert error",
						slog.String("severity", "ERROR"),
						slog.String("message", err.Error()),
						slog.String("token", token),
					)
					http.Error(w, "登録に失敗しました。", http.StatusInternalServerError)
					return
				}
				if user.Keyword && user.KeywordText != "" {
					err = fcm.SetTopic(token, b.KeywordText)
					if err != nil {
						slog.Error("SetTopic error",
							slog.String("severity", "ERROR"),
							slog.String("message", err.Error()),
							slog.String("token", token),
							slog.String("topic", b.KeywordText),
						)
						http.Error(w, "キーワードの登録に失敗しました", http.StatusInternalServerError)
						return
					}
				}
				return
			}
			if err != nil {
				slog.Error("select user error",
					slog.String("severity", "ERROR"),
					slog.String("message", err.Error()),
					slog.String("token", token),
				)
				http.Error(w, "ユーザ情報の取得に失敗しました", http.StatusInternalServerError)
				return
			}

			// 変更の場合
			if !b.Keyword || user.KeywordText != "" {
				err = fcm.DeleteTopic(token, user.KeywordText)
				if err != nil {
					slog.Error("DeleteTopic error",
						slog.String("severity", "ERROR"),
						slog.String("message", err.Error()),
						slog.String("token", token),
						slog.String("topic", user.KeywordText),
					)
					http.Error(w, "キーワードの更新に失敗しました", http.StatusInternalServerError)
					return
				}
			}
			if b.Keyword && b.KeywordText != "" {
				err = fcm.SetTopic(token, b.KeywordText)
				if err != nil {
					slog.Error("Delete & SetTopic error",
						slog.String("severity", "ERROR"),
						slog.String("message", err.Error()),
						slog.String("token", token),
						slog.String("before_topic", user.KeywordText),
						slog.String("after_topic", b.KeywordText),
					)
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			}

			_, err = db.NewUpdate().
				Model(&nsa.User{Token: token, Song: b.Song, Keyword: b.Keyword, KeywordText: b.KeywordText}).
				Column("song", "keyword", "keyword_text").
				WherePK().Exec(ctx)
			if err != nil {
				slog.Error("update user error",
					slog.String("severity", "ERROR"),
					slog.String("message", err.Error()),
					slog.String("token", token),
					slog.String("before_topic", user.KeywordText),
					slog.String("after_topic", b.KeywordText),
				)
				http.Error(w, err.Error(), http.StatusInternalServerError)
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
	err := db.NewSelect().Column("song", "keyword", "keyword_text").Model(user).Where("token = ?", token).Scan(ctx)
	if err == sql.ErrNoRows {
		w.WriteHeader(http.StatusNoContent)
		w.Write([]byte("No Content"))
		return
	}
	if err != nil {
		slog.Error("insert token error",
			slog.String("severity", "ERROR"),
			slog.String("message", err.Error()),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("content-Type", "application/json")

	v, err := json.Marshal(&ReqBody{user.Song, user.Keyword, user.KeywordText})
	if err != nil {
		slog.Error("insert token error",
			slog.String("severity", "ERROR"),
			slog.String("message", err.Error()),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(v)
}
