package nsa

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

type Youtube struct {
	Service *youtube.Service
}

func NewYoutube(key string) *Youtube {
	ctx := context.Background()
	yt, err := youtube.NewService(ctx, option.WithAPIKey(key))
	if err != nil {
		panic(err)
	}
	return &Youtube{yt}
}

// チャンネルIDをキー、プレイリストに含まれている動画数を値とした連想配列を返す
func (y *Youtube) Playlists(plist []string) (map[string]int64, error) {
	newlist := make(map[string]int64, 500)
	for i := 0; i*50 <= len(plist); i++ {
		var id string
		if len(plist) > 50*(i+1) {
			id = strings.Join(plist[50*i:50*(i+1)], ",")
		} else {
			id = strings.Join(plist[50*i:], ",")
		}
		call := y.Service.Playlists.List([]string{"snippet", "contentDetails"}).MaxResults(50).Id(id)
		res, err := call.Do()
		if err != nil {
			slog.Error("playlists-list",
				slog.String("severity", "ERROR"),
				slog.String("message", err.Error()),
			)
			return nil, err
		}

		for _, item := range res.Items {
			newlist[item.Id] = item.ContentDetails.ItemCount
			slog.Debug("youtube-playlists-list",
				slog.String("severity", "DEBUG"),
				slog.String("PlaylistId", item.Id),
				slog.Int64("ItemCount", item.ContentDetails.ItemCount),
			)
		}
	}
	return newlist, nil
}

func (y *Youtube) PlaylistItems(plist []string) ([]string, error) {
	// 動画IDを格納する文字列型配列を宣言
	vidList := make([]string, 0, 1500)

	for _, pid := range plist {
		// 取得した動画IDをログに出力するための変数
		fmt.Println(pid)
		var rid []string
		call := y.Service.PlaylistItems.List([]string{"snippet"}).PlaylistId(pid).MaxResults(3)
		res, err := call.Do()
		if err != nil {
			slog.Error("playlistitems-list",
				slog.String("severity", "ERROR"),
				slog.String("message", err.Error()),
			)
			return []string{}, err
		}

		for _, item := range res.Items {
			rid = append(rid, item.Snippet.ResourceId.VideoId)
			vidList = append(vidList, item.Snippet.ResourceId.VideoId)
		}

		slog.Debug("youtube-playlistitems-list",
			slog.String("severity", "DEBUG"),
			slog.String("PlaylistId", pid),
			slog.String("videoId", strings.Join(rid, ",")),
		)
	}
	return vidList, nil
}

// Youtube Data API から動画情報を取得
func (y *Youtube) Videos(vidList []string) ([]youtube.Video, error) {
	var rlist []youtube.Video
	for i := 0; i*50 <= len(vidList); i++ {
		var id string
		if len(vidList) > 50*(i+1) {
			id = strings.Join(vidList[50*i:50*(i+1)], ",")
		} else {
			id = strings.Join(vidList[50*i:], ",")
		}
		call := y.Service.Videos.List([]string{"snippet", "contentDetails", "liveStreamingDetails"}).Id(id).MaxResults(50)
		res, err := call.Do()
		if err != nil {
			slog.Error("videos-list",
				slog.String("severity", "ERROR"),
				slog.String("message", err.Error()),
			)
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

			slog.Info("youtube-video-list",
				slog.String("severity", "INFO"),
				slog.String("id", video.Id),
				slog.String("title", video.Snippet.Title),
				slog.String("duration", video.ContentDetails.Duration),
				slog.String("schedule", scheduledStartTime),
				slog.String("channel_id", video.Snippet.ChannelId),
			)
		}
	}
	return rlist, nil
}

// 放送前、放送中のプレミア動画、ライプ　普通動画の公開直後の動画に絞る
func (y *Youtube) FilterVideos(vlist []youtube.Video) []youtube.Video {
	var filtedVideoList []youtube.Video
	for _, v := range vlist {
		// プレミア公開、生放送終了した動画
		if v.LiveStreamingDetails != nil && v.Snippet.LiveBroadcastContent == "none" {
			continue
		}

		// プレミア公開、生放送をする予定、している最中の動画
		if v.Snippet.LiveBroadcastContent != "none" {
			filtedVideoList = append(filtedVideoList, v)
			continue
		}

		// プレミア公開、生放送ではなく、10分以内に公開された動画
		t, _ := time.Parse("2006-01-02T15:04:05Z", v.Snippet.PublishedAt)
		if time.Now().UTC().Add(-10*time.Minute).Compare(t) < 0 {
			filtedVideoList = append(filtedVideoList, v)
			continue
		}
	}

	return filtedVideoList
}

// 歌ってみた動画のタイトルによく含まれるキーワードが 指定した動画に含まれているか
func (y *Youtube) FindSongKeyword(video youtube.Video) bool {
	for _, word := range getSongWordList() {
		if strings.Contains(strings.ToLower(video.Snippet.Title), strings.ToLower(word)) {
			return true
		}
	}
	return false
}

// 無視するキーワードが 指定した動画に含まれているか
func (y *Youtube) FindIgnoreKeyword(video youtube.Video) bool {
	for _, word := range []string{"切り抜き", "ラジオ"} {
		if strings.Contains(strings.ToLower(video.Snippet.Title), strings.ToLower(word)) {
			return true
		}
	}
	return false
}