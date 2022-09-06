package main

import (
	"regexp"
	"strings"

	"github.com/rs/zerolog/log"
)

type Youtube struct{}

type YouTubeCheckResponse struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Schedule  string `json:"schedule"`
	TwitterID string `json:"twitter_id"`
}

// 動画IDから動画詳細情報を取得して歌動画か判断する
func (yt *Youtube) Check(tsr []TwitterSearchResponse) ([]YouTubeCheckResponse, error) {
	var ytcr []YouTubeCheckResponse
	var id []string
	yttw := make(map[string]string)
	for _, v := range tsr {
		id = append(id, v.YouTubeID)
		yttw[v.YouTubeID] = v.ID
	}

	call := YoutubeService.Videos.List([]string{"snippet", "contentDetails", "liveStreamingDetails"}).Id(strings.Join(id, ",")).MaxResults(50)
	res, err := call.Do()
	if err != nil {
		log.Error().Str("severity", "ERROR").Err(err).Msg("videos-list call error")
		return nil, err
	}

	// 歌動画か判断する
	for _, video := range res.Items {
		// プレミア公開する動画か
		scheduledStartTime := "" // 例 2022-03-28T11:00:00Z
		if video.LiveStreamingDetails != nil {
			// "2022-03-28 11:00:00"形式に変換
			rep1 := strings.Replace(video.LiveStreamingDetails.ScheduledStartTime, "T", " ", 1)
			scheduledStartTime = strings.Replace(rep1, "Z", "", 1)
		} else {
			continue
		}
		// 動画の長さが9分59秒以下ではない場合
		if !regexp.MustCompile(`^PT([1-9]M[1-5]?[0-9]S|[1-5]?[0-9]S)`).Match([]byte(video.ContentDetails.Duration)) {
			continue
		}
		// 切り抜き動画である場合
		if regexp.MustCompile(`.*切り抜き.*`).Match([]byte(video.Snippet.Title)) {
			continue
		}
		// タイトルに特定の文字列が含まれているか
		if regexp.MustCompile(`.*cover|Cover|歌って|MV.*`).Match([]byte(video.Snippet.Title)) {
			ytcr = append(ytcr, YouTubeCheckResponse{ID: video.Id, Title: video.Snippet.Title, Schedule: video.LiveStreamingDetails.ScheduledStartTime, TwitterID: yttw[video.Id]})

			log.Info().
				Str("severity", "INFO").
				Str("service", "youtube-video-check").
				Str("id", video.Id).
				Str("title", video.Snippet.Title).
				Str("duration", video.ContentDetails.Duration).
				Str("schedule", scheduledStartTime).
				Send()
			continue
		}
		// 動画概要欄に特定の文字が含まれているか
		if !regexp.MustCompile(`.*vocal|Vocal|song|Song|歌|MV.*`).Match([]byte(video.Snippet.Description)) {
			continue
		}

		ytcr = append(ytcr, YouTubeCheckResponse{ID: video.Id, Title: video.Snippet.Title, Schedule: video.LiveStreamingDetails.ScheduledStartTime, TwitterID: yttw[video.Id]})

		log.Info().
			Str("severity", "INFO").
			Str("service", "youtube-video-check").
			Str("id", video.Id).
			Str("title", video.Snippet.Title).
			Str("duration", video.ContentDetails.Duration).
			Str("schedule", scheduledStartTime).
			Send()
	}
	return ytcr, nil
}

// 動画IDからAPIを叩いて動画情報を取得し、DBに保存されている動画情報と異なるデータがある場合、新しい情報に上書きする
// 上書きした場合やデータを削除した場合は true を返す
func (yt *Youtube) CheckVideo(vid string) (bool, error) {
	var title string
	var scheTime string
	// エラー表示に使うカスタムログの設定
	clog := log.Error().Str("severity", "ERROR").Str("service", "youtube-check-video")
	// Youtube Data API にリクエスト送信して動画情報を取得
	call := YoutubeService.Videos.List([]string{"snippet", "contentDetails", "liveStreamingDetails"}).Id(vid).MaxResults(1)
	res, err := call.Do()
	if err != nil {
		clog.Err(err).Msg("videos-list call error")
		return false, err
	}
	// 動画が削除されていた場合、DBに保存されている動画情報も削除する
	if len(res.Items) == 0 {
		_, err := DB.Exec("DELETE FROM videos WHERE id = ?", vid)
		if err != nil {
			clog.Err(err).Msg("delete video error")
			return false, err
		}
		return true, nil
	}
	// DBから動画情報を取得
	err = DB.QueryRow("SELECT title, scheduled_start_time FROM videos WHERE id = ?", vid).Scan(&title, &scheTime)
	if err != nil {
		clog.Err(err).Msg("select video error")
		return false, err
	}
	// DBに保存されている動画情報がAPIから取得したデータと異なる場合
	if title != res.Items[0].Snippet.Title || scheTime != res.Items[0].LiveStreamingDetails.ScheduledStartTime {
		// 新しい動画情報に上書きする
		_, err := DB.Exec("UPDATE videos SET title = ?, scheduled_start_time = ? WHERE id = ?", title, scheTime, vid)
		if err != nil {
			clog.Err(err).Msg("Failed to overwrite")
			return false, err
		}
		return true, nil
	}
	// DBに保存されている動画情報がAPIから取得したデータと同じ場合
	return false, nil
}
