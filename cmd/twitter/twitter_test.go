package twitter_test

import (
	"testing"

	"github.com/aopontann/nijisanji-songs-announcement/cmd/twitter"
)

func TestTweet(t *testing.T) {
	tw := twitter.New()
	err := tw.Id("test").Title("title").Tweet()
	if err != nil {
		t.Errorf("Tweets Error")
	}
}