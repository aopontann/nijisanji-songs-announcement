package nsa

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	ndb "github.com/aopontann/nijisanji-songs-announcement/db"
	_ "github.com/go-sql-driver/mysql"
	"github.com/rs/zerolog/log"
)

func TestMisskeyTask(t *testing.T) {
	db, err := sql.Open("mysql", os.Getenv("DSN"))
	if err != nil {
		log.Fatal().Str("severity", "ERROR").Str("service", "sql.Open").Msg(err.Error())
	}

	now, _ := time.Parse(time.RFC3339, time.Now().UTC().Format("2006-01-02T15:04:00Z"))
	tAfter := now.Add(1 * time.Second)
	tBefore := now.Add(5 * time.Minute)

	log.Info().Str("severity", "INFO").Str("service", "misskey").Str("datetime", fmt.Sprintf("%v ~ %v\n", tAfter, tBefore)).Send()

	queries := ndb.New(db)
	ctx := context.Background()
	vList, err := queries.ListVideoIdTitle(ctx, ndb.ListVideoIdTitleParams{
		ScheduledStartTime:   tAfter,
		ScheduledStartTime_2: tBefore,
	})
	if err != nil {
		t.Errorf("Tweets Error")
	}
	fmt.Println(vList)

}
