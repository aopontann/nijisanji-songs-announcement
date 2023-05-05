package main

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"

	"github.com/rs/zerolog/log"
	"google.golang.org/api/youtube/v3"
)

type GetVideoInfo struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

// 歌ってみた動画かフィルターをかける処理
func (list YTVRList) Select() YTVRList {
	var rlist []youtube.Video
	// 歌動画か判断する
	for _, video := range list {
		// プレミア公開する動画か
		if video.LiveStreamingDetails == nil {
			continue
		}
		// 放送が終了している場合
		if video.Snippet.LiveBroadcastContent == "none" {
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
		// 動画タイトルに特定の文字が含まれているか
		// checkRes := TitleCheck(video.Title)
		// checkRes := true
		rlist = append(rlist, video)

		log.Info().
			Str("severity", "INFO").
			Str("service", "youtube-video-select").
			Str("id", video.Id).
			Str("title", video.Snippet.Title).
			Str("duration", video.ContentDetails.Duration).
			Str("schedule", video.LiveStreamingDetails.ScheduledStartTime).
			Send()
	}
	return rlist
}

// にじさんじライバーのチャンネルでアップロードされた動画か
func (list YTVRList) IsNijisanji() (YTVRList, error) {
	var rlist YTVRList
	// にじさんじライバーのチャンネルリストを取得
	channelIdList, err := GetChannelIdList()
	if err != nil {
		return nil, err
	}
	// 歌動画か判断する
	for _, video := range list {
		for _, cid := range channelIdList {
			if video.Snippet.ChannelId == cid {
				rlist = append(rlist, video)
				break
			}
		}
	}
	return rlist, nil
}

// DBに保存されていない動画か
func (list YTVRList) NotExist() (YTVRList, error) {
	var rlist YTVRList
	for _, video := range list {
		err := DB.QueryRow("SELECT id FROM videos WHERE id = ?", video.Id).Scan()
		if errors.Is(err, sql.ErrNoRows) {
			rlist = append(rlist, video)
			continue
		}
		if err != nil {
			return nil, err
		}
	}
	return rlist, nil
}

// 時間を指定して動画を取得する
func GetVideos(at string, bt string) ([]GetVideoInfo, error) {
	var (
		id        string
		title     string
		videoList []GetVideoInfo
	)
	rows, err := DB.Query("SELECT id, title FROM videos WHERE songConfirm = 1 AND scheduled_start_time >= ? AND scheduled_start_time <= ?", at, bt)
	if err != nil {
		log.Error().Str("severity", "ERROR").Err(err).Msg("select videos failed")
		return []GetVideoInfo{}, err
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&id, &title)
		if err != nil {
			return []GetVideoInfo{}, err
		}
		videoList = append(videoList, GetVideoInfo{ID: id, Title: title})
	}
	err = rows.Err()
	if err != nil {
		return []GetVideoInfo{}, err
	}
	return videoList, nil
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

// タイトルに特定の文字が含まれているかチェックする
func TitleCheck(title string) bool {
	for _, cstr := range definiteList {
		reg := fmt.Sprintf(`.*%s.*`, cstr)
		if regexp.MustCompile(reg).Match([]byte(title)) {
			return true
		}
	}
	return false
}
