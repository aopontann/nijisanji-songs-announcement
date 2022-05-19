package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

func YoutubeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed) // 405
		w.Write([]byte("POSTだけだよ"))
		return
	}

	vid, err := YoutubeSearchList()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	err = YoutubeVideoList(vid)
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
	dtAfter := time.Now().UTC().Format("2006-01-02 15:04:05")
	dtBefore := time.Now().UTC().Add(5 * time.Minute).Format("2006-01-02 15:04:00")

	log.Info().Str("service", "tweet").Str("datetime", fmt.Sprintf("%s ~ %s\n", dtAfter, dtBefore)).Send()

	videoList, err := GetVideos(dtAfter, dtBefore)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	for _, video := range videoList {
		log.Info().Str("service", "tweet").Str("id", video.Id).Str("title", video.Title).Send()
		err := PostTweet(video.Id, video.Title)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
	}
	w.Write([]byte("Twitter OK"))
}
