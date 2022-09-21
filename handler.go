package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

var tw = Twitter{}
var yt = Youtube{}

func YoutubeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed) // 405
		w.Write([]byte("POSTだけだよ"))
		return
	}

	vid, err := yt.Search()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	yvr, err := yt.Video(vid)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	ysr, err := yt.Select(yvr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	err = yt.Save(ysr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.Write([]byte("Youtube OK"))
}

func TwitterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("POSTだけだよ"))
		return
	}
	dtAfter := time.Now().UTC().Add(1 * time.Second).Format("2006-01-02 15:04:05")
	dtBefore := time.Now().UTC().Add(5 * time.Minute).Format("2006-01-02 15:04:00")

	log.Info().Str("severity", "INFO").Str("service", "tweet").Str("datetime", fmt.Sprintf("%s ~ %s\n", dtAfter, dtBefore)).Send()

	videoList, err := GetVideos(dtAfter, dtBefore)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	for _, video := range videoList {
		// changed, err := yt.CheckVideo(video.Id)
		log.Info().Str("severity", "INFO").Str("service", "tweet").Str("id", video.Id).Str("title", video.Title).Send()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		// if changed {
		// 	continue
		// }
		err = tw.Post(video.Id, video.Title)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
	}
	w.Write([]byte("Twitter OK"))
}

func TwitterSearchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("POSTだけだよ"))
		return
	}

	tsr, err := tw.Search()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	ytcr, err := tw.Select(tsr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	err = DBSave(ytcr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.Write([]byte("Twitter OK"))
}
