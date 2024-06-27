package nsa

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

func TestDBPlaylists(t *testing.T) {
	// 事前準備
	bunDB := setup()
	defer bunDB.Close()
	defer down(bunDB)
	db := NewDB(bunDB)
	ctx := context.Background()

	vtubers := []Vtuber{
		{ID: "UCe_p3YEuYJb8Np0Ip9dk-FQ", Name: "AAA", ItemCount: 1, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: "UUveZ9Ic1VtcXbsyaBgxPMvg", Name: "BBB", ItemCount: 18, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}
	_, err := bunDB.NewInsert().Model(&vtubers).Exec(ctx)
	if err != nil {
		t.Error(err)
	}

	// テスト
	playlists, err := db.Playlists()
	if err != nil {
		t.Error(err)
	}

	if playlists["UUe_p3YEuYJb8Np0Ip9dk-FQ"] != 1 {
		t.Errorf("except 1, but %d", playlists["UUe_p3YEuYJb8Np0Ip9dk-FQ"])
	}

	if playlists["UUveZ9Ic1VtcXbsyaBgxPMvg"] != 18 {
		t.Errorf("except 18, but %d", playlists["UUveZ9Ic1VtcXbsyaBgxPMvg"])
	}
}

func TestUpdatePlaylistItem(t *testing.T) {
	//////////////////// 事前準備 ////////////////////
	bunDB := setup()
	defer bunDB.Close()
	defer down(bunDB)
	db := NewDB(bunDB)
	ctx := context.Background()

	vtubers := []Vtuber{
		{ID: "UCe_p3YEuYJb8Np0Ip9dk-FQ", Name: "AAA", ItemCount: 1, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: "UCveZ9Ic1VtcXbsyaBgxPMvg", Name: "BBB", ItemCount: 18, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}
	_, err := bunDB.NewInsert().Model(&vtubers).Exec(ctx)
	if err != nil {
		t.Error(err)
	}
	//////////////////////////////////////////////////

	tx, err := db.Service.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		t.Error(err)
	}

	plist := map[string]int64{
		"UUe_p3YEuYJb8Np0Ip9dk-FQ": 2,
		"UUveZ9Ic1VtcXbsyaBgxPMvg": 19,
	}

	// テスト
	err = db.UpdatePlaylistItem(tx, plist)
	if err != nil {
		t.Error(err)
	}
	err = tx.Commit()
	if err != nil {
		t.Error(err)
	}

	playlists, err := db.Playlists()
	if err != nil {
		t.Error(err)
	}

	if playlists["UUe_p3YEuYJb8Np0Ip9dk-FQ"] != 2 {
		t.Errorf("except 2, but %d", playlists["UUe_p3YEuYJb8Np0Ip9dk-FQ"])
	}

	if playlists["UUveZ9Ic1VtcXbsyaBgxPMvg"] != 19 {
		t.Errorf("except 19, but %d", playlists["UUveZ9Ic1VtcXbsyaBgxPMvg"])
	}
}

func TestGetKeywordTextList(t *testing.T) {
	//////////////////// 事前準備 ////////////////////
	bunDB := setup()
	defer bunDB.Close()
	db := NewDB(bunDB)
	ctx := context.Background()

	vtubers := []User{
		{Token: "aaa", Song: true, Keyword: true, KeywordText: "aaa", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{Token: "bbb", Song: true, Keyword: true, KeywordText: "", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{Token: "ccc", Song: true, Keyword: true, KeywordText: "ccc", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{Token: "ddd", Song: true, Keyword: false, KeywordText: "ddd", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{Token: "eee", Song: true, Keyword: true, KeywordText: "ccc", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}
	_, err := bunDB.NewInsert().Model(&vtubers).Exec(ctx)
	if err != nil {
		t.Error(err)
	}
	//////////////////////////////////////////////////

	// テスト
	list, err := db.getKeywordTextList()
	if err != nil {
		t.Error(err)
	}
	for _, r := range list {
		fmt.Println(r)
	}
	if len(list) != 2 {
		t.Errorf("except 2, but %d", len(list))
	}

	_, err = bunDB.NewDelete().Model(&vtubers).WherePK().Exec(ctx)
	if err != nil {
		t.Error(err)
	}
}

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

// 動画が登録されているか

func setup() *bun.DB {
	config, err := pgx.ParseConfig(os.Getenv("DSN"))
	if err != nil {
		panic(err)
	}
	sqldb := stdlib.OpenDB(*config)
	return bun.NewDB(sqldb, pgdialect.New())
}

func down(db *bun.DB) {
	defer db.Close()
	ctx := context.Background()
	_, err := db.NewTruncateTable().Model((*Vtuber)(nil)).Exec(ctx)
	if err != nil {
		panic(err)
	}
}
