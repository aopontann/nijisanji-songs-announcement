package nsa

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

func TestSongVideos5m(t *testing.T) {
	bunDB := setup()
	defer bunDB.Close()
	db := NewDB(bunDB)
	ctx := context.Background()

	videos := []Video{
		{ID: "aaa", Title: "MV", Duration: "PT4M", Song: false, Viewers: 0, Content: "upcoming", StartTime: time.Date(2024, 6, 27, 12, 10, 0, 0, time.UTC)},
		{ID: "bbb", Title: "aaa", Duration: "PT4M", Song: false, Viewers: 0, Content: "upcoming", StartTime: time.Date(2024, 6, 27, 12, 10, 0, 0, time.UTC)},
		{ID: "ccc", Title: "歌ってみた", Duration: "PT4M", Song: false, Viewers: 0, Content: "upcoming", StartTime: time.Date(2024, 6, 27, 12, 10, 0, 0, time.UTC)},
		{ID: "ddd", Title: "bbb", Duration: "PT4M", Song: true, Viewers: 0, Content: "upcoming", StartTime: time.Date(2024, 6, 27, 12, 10, 0, 0, time.UTC)},
	}

	_, err := bunDB.NewInsert().Model(&videos).Exec(ctx)
	if err != nil {
		t.Error(err)
	}

	res, err := db.songVideos5m()
	if err != nil {
		t.Error(err)
	}
	for _, v := range res {
		fmt.Println(v.Title)
	}
	_, err = bunDB.NewDelete().Model(&videos).WherePK().Exec(ctx)
	if err != nil {
		t.Error(err)
	}
}

func TestSaveVideos(t *testing.T) {
	youtubeApiKey := os.Getenv("YOUTUBE_API_KEY")
	yt := NewYoutube(youtubeApiKey)

	bunDB := setup()
	defer bunDB.Close()
	db := NewDB(bunDB)
	ctx := context.Background()

	videos := []Video{
		{ID: "o4Xhm5fVMBA", Title: "aaa", Duration: "PT4M", Song: false, Viewers: 0, Content: "upcoming", StartTime: time.Date(2024, 6, 27, 12, 10, 0, 0, time.UTC)},
		{ID: "jUdRrvEFZXc", Title: "bbb", Duration: "PT4M", Song: false, Viewers: 0, Content: "upcoming", StartTime: time.Date(2024, 6, 27, 12, 10, 0, 0, time.UTC)},
		{ID: "pla0iR5CG4M", Title: "ccc", Duration: "PT4M", Song: false, Viewers: 0, Content: "upcoming", StartTime: time.Date(2024, 6, 27, 12, 10, 0, 0, time.UTC)},
	}

	_, err := bunDB.NewInsert().Model(&videos).Exec(ctx)
	if err != nil {
		t.Error(err)
	}

	newVideos, err := yt.Videos([]string{"o4Xhm5fVMBA", "jUdRrvEFZXc", "pla0iR5CG4M"})
	if err != nil {
		t.Error(err)
	}

	err = db.SaveVideos(newVideos)
	if err != nil {
		t.Error(err)
	}

	var cids []string
	var titles []string
	err = db.Service.NewSelect().Model((*Video)(nil)).Column("id", "title").Scan(ctx, &cids, &titles)
	if err != nil {
		t.Error(err)
	}

	fmt.Println(cids)
	fmt.Println(titles)
}

func TestNotExistsVideos(t *testing.T) {
	bunDB := setup()
	defer bunDB.Close()
	db := NewDB(bunDB)
	yt := NewYoutube(os.Getenv("YOUTUBE_API_KEY"))
	ctx := context.Background()

	videos := []Video{
		{ID: "SDr4sxCuMf0", Title: "MV", Duration: "PT4M", Song: false, Viewers: 0, Content: "upcoming", StartTime: time.Date(2024, 6, 27, 12, 10, 0, 0, time.UTC)},
	}
	_, err := bunDB.NewInsert().Model(&videos).Exec(ctx)
	if err != nil {
		t.Error(err)
	}

	ytVideos, err := yt.Videos([]string{"SDr4sxCuMf0", "YIQFuRXF3tQ"})
	if err != nil {
		t.Error(err)
	}

	notExistsVideos, err := db.NotExistsVideos(ytVideos)
	if err != nil {
		t.Error(err)
	}

	for _, v := range notExistsVideos {
		fmt.Println(v.Id)
	}

	if notExistsVideos[0].Id != "YIQFuRXF3tQ" {
		t.Error(err)
	}

	_, err = bunDB.NewDelete().Model(&videos).WherePK().Exec(ctx)
	if err != nil {
		t.Error(err)
	}
}

func setup() *bun.DB {
	config, err := pgx.ParseConfig(os.Getenv("DSN"))
	if err != nil {
		panic(err)
	}
	sqldb := stdlib.OpenDB(*config)
	return bun.NewDB(sqldb, pgdialect.New())
}
