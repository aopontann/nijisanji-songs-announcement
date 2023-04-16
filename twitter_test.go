package main

import (
	"testing"
)

func TestTweets(t *testing.T) {
	video := GetVideoInfo{ID: "test", Title: "test"}
	err := video.Tweets()
	if err != nil {
		t.Errorf("Tweets Error")
	}
}