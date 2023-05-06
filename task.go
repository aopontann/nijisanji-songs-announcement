package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

func UpdateItemCountTask() error {
	plist, err := Playlists()
	if err != nil {
		return err
	}
	err = plist.Save()
	if err != nil {
		return err
	}
	return nil
}

func CheckNewVideoTask() error {
	plist, err := Playlists()
	if err != nil {
		return err
	}

	slist, err := plist.Select()
	if err != nil {
		return err
	}

	vid, err := slist.Items()
	if err != nil {
		return err
	}

	vlist, err := vid.Video()
	if err != nil {
		return err
	}

	vlist, err = vlist.Select().IsNijisanji()
	if err != nil {
		return err
	}

	vlist, err = vlist.NotExist()
	if err != nil {
		return err
	}

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

	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("DB.Begin failed")
	}

	stmt, err := tx.Prepare("INSERT IGNORE INTO videos(id, title, songConfirm, scheduled_start_time) VALUES(?,?,?,?)")
	if err != nil {
		return fmt.Errorf("tx.Prepare failed")
	}
	for _, video := range vlist {
		// "2022-03-28 11:00:00"形式に変換
		rep1 := strings.Replace(video.LiveStreamingDetails.ScheduledStartTime, "T", " ", 1)
		scheduledStartTime := strings.Replace(rep1, "Z", "", 1)
		// DBに動画情報を保存
		_, err := stmt.Exec(video.Id, video.Snippet.Title, true, scheduledStartTime)
		if err != nil {
			if tx.Rollback() != nil {
				return fmt.Errorf("tx.Rollback() failed")
			}
			return fmt.Errorf("stmt.Exec failed")
		}
	}

	// 動画が削除されて動画数が減っていても、上書きする
	stmt, err = tx.Prepare("UPDATE vtubers SET item_count = ? WHERE id = ? AND item_count != ?")
	if err != nil {
		return err
	}
	for _, list := range plist {
		_, err := stmt.Exec(list.ItemCount, list.ID, list.ItemCount)
		if err != nil {
			if tx.Rollback() != nil {
				return fmt.Errorf("tx.Rollback() failed")
			}
			return fmt.Errorf("stmt.Exec failed")
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
