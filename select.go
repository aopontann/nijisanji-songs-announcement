package nsa

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"

	ndb "github.com/aopontann/nijisanji-songs-announcement/db"
	"github.com/rs/zerolog/log"
	"google.golang.org/api/youtube/v3"
)

type Video struct {
	list    VideoList
	queries *ndb.Queries
}

type VideoList map[string]youtube.Video

func NewSelect(vList []youtube.Video, db *sql.DB) *Video {
	m := make(map[string]youtube.Video, 1500)
	for _, v := range vList {
		m[v.Id] = v
	}

	queries := ndb.New(db)

	return &Video{
		list:    m,
		queries: queries,
	}
}

func (v *Video) GetList() VideoList {
	return v.list
}

// 動画リストをログ表示する
func (v *Video) Log(service string) *Video {
	for _, video := range v.list {
		log.Info().
			Str("severity", "INFO").
			Str("service", service).
			Str("id", video.Id).
			Str("title", video.Snippet.Title).
			Str("duration", video.ContentDetails.Duration).
			Str("schedule", video.LiveStreamingDetails.ScheduledStartTime).
			Send()
	}
	return v
}

// プレミア公開する動画か
func (v *Video) IsLiveStreaming() *Video {
	for _, video := range v.list {
		if video.LiveStreamingDetails == nil {
			delete(v.list, video.Id)
		}
	}
	return v
}

// 放送が終了していないか
func (v *Video) IsNotLiveFinished() *Video {
	for _, video := range v.list {
		if video.Snippet.LiveBroadcastContent == "none" {
			delete(v.list, video.Id)
		}
	}
	return v
}

// 動画の長さが10分未満か
func (v *Video) IsUnder10min() *Video {
	for _, video := range v.list {
		if !regexp.MustCompile(`^PT([1-9]M[1-5]?[0-9]S|[1-5]?[0-9]S)`).Match([]byte(video.ContentDetails.Duration)) {
			delete(v.list, video.Id)
		}
	}
	return v
}

// 切り抜き動画ではないか
func (v *Video) IsNotClipped() *Video {
	for _, video := range v.list {
		if regexp.MustCompile(`.*切り抜き.*`).Match([]byte(video.Snippet.Title)) {
			delete(v.list, video.Id)
		}
	}
	return v
}

// タイトルに特定のキーワードが含まれているか（この関数はまだ使わない）
func (v *Video) IsExistsKeyword() *Video {
	for _, video := range v.list {
		var flag = false
		for _, cstr := range definiteList {
			reg := fmt.Sprintf(`.*%s.*`, cstr)
			if regexp.MustCompile(reg).Match([]byte(video.Snippet.Title)) {
				flag = true
				break
			}
		}
		// 特定のキーワードが含まれていなかった場合
		if !flag {
			delete(v.list, video.Id)
		}
	}
	return v
}

// にじさんじライバーのチャンネルでアップロードされた動画か
func (v *Video) IsNijisanji() (*Video, error) {
	ctx := context.Background()
	chIdList, err := v.queries.ListVtuberID(ctx)
	if err != nil {
		return nil, err
	}

	for _, video := range v.list {
		var flag = false
		for _, cid := range chIdList {
			if video.Snippet.ChannelId == cid {
				flag = true
				break
			}
		}
		// にじさんじライバーのチャンネルでアップロードされた動画でなかった場合
		if !flag {
			delete(v.list, video.Id)
		}
	}
	return v, nil
}

// DBに保存されていない動画か
func (v *Video) IsNotExists() (*Video, error) {
	var vidList []string
	ctx := context.Background()
	for _, video := range v.list {
		vidList = append(vidList, video.Id)
	}

	// DBに保存されていた動画IDリストを取得
	existsVidList, err := v.queries.ExistsVideos(ctx, vidList)
	if err != nil {
		return nil, err
	}

	for _, vid := range existsVidList {
		delete(v.list, vid)
	}
	return v, nil
}

// タイトルにこの文字が含まれていると歌動画確定
var definiteList = []string{
	"歌ってみた",
	"歌わせていただきました",
	"歌って踊ってみた",
	"cover",
	"Cover",
	"COVER",
	"MV",
	"Music Video",
	// "ソング",
	// "song",
	"オリジナル曲",
	"オリジナルMV",
	"Official Lyric Video",
}
