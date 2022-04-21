package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"
)

func YoutubeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
        w.WriteHeader(http.StatusMethodNotAllowed) // 405
        w.Write([]byte("POSTだけだよ"))
        return
    }

	ctx := context.Background()
	videoId, err := YoutubeSearchList(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	fmt.Printf("videoId=%s\n", videoId)

	err = YoutubeVideoList(ctx, videoId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.Write([]byte("Youtube OK"))
}

const endpoint = "https://api.twitter.com/2/tweets"

func TwitterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
        w.WriteHeader(http.StatusMethodNotAllowed)
        w.Write([]byte("POSTだけだよ"))
        return
    }
	dtAfter := time.Now().UTC().Format("2006-01-02 15:04:00")
	dtBefore := time.Now().UTC().Add(5 * time.Minute).Format("2006-01-02 15:04:00")

	log.Printf("twitter-get-youtube-video: %s ~ %s\n", dtAfter, dtBefore)

	videoList, err := GetVideos(dtAfter, dtBefore)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	for _, video := range videoList {
		err := PostTweet(video.Id, video.Title)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
	}
	w.Write([]byte("Twitter OK"))
}
