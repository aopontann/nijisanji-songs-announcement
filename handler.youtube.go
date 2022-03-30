package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

func YoutubeHandler(w http.ResponseWriter, _ *http.Request) {
	ctx := context.Background()
	youtubeService, err := youtube.NewService(ctx, option.WithAPIKey(os.Getenv("YOUTUBE_API_KEY")))
	if err != nil {
		log.Fatalf("Error creating YouTube client: %v", err)
	}
	// 動画検索
	searchCall := youtubeService.Search.List([]string{"id"}).
		MaxResults(50).
		Q("にじさんじ + 歌って|cover|歌").
		PublishedAfter("2022-03-28T00:00:00Z").
		PublishedBefore("2022-03-28T23:59:59Z")
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

	// にじさんじのライバーのチャンネルリストを取得する
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
		checkRes := TitleCheck(video.Snippet.Title)
		// にじさんじライバーのチャンネルで公開されたか
		if !NijisanjiCheck(channelIdList, video.Snippet.ChannelId) {
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
