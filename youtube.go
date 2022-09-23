package main

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

type Youtube struct{}

type YoutubeVideoResponse struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Schedule    string `json:"schedule"`
	Duration    string `json:"duration"`
	ChannelID   string `json:"channel_id"`
}

type YoutubeSelectResponse struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Schedule    string `json:"schedule"`
	SongConfirm bool   `json:"song_confirm"`
}

type YouTubeCheckResponse struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Schedule  string `json:"schedule"`
	TwitterID string `json:"twitter_id"`
}

// Youtube Data API Search List を叩いて検索結果の動画IDを取得
func (yt *Youtube) Search() ([]string, error) {
	// 動画検索範囲
	dtAfter := time.Now().UTC().Add(-120 * time.Minute).Format("2006-01-02T15:04:00Z")
	dtBefore := time.Now().UTC().Add(-60 * time.Minute).Format("2006-01-02T15:04:00Z")
	// 動画IDを格納する文字列型配列を宣言
	vid := make([]string, 0, 600)

	for _, q := range []string{"にじさんじ", "NIJISANJI"} {
		pt := ""
		for {
			// youtube data api search list にリクエストを送る
			call := YoutubeService.Search.List([]string{"id"}).MaxResults(50).Q(q).PublishedAfter(dtAfter).PublishedBefore(dtBefore).PageToken(pt)
			res, err := call.Do()
			if err != nil {
				log.Error().Str("severity", "ERROR").Err(err).Msg("search-list call error")
				return []string{}, err
			}

			for _, item := range res.Items {
				vid = append(vid, item.Id.VideoId)
			}

			log.Info().
				Str("severity", "INFO").
				Str("service", "youtube-search-list").
				Str("published", fmt.Sprintf("%s ~ %s", dtAfter, dtBefore)).
				Str("q", q).
				Str("pageInfo", fmt.Sprintf("perPage=%d total=%d nextPage=%s\n", res.PageInfo.ResultsPerPage, res.PageInfo.TotalResults, res.NextPageToken)).
				Strs("id", vid).
				Send()

			if res.NextPageToken == "" {
				break
			}
			pt = res.NextPageToken
		}
	}
	return vid, nil
}

// Youtube Data API から動画情報を取得
func (yt *Youtube) Video(vid []string) ([]YoutubeVideoResponse, error) {
	var yvs []YoutubeVideoResponse
	for i := 0; i*50 <= len(vid); i++ {
		var id string
		if len(vid) > 50*(i+1) {
			id = strings.Join(vid[50*i:50*(i+1)], ",")
		} else {
			id = strings.Join(vid[50*i:], ",")
		}
		call := YoutubeService.Videos.List([]string{"snippet", "contentDetails", "liveStreamingDetails"}).Id(id).MaxResults(50)
		res, err := call.Do()
		if err != nil {
			log.Error().Str("severity", "ERROR").Err(err).Msg("videos-list call error")
			return []YoutubeVideoResponse{}, err
		}

		for _, video := range res.Items {
			scheduledStartTime := "" // 例 2022-03-28T11:00:00Z
			if video.LiveStreamingDetails != nil {
				// "2022-03-28 11:00:00"形式に変換
				rep1 := strings.Replace(video.LiveStreamingDetails.ScheduledStartTime, "T", " ", 1)
				scheduledStartTime = strings.Replace(rep1, "Z", "", 1)
			}

			yvs = append(yvs, YoutubeVideoResponse{ID: video.Id, Title: video.Snippet.Title, Description: video.Snippet.Description, Schedule: scheduledStartTime, Duration: video.ContentDetails.Duration, ChannelID: video.Snippet.ChannelId})

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
	return yvs, nil
}

// 歌ってみた動画かフィルターをかける処理
func (yt *Youtube) Select(yvs []YoutubeVideoResponse) ([]YoutubeSelectResponse, error) {
	var ysr []YoutubeSelectResponse
	// にじさんじライバーのチャンネルリストを取得
	channelIdList, err := GetChannelIdList()
	if err != nil {
		return nil, err
	}
	// 歌動画か判断する
	for _, video := range yvs {
		// プレミア公開する動画か
		if video.Schedule == "" {
			continue
		}
		// 動画の長さが9分59秒以下ではない場合
		if !regexp.MustCompile(`^PT([1-9]M[1-5]?[0-9]S|[1-5]?[0-9]S)`).Match([]byte(video.Duration)) {
			continue
		}
		// 切り抜き動画である場合
		if regexp.MustCompile(`.*切り抜き.*`).Match([]byte(video.Title)) {
			continue
		}
		// 動画タイトルに特定の文字が含まれているか
		// checkRes := TitleCheck(video.Title)
		checkRes := true

		// にじさんじライバーのチャンネルで公開されたか
		if !NijisanjiCheck(channelIdList, video.ChannelID) {
			continue
		}

		ysr = append(ysr, YoutubeSelectResponse{ID: video.ID, Title: video.Title, Schedule: video.Schedule, SongConfirm: checkRes})

		log.Info().
			Str("severity", "INFO").
			Str("service", "youtube-video-list").
			Str("id", video.ID).
			Str("title", video.Title).
			Str("duration", video.Duration).
			Str("schedule", video.Schedule).
			Send()
	}
	return ysr, nil
}

// 動画情報をDBに保存
func (yt *Youtube) Save(ysr []YoutubeSelectResponse) error {
	tx, err := DB.Begin()
	if err != nil {
		log.Error().Str("severity", "ERROR").Err(err).Msg("DB.Begin error")
		return err
	}
	// DB準備
	stmt, err := tx.Prepare("INSERT IGNORE INTO videos(id, title, songConfirm, scheduled_start_time) VALUES(?,?,?,?)")
	if err != nil {
		log.Error().Str("severity", "ERROR").Err(err).Msg("DB.Prepare error")
		return err
	}

	for _, video := range ysr {
		// DBに動画情報を保存
		_, err := stmt.Exec(video.ID, video.Title, video.SongConfirm, video.Schedule)
		if err != nil {
			log.Error().Str("severity", "ERROR").Err(err).Msg("Save videos failed")
			if tx.Rollback() != nil {
				log.Error().Str("severity", "ERROR").Err(err).Msg("Rollback error")
			}
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Error().Str("severity", "ERROR").Err(err).Msg("tx.Commit error")
		return err
	}
	return nil
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
