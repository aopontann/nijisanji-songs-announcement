package main

import (
	"context"
	"database/sql"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/rs/zerolog/log"
)

type FireStore struct{}

var DB *sql.DB

func DBInit() {
	var err error
	DB, err = sql.Open("mysql", os.Getenv("DSN"))
	if err != nil {
		log.Fatal().Err(err).Msg("db connection failed")
	}

	err = DB.Ping()
	if err != nil {
		log.Fatal().Err(err).Msg("db ping failed")
	}
}

func (fs *FireStore) Save(ytcr []YouTubeCheckResponse) error {
	ctx := context.Background()
	for _, v := range ytcr {
		_, _, err := FireStoreService.Collection("demo").Add(ctx, map[string]interface{}{
			"twitter_id":  v.TwitterID,
			"schedule": v.Schedule,
			"youtube_id":   v.ID,
		})
		if err != nil {
			return err
		}
	}
	return nil
}
