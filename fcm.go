package nsa

import (
	"context"
	"log"
	"log/slog"
	"time"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"github.com/avast/retry-go/v4"
)

type FCM struct {
	Client *messaging.Client
}

type NotificationVideo struct {
	ID        string
	Title     string
	Thumbnail string
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

func (c *FCM) Notification(title string, tokens []string, video *NotificationVideo) error {
	nofication := &messaging.Notification{
		Title:    title,
		Body:     video.Title,
		ImageURL: video.Thumbnail,
	}
	webpush := &messaging.WebpushConfig{
		Headers: map[string]string{
			"Urgency": "high",
		},
		FCMOptions: &messaging.WebpushFCMOptions{
			Link: "https://youtu.be/" + video.ID,
		},
	}

	for i := 0; i*500 <= len(tokens); i++ {
		var t []string
		if len(tokens) > 500*(i+1) {
			t = tokens[i*500 : (i+1)*500]
		} else {
			t = tokens[500*i:]
		}
		message := &messaging.MulticastMessage{
			Notification: nofication,
			Tokens:       t,
			Webpush:      webpush,
		}

		// 3回までリトライ　1秒後にリトライ
		err := retry.Do(
			func() error {
				response, err := c.Client.SendEachForMulticast(context.Background(), message)
				if err != nil {
					slog.Error("Notification error",
						slog.String("severity", "ERROR"),
						slog.String("message", err.Error()),
					)
					return err
				}
				for _, r := range response.Responses {
					if r.Error != nil {
						slog.Warn("Notification warning",
							slog.String("severity", "WARNING"),
							slog.String("message", r.Error.Error()),
						)
					}
				}
				return nil
			},
			retry.Attempts(3),
			retry.Delay(2*time.Second),
		)
		if err != nil {
			slog.Error("Notification error (retry)",
				slog.String("severity", "ERROR"),
				slog.String("message", err.Error()),
			)
			return err
		}

	}

	return nil
}
