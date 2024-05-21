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
	subscription := []string{
		`{"endpoint":"https://fcm.googleapis.com/fcm/send/eTp97D5raIc:APA91bH1lLV71EcU0wPOR6-lEQvo5t4p2HyBPM-Ucx5O8ZYqq_LhMd_MaQyX38qYXcBwVhqZGPaFIFV5Ud6-ESlPoS5OT_zgW9vuWEyEnpXLrY0Uu5-lTzgZb1BnsHsTidlsP37mTo3x","expirationTime":null,"keys":{"p256dh":"BHEAyCxb4fTj7B9twvOI_afrMCc4_7okw9EQJK7o3P2_jU1LFz4jBf-tVe0maS2Vb0jv_mASZIjhyh6lEr5-hHE","auth":"9kCExV0RucBCO9UdeNtCuQ"}}`,
		`{"endpoint":"https://fcm.googleapis.com/fcm/send/eJdotu-BVd4:APA91bGsthfG8hrpAmrXVbDsteJ1_Yzk-O2Fyx6617Ygup1xuxZ7gWOH4myrFRz5AX5U3dT7tEWGMtkfc587qOWW-131LX-tBGWT_TMX7mSMu9pTxYFsu_EMnw6h1LdwBm_PSZeqBZpx","expirationTime":null,"keys":{"p256dh":"BMm0b-KaJU8sDicHB-TlSaSGjXzZUe-qCTPlPn5Ia1MouNYM7te6_V2MNvmaqrGoXkPxN7YjrE3mIyT0MYGDu3g","auth":"0hZxVMyof_CyrLVSkbEh6g"}}`,
		`{
			"endpoint": "https://fcm.googleapis.com/fcm/send/cNuFo9MI5Io:APA91bH4dk0XPWmRqLWN09Dt4xYxjrgGEjR5DhPtAuD94KRGfBujVzNc1F1njoOxAZNZm9XtDz3gkHcWWbpDxyXOhwYwTXF-XLJO-zNRTurExAvXxSjKPBoRuPB1Y8YVHHGfJRlr_oQE",
			"expirationTime": null,
			"keys": {
				"p256dh": "BHOEapFpyOw1m2Gtq8-5R4NXDYx6GC6uKN_X9JiqgnKED06I5sQdFTTQSEnvyGmzLBx3dxDZa4a-ZvlFwCggTG4",
				"auth": "e9YqgIPIoDje3TTEHqNpJA"
			}
		}`,
	}
	json.Unmarshal([]byte(subscription[2]), s)

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
		TTL:             3600,
	})
	// 410 or 210
	fmt.Println(resp.StatusCode)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()
}
