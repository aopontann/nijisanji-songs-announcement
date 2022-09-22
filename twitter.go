package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

type Twitter struct{}

type PostTweetContext struct {
	Text string `json:"text"`
}

type GetTweetContext struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

type ListTweetsEntities struct {
	Annotations []struct {
		Start          int     `json:"start"`
		End            int     `json:"end"`
		Probability    float64 `json:"probability"`
		Type           string  `json:"type"`
		NormalizedText string  `json:"normalized_text"`
	} `json:"annotations"`
	Cashtags []struct {
		Start int    `json:"start"`
		End   int    `json:"end"`
		Tag   string `json:"tag"`
	} `json:"cashtags"`
	Hashtags []struct {
		Start int    `json:"start"`
		End   int    `json:"end"`
		Tag   string `json:"tag"`
	} `json:"hashtags"`
	Mentions []struct {
		Start int    `json:"start"`
		End   int    `json:"end"`
		Tag   string `json:"tag"`
	} `json:"mentions"`
	Urls []struct {
		Start       int    `json:"start"`
		End         int    `json:"end"`
		URL         string `json:"url"`
		ExpandedURL string `json:"expanded_url"`
		DisplayURL  string `json:"display_url"`
		Status      string `json:"status"`
		Title       string `json:"title"`
		Description string `json:"description"`
		UnwoundURL  string `json:"unwound_url"`
	} `json:"urls"`
}

type ListTweetsResponse struct {
	Data []struct {
		Entities ListTweetsEntities `json:"entities"`
		ID       string             `json:"id"`
		Text     string             `json:"text"`
	} `json:"data"`
	Meta struct {
		ResultCount int    `json:"result_count"`
		NextToken   string `json:"next_token"`
	} `json:"meta"`
}

type TwitterSearchResponse struct {
	ID        string `json:"id"`
	YouTubeID string `json:"youtube_id"`
	Text      string `json:"text"`
}

var clint = &http.Client{
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

// Searchで使用するカスタムエラーログ
var twlog = log.Info().Str("service", "twitter-search").Str("severity", "ERROR")

// getRedirectで使用するカスタムエラーログ
var tgrlog = log.Info().Str("service", "twitter-getRedirect").Str("severity", "ERROR")

// 歌動画の告知ツイート
func (tw *Twitter) Post(id string, text string) error {
	const endpoint = "https://api.twitter.com/2/tweets"
	text = fmt.Sprintf("5分後に公開予定\n\n%s\n\nhttps://www.youtube.com/watch?v=%s", text, id)
	requestBody := &PostTweetContext{
		Text: text,
	}
	jsonString, err := json.Marshal(requestBody)
	if err != nil {
		return err
	}
	t := time.Now().Unix()

	// 認証に使うnonceを生成する
	key := make([]byte, 32)
	_, err = rand.Read(key)
	if err != nil {
		return err
	}
	nonce := base64.StdEncoding.EncodeToString(key)
	symbols := []string{"+", "/", "="}
	for _, s := range symbols {
		nonce = strings.Replace(nonce, s, "", -1)
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
		log.Error().Str("severity", "ERROR").Err(err).Msg("twitter call failed")
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

// にじさんじライバーのツイートを取得する
func (tw *Twitter) Search() ([]TwitterSearchResponse, error) {
	endpoint := "https://api.twitter.com/2/lists/1538799448679395328/tweets?tweet.fields=entities&expansions=referenced_tweets.id&max_results=30"
	client := &http.Client{}
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		twlog.Msg("http.NewRequest error")
		return nil, err
	}

	token := fmt.Sprintf("Bearer %s", os.Getenv("TWITTER_BEARER_TOKEN"))
	req.Header.Add("Authorization", token)

	res, err := client.Do(req)
	if err != nil {
		twlog.Msg(err.Error())
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		twlog.Msg(err.Error())
		return nil, err
	}

	var gtc ListTweetsResponse
	if err := json.Unmarshal(body, &gtc); err != nil {
		twlog.Msg(err.Error())
		return nil, err
	}

	var tsr []TwitterSearchResponse
	for _, tweet := range gtc.Data {
		// ツイート内容に"公開"の文字が含まれている場合、メールを送る
		if strings.Contains(tweet.Text, "公開") {
			err := sendMail(tweet.ID, tweet.Text)
			if err != nil {
				twlog.Msg(err.Error())
				return nil, err
			}
		}

		yid, err := getRedirect(tweet.ID, tweet.Text)
		if err != nil {
			twlog.Msg(err.Error())
			return nil, err
		}
		if yid != "" {
			tsr = append(tsr, TwitterSearchResponse{ID: tweet.ID, YouTubeID: yid, Text: tweet.Text})
		}
	}

	return tsr, nil
}

// プレミア公開される歌ってみた動画のみ返ってくる
// 動画IDから動画詳細情報を取得して歌動画か判断する
func (tw *Twitter) Select(tsr []TwitterSearchResponse) ([]YouTubeCheckResponse, error) {
	var ytcr []YouTubeCheckResponse
	var id []string
	yttw := make(map[string]string)
	for _, v := range tsr {
		id = append(id, v.YouTubeID)
		yttw[v.YouTubeID] = v.ID
	}
	// にじさんじライバーのチャンネルリストを取得
	channelIdList, err := GetChannelIdList()

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
		// if regexp.MustCompile(`.*cover|Cover|歌って|MV.*`).Match([]byte(video.Snippet.Title)) {
		// 	ytcr = append(ytcr, YouTubeCheckResponse{ID: video.Id, Title: video.Snippet.Title, Schedule: video.LiveStreamingDetails.ScheduledStartTime, TwitterID: yttw[video.Id]})

		// 	log.Info().
		// 		Str("severity", "INFO").
		// 		Str("service", "youtube-video-check").
		// 		Str("id", video.Id).
		// 		Str("title", video.Snippet.Title).
		// 		Str("duration", video.ContentDetails.Duration).
		// 		Str("schedule", scheduledStartTime).
		// 		Send()
		// 	continue
		// }
		// // 動画概要欄に特定の文字が含まれているか
		// if !regexp.MustCompile(`.*vocal|Vocal|song|Song|歌|MV|作曲.*`).Match([]byte(video.Snippet.Description)) {
		// 	continue
		// }
		// にじさんじライバーのチャンネルで公開されたか
		if !NijisanjiCheck(channelIdList, video.Snippet.ChannelId) {
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

// ツイート内容にある短縮URLからリダイレクト先URLを取得する
// Youtube動画リンクではない場合は""を返す
func getRedirect(id string, text string) (string, error) {
	idx := strings.Index(text, "https://t.co/")
	// 短縮URLがない場合
	if idx == -1 {
		return "", nil
	}
	// 192行目のエラー回避処理
	if len(text) < idx+23 {
		tgrlog.Str("id", id).Str("text", "text").Msg("len(text) < idx+23 error")
		return "", nil
	}
	// ツイート内容から短縮URL部分を抽出
	sid := text[idx : idx+23]

	// Redirect先のURLを取得
	req, err := http.NewRequest("GET", sid, nil)
	if err != nil {
		return "", err
	}
	resp, err := clint.Do(req)
	if err != nil {
		tgrlog.Str("id", id).Str("text", "text").Msg("client.Do error")
		return "", err
	}
	rid := resp.Header.Get("Location")

	// リダイレクト先URLからYouTubeの動画ID部分を抽出する
	idx = strings.Index(rid, "youtu.be")
	if idx != -1 {
		return rid[17:28], nil
	}
	idx = strings.Index(rid, "youtube.com")
	if idx != -1 {
		return rid[30:41], nil
	}
	// youtube動画のリンクでなかった場合
	return "", nil
}

func getUrl()
