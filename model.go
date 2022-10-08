package main

import (
	"database/sql"
	"fmt"
	"regexp"

	"github.com/rs/zerolog/log"
)

type getVideoInfo struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	TwitterID string `json:"twitter_id"`
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

// にじさんじライバーのチャンネルから投稿された動画かチェックする
func NijisanjiCheck(channelIdList []string, id string) bool {
	for _, channelId := range channelIdList {
		if id == channelId {
			return true
		}
	}
	return false
}

// にじさんじライバーのチャンネルID一覧を取得する
func GetChannelIdList() ([]string, error) {
	var (
		channelId     string
		channelIdList []string
	)
	rows, err := DB.Query("select id from vtubers")
	if err != nil {
		log.Error().Str("severity", "ERROR").Err(err).Msg("select vtuber failed")
		return channelIdList, err
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&channelId)
		if err != nil {
			return channelIdList, err
		}
		channelIdList = append(channelIdList, channelId)
	}
	err = rows.Err()
	if err != nil {
		return channelIdList, err
	}
	return channelIdList, nil
}

// 時間を指定して動画を取得する
func GetVideos(at string, bt string) ([]getVideoInfo, error) {
	var (
		id        string
		title     string
		tid       sql.NullString
		videoList []getVideoInfo
	)
	rows, err := DB.Query("SELECT id, title, twitter_id FROM videos WHERE songConfirm = 1 AND scheduled_start_time >= ? AND scheduled_start_time <= ?", at, bt)
	if err != nil {
		log.Error().Str("severity", "ERROR").Err(err).Msg("select videos failed")
		return []getVideoInfo{}, err
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&id, &title, &tid)
		if err != nil {
			return []getVideoInfo{}, err
		}
		if tid.Valid {
			videoList = append(videoList, getVideoInfo{ID: id, Title: title, TwitterID: tid.String})
		} else {
			videoList = append(videoList, getVideoInfo{ID: id, Title: title, TwitterID: ""})
		}
	}
	err = rows.Err()
	if err != nil {
		return []getVideoInfo{}, err
	}
	return videoList, nil
}
