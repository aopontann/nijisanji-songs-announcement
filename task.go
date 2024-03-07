package nsa

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	_ "github.com/go-sql-driver/mysql"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/mysqldialect"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

type Vtuber struct {
	bun.BaseModel `bun:"table:vtubers"`

	ID        string    `bun:"id,type:varchar(24),pk"`
	Name      string    `bun:"name,notnull,type:varchar"`
	ItemCount int64     `bun:"item_count,default:0,type:integer"`
	CreatedAt time.Time `bun:",nullzero,notnull,default:current_timestamp()"`
	UpdatedAt time.Time `bun:",nullzero,notnull,default:current_timestamp() ON UPDATE current_timestamp()"`
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
	CreatedAt time.Time `bun:",nullzero,notnull,default:current_timestamp()"`
	UpdatedAt time.Time `bun:",nullzero,notnull,default:current_timestamp() ON UPDATE current_timestamp()"`
}

type User struct {
	bun.BaseModel `bun:"table:users"`

	Token     string    `bun:"token,type:varchar(200),pk"`
	CreatedAt time.Time `bun:",nullzero,notnull,default:current_timestamp()"`
	UpdatedAt time.Time `bun:",nullzero,notnull,default:current_timestamp() ON UPDATE current_timestamp()"`
}

type ChannelBody struct {
	CIDList []string `json:"channelId"`
}
type UpdatedChannel struct {
	CID       string `json:"channelId"`
	ItemCount int64  `json:"item_count"`
}
type CheckResBody struct {
	Data []UpdatedChannel `json:"data"`
}
type VideoBody struct {
	Vlist []youtube.Video `json:"video_list"`
}

func Hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "hello!")
}

func CheckNewVideo(w http.ResponseWriter, r *http.Request) {
	// 初期化処理
	sqldb, err := sql.Open("mysql", os.Getenv("DSN"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	ctx := context.Background()
	db := bun.NewDB(sqldb, mysqldialect.New())
	yt, err := youtube.NewService(ctx, option.WithAPIKey(os.Getenv("YOUTUBE_API_KEY")))
	if err != nil {
		slog.Error("youtube.NewService",
			slog.String("severity", "ERROR"),
			slog.String("message", err.Error()),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// DBからチャンネルID、チャンネルごとの動画数を取得
	var ids []string
	var itemCount []int64
	err = db.NewSelect().Model((*Vtuber)(nil)).Column("id", "item_count").Scan(ctx, &ids, &itemCount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
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
	newlist, err := CustomPlaylists(yt, plist)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// DBを新しく取得したデータに更新
	var upIdList []string
	var upItemCountList []int64
	for k, v := range newlist {
		if oldList[k] != v {
			upIdList = append(upIdList, k)
			upItemCountList = append(upItemCountList, v)
		}
	}
	// 以下のコメントアウトされたAPIで生成されるSQLではバルクアップデートができなかった(planetscale上)ため、ELT,FIELDを使うようにした
	// query := db.NewUpdate().Model(&VtuberList).Column("item_count").Bulk()
	query := db.NewUpdate().Model((*Vtuber)(nil)).
		Set("item_count = ELT(FIELD(id, ?), ?)", bun.In(upIdList), bun.In(upItemCountList)).
		Where("id IN (?)", bun.In(upIdList))
	rawQuery, err := query.AppendQuery(db.Formatter(), nil)
	if err != nil {
		slog.Error("query.AppendQuery",
			slog.String("severity", "ERROR"),
			slog.String("message", err.Error()),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Println(string(rawQuery))

	_, err = query.Exec(ctx)
	if err != nil {
		slog.Error("update-itemCount",
			slog.String("severity", "ERROR"),
			slog.String("message", err.Error()),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 新着動画をアップロードしたチャンネルIDリストをレスポンスとして返す
	var cidList []string
	for k, v := range newlist {
		if v > oldList[k] {
			cidList = append(cidList, k)
		}
	}
	resp := ChannelBody{cidList}
	bytes, err := json.Marshal(resp)
	if err != nil {
		slog.Error("json.Marshal",
			slog.String("severity", "ERROR"),
			slog.String("message", err.Error()),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
}

func SaveVideos(w http.ResponseWriter, r *http.Request) {
	// 初期化処理
	sqldb, err := sql.Open("mysql", os.Getenv("DSN"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	ctx := context.Background()
	db := bun.NewDB(sqldb, mysqldialect.New())
	yt, err := youtube.NewService(ctx, option.WithAPIKey(os.Getenv("YOUTUBE_API_KEY")))
	if err != nil {
		slog.Error("youtube.NewService",
			slog.String("severity", "ERROR"),
			slog.String("message", err.Error()),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var b ChannelBody
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// list が空のときにインサートエラーが発生してしまうため、その回避処理
	if len(b.CIDList) == 0 {
		resp := VideoBody{[]youtube.Video{}}
		bytes, _ := json.Marshal(resp)
		w.Header().Set("Content-Type", "application/json")
		w.Write(bytes)
		return
	}

	vidList, err := CustomPlaylistItems(yt, b.CIDList)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		slog.Error("CustomPlaylistItems",
			slog.String("severity", "ERROR"),
			slog.String("message", err.Error()),
		)
		return
	}
	vlist, err := CustomVideo(yt, vidList)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		slog.Error("CustomVideo",
			slog.String("severity", "ERROR"),
			slog.String("message", err.Error()),
		)
		return
	}

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
		})
	}

	_, err = db.NewInsert().
		Model(&Videos).
		Ignore().
		Exec(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		slog.Error("NewInsert Videos",
			slog.String("severity", "ERROR"),
			slog.String("message", err.Error()),
		)
		return
	}

	resp := VideoBody{vlist}
	bytes, err := json.Marshal(resp)
	if err != nil {
		slog.Error("json.Marshal",
			slog.String("severity", "ERROR"),
			slog.String("message", err.Error()),
		)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
}

func SongVideoAnnounce(w http.ResponseWriter, r *http.Request) {
	// 初期化処理
	sqldb, err := sql.Open("mysql", os.Getenv("DSN"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	ctx := context.Background()
	db := bun.NewDB(sqldb, mysqldialect.New())
	app, err := firebase.NewApp(ctx, nil)
	if err != nil {
		slog.Error("firebase.NewApp error",
			slog.String("severity", "ERROR"),
			slog.String("message", err.Error()),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	client, err := app.Messaging(ctx)
	if err != nil {
		slog.Error("app.Messaging error",
			slog.String("severity", "ERROR"),
			slog.String("message", err.Error()),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	now, _ := time.Parse(time.RFC3339, time.Now().UTC().Format("2006-01-02T15:04:00Z"))
	tAfter := now.Add(1 * time.Second)
	tBefore := now.Add(5 * time.Minute)
	var videos []Video
	err = db.NewSelect().Model(&videos).Where("? BETWEEN ? AND ?", bun.Ident("scheduled_start_time"), tAfter, tBefore).Scan(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// tw := NewTwitter()
	// mk := NewMisskey(os.Getenv("MISSKEY_TOKEN"))

	var tokens []string
	err = db.NewSelect().Model((*User)(nil)).Column("token").Scan(ctx, &tokens)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
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
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
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
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
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
	w.Write([]byte("OK!!"))
}

func KeywordAnnounce(w http.ResponseWriter, r *http.Request) {
	var reqBody VideoBody
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	for _, v := range reqBody.Vlist {
		if regexp.MustCompile(`.*Lethal Company.*`).Match([]byte(v.Snippet.Title)) {
			err := SendMail("Lethal Company やるよ！", fmt.Sprintf("https://www.youtube.com/watch?v=%s", v.Id))
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
	}
}

// func DeleteVideos(w http.ResponseWriter, r *http.Request) {
// 	// 初期化処理
// 	sqldb, err := sql.Open("mysql", os.Getenv("DSN"))
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}
// 	ctx := context.Background()
// 	db := bun.NewDB(sqldb, mysqldialect.New())
// 	yt, err := youtube.NewService(ctx, option.WithAPIKey(os.Getenv("YOUTUBE_API_KEY")))
// 	if err != nil {
// 		slog.Error("youtube.NewService",
// 			slog.String("severity", "ERROR"),
// 			slog.String("message", err.Error()),
// 		)
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}
// }
