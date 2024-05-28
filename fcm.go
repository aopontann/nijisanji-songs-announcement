package nsa

import (
	"context"
	"fmt"
	"log"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
)

type FCM struct {
	Client *messaging.Client
}

func NewFCM() *FCM {
	ctx := context.Background()
	app, err := firebase.NewApp(context.Background(), nil)
	if err != nil {
		log.Fatalf("error initializing app: %v\n", err)
	}

	client, err := app.Messaging(ctx)
	if err != nil {
		log.Fatalf("error getting Messaging client: %v\n", err)
	}
	return &FCM{client}
}

func (c *FCM) Send(video Video, title string, tokens []string, urgency string) error {
	message := &messaging.MulticastMessage{
		Data: map[string]string{
			"title": title,
			"body":  video.Title,
			"url":   "https://youtu.be/" + video.ID,
			"icon":  video.Thumbnail,
		},
		Tokens: tokens,
		Webpush: &messaging.WebpushConfig{
			Headers: map[string]string{
				"Urgency": urgency,
			},
		},
	}

	fmt.Println("tokens", tokens)

	response, err := c.Client.SendEachForMulticast(context.Background(), message)
	if err != nil {
		return err
	}
	for _, r := range response.Responses {
		fmt.Println(r)
	}
	// Response is a message ID string.
	fmt.Println("Successfully sent message:", response)
	return nil
}
