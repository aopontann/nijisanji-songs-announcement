package main

import (
	"database/sql"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/rs/zerolog/log"
)

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

func PlaylistsSeed() error {
	cList, err := Channels()
	if err != nil {
		return err
	}

	tx, err := DB.Begin()
	if err != nil {
		log.Error().Str("severity", "ERROR").Err(err).Msg("DB.Begin error")
		return err
	}

	stmt, err := tx.Prepare("UPDATE vtubers SET playlist_id = ? WHERE id = ?")
	if err != nil {
		log.Error().Str("severity", "ERROR").Err(err).Msg("DB.Prepare error")
		return err
	}

	for _, c := range cList {
		_, err = stmt.Exec(c.PlaylistID, c.ID)
		if err != nil {
			log.Error().Str("severity", "ERROR").Err(err).Msg("DB.Exec error")
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Error().Str("severity", "ERROR").Err(err).Msg("tx.commit error")
		return err
	}

	return nil
}
