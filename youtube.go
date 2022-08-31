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
