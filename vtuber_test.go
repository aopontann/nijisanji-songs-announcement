package main

import (
	"fmt"
	"testing"
)

func TestGetPlaylistsID(t *testing.T) {
	// DB接続初期化
	DBInit()
	defer DB.Close()

	plist, err := GetPlaylistsID()
	if err != nil {
		t.Errorf("Playlists Error")
	}
	fmt.Println(plist)
}
