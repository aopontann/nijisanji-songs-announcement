package main

import (
	"fmt"

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
	err := nsa.NewMail().Subject("歌みた動画判定").Id("id001").Title("タイトル").Send()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	// fcm := nsa.NewFCM()
	// st, _ := time.Parse("2006-01-02 15:04:05", "2024-05-26 13:20:00")
	// video := nsa.Video{
	// 	ID:        "Sqpmvv8uulM",
	// 	Title:     "心予報/歌わせていただきました。",
	// 	Duration:  "PT3M24S",
	// 	Viewers:   0,
	// 	Content:   "upcoming",
	// 	StartTime: st,
	// 	Thumbnail: "https://i.ytimg.com/vi/0Jh4HIL43uQ/hqdefault_live.jpg",
	// 	CreatedAt: time.Now(),
	// 	UpdatedAt: time.Now(),
	// }

	// // fmt.Println(b)
	// tokens := []string{"fs7MrEyieVJVisNjg32jLp:APA91bFLZdFU53W9k6biHDa19onI7us40J5TrW_SqIp_IRFhUJzCxFu7GLGchWElkWNiwUl-zJnoOn0UG00ZrxF6kaZ0CARKLDxqmLrxY3OexRGTz9GdY7LPX-4MzsCHQ0x7v1EzYh1f"}

	// // err := fcm.SetTopic(token, strToByte("石神のぞみ"))
	// err := fcm.Notification("5分後に公開", video, tokens)
	// // err := fcm.DeleteTopic(token, strToByte("石神のぞみ"))
	// // err := fcm.SendWithTopic(strToByte("石神のぞみ"))
	// if err != nil {
	// 	log.Fatalln(err)
	// }

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
