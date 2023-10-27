package main_test

import (
	"testing"
	"github.com/aopontann/nijisanji-songs-announcement"
)

func TestTweet(t *testing.T) {
	tw := main.NewTwitter()
	err := tw.Id("test").Title("title").Tweet()
	if err != nil {
		t.Errorf("Tweets Error")
	}
}

