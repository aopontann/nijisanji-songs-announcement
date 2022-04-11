package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

type getVideoInfo struct {
	Id    string `json:"id"`
	Title string `json:"title"`
}

func YoutubeHandler(w http.ResponseWriter, _ *http.Request) {
	ctx := context.Background()
	youtubeService, err := youtube.NewService(ctx, option.WithAPIKey(os.Getenv("YOUTUBE_API_KEY")))
	if err != nil {
		log.Fatalf("Error creating YouTube client: %v", err)
	}
	// 動画検索
	dtAfter := time.Now().Add(-1 * time.Hour).Format("2006-01-02T15:00:00Z")
	dtBefore := time.Now().Format("2006-01-02T15:04:05Z")
	searchCall := youtubeService.Search.List([]string{"id"}).
		MaxResults(50).
		Q("にじさんじ + 歌って|cover|歌").
		PublishedAfter(dtAfter).
		PublishedBefore(dtBefore)
	searchRes, err := searchCall.Do()
	if err != nil {
		log.Fatalf("Error making API call to list channels: %v\n", err.Error())
	}
	// これもっといいロジックがある https://qiita.com/ono_matope/items/d5e70d8a9ff2b54d5c37
	videoId := ""
	for _, searchItem := range searchRes.Items {
		if videoId == "" {
			videoId = searchItem.Id.VideoId
			continue
		}
		videoId = videoId + "," + searchItem.Id.VideoId
	}
	fmt.Printf("videoId=%s\n", videoId)

	// にじさんじのライバーのチャンネルリスト
	var (
		channelId     string
		channelIdList []string
	)
	rows, err := DB.Query("select id from vtubers")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&channelId)
		if err != nil {
			log.Fatal(err)
		}
		channelIdList = append(channelIdList, channelId)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	// 動画詳細情報取得
	videoscall := youtubeService.Videos.List([]string{"snippet", "contentDetails", "liveStreamingDetails"}).Id(videoId)
	response, err := videoscall.Do()
	if err != nil {
		// The channels.list method call returned an error.
		log.Fatalf("Error making API call to list channels: %v\n", err.Error())
	}
	// DB準備
	stmt, err := DB.Prepare("INSERT IGNORE INTO videos(id, title, songConfirm, scheduled_start_time) VALUES(?,?,?,?)")
	if err != nil {
		log.Fatal(err)
	}
	// 歌動画か判断する
	for _, video := range response.Items {
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
		// 動画タイトルに特定の文字が含まれているか
		checkRes := titleCheck(video.Snippet.Title)
		// にじさんじライバーのチャンネルで公開されたか
		if !nijisanjiCheck(channelIdList, video.Snippet.ChannelId) {
			// にじさんじライバーのチャンネルでもなく、特定の文字が含まれていない場合
			if !checkRes {
				continue
			}
			// にじさんじライバーのチャンネルではないが、特定の文字が含まれている場合（外部コラボの可能性がある）
			checkRes = false
		}
		// DBに動画情報を保存
		_, err := stmt.Exec(video.Id, video.Snippet.Title, checkRes, scheduledStartTime)
		if err != nil {
			log.Fatal(err)
		}
		// テストログ
		fmt.Printf("id=%s  title=%s duration=%s schedule=%s\n", video.Id, video.Snippet.Title, video.ContentDetails.Duration, scheduledStartTime)
	}
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
func titleCheck(title string) bool {
	for _, cstr := range definiteList {
		reg := fmt.Sprintf(`.*%s.*`, cstr)
		if regexp.MustCompile(reg).Match([]byte(title)) {
			return true
		}
	}
	return false
}

// にじさんじライバーのチャンネルから投稿された動画かチェックする
func nijisanjiCheck(channelIdList []string, id string) bool {
	for _, channelId := range channelIdList {
		if id == channelId {
			return true
		}
	}
	return false
}

// 時間を指定して動画を取得する
func getVideos(at string, bt string) ([]getVideoInfo, error) {
	var (
		id        string
		title     string
		videoList []getVideoInfo
	)
	rows, err := DB.Query("SELECT id, title FROM videos WHERE songConfirm = 1 AND scheduled_start_time >= ? AND scheduled_start_time <= ?", at, bt)
	if err != nil {
		return []getVideoInfo{}, err
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&id, &title)
		if err != nil {
			return []getVideoInfo{}, err
		}
		videoList = append(videoList, getVideoInfo{Id: id, Title: title})
	}
	err = rows.Err()
	if err != nil {
		return []getVideoInfo{}, err
	}
	return videoList, nil
}
