package main

import (
	"context"
	"fmt"
	"log"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
)

func main() {
	app, err := firebase.NewApp(context.Background(), nil)
	if err != nil {
		log.Fatalf("error initializing app: %v\n", err)
	}
	ctx := context.Background()
	client, err := app.Messaging(ctx)
	if err != nil {
		log.Fatalf("error getting Messaging client: %v\n", err)
	}

	// This registration token comes from the client FCM SDKs.
	tokens := []string{
		"dVB4sxyDVFizThW0JqZCeo:APA91bG1S30VUEYiCxA_W0ZOS8OyS6Wifufu4xvj_XnVP6poXapOFB9jucHn_3Z0bYSKDRJ1I5Wq3Q7CZn0WXm3hT6JSAATDspNRroVOSuzxzrpogtyAW6PHjiGgNav8BWFBUDFeJY24",
		"dAbLpFsgrEvEsYUHoQhdAZ:APA91bHv2uy6T-XQDH9JoM63X_tEyxlip-90jM_ZF3-VuSb9PWbARLF2o7HJo9YcfXUvE5x351kuWQBIRflKGeere7ZInDbUteKsb6yqLH31XZmXoCUBOMds7biOZ7bUnr_st6aorSmL",
	}
	// See documentation on defining a message payload.
	message := &messaging.MulticastMessage{
		Data: map[string]string{
			"title": "5分後に公開",
			"body": "動画タイトル",
			"icon": "https://avatars.githubusercontent.com/u/74401628?v=4",
		},
		Tokens: tokens,
		Webpush: &messaging.WebpushConfig{
			Headers: map[string]string{
				"Urgency": "high",
			},
		},
	}

	// Send a message to the device corresponding to the provided
	// registration token.
	response, err := client.SendEachForMulticast(ctx, message)
	if err != nil {
		log.Fatalln(err)
	}
	// Response is a message ID string.
	fmt.Println("Successfully sent message:", response)
}

// func main2() {
// 	addr := os.Getenv("MAIL_ADDRESS")
// 	publicKey := os.Getenv("WEBPUSH_PUBLIC_KEY")
// 	privateKey := os.Getenv("WEBPUSH_PRIVATE_KEY")

// 	// Decode subscription
// 	s := &webpush.Subscription{}
// 	subscription := []string{
// 		`{"endpoint":"https://fcm.googleapis.com/fcm/send/eTp97D5raIc:APA91bH1lLV71EcU0wPOR6-lEQvo5t4p2HyBPM-Ucx5O8ZYqq_LhMd_MaQyX38qYXcBwVhqZGPaFIFV5Ud6-ESlPoS5OT_zgW9vuWEyEnpXLrY0Uu5-lTzgZb1BnsHsTidlsP37mTo3x","expirationTime":null,"keys":{"p256dh":"BHEAyCxb4fTj7B9twvOI_afrMCc4_7okw9EQJK7o3P2_jU1LFz4jBf-tVe0maS2Vb0jv_mASZIjhyh6lEr5-hHE","auth":"9kCExV0RucBCO9UdeNtCuQ"}}`,
// 		`{"endpoint":"https://fcm.googleapis.com/fcm/send/eJdotu-BVd4:APA91bGsthfG8hrpAmrXVbDsteJ1_Yzk-O2Fyx6617Ygup1xuxZ7gWOH4myrFRz5AX5U3dT7tEWGMtkfc587qOWW-131LX-tBGWT_TMX7mSMu9pTxYFsu_EMnw6h1LdwBm_PSZeqBZpx","expirationTime":null,"keys":{"p256dh":"BMm0b-KaJU8sDicHB-TlSaSGjXzZUe-qCTPlPn5Ia1MouNYM7te6_V2MNvmaqrGoXkPxN7YjrE3mIyT0MYGDu3g","auth":"0hZxVMyof_CyrLVSkbEh6g"}}`,
// 		`{"endpoint":"https://fcm.googleapis.com/fcm/send/f0kSjQuO6K0:APA91bFvuPpZhCAgCRzw46f7lUm4sHdok_DSBKoYm1hPtMX5NG4eIzF-0tvjBVLSXtpTiOvK_ghQmBPRnKm4Bcap4ytv3TQnCTuxjpDmlVAAswmlA2mksZejrJt99R3j_WF5sxSz2aO-","expirationTime":null,"keys":{"p256dh":"BLtb719EYghhmIt-f6-2f1LSXITJKKymxnlDZQEJmpapmORwmDmTMlCoFcV8cc78eio5F2uhKpH5uwx5jjnj4bY","auth":"Vxm-QF74P0xEkZVHdwt3Zg"}}`,
// 	}
// 	json.Unmarshal([]byte(subscription[2]), s)

// 	req_message := `
// 	{
// 		"title": "あたらしいWeb Pushが送信されました",
// 		"body": "Web Pushです",
// 		"data": {
// 		  "url": "https://www.youtube.com/watch?v=gJk_vTwiduU"
// 		},
// 		"icon": "https://avatars.githubusercontent.com/u/74401628?v=4"
// 	  }
// 	`

// 	// Send Notification
// 	resp, err := webpush.SendNotification([]byte(req_message), s, &webpush.Options{
// 		Subscriber:      addr,
// 		VAPIDPublicKey:  publicKey,
// 		VAPIDPrivateKey: privateKey,
// 		TTL:             3600,
// 	})
// 	// 410 or 210
// 	fmt.Println(resp.StatusCode)
// 	fmt.Println(resp.Header)
// 	if err != nil {
// 		fmt.Println(err)
// 	}
// 	defer resp.Body.Close()
// }
