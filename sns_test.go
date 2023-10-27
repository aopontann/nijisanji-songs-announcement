package main

import (
	"testing"
)

func TestTweet(t *testing.T) {
	tw := NewTwitter()
	err := tw.Id("test").Title("title").Tweet()
	if err != nil {
		t.Errorf("Tweets Error")
	}
}
