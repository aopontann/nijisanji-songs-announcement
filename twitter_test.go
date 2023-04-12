package main

import (
	"testing"
)

func TestTweets(t *testing.T) {
	err := Tweets()
	if err != nil {
		t.Errorf("Tweets Error")
	}
}