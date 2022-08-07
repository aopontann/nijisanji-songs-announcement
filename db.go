package main

import (
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

func DBSave(ytcr []YouTubeCheckResponse) error {
	tx, err := DB.Begin()
	if err != nil {
		log.Error().Str("severity", "ERROR").Err(err).Msg("DB.Begin error")
		return err
	}
	stmt, err := tx.Prepare("INSERT IGNORE INTO videos(id, title, songConfirm, scheduled_start_time, twitter_id) VALUES(?,?,?,?,?)")
	if err != nil {
		log.Error().Str("severity", "ERROR").Err(err).Msg("DB.Prepare error")
		return err
	}

	for _, v := range ytcr {
		_, err := stmt.Exec(v.ID, v.Title, true, v.Schedule, v.TwitterID)
		if err != nil {
			log.Error().Str("severity", "ERROR").Err(err).Msg("Save videos failed")
			tx.Rollback()
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Error().Str("severity", "ERROR").Err(err).Msg("DB.Commit error")
		return err
	}
	return nil
}
