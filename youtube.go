package main

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

type Youtube struct{}

type YouTubeVideoResponse struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Schedule    string `json:"schedule"`
	Duration    string `json:"duration"`
	ChannelID   string `json:"channel_id"`
}

type YouTubeSelectResponse struct {
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

type YouTubeChannelsResponse struct {
	ID         string `json:"id"`
	VideoCount uint64 `json:"video_count"`
}

type NewUploadChannelList struct {
	ID            string `json:"id"`
	NewVideoCount uint64 `json:"new_video_count"`
	OldVideoCount uint64 `json:"old_video_count"`
}

type VideoIDList []string
type YTVRList []YouTubeVideoResponse
type YTSRList []YouTubeSelectResponse

// チャンネルIDと最新の動画数を格納する配列の型
type YTCRList []YouTubeChannelsResponse

// 新しく動画をアップロードしたチャンネルIDと古い動画と新しい動画の数を格納する配列の型
type NewUpChList []NewUploadChannelList

var YouTubeService *youtube.Service

// Youtube Data API の 初期化処理
func YTNew() {
	var err error
	ctx := context.Background()
	YouTubeService, err = youtube.NewService(ctx, option.WithAPIKey(os.Getenv("YOUTUBE_API_KEY")))
	if err != nil {
		log.Fatal().Err(err).Msg("youtube.NewService create failed")
	}
}

// Youtube Data API Search List を叩いて検索結果の動画IDを取得
func Search() (VideoIDList, error) {
	// 動画検索範囲
	dtAfter := time.Now().UTC().Add(-120 * time.Minute).Format("2006-01-02T15:04:00Z")
	dtBefore := time.Now().UTC().Add(-60 * time.Minute).Format("2006-01-02T15:04:00Z")
	// 動画IDを格納する文字列型配列を宣言
	vid := make([]string, 0, 600)

	for _, q := range []string{"にじさんじ", "NIJISANJI"} {
		pt := ""
		for {
			// youtube data api search list にリクエストを送る
			call := YouTubeService.Search.List([]string{"id"}).MaxResults(50).Q(q).PublishedAfter(dtAfter).PublishedBefore(dtBefore).PageToken(pt)
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
func (list VideoIDList) Video() (YTVRList, error) {
	var yvs []YouTubeVideoResponse
	for i := 0; i*50 <= len(list); i++ {
		var id string
		if len(list) > 50*(i+1) {
			id = strings.Join(list[50*i:50*(i+1)], ",")
		} else {
			id = strings.Join(list[50*i:], ",")
		}
		call := YouTubeService.Videos.List([]string{"snippet", "contentDetails", "liveStreamingDetails"}).Id(id).MaxResults(50)
		res, err := call.Do()
		if err != nil {
			log.Error().Str("severity", "ERROR").Err(err).Msg("videos-list call error")
			return []YouTubeVideoResponse{}, err
		}

		for _, video := range res.Items {
			scheduledStartTime := "" // 例 2022-03-28T11:00:00Z
			if video.LiveStreamingDetails != nil {
				// "2022-03-28 11:00:00"形式に変換
				rep1 := strings.Replace(video.LiveStreamingDetails.ScheduledStartTime, "T", " ", 1)
				scheduledStartTime = strings.Replace(rep1, "Z", "", 1)
			}

			yvs = append(yvs, YouTubeVideoResponse{ID: video.Id, Title: video.Snippet.Title, Description: video.Snippet.Description, Schedule: scheduledStartTime, Duration: video.ContentDetails.Duration, ChannelID: video.Snippet.ChannelId})

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
func (list YTVRList) Select() (YTSRList, error) {
	var ysr []YouTubeSelectResponse
	// にじさんじライバーのチャンネルリストを取得
	channelIdList, err := GetChannelIdList()
	if err != nil {
		return nil, err
	}
	// 歌動画か判断する
	for _, video := range list {
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

		ysr = append(ysr, YouTubeSelectResponse{ID: video.ID, Title: video.Title, Schedule: video.Schedule, SongConfirm: checkRes})

		log.Info().
			Str("severity", "INFO").
			Str("service", "youtube-video-select").
			Str("id", video.ID).
			Str("title", video.Title).
			Str("duration", video.Duration).
			Str("schedule", video.Schedule).
			Send()
	}
	return ysr, nil
}

// 動画情報をDBに保存
func (list YTSRList) Save() error {
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

	for _, video := range list {
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
	call := YouTubeService.Videos.List([]string{"snippet", "contentDetails", "liveStreamingDetails"}).Id(vid).MaxResults(1)
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

func Activities() (VideoIDList, error) {
	// 動画IDを格納する文字列型配列を宣言
	vid := make([]string, 0, 600)

	// にじさんじライバーのチャンネルリストを取得
	channelIdList, err := GetChannelIdList()
	if err != nil {
		return nil, err
	}

	for _, cid := range channelIdList {
		// 取得した動画IDをログに出力するための変数
		var rid []string
		call := YouTubeService.Activities.List([]string{"snippet", "contentDetails"}).ChannelId(cid).MaxResults(5)
		res, err := call.Do()
		if err != nil {
			log.Error().Str("severity", "ERROR").Err(err).Msg("activities-list call error")
			return []string{}, err
		}

		for _, item := range res.Items {
			if item.Snippet.Type != "upload" {
				continue
			}
			rid = append(rid, item.ContentDetails.Upload.VideoId)
			vid = append(vid, item.ContentDetails.Upload.VideoId)
		}

		log.Info().
			Str("severity", "INFO").
			Str("service", "youtube-activities-list").
			Str("ChannelId", cid).
			Strs("rid", rid).
			Str("pageInfo", fmt.Sprintf("perPage=%d total=%d nextPage=%s\n", res.PageInfo.ResultsPerPage, res.PageInfo.TotalResults, res.NextPageToken)).
			Send()
	}
	return vid, nil
}

// チャンネルIDとチャンネルにアップロードされた動画の数を取得
func Channels() (YTCRList, error) {
	var cResList []YouTubeChannelsResponse
	// にじさんじライバーのチャンネルリストを取得
	cList, err := GetChannelIdList()
	if err != nil {
		return nil, err
	}

	for i := 0; i*50 <= len(cList); i++ {
		var id string
		if len(cList) > 50*(i+1) {
			id = strings.Join(cList[50*i:50*(i+1)], ",")
		} else {
			id = strings.Join(cList[50*i:], ",")
		}
		call := YouTubeService.Channels.List([]string{"statistics"}).MaxResults(50).Id(id)
		res, err := call.Do()
		if err != nil {
			log.Error().Str("severity", "ERROR").Err(err).Msg("channels-list call error")
			return []YouTubeChannelsResponse{}, err
		}

		for _, item := range res.Items {
			cResList = append(cResList, YouTubeChannelsResponse{ID: item.Id, VideoCount: item.Statistics.VideoCount})
		}
	}
	return cResList, nil
}

// 新しく取得したチャンネルの動画数をDBに保存
func (clist YTCRList) Save() error {
	tx, err := DB.Begin()
	if err != nil {
		log.Error().Str("severity", "ERROR").Err(err).Msg("DB.Begin error")
		return err
	}
	// 動画が削除されて動画数が減っていても、上書きする
	stmt, err := tx.Prepare("UPDATE vtubers SET video_count = ? WHERE id = ? AND video_count != ?")
	if err != nil {
		log.Error().Str("severity", "ERROR").Err(err).Msg("DB.Prepare error")
		return err
	}

	for _, list := range clist {
		_, err := stmt.Exec(list.VideoCount, list.ID, list.VideoCount)
		if err != nil {
			log.Error().Str("severity", "ERROR").Err(err).Msg("Save video_count failed")
			if tx.Rollback() != nil {
				log.Error().Str("severity", "ERROR").Err(err).Msg("Rollback error")
			}
			return err
		}
		log.Info().
			Str("severity", "INFO").
			Str("service", "youtube-channels-save").
			Str("channelId", list.ID).
			Uint64("videoCount", list.VideoCount).
			Send()
	}

	err = tx.Commit()
	if err != nil {
		log.Error().Str("severity", "ERROR").Err(err).Msg("tx.Commit error")
		return err
	}

	return nil
}

// 新しく動画をアップロードしたチャンネルIDと古い動画数と新しい動画の数を返す
func (newlist YTCRList) CheckUpload() (NewUpChList, error) {
	// 新しく動画をアップロードしたチャンネルIDを格納する変数
	var diffCList NewUpChList
	// DBに保存されているチャンネルIDとそのチャンネルの動画本数の一覧を取得 map[string] uint64
	vvcList, err := GetVideoCountList()
	if err != nil {
		return nil, err
	}

	for _, newCh := range newlist {
		// DBに保存されている動画の本数より、新しく取得した動画の本数が多い場合　動画が削除されて数が減っている場合は返さない
		if vvcList[newCh.ID] < newCh.VideoCount {
			diffCList = append(diffCList, NewUploadChannelList{ID: newCh.ID, NewVideoCount: newCh.VideoCount, OldVideoCount: vvcList[newCh.ID]})
		}
	}
	return diffCList, nil
}

// 新しくアップロードされた動画のIDを取得
func (list NewUpChList) Search() (VideoIDList, error) {
	// 動画IDを格納する文字列型配列を宣言
	vid := make([]string, 0, 600)

	for _, ch := range list {
		// 取得した動画IDをログに出力するための変数
		var rid []string
		call := YouTubeService.Search.List([]string{"snippet"}).ChannelId(ch.ID).MaxResults(int64(ch.NewVideoCount - ch.OldVideoCount)).Order("date")
		res, err := call.Do()
		if err != nil {
			log.Error().Str("severity", "ERROR").Err(err).Msg("search-list call error")
			return []string{}, err
		}

		for _, item := range res.Items {
			if item.Id.Kind != "youtube#video" {
				continue
			}
			rid = append(rid, item.Id.VideoId)
			vid = append(vid, item.Id.VideoId)
		}

		log.Info().
			Str("severity", "INFO").
			Str("service", "youtube-search-list").
			Str("ChannelId", ch.ID).
			Strs("videoId", rid).
			Send()
	}
	return vid, nil
}

