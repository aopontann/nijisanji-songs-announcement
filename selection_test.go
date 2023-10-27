package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	ndb "github.com/aopontann/nijisanji-songs-announcement/db"
)

func TestIsNijisanji(t *testing.T) {
	db, err := sql.Open("mysql", os.Getenv("DSN"))
	if err != nil {
		t.Errorf("sql.Open")
	}

	yt, err := NewYoutube(db)
	if err != nil {
		t.Errorf("youtube.New(queries)")
	}
	vList, err := yt.Video([]string{"Grs5sJ6DlnI", "Nr7XM4187H4"})
	v := NewSelect(vList, db)
	if err != nil {
		t.Errorf("video.New(vList, queries)")
	}

	v, err = v.IsNijisanji()
	if err != nil {
		t.Errorf("video.New(vList, queries)")
	}

	res := v.GetList()
	for _, v := range res {
		fmt.Println(v.Id)
	}
}

func TestNotExists(t *testing.T) {
	db, err := sql.Open("mysql", os.Getenv("DSN"))
	if err != nil {
		t.Errorf("sql.Open")
	}

	yt, err := NewYoutube(db)
	if err != nil {
		t.Errorf("youtube.New(queries)")
	}
	vList, err := yt.Video([]string{"Grs5sJ6DlnI", "G8yMXxscPFg"})
	v := NewSelect(vList, db)
	if err != nil {
		t.Errorf("video.New(vList, queries)")
	}

	v, err = v.IsNotExists()
	if err != nil {
		t.Errorf("v.NotExists()")
	}

	res := v.GetList()
	for _, v := range res {
		fmt.Println(v.Id)
	}
}

func TestMap(t *testing.T) {
	m := map[string]int{"foo": 32, "bar": 12, "ooo": 45}
	for i, a := range m {
		fmt.Println(i)
		fmt.Println(a)
	}
	delete(m, "bbb")
}

func TestCreateVideo(t *testing.T) {
	ctx := context.Background()
	db, err := sql.Open("mysql", os.Getenv("DSN"))
	if err != nil {
		t.Errorf("sql.Open")
	}

	queries := ndb.New(db)
	st, _ := time.Parse(time.RFC3339, "2024-04-19T12:00:16Z")

	err = queries.CreateVideo(ctx, ndb.CreateVideoParams{
		ID:                 "id0001",
		Title:              "title0001",
		Songconfirm:        1,
		ScheduledStartTime: st,
	})
	if err != nil {
		t.Errorf("queries.CreateVideo")
	}
}
