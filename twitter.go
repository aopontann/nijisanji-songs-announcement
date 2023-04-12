package main

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/dghubble/oauth1"
	"github.com/rs/zerolog/log"
)

type Twitter struct{}

type GetTweetContext struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

// 歌動画の告知ツイート
func (tw *Twitter) Post(v getVideoInfo) error {
	const endpoint = "https://api.twitter.com/2/tweets"
	var requestBody string

	if v.TwitterID != "" {
		requestBody = fmt.Sprintf(`{
			"text": "5分後に公開",
			"quote_tweet_id": "%s"
		}`, v.TwitterID)
	} else {
		requestBody = fmt.Sprintf(`{
			"text": "5分後に公開予定\n\n%s\n\nhttps://www.youtube.com/watch?v=%s"
		}`, v.Title, v.ID)
	}
	payload := strings.NewReader(requestBody)

	t := time.Now().Unix()

	// 認証に使うnonceを生成する
	key := make([]byte, 32)
	_, err := rand.Read(key)
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

	req, err := http.NewRequest("POST", endpoint, payload)
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

func Tweets() error {
	url := "https://api.twitter.com/2/tweets"
	config := oauth1.NewConfig(os.Getenv("TWITTER_API_KEY"), os.Getenv("TWITTER_API_SECRET_KEY"))
	token := oauth1.NewToken(os.Getenv("TWITTER_ACCESS_TOKEN"), os.Getenv("TWITTER_ACCESS_TOKEN_SECRET"))

	payload := strings.NewReader(`{"text": "自動ツイートテスト1"}`)

	httpClient := config.Client(oauth1.NoContext, token)

	resp, err := httpClient.Post(url, "application/json", payload)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Printf("Raw Response Body:\n%v\n", string(body))
	return nil
}
