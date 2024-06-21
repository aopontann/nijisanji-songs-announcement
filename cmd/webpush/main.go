package main

import (
	"log"
	"time"

	nsa "github.com/aopontann/nijisanji-songs-announcement"
)

// func strToByte(text string) string {
// 	strList := []string{}
// 	for _, b := range []byte(text) {
// 		strList = append(strList, strconv.Itoa(int(b)))
// 	}
// 	return strings.Join(strList, "_")
// }

func main() {
	fcm := nsa.NewFCM()
	st, _ := time.Parse("2006-01-02 15:04:05", "2024-05-26 13:20:00")
	video := nsa.Video{
		ID:        "Sqpmvv8uulM",
		Title:     "心予報/歌わせていただきました。",
		Duration:  "PT3M24S",
		Viewers:   0,
		Content:   "upcoming",
		Announced: false,
		StartTime: st,
		Thumbnail: "https://i.ytimg.com/vi/OPzbUoLxYyE/default.jpg",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// fmt.Println(b)
	tokens := []string{"fJoacu4b1JPNfE2rW73N_R:APA91bFcC-yOj4ZB8J31QHrGGrh1eHUVrxYf4ZkKsmnrzA9O_fbkb_Ml5KP33i7YpvoDT7Wd9MVt6_eMvrHTTfkhdEvgWF3CyTOQvOlWIOvbl84vbD84oiuxDacJWBShXwz52TFlcO7G"}

	// err := fcm.SetTopic(token, strToByte("石神のぞみ"))
	err := fcm.SongNotification2(video, tokens)
	// err := fcm.DeleteTopic(token, strToByte("石神のぞみ"))
	// err := fcm.SendWithTopic(strToByte("石神のぞみ"))
	if err != nil {
		log.Fatalln(err)
	}

	// This registration token comes from the client FCM SDKs.
	// tokens := []string{
	// 	"fJoacu4b1JPNfE2rW73N_R:APA91bFOuF-aDC2LFxoTYdSkEzZ9FJHzirqi8frhOK00tjZHCzuQlw3Lfmr3ByxtTM_MiXhbSgLUlwEJh_sGEeNCrbLfxN73eIbIrWTnxiw6cOxFBA7vioyannb_eSl-T1-NjwhGE9cO",
	// }
	// See documentation on defining a message payload.
	// message := &messaging.MulticastMessage{
	// 	Data: map[string]string{
	// 		"title": "5分後に公開",
	// 		"body": "動画タイトル",
	// 		"icon": "https://avatars.githubusercontent.com/u/74401628?v=4",
	// 	},
	// 	Tokens: tokens,
	// 	Webpush: &messaging.WebpushConfig{
	// 		Headers: map[string]string{
	// 			"Urgency": "high",
	// 		},
	// 	},

	// }
	// message := &messaging.Message{
	// 	Data: map[string]string{
	// 		"title": "5分後に公開",
	// 		"body":  "動画タイトル",
	// 		"icon":  "https://avatars.githubusercontent.com/u/74401628?v=4",
	// 	},
	// 	Topic: "石神のぞみ",
	// }

	// Send a message to the device corresponding to the provided
	// registration token.
	// response, err := client.SendEachForMulticast(ctx, message)
	// response, err := client.Send(ctx, message)
	// if err != nil {
	// 	log.Fatalln(err)
	// }
	// Response is a message ID string.
	// fmt.Println("Successfully sent message:", response)
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
