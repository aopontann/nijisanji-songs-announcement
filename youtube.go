package nsa

import (
	"log/slog"
	"strings"

	"google.golang.org/api/youtube/v3"
)

// チャンネルIDをキー、プレイリストに含まれている動画数を値とした連想配列を返す
func CustomPlaylists(yt *youtube.Service, plist []string) (map[string]int64, error) {
	newlist := make(map[string]int64, 500)
	for i := 0; i*50 <= len(plist); i++ {
		var id string
		if len(plist) > 50*(i+1) {
			id = strings.Join(plist[50*i:50*(i+1)], ",")
		} else {
			id = strings.Join(plist[50*i:], ",")
		}
		call := yt.Playlists.List([]string{"snippet", "contentDetails"}).MaxResults(50).Id(id)
		res, err := call.Do()
		if err != nil {
			slog.Error("playlists-list",
				slog.String("severity", "ERROR"),
				slog.String("message", err.Error()),
			)
			return nil, err
		}

		for _, item := range res.Items {
			newlist[item.Snippet.ChannelId] = item.ContentDetails.ItemCount
			slog.Debug("youtube-playlists-list",
				slog.String("severity", "DEBUG"),
				slog.String("PlaylistId", item.Id),
				slog.Int64("ItemCount", item.ContentDetails.ItemCount),
			)
		}
	}
	return newlist, nil
}

func CustomPlaylistItems(yt *youtube.Service, clist []string) ([]string, error) {
	// 動画IDを格納する文字列型配列を宣言
	vidList := make([]string, 0, 1500)

	for _, cid := range clist {
		// 取得した動画IDをログに出力するための変数
		var rid []string
		pid := strings.Replace(cid, "UC", "UU", 1)
		call := yt.PlaylistItems.List([]string{"snippet"}).PlaylistId(pid).MaxResults(1)
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
func CustomVideo(yt *youtube.Service, vidList []string) ([]youtube.Video, error) {
	var rlist []youtube.Video
	for i := 0; i*50 <= len(vidList); i++ {
		var id string
		if len(vidList) > 50*(i+1) {
			id = strings.Join(vidList[50*i:50*(i+1)], ",")
		} else {
			id = strings.Join(vidList[50*i:], ",")
		}
		call := yt.Videos.List([]string{"snippet", "contentDetails", "liveStreamingDetails"}).Id(id).MaxResults(50)
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
