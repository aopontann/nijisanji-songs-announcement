package youtube

import (
	"context"
	"database/sql"
	"os"
	"strings"

	ndb "github.com/aopontann/nijisanji-songs-announcement/db"
	"github.com/rs/zerolog/log"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

type Youtube struct {
	service *youtube.Service
	queries *ndb.Queries
}

func New(db *sql.DB) (*Youtube, error) {
	ctx := context.Background()
	s, err := youtube.NewService(ctx, option.WithAPIKey(os.Getenv("YOUTUBE_API_KEY")))
	if err != nil {
		log.Fatal().Err(err).Msg("youtube.NewService create failed")
		return nil, err
	}

	queries := ndb.New(db)

	return &Youtube{
		service: s,
		queries: queries,
	}, nil
}

// プレイリストIDとキー、プレイリストに含まれている動画数を値とした連想配列を返す
func (yt *Youtube) Playlists() (map[string]int64, error) {
	list := make(map[string]int64, 500)
	ctx := context.Background()
	pidList, err := yt.queries.ListPlaylistID(ctx)
	if err != nil {
		return nil, err
	}

	for i := 0; i*50 <= len(pidList); i++ {
		var id string
		if len(pidList) > 50*(i+1) {
			id = strings.Join(pidList[50*i:50*(i+1)], ",")
		} else {
			id = strings.Join(pidList[50*i:], ",")
		}
		call := yt.service.Playlists.List([]string{"snippet", "contentDetails"}).MaxResults(50).Id(id)
		res, err := call.Do()
		if err != nil {
			log.Error().Str("severity", "ERROR").Err(err).Msg("playlists-list call error")
			return nil, err
		}

		for _, item := range res.Items {
			list[item.Id] = item.ContentDetails.ItemCount
			log.Info().
				Str("severity", "INFO").
				Str("service", "youtube-playlists-list").
				Str("PlaylistId", item.Id).
				Int64("ItemCount", item.ContentDetails.ItemCount).
				Send()
		}
	}
	return list, nil
}

// DBに保存されているプレイリストの動画数とAPIから取得したプレイリストの動画数を比較し、動画数が変わっているプレイリストIDを返す
func (yt *Youtube) CheckItemCount(list map[string]int64) ([]string, error) {
	// 動画数が変わっているプレイリストIDを入れる配列
	var pidList []string
	ctx := context.Background()
	itemCountList, err := yt.queries.ListItemCount(ctx)
	if err != nil {
		return nil, err
	}

	for _, row := range itemCountList {
		value, isThere := list[row.ID]
		if isThere && value != int64(row.ItemCount) {
			pidList = append(pidList, row.ID)
		}
	}

	return pidList, nil
}

// プレイリストに含まれている動画IDを取得する
func (yt *Youtube) Items(pidList []string) ([]string, error) {
	// 動画IDを格納する文字列型配列を宣言
	vidList := make([]string, 0, 1500)

	for _, pid := range pidList {
		// 取得した動画IDをログに出力するための変数
		var rid []string
		call := yt.service.PlaylistItems.List([]string{"snippet"}).PlaylistId(pid).MaxResults(3)
		res, err := call.Do()
		if err != nil {
			log.Error().Str("severity", "ERROR").Err(err).Msg("playlistitems-list call error")
			return []string{}, err
		}

		for _, item := range res.Items {
			rid = append(rid, item.Snippet.ResourceId.VideoId)
			vidList = append(vidList, item.Snippet.ResourceId.VideoId)
		}

		log.Info().
			Str("severity", "INFO").
			Str("service", "youtube-playlistitems-list").
			Str("PlaylistId", pid).
			Strs("videoId", rid).
			Send()
	}
	return vidList, nil
}

// Youtube Data API から動画情報を取得
func (yt *Youtube) Video(vidList []string) ([]youtube.Video, error) {
	var rlist []youtube.Video
	for i := 0; i*50 <= len(vidList); i++ {
		var id string
		if len(vidList) > 50*(i+1) {
			id = strings.Join(vidList[50*i:50*(i+1)], ",")
		} else {
			id = strings.Join(vidList[50*i:], ",")
		}
		call := yt.service.Videos.List([]string{"snippet", "contentDetails", "liveStreamingDetails"}).Id(id).MaxResults(50)
		res, err := call.Do()
		if err != nil {
			log.Error().Str("severity", "ERROR").Err(err).Msg("videos-list call error")
			return nil, err
		}

		for _, video := range res.Items {
			scheduledStartTime := "" // 例 2022-03-28T11:00:00Z
			if video.LiveStreamingDetails != nil {
				// "2022-03-28 11:00:00"形式に変換
				rep1 := strings.Replace(video.LiveStreamingDetails.ScheduledStartTime, "T", " ", 1)
				scheduledStartTime = strings.Replace(rep1, "Z", "", 1)
			}

			rlist = append(rlist, *video)

			log.Info().
				Str("severity", "INFO").
				Str("service", "youtube-video-list").
				Str("id", video.Id).
				Str("title", video.Snippet.Title).
				Str("duration", video.ContentDetails.Duration).
				Str("schedule", scheduledStartTime).
				Str("channel_id", video.Snippet.ChannelId).
				Send()
		}
	}
	return rlist, nil
}
