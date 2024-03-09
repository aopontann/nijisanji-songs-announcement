package nsa

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"slices"
	"strings"
	"time"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

type Job struct {
	db *bun.DB
	yt *youtube.Service
}

type Vtuber struct {
	bun.BaseModel `bun:"table:vtubers"`

	ID        string    `bun:"id,type:varchar(24),pk"`
	Name      string    `bun:"name,notnull,type:varchar"`
	ItemCount int64     `bun:"item_count,default:0,type:integer"`
	CreatedAt time.Time `bun:"created_at,type:TIMESTAMP(0),nullzero,notnull,default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `bun:"updated_at,type:TIMESTAMP(0),nullzero,notnull,default:CURRENT_TIMESTAMP"`
}

type Video struct {
	bun.BaseModel `bun:"table:videos"`

	ID        string    `bun:"id,type:varchar(11),pk"`
	Title     string    `bun:"title,notnull,type:varchar"`
	Duration  string    `bun:"duration,notnull,type:varchar"` //int型に変換したほうがいいか？
	Viewers   int64     `bun:"viewers,notnull,type:integer"`
	Content   string    `bun:"content,notnull,type:varchar"`
	Announced bool      `bun:"announced,default:false,type:boolean"`
	StartTime time.Time `bun:"scheduled_start_time,type:timestamp"`
	CreatedAt time.Time `bun:"created_at,type:TIMESTAMP(0),nullzero,notnull,default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `bun:"updated_at,type:TIMESTAMP(0),nullzero,notnull,default:CURRENT_TIMESTAMP"`
}

type User struct {
	bun.BaseModel `bun:"table:users"`

	Token     string    `bun:"token,type:varchar(200),pk"`
	CreatedAt time.Time `bun:"created_at,type:TIMESTAMP(0),nullzero,notnull,default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `bun:"updated_at,type:TIMESTAMP(0),nullzero,notnull,default:CURRENT_TIMESTAMP"`
}

func UpdatePlaylistItemJob() error {
	ctx := context.Background()
	config, err := pgx.ParseConfig(os.Getenv("DSN"))
	if err != nil {
		return err
	}
	sqldb := stdlib.OpenDB(*config)
	db := bun.NewDB(sqldb, pgdialect.New())

	yt, err := youtube.NewService(ctx, option.WithAPIKey(os.Getenv("YOUTUBE_API_KEY")))
	if err != nil {
		return err
	}

	job := Job{db, yt}

	// 動画が投稿されたチャンネルIDと動画数を取得
	newlist, err := job.NewPlaylistItem()
	if err != nil {
		return err
	}

	tx, err := job.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	err = job.UpdatePlaylistItem(tx, newlist)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func CheckNewVideoJob() error {
	ctx := context.Background()
	config, err := pgx.ParseConfig(os.Getenv("DSN"))
	if err != nil {
		return err
	}
	sqldb := stdlib.OpenDB(*config)
	db := bun.NewDB(sqldb, pgdialect.New())

	yt, err := youtube.NewService(ctx, option.WithAPIKey(os.Getenv("YOUTUBE_API_KEY")))
	if err != nil {
		return err
	}

	job := Job{db, yt}

	// 動画が投稿されたチャンネルIDと動画数を取得
	newlist, err := job.NewPlaylistItem()
	if err != nil {
		return err
	}

	var cidList []string
	for cid := range newlist {
		cidList = append(cidList, cid)
	}

	// 動画IDリストを取得
	vidList, err := CustomPlaylistItems(yt, cidList)
	if err != nil {
		return err
	}

	vlist, err := CustomVideo(yt, vidList)
	if err != nil {
		return err
	}

	filtedVlist, err := FilterVideo(vlist)
	if err != nil {
		return err
	}

	tx, err := job.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	err = job.UpdatePlaylistItem(tx, newlist)
	if err != nil {
		return err
	}
	if len(filtedVlist) == 0 {
		err := tx.Commit()
		if err != nil {
			return err
		}
		return nil
	}
	err = job.SaveVideos(tx, filtedVlist)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func SongVideoAnnounceJob() error {
	ctx := context.Background()
	config, err := pgx.ParseConfig(os.Getenv("DSN"))
	if err != nil {
		return err
	}
	sqldb := stdlib.OpenDB(*config)
	db := bun.NewDB(sqldb, pgdialect.New())

	app, err := firebase.NewApp(ctx, nil)
	if err != nil {
		slog.Error("firebase.NewApp error",
			slog.String("severity", "ERROR"),
			slog.String("message", err.Error()),
		)
		return err
	}
	client, err := app.Messaging(ctx)
	if err != nil {
		slog.Error("app.Messaging error",
			slog.String("severity", "ERROR"),
			slog.String("message", err.Error()),
		)
		return err
	}

	now, _ := time.Parse(time.RFC3339, time.Now().UTC().Format("2006-01-02T15:04:00Z"))
	tAfter := now.Add(1 * time.Second)
	tBefore := now.Add(5 * time.Minute)
	var videos []Video
	err = db.NewSelect().Model(&videos).Where("? BETWEEN ? AND ?", bun.Ident("scheduled_start_time"), tAfter, tBefore).Scan(ctx)
	if err != nil {
		return err
	}

	// tw := NewTwitter()
	// mk := NewMisskey(os.Getenv("MISSKEY_TOKEN"))

	var tokens []string
	err = db.NewSelect().Model((*User)(nil)).Column("token").Scan(ctx, &tokens)
	if err != nil {
		return err
	}

	for _, v := range videos {
		// 放送前の動画か
		if v.Content != "upcoming" {
			continue
		}
		// 動画の長さが10分未満か
		if !regexp.MustCompile(`^PT([1-9]M[1-5]?[0-9]S|[1-5]?[0-9]S)`).Match([]byte(v.Duration)) {
			continue
		}
		// 切り抜き動画ではないか
		if regexp.MustCompile(`.*切り抜き.*`).Match([]byte(v.Title)) {
			continue
		}
		// ショート動画ではないか
		if regexp.MustCompile(`.*shorts.*`).Match([]byte(v.Title)) {
			continue
		}
		// 試聴動画ではないか
		if regexp.MustCompile(`.*試聴.*`).Match([]byte(v.Title)) {
			continue
		}

		err = SendMail("検証 5分後に公開", fmt.Sprintf("https://www.youtube.com/watch?v=%s", v.ID))
		if err != nil {
			return err
		}

		// FCM
		message := &messaging.MulticastMessage{
			Notification: &messaging.Notification{
				Title: v.Title,
				Body:  fmt.Sprintf("https://www.youtube.com/watch?v=%s", v.ID),
			},
			Tokens: tokens,
		}
		// 500通まで
		_, err := client.SendEachForMulticast(ctx, message)
		if err != nil {
			slog.Error("SendEachForMulticast error",
				slog.String("severity", "ERROR"),
				slog.String("message", err.Error()),
			)
			return err
		}

		// ツイート
		// err = tw.Id(v.ID).Title(v.Title).Tweet()
		// if err != nil {
		// 	http.Error(w, err.Error(), http.StatusInternalServerError)
		// 	return
		// }

		// Missky post
		// err = mk.Post(v.ID, v.Title)
		// if err != nil {
		// 	http.Error(w, err.Error(), http.StatusInternalServerError)
		// 	return
		// }
	}
	return nil
}

// 公開済みの動画、Youtube上で削除された動画をDBから削除する
func DeleteVideoJob() error {
	ctx := context.Background()
	config, err := pgx.ParseConfig(os.Getenv("DSN"))
	if err != nil {
		return err
	}
	sqldb := stdlib.OpenDB(*config)
	db := bun.NewDB(sqldb, pgdialect.New())

	yt, err := youtube.NewService(ctx, option.WithAPIKey(os.Getenv("YOUTUBE_API_KEY")))
	if err != nil {
		slog.Error("youtube.NewService",
			slog.String("severity", "ERROR"),
			slog.String("message", err.Error()),
		)
		return err
	}
	var vidList []string
	err = db.NewSelect().Model((*Video)(nil)).Column("id").Scan(ctx, &vidList)
	if err != nil {
		return err
	}

	vlist, err := CustomVideo(yt, vidList)
	if err != nil {
		return err
	}

	// 削除する動画ID
	var delVidList []string
	for _, v := range vlist {
		// 公開前、公開中は削除しない
		if v.Snippet.LiveBroadcastContent != "none" {
			continue
		}
		// プレミア公開、生放送が終了した動画は削除する
		if v.LiveStreamingDetails != nil {
			delVidList = append(delVidList, v.Id)
			continue
		}

		// プレミア公開、生放送ではなく、24時間以内に公開された動画は削除しない
		t, _ := time.Parse("2006-01-02T15:04:05Z", v.Snippet.PublishedAt)
		if time.Now().AddDate(0, 0, -1).Compare(t) < 0 {
			continue
		}
		delVidList = append(delVidList, v.Id)
	}

	// youtube上で削除された動画も削除
	var resVidList []string
	for _, v := range vlist {
		resVidList = append(resVidList, v.Id)
	}
	for _, id := range vidList {
		if !slices.Contains(resVidList, id) {
			delVidList = append(delVidList, id)
		}
	}

	_, err = db.NewDelete().Model((*Video)(nil)).Where("id IN (?)", bun.In(delVidList)).Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

// キーワード告知（checkNewVideoJobとは別で処理することになった場合に記載）
func KeywordAnnounceJob() error {
	ctx := context.Background()
	config, err := pgx.ParseConfig(os.Getenv("DSN"))
	if err != nil {
		panic(err)
	}
	sqldb := stdlib.OpenDB(*config)
	db := bun.NewDB(sqldb, pgdialect.New())

	now, _ := time.Parse(time.RFC3339, time.Now().UTC().Format("2006-01-02T15:04:00Z"))
	tAfter := now.Add(-10 * time.Minute)
	tBefore := now.Add(-5 * time.Minute)
	var videos []Video
	err = db.NewSelect().Model(&videos).Where("? BETWEEN ? AND ?", bun.Ident("created_at"), tAfter, tBefore).Scan(ctx)
	if err != nil {
		return err
	}

	for _, v := range videos {
		if regexp.MustCompile(`.*Lethal Company.*`).Match([]byte(v.Title)) {
			err := SendMail("Lethal Company やるよ！", fmt.Sprintf("https://www.youtube.com/watch?v=%s", v.ID))
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// チャンネルIDをキー、プレイリストに含まれている動画数を値とした連想配列を返す
func (j Job) NewPlaylistItem() (map[string]int64, error) {
	// DBからチャンネルID、チャンネルごとの動画数を取得
	var ids []string
	var itemCount []int64
	ctx := context.Background()
	err := j.db.NewSelect().Model((*Vtuber)(nil)).Column("id", "item_count").Scan(ctx, &ids, &itemCount)
	if err != nil {
		return nil, err
	}

	// プレイリスト一覧と比較用のMAPを作成
	var plist []string
	oldList := make(map[string]int64, 500)
	for i := range ids {
		oldList[ids[i]] = itemCount[i]
		pid := strings.Replace(ids[i], "UC", "UU", 1)
		plist = append(plist, pid)
	}

	// チャンネルIDをキー、プレイリストに含まれている動画数を値とした連想配列を返す
	// TODO: j.yt 検討
	newlist, err := CustomPlaylists(j.yt, plist)
	if err != nil {
		return nil, err
	}

	// 動画数が変わっていないチャンネルは返り値に含めない
	for i, id := range ids {
		if itemCount[i] == newlist[id] {
			delete(newlist, id)
		}
	}

	return newlist, nil
}

func (j Job) UpdatePlaylistItem(tx bun.Tx, newlist map[string]int64) error {
	ctx := context.Background()
	// DBを新しく取得したデータに更新
	var updateVideo []Vtuber
	for k, v := range newlist {
		updateVideo = append(updateVideo, Vtuber{ID: k, ItemCount: v, UpdatedAt: time.Now()})
	}

	_, err := tx.NewUpdate().Model(&updateVideo).Column("item_count", "updated_at").Bulk().Exec(ctx)
	if err != nil {
		slog.Error("update-itemCount",
			slog.String("severity", "ERROR"),
			slog.String("message", err.Error()),
		)
		return err
	}

	return nil
}

func (j Job) SaveVideos(tx bun.Tx, vlist []youtube.Video) error {
	var Videos []Video
	for _, v := range vlist {
		var Viewers int64
		Viewers = 0
		scheduledStartTime := "1998-01-01 15:04:05" // 例 2022-03-28T11:00:00Z
		if v.LiveStreamingDetails != nil {
			Viewers = int64(v.LiveStreamingDetails.ConcurrentViewers)
			// "2022-03-28 11:00:00"形式に変換
			rep1 := strings.Replace(v.LiveStreamingDetails.ScheduledStartTime, "T", " ", 1)
			scheduledStartTime = strings.Replace(rep1, "Z", "", 1)
		}
		t, _ := time.Parse("2006-01-02 15:04:05", scheduledStartTime)
		Videos = append(Videos, Video{
			ID:        v.Id,
			Title:     v.Snippet.Title,
			Duration:  v.ContentDetails.Duration,
			Content:   v.Snippet.LiveBroadcastContent,
			Viewers:   Viewers,
			StartTime: t,
			UpdatedAt: time.Now(),
		})
	}

	if len(Videos) == 0 {
		return nil
	}

	ctx := context.Background()
	_, err := tx.NewInsert().Model(&Videos).Ignore().Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}
