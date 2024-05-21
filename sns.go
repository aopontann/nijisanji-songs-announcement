package nsa

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/smtp"
	"os"
	"strings"

	"github.com/dghubble/oauth1"
	webpush "github.com/SherClockHolmes/webpush-go"
)

// Misskey関連の機能のレシーバを登録する構造体
type Misskey struct {
	token string
}

// misskey投稿するときのリクエストボディの構造体
type ReqBody struct {
	I      string `json:"i"`
	Text   string `json:"text"`
	Detail bool   `json:"detail"`
}

// Twitter関連の機能のレシーバを登録する構造体
type Twitter struct {
	vid   string
	title string
}

func NewTwitter() *Twitter {
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
	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("Raw Response Body:\n%v\n", string(body))
	return nil
}

func NewMisskey(token string) *Misskey {
	return &Misskey{token: token}
}

// misskey に投稿する
func (m *Misskey) Post(id string, title string) error {
	url := "https://@aopontan@misskey.io/api/notes/create"
	content := fmt.Sprintf(`
	【5分後に公開】
	%s
	https://www.youtube.com/watch?v=%s
	`, title, id)

	resb := ReqBody{
		I:      m.token,
		Text:   content,
		Detail: false,
	}

	payload, err := json.Marshal(resb)
	if err != nil {
		return err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func SendMail(subject string, message string) error {
	addr := os.Getenv("MAIL_ADDRESS")
	auth := smtp.PlainAuth(
		"",
		addr,                       // 送信に使うアカウント
		os.Getenv("SMTP_PASSWORD"), // アカウントのパスワード or アプリケーションパスワード
		"smtp.gmail.com",
	)

	slog.Info(
		"send-mail",
		slog.String("severity", "INFO"),
		slog.String("subject", subject),
		slog.String("message", message),
	)

	return smtp.SendMail(
		"smtp.gmail.com:587",
		auth,
		addr,           // 送信元
		[]string{addr}, // 送信先
		[]byte(
			"To: "+addr+"\r\n"+
				"Subject:"+subject+"\r\n"+
				"\r\n"+
				message+"\r\n"),
	)
}

func WebPush(video Video, subscriptionList []string) error {
	addr := os.Getenv("MAIL_ADDRESS")
	publicKey := os.Getenv("WEBPUSH_PUBLIC_KEY")
	privateKey := os.Getenv("WEBPUSH_PRIVATE_KEY")

	reqformatMessage := `
	{
		"title": "5分後に公開",
		"body": "%s",
		"data": {
		  "url": "https://youtu.be/%s"
		},
		"icon": "%s"
	  }
	`
	req_message := fmt.Sprintf(reqformatMessage, video.Title, video.ID, video.Thumbnail)

	s := &webpush.Subscription{}
	for _, subscription := range subscriptionList {
		json.Unmarshal([]byte(subscription), s)
		// Send Notification
		res, err := webpush.SendNotification([]byte(req_message), s, &webpush.Options{
			Subscriber:      addr,
			VAPIDPublicKey:  publicKey,
			VAPIDPrivateKey: privateKey,
			TTL:             300,
			Urgency: webpush.UrgencyHigh,
		})
		if err != nil {
			fmt.Println(err)
			return err
		}
		slog.Info(
			"song-webpush",
			slog.String("severity", "INFO"),
			slog.String("vid", video.ID),
			slog.String("status_code", res.Status),
		)
	}
	return nil
}