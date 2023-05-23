package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/aopontann/nijisanji-songs-announcement/cmd/selection"
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

	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("DB.Begin failed")
	}
	qtx := queries.WithTx(tx)

	// 動画が削除されて動画数が減っていても、上書きする
	for pid, count := range itemCountList {
		err = qtx.UpdatePlaylistItemCount(ctx, ndb.UpdatePlaylistItemCountParams{
			ItemCount: int32(count),
			ID: pid,
			ItemCount_2: int32(count),
		})
		if err != nil {
			if tx.Rollback() != nil {
				return fmt.Errorf("tx.Rollback() failed")
			}
			return fmt.Errorf("UpdatePlaylistItemCount failed")
		}
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
			err = SendMail("【開発用】新しい動画がアップロードされました", fmt.Sprintf("https://www.youtube.com/watch?v=%s", v.Id))
		} else {
			err = SendMail("新しい動画がアップロードされました", fmt.Sprintf("https://www.youtube.com/watch?v=%s", v.Id))
		}
		if err != nil {
			return err
		}
	}

	ctx := context.Background()
	queries := ndb.New(db)

	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("DB.Begin failed")
	}
	qtx := queries.WithTx(tx)

	for _, video := range vlist {
		t, _ := time.Parse(time.RFC3339, video.LiveStreamingDetails.ScheduledStartTime)
		// DBに動画情報を保存
		err := qtx.CreateVideo(ctx, ndb.CreateVideoParams{
			ID:                 video.Id,
			Title:              video.Snippet.Title,
			Songconfirm:        1,
			ScheduledStartTime: t,
		})
		if err != nil {
			if tx.Rollback() != nil {
				return fmt.Errorf("tx.Rollback() failed")
			}
			return fmt.Errorf("CreateVideo failed")
		}
	}

	// 動画が削除されて動画数が減っていても、上書きする
	for pid, count := range itemCountList {
		err = qtx.UpdatePlaylistItemCount(ctx, ndb.UpdatePlaylistItemCountParams{
			ItemCount: int32(count),
			ID: pid,
			ItemCount_2: int32(count),
		})
		if err != nil {
			if tx.Rollback() != nil {
				return fmt.Errorf("tx.Rollback() failed")
			}
			return fmt.Errorf("UpdatePlaylistItemCount failed")
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func TweetTask() error {
	dtAfter := time.Now().UTC().Add(1 * time.Second).Format("2006-01-02 15:04:05")
	dtBefore := time.Now().UTC().Add(5 * time.Minute).Format("2006-01-02 15:04:00")

	log.Info().Str("severity", "INFO").Str("service", "tweet").Str("datetime", fmt.Sprintf("%s ~ %s\n", dtAfter, dtBefore)).Send()

	videoList, err := GetVideos(dtAfter, dtBefore)
	if err != nil {
		return err
	}

	for _, video := range videoList {
		// changed, err := yt.CheckVideo(video.Id)
		log.Info().Str("severity", "INFO").Str("service", "tweet").Str("id", video.ID).Str("title", video.Title).Send()
		if err != nil {
			return err
		}
		// if changed {
		// 	continue
		// }
		err = video.Tweets()
		if err != nil {
			return err
		}
	}
	return nil
}
