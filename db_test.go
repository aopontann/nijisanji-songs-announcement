package nsa

import (
	"context"
	"database/sql"
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
