package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

func YoutubeHandler(w http.ResponseWriter, _ *http.Request) {
	ctx := context.Background()
	youtubeService, err := youtube.NewService(ctx, option.WithAPIKey(os.Getenv("YOUTUBE_API_KEY")))
	if err != nil {
		log.Fatalf("Error creating YouTube client: %v", err)
	}
	part := []string{"id", "snippet"}
	call := youtubeService.Videos.List(part).Id("2yD0iNl2E-I,CT7tJZwWLdQ")
	response, err := call.Do()
	if err != nil {
		// The channels.list method call returned an error.
		log.Fatalf("Error making API call to list channels: %v\n", err.Error())
	}
	for _, video := range response.Items {
		fmt.Printf("id=%v title%v", video.Id, video.Snippet.Title)
	}
}
