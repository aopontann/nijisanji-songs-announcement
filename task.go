package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aopontann/nijisanji-songs-announcement/cmd/mail"
	"github.com/aopontann/nijisanji-songs-announcement/cmd/misskey"
	"github.com/aopontann/nijisanji-songs-announcement/cmd/selection"
	"github.com/aopontann/nijisanji-songs-announcement/cmd/twitter"
	"github.com/aopontann/nijisanji-songs-announcement/cmd/youtube"
	ndb "github.com/aopontann/nijisanji-songs-announcement/db"
	"github.com/rs/zerolog/log"
)

func UpdateItemCountTask(db *sql.DB) error {
	yt, err := youtube.New(db)
	if err != nil {
		return err
	}

	itemCountList, err := yt.Playlists()
	if err != nil {
		return err
	}

	ctx := context.Background()
	queries := ndb.New(db)

	// 動画が削除されて動画数が減っていても、上書きする
	for pid, count := range itemCountList {
		err = queries.UpdatePlaylistItemCount(ctx, ndb.UpdatePlaylistItemCountParams{
			ItemCount:   int32(count),
			ID:          strings.Replace(pid, "UU", "UC", 1),
			ItemCount_2: int32(count),
		})
		if err != nil {
			return fmt.Errorf(err.Error())
		}
		log.Info().
			Str("severity", "INFO").
			Str("service", "db-update-playlist-count").
			Str("PlaylistId", pid).
			Int64("ItemCount",count).
			Send()
	}
	return nil
}

func CheckNewVideoTask(db *sql.DB) error {
	yt, err := youtube.New(db)
	if err != nil {
		return err
	}

	itemCountList, err := yt.Playlists()
	if err != nil {
		return err
	}

	pidList, err := yt.CheckItemCount(itemCountList)
	if err != nil {
		return err
	}

	vidList, err := yt.Items(pidList)
	if err != nil {
		return err
	}

	vList, err := yt.Video(vidList)
	if err != nil {
		return err
	}

	selc := selection.New(vList, db)

	selc, err = selc.IsLiveStreaming().IsNotClipped().IsNotLiveFinished().IsUnder10min().IsNijisanji()
	if err != nil {
		return err
	}
	selc, err = selc.IsNotExists()
	if err != nil {
		return err
	}

	vlist := selc.GetList()

	for _, v := range vlist {
		var err error
		if os.Getenv("ENV") == "dev" {
			err = mail.Send("【開発用】新しい動画がアップロードされました", fmt.Sprintf("https://www.youtube.com/watch?v=%s", v.Id))
		} else {
			err = mail.Send("新しい動画がアップロードされました", fmt.Sprintf("https://www.youtube.com/watch?v=%s", v.Id))
		}
		if err != nil {
			return err
		}
	}

	ctx := context.Background()
	queries := ndb.New(db)

	for _, video := range vlist {
		t, _ := time.Parse(time.RFC3339, video.LiveStreamingDetails.ScheduledStartTime)
		// DBに動画情報を保存
		err := queries.CreateVideo(ctx, ndb.CreateVideoParams{
			ID:                 video.Id,
			Title:              video.Snippet.Title,
			Songconfirm:        1,
			ScheduledStartTime: t,
		})
		if err != nil {
			return fmt.Errorf("CreateVideo failed")
		}
	}

	// 動画が削除されて動画数が減っていても、上書きする
	for pid, count := range itemCountList {
		cid := strings.Replace(pid, "UU", "UC", 1)
		err = queries.UpdatePlaylistItemCount(ctx, ndb.UpdatePlaylistItemCountParams{
			ItemCount:   int32(count),
			ID:          cid,
			ItemCount_2: int32(count),
		})
		if err != nil {
			return fmt.Errorf(err.Error())
		}
	}

	return nil
}

func TweetTask(db *sql.DB) error {
	queries := ndb.New(db)
	now, _ := time.Parse(time.RFC3339, time.Now().UTC().Format("2006-01-02T15:04:00Z"))
	tAfter := now.Add(1 * time.Second)
	tBefore := now.Add(5 * time.Minute)

	log.Info().Str("severity", "INFO").Str("service", "tweet").Str("datetime", fmt.Sprintf("%v ~ %v\n", tAfter, tBefore)).Send()

	ctx := context.Background()
	vList, err := queries.ListVideoIdTitle(ctx, ndb.ListVideoIdTitleParams{
		ScheduledStartTime: tAfter,
		ScheduledStartTime_2: tBefore,
	})
	if err != nil {
		return err
	}

	tw := twitter.New()

	for _, video := range vList {
		// changed, err := yt.CheckVideo(video.Id)
		log.Info().Str("severity", "INFO").Str("service", "tweet").Str("id", video.ID).Str("title", video.Title).Send()
		if err != nil {
			return err
		}

		err = tw.Id(video.ID).Title(video.Title).Tweet()
		if err != nil {
			return err
		}
	}
	return nil
}

func MisskeyPostTask(db *sql.DB) error {
	queries := ndb.New(db)
	now, _ := time.Parse(time.RFC3339, time.Now().UTC().Format("2006-01-02T15:04:00Z"))
	tAfter := now.Add(1 * time.Second)
	tBefore := now.Add(5 * time.Minute)

	log.Info().Str("severity", "INFO").Str("service", "misskey").Str("datetime", fmt.Sprintf("%v ~ %v\n", tAfter, tBefore)).Send()

	ctx := context.Background()
	vList, err := queries.ListVideoIdTitle(ctx, ndb.ListVideoIdTitleParams{
		ScheduledStartTime: tAfter,
		ScheduledStartTime_2: tBefore,
	})
	if err != nil {
		return err
	}

	mk := misskey.New(os.Getenv("MISSKEY_TOKEN"))

	for _, v := range vList {
		log.Info().Str("severity", "INFO").Str("service", "tweet").Str("id", v.ID).Str("title", v.Title).Send()
		err = mk.Post(v.ID, v.Title)
		if err != nil {
			return err
		}
	}
	return nil
}