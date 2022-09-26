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
	"google.golang.org/api/youtube/v3"
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
		Status      int    `json:"status"`
		Title       string `json:"title"`
		Description string `json:"description"`
		UnwoundURL  string `json:"unwound_url"`
	} `json:"urls"`
}

type ListTweetsData struct {
	Entities  ListTweetsEntities `json:"entities"`
	CreatedAt string             `json:"created_at"`
	ID        string             `json:"id"`
	Text      string             `json:"text"`
}

type ListTweetsResponse struct {
	Data []ListTweetsData `json:"data"`
	Meta struct {
		ResultCount int    `json:"result_count"`
		NextToken   string `json:"next_token"`
	} `json:"meta"`
}

type TwitterSearchResponse struct {
	ID          string         `json:"id"`
	Text        string         `json:"text"`
	YoutubeData *youtube.Video `json:"youtube_data"`
}

// Searchで使用するカスタムエラーログ
var twlog = log.Info().Str("service", "twitter-search").Str("severity", "ERROR")

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

// 過去10分間に投稿されたにじさんじライバーのツイートを取得する
func (tw *Twitter) Search() ([]TwitterSearchResponse, error) {
	endpoint := "https://api.twitter.com/2/lists/1538799448679395328/tweets?tweet.fields=entities,created_at&max_results=30"
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

	twData := tweetsFilter(gtc.Data)

	// ツイートに含まれている可能性がある動画IDを保存する配列
	var yidList []string
	for _, tweet := range twData {
		yid := getUrl(tweet.Entities)
		if yid != "" {
			yidList = append(yidList, yid)
		}
	}

	// Youtube Data APIから動画情報を取得
	videos, err := yt.Video2(yidList)
	if err != nil {
		twlog.Msg(err.Error())
		return nil, err
	}
	// youtubeIDをキーとし、動画情報を値をとしたmap配列
	mapVideos := make(map[string]*youtube.Video)
	for _, video := range videos {
		mapVideos[video.Id] = video
	}

	// 返り値にYoutube Data APIのレスポンス結果も付属させる
	var tsr []TwitterSearchResponse
	for _, tweet := range twData {
		log.Info().Str("severity", "INFO").Str("service", "tweet-search").Str("id", tweet.ID).Str("text", tweet.Text).Send()

		yid := getUrl(tweet.Entities)
		if yid != "" {
			tsr = append(tsr, TwitterSearchResponse{ID: tweet.ID, Text: tweet.Text, YoutubeData: mapVideos[yid]})
		} else {
			tsr = append(tsr, TwitterSearchResponse{ID: tweet.ID, Text: tweet.Text})
		}
	}
	return tsr, nil
}

// プレミア公開される歌ってみた動画のみ返ってくる
// 動画IDから動画詳細情報を取得して歌動画か判断する
func (tw *Twitter) Select(tsr []TwitterSearchResponse) ([]TwitterSearchResponse, error) {
	var ytcr []TwitterSearchResponse
	// にじさんじライバーのチャンネルリストを取得
	channelIdList, err := GetChannelIdList()
	if err != nil {
		log.Info().Str("service", "twitter-select").Str("severity", "ERROR").Msg("GetChannelIdList() error")
	}

	for _, t := range tsr {
		// ツイートに動画IDがなかった場合
		if t.YoutubeData == nil {
			continue
		}
		video := t.YoutubeData
		// プレミア公開する動画ではない場合
		if video.LiveStreamingDetails == nil {
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
		// にじさんじライバーのチャンネルで公開されていない場合
		if !NijisanjiCheck(channelIdList, video.Snippet.ChannelId) {
			continue
		}

		// フィルターをくくり抜けたデータのみログを表示
		log.Info().
			Str("severity", "INFO").
			Str("service", "twitter-select").
			Str("twitter_id", t.ID).
			Str("id", video.Id).
			Str("title", video.Snippet.Title).
			Str("duration", video.ContentDetails.Duration).
			Str("schedule", video.LiveStreamingDetails.ScheduledStartTime).
			Send()

		ytcr = append(ytcr, t)
	}
	return ytcr, nil
}

// ツイート内容からURLを取得する
// Youtube動画リンクではない場合は""を返す
func getUrl(entities ListTweetsEntities) string {
	if entities.Urls == nil {
		return ""
	}
	vid := ""
	for _, url := range entities.Urls {
		idx := strings.Index(url.ExpandedURL, "youtu.be")
		if idx != -1 {
			vid = url.ExpandedURL[idx+9 : idx+20]
		}
		// https://www.youtube.com/watch?v=PBf70efxSkI
		if strings.Contains(url.ExpandedURL, "youtube.com/watch") {
			vid = url.ExpandedURL[32:43]
		}
		log.Info().Str("severity", "INFO").Str("service", "tweet-getURL").Str("url", url.ExpandedURL).Str("vid", vid).Send()
		if vid != "" {
			return vid
		}
	}
	return ""
}

// 取得した複数のツイートをフィルターにかける
func tweetsFilter(ltd []ListTweetsData) []ListTweetsData {
	// ツイートの情報を格納する
	var twData []ListTweetsData
	for _, tweet := range ltd {
		ago10m := time.Now().Add(-10 * time.Minute)
		t, _ := time.Parse(time.RFC3339, tweet.CreatedAt)
		// 10分前以上前にツイートの場合
		if t.Before(ago10m) {
			continue
		}
		twData = append(twData, tweet)
	}
	return twData
}

func (tw *Twitter) Mail(tsr []TwitterSearchResponse) error {
	clog := log.Error().Str("severity", "ERROR").Str("service", "twitter-mail")
	for _, t := range tsr {
		// ツイート内容に"公開"の文字が含まれている場合、メールを送る
		if strings.Contains(t.Text, "公開") {
			err := sendMail(t.ID, t.Text)
			if err != nil {
				clog.Str("id", t.ID).Str("text", t.Text).Msg(err.Error())
				return err
			}
			continue
		}
		// ツイートにYoutubeのURLが含まれていない場合
		if t.YoutubeData == nil {
			continue
		}
		// Youtube動画の情報
		var video = t.YoutubeData
		if video.LiveStreamingDetails == nil {
			continue
		}
		// 放送前ではない場合
		if video.Snippet.LiveBroadcastContent != "upcoming" {
			continue
		}
		// 生放送の動画であった場合
		// 生放送でも終了後に"PT3H12M23S"のようになるが、LiveBroadcastContentが"upcoming"ではなくなるため上の条件に当てはまる
		if video.ContentDetails.Duration == "P0D" {
			continue
		}
		// ツイートIDとツイート内容をメールの送信
		err := sendMail(t.ID, t.Text)
		if err != nil {
			clog.Str("id", t.ID).Str("text", t.Text).Msg(err.Error())
			return err
		}
	}
	return nil
}
