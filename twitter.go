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
type TwitterSearchRef struct {
	data []GetTweetContext
}

type PostTweetContext struct {
	Text string `json:"text"`
}

type GetTweetContext struct {
	AuthorID string `json:"author_id"`
	ID       string `json:"id"`
	Text     string `json:"text"`
}

type ListsResponse struct {
	Data     []GetTweetContext `json:"data"`
	Includes struct {
		Users []struct {
			CreatedAt time.Time `json:"created_at"`
			ID        string    `json:"id"`
			Name      string    `json:"name"`
			Username  string    `json:"username"`
		} `json:"users"`
	} `json:"includes"`
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
func (tw *Twitter) Search() (*TwitterSearchRef, error) {
	endpoint := "https://api.twitter.com/2/lists/1538799448679395328/tweets?expansions=author_id&user.fields=created_at&max_results=50"
	client := &http.Client{}
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	token := fmt.Sprintf("Bearer %s", os.Getenv("TWITTER_BEARER_TOKEN"))
	req.Header.Add("Authorization", token)

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	var gtc ListsResponse
	if err := json.Unmarshal(body, &gtc); err != nil {
		return nil, err
	}

	return &TwitterSearchRef{
		data: gtc.Data,
	}, nil
}

// ツイート内容に特定の文字列が含まれているかチェック
func (tws *TwitterSearchRef) Select() ([]TwitterSearchResponse, error) {
	var tsr []TwitterSearchResponse
	clint := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	for _, v := range tws.data {
		clog := log.Info().Str("service", "twitter-select").Str("twitter_id", v.ID).Str("Text", v.Text)
		if !regexp.MustCompile(".*song|プレミア公開|MV.*").Match([]byte(v.Text)) {
			continue
		}
		if regexp.MustCompile(".*RT.*").Match([]byte(v.Text)) {
			continue
		}
		// ツイート内容に含まれるURLが短縮版のため、リダイレクト先のURLを取得する
		idx := strings.Index(v.Text, "https://t.co/")
		if idx == -1 {
			continue
		}
		if len(v.Text) < idx+23 {
			clog.Str("severity", "WARNING").Send()
			continue
		}
		sid := v.Text[idx : idx+23]
		req, err := http.NewRequest("GET", sid, nil)
		if err != nil {
			return nil, err
		}
		resp, err := clint.Do(req)
		if err != nil {
			return nil, err
		}
		// Redirect先のURLを取得
		rid := resp.Header.Get("Location")

		// URLからYouTubeの動画ID部分を抽出する
		var yid string
		idx = strings.Index(rid, "youtu.be")
		if idx != -1 {
			yid = rid[17:28]
		}
		if idx == -1 {
			idx = strings.Index(rid, "youtube.com")
			if idx == -1 {
				continue
			}
			yid = rid[32:43]
		}

		tsr = append(tsr, TwitterSearchResponse{ID: v.ID, YouTubeID: yid, Text: v.Text})
		clog.Str("severity", "INFO").Str("youtube_id", yid).Send()
	}
	return tsr, nil
}
