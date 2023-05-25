package twitter

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/dghubble/oauth1"
)

type Twitter struct {
	vid   string
	title string
}

func New() *Twitter {
	return &Twitter{}
}

func (tw *Twitter) Id(vid string) *Twitter {
	tw.vid = vid
	return tw
}

func (tw *Twitter) Title(title string) *Twitter {
	tw.title = title
	return tw
}

func (tw *Twitter) Tweet() error {
	url := "https://api.twitter.com/2/tweets"
	config := oauth1.NewConfig(os.Getenv("TWITTER_API_KEY"), os.Getenv("TWITTER_API_SECRET_KEY"))
	token := oauth1.NewToken(os.Getenv("TWITTER_ACCESS_TOKEN"), os.Getenv("TWITTER_ACCESS_TOKEN_SECRET"))

	reqBody := fmt.Sprintf(`{"text": "【5分後に公開】\n\n%s\n\nhttps://www.youtube.com/watch?v=%s"}`, tw.title, tw.vid)
	payload := strings.NewReader(reqBody)

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
