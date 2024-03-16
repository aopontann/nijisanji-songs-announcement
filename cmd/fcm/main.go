package main

import (
	"context"
	"fmt"
	"log"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	// "google.golang.org/api/option"
)

func main() {
	// opt := option.WithCredentialsFile("token.json")
	ctx := context.Background()
	app, err := firebase.NewApp(ctx, nil)
	// app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatalf("error initializing app: %v\n", err)
	}

	client, err := app.Messaging(ctx)
	if err != nil {
		log.Fatalf("error getting Messaging client: %v\n", err)
	}

	// This registration token comes from the client FCM SDKs.
	registrationToken := ""

	// See documentation on defining a message payload.
	message := &messaging.Message{
		Data: map[string]string{
			"title": "video_title",
			"url":  "http://example.com",
		},
		// Notification: &messaging.Notification{
		// 	Title: "Price drop22",
		// 	Body:  "5% off all electronics",
		// },
		Token: registrationToken,
	}

	// Send a message to the device corresponding to the provided
	// registration token.
	response, err := client.Send(ctx, message)
	if err != nil {
		log.Fatalln(err.Error())
	}
	// Response is a message ID string.
	fmt.Println("Successfully sent message:", response)
}
