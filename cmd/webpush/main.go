package main

import (
	"encoding/json"
	"fmt"
	"os"

	webpush "github.com/SherClockHolmes/webpush-go"
)

func main() {
	addr := os.Getenv("MAIL_ADDRESS")
	publicKey := os.Getenv("WEBPUSH_PUBLIC_KEY")
	privateKey := os.Getenv("WEBPUSH_PRIVATE_KEY")

	// Decode subscription
	s := &webpush.Subscription{}
	subscription := `
	`
	json.Unmarshal([]byte(subscription), s)

	req_message := `
	{
		"title": "あたらしいWeb Pushが送信されました",
		"body": "Web Pushです",
		"data": {
		  "url": "https://www.youtube.com/watch?v=gJk_vTwiduU"
		},
		"icon": "https://avatars.githubusercontent.com/u/74401628?v=4"
	  }
	`

	// Send Notification
	resp, err := webpush.SendNotification([]byte(req_message), s, &webpush.Options{
		Subscriber:      addr,
		VAPIDPublicKey:  publicKey,
		VAPIDPrivateKey: privateKey,
		TTL:             30,
	})
	// 410 or 210
	fmt.Println(resp.StatusCode)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()
}
