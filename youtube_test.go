package main

import (
	"testing"
)

func TestSearch(t *testing.T) {
	// DB接続初期化
	DBInit()
	defer DB.Close()
	// YouTube Data API 初期化
	YTNew()

	vid, err := Search()
	if err != nil {
		t.Errorf("Search Error")
	}
	vlist, err := vid.Video()
	if err != nil {
		t.Errorf("Video Error")
	}
	_, err = vlist.Select()
	if err != nil {
		t.Errorf("Select Error")
	}
}

func TestActivities(t *testing.T) {
	// DB接続初期化
	DBInit()
	defer DB.Close()
	// YouTube Data API 初期化
	YTNew()

	vid, err := Activities()
	if err != nil {
		t.Errorf("Search Error")
	}
	vlist, err := vid.Video()
	if err != nil {
		t.Errorf("Video Error")
	}
	slist, err := vlist.Select()
	if err != nil {
		t.Errorf("Select Error")
	}
	err = slist.Save()
	if err != nil {
		t.Errorf("Save Error")
	}
}