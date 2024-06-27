package nsa

import (
	"context"
	"log/slog"
	"regexp"
	"strings"
	"time"

	"github.com/uptrace/bun"
	"google.golang.org/api/youtube/v3"
)

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
	Thumbnail string    `bun:"thumbnail,notnull,type:varchar"`
	CreatedAt time.Time `bun:"created_at,type:TIMESTAMP(0),nullzero,notnull,default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `bun:"updated_at,type:TIMESTAMP(0),nullzero,notnull,default:CURRENT_TIMESTAMP"`
}

type User struct {
	bun.BaseModel `bun:"table:users"`

	Token       string    `json:"token" bun:"token,type:varchar(1000),pk"`
	Song        bool      `json:"song" bun:"song,default:false,type:boolean"`
	Keyword     bool      `json:"keyword" bun:"keyword,default:false,type:boolean"`
	KeywordText string    `json:"keyword_text" bun:"keyword_text,type:varchar(100)"`
	CreatedAt   time.Time `bun:"created_at,type:TIMESTAMP(0),nullzero,notnull,default:CURRENT_TIMESTAMP"`
	UpdatedAt   time.Time `bun:"updated_at,type:TIMESTAMP(0),nullzero,notnull,default:CURRENT_TIMESTAMP"`
}

type DB struct {
	Service *bun.DB
}

func NewDB(db *bun.DB) *DB {
	return &DB{db}
}

// DBに登録されているPlaylistsの動画数を取得
// 返り値：map （キー：プレイリストID　値：動画数）
func (db *DB) Playlists() (map[string]int64, error) {
	// DBからチャンネルID、チャンネルごとの動画数を取得
	var ids []string
	var itemCount []int64
	ctx := context.Background()
	err := db.Service.NewSelect().Model((*Vtuber)(nil)).Column("id", "item_count").Scan(ctx, &ids, &itemCount)
	if err != nil {
		return nil, err
	}

	list := make(map[string]int64, 500)
	for i := range ids {
		pid := strings.Replace(ids[i], "UC", "UU", 1)
		list[pid] = itemCount[i]
	}

	return list, nil
}

func (db *DB) UpdatePlaylistItem(tx bun.Tx, newlist map[string]int64) error {
	ctx := context.Background()
	// DBを新しく取得したデータに更新
	var updateVideo []Vtuber
	for pid, v := range newlist {
		cid := strings.Replace(pid, "UU", "UC", 1)
		updateVideo = append(updateVideo, Vtuber{ID: cid, ItemCount: v, UpdatedAt: time.Now()})
	}

	if len(updateVideo) == 0 {
		return nil
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

func (db *DB) SaveVideos(tx bun.Tx, vlist []youtube.Video) error {
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
			Thumbnail: v.Snippet.Thumbnails.High.Url,
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

// 5分後にプレミア公開される動画を取得
func (db *DB) songVideos5m() ([]Video, error) {
	ctx := context.Background()
	now, _ := time.Parse(time.RFC3339, time.Now().UTC().Format("2006-01-02T15:04:00Z"))
	tAfter := now.Add(1 * time.Second)
	tBefore := now.Add(5 * time.Minute)
	var videos []Video
	err := db.Service.NewSelect().Model(&videos).Where("? BETWEEN ? AND ?", bun.Ident("scheduled_start_time"), tAfter, tBefore).Scan(ctx)
	if err != nil {
		return nil, err
	}

	if len(videos) == 0 {
		return []Video{}, nil
	}

	var filtedVideos []Video

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
		filtedVideos = append(filtedVideos, v)
	}

	return filtedVideos, nil
}

func (db *DB) getSongTokens() ([]string, error) {
	// DBからチャンネルID、チャンネルごとの動画数を取得
	var tokens []string
	ctx := context.Background()
	err := db.Service.NewSelect().Model((*User)(nil)).Column("token").Where("song = true").Scan(ctx, &tokens)
	if err != nil {
		return nil, err
	}

	return tokens, nil
}

func (db *DB) getKeywordTextList() ([]string, error) {
	// DBからチャンネルID、チャンネルごとの動画数を取得
	var list []string
	ctx := context.Background()
	err := db.Service.NewSelect().Model((*User)(nil)).Column("keyword_text").Where("keyword = true").Where("keyword_text != ''").Group("keyword_text").Scan(ctx, &list)
	if err != nil {
		return nil, err
	}

	return list, nil
}

// func (db *DB) GetKeywordRegisterList() ([]Result, error) {

// 	url := os.Getenv("D1_URL")
// 	method := "POST"
// 	token := os.Getenv("D1_TOKEN")

// 	payload := strings.NewReader(`
//   {
//     "sql": "SELECT * FROM users WHERE not word = '';"
//   }`)

// 	client := &http.Client{}
// 	req, err := http.NewRequest(method, url, payload)

// 	if err != nil {
// 		fmt.Println(err)
// 		return nil, err
// 	}
// 	req.Header.Add("Content-Type", "application/json")
// 	req.Header.Add("Authorization", "Bearer "+token)

// 	res, err := client.Do(req)
// 	if err != nil {
// 		fmt.Println(err)
// 		return nil, err
// 	}
// 	defer res.Body.Close()

// 	s := &D1Response{}
// 	body, err := io.ReadAll(res.Body)
// 	if err != nil {
// 		fmt.Println(err)
// 		return nil, err
// 	}
// 	json.Unmarshal(body, s)

// 	slog.Info(
// 		"get-keyword-d1",
// 		slog.String("severity", "INFO"),
// 		slog.Any("res", s),
// 		slog.String("status_code", res.Status),
// 	)

// 	return s.Result[0].Results, nil
// }
