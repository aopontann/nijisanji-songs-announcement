package nsa

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
)

func TestSearch(t *testing.T) {
	db, err := sql.Open("mysql", os.Getenv("DSN"))
	if err != nil {
		t.Errorf("sql.Open")
	}

	yt, err := NewYoutube(db)
	if err != nil {
		t.Errorf("youtube.New(queries)")
	}

	m := map[string]int64{"UU_4tXjqecqox5Uc05ncxpxg": 100, "UC_82HBGtvwN1hcGeOGHzUBQ": 469}

	list, err := yt.CheckItemCount(m)
	if err != nil {
		t.Errorf("yt.NewUpload(m)")
	}
	fmt.Println(list)
}

func TestItems(t *testing.T) {
	db, err := sql.Open("mysql", os.Getenv("DSN"))
	if err != nil {
		t.Errorf("sql.Open")
	}

	yt, err := NewYoutube(db)
	if err != nil {
		t.Errorf("youtube.New(queries)")
	}

	vidList, err := yt.Items([]string{"UU_4tXjqecqox5Uc05ncxpxg", "UU_82HBGtvwN1hcGeOGHzUBQ"})
	if err != nil {
		t.Errorf("youtube.New(queries)")
	}
	fmt.Println(vidList)
}
