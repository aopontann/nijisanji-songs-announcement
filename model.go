package main

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
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

type RequestBody struct {
	Text string `json:"text"`
}

// 過去30分間までにYouTubeにアップロードされた動画を取得する
func YoutubeSearchList(ctx context.Context, q string) (string, error) {
	youtubeService, err := youtube.NewService(ctx, option.WithAPIKey(os.Getenv("YOUTUBE_API_KEY")))
	if err != nil {
		return "", err
	}
	// 動画検索
	dtAfter := time.Now().UTC().Add(-30 * time.Minute).Format("2006-01-02T15:04:00Z")
	dtBefore := time.Now().UTC().Format("2006-01-02T15:04:00Z")

	searchCall := youtubeService.Search.List([]string{"id"}).
	MaxResults(50).
	Q(q).
	PublishedAfter(dtAfter).
	PublishedBefore(dtBefore)
	searchRes, err := searchCall.Do()
	if err != nil {
		return "", err
	}

	log.Printf("youtube-search-list: %s ~ %s item: %d\n", dtAfter, dtBefore,searchRes.PageInfo.ResultsPerPage)

	// これもっといいロジックがある https://qiita.com/ono_matope/items/d5e70d8a9ff2b54d5c37
	videoId := ""
	for _, searchItem := range searchRes.Items {
		if videoId == "" {
			videoId = searchItem.Id.VideoId
			continue
		}
		videoId = videoId + "," + searchItem.Id.VideoId
	}
	return videoId, nil
}

// Youtube Data API から動画情報を取得し、歌動画か判別してDBに動画情報を保存する
func YoutubeVideoList(ctx context.Context, videoIdList string) error {

	// Youtube Data API から動画情報を取得する
	youtubeService, err := youtube.NewService(ctx, option.WithAPIKey(os.Getenv("YOUTUBE_API_KEY")))
	videoscall := youtubeService.Videos.List([]string{"snippet", "contentDetails", "liveStreamingDetails"}).Id(videoIdList).MaxResults(50)
	response, err := videoscall.Do()
	if err != nil {
		return err
	}

	// にじさんじライバーのチャンネルリストを取得する
	channelIdList, err := GetChannelIdList()
	if err != nil {
		return err
	}

	// DB準備
	stmt, err := DB.Prepare("INSERT IGNORE INTO videos(id, title, songConfirm, scheduled_start_time) VALUES(?,?,?,?)")
	if err != nil {
		return err
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
			return err
		}
		// テストログ
		log.Printf("id=%s  title=%s duration=%s schedule=%s\n", video.Id, video.Snippet.Title, video.ContentDetails.Duration, scheduledStartTime)
	}
	return nil
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

// 認証に使うnonceを生成する
func GenerateOAthNonce() (string, error) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		return "", err
	}

	nonce := base64.StdEncoding.EncodeToString(key)
	symbols := []string{"+", "/", "="}
	for _, s := range symbols {
		nonce = strings.Replace(nonce, s, "", -1)
	}
	return nonce, nil
}

// 歌動画の告知ツイート
func PostTweet(id string, text string) error {
	text = fmt.Sprintf("5分後に公開予定\n\n%s\n\nhttps://www.youtube.com/watch?v=%s", text, id)
	requestBody := &RequestBody{
		Text: text,
	}
	jsonString, err := json.Marshal(requestBody)
	if err != nil {
		return err
	}
	t := time.Now().Unix()
	nonce, err := GenerateOAthNonce()
	if err != nil {
		return err
	}

	const parameterBaseString = `oauth_consumer_key=%s&oauth_nonce=%s&oauth_signature_method=HMAC-SHA1&oauth_timestamp=%d&oauth_token=%s&oauth_version=1.0`
	parameter := fmt.Sprintf(
		parameterBaseString,
		os.Getenv("TWITTER_API_KEY"),
		nonce,
		t,
		os.Getenv("TWITTER_ACCESS_TOKEN"))
	baseString := fmt.Sprintf("POST&%s&%s", url.QueryEscape(endpoint), url.QueryEscape(parameter))
	signingKey := fmt.Sprintf("%s&%s", os.Getenv("TWITTER_API_SECRET_KEY"), os.Getenv("TWITTER_ACCESS_TOKEN_SECRET"))

	h := hmac.New(sha1.New, []byte(signingKey))
	io.WriteString(h, baseString)
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))

	const oauth1header = `OAuth oauth_consumer_key="%s",oauth_nonce="%s",oauth_signature="%s",oauth_signature_method="%s",oauth_timestamp="%d",oauth_token="%s",oauth_version="%s"`

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonString))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf(oauth1header,
		os.Getenv("TWITTER_API_KEY"),
		nonce,
		url.QueryEscape(signature),
		"HMAC-SHA1",
		t,
		os.Getenv("TWITTER_ACCESS_TOKEN"),
		"1.0",
	))

	client := new(http.Client)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	byteArray, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	fmt.Printf("%#v", string(byteArray))
	return nil
}
