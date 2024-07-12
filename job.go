package nsa

import (
	"context"
	"log/slog"
	"slices"
	"time"

	"github.com/avast/retry-go"
	"github.com/uptrace/bun"
)

type Job struct {
	db  *DB
	yt  *Youtube
	fcm *FCM
}

func NewJobs(youtubeApiKey string, db *bun.DB) *Job {
	return &Job{
		yt:  NewYoutube(youtubeApiKey),
		db:  NewDB(db),
		fcm: NewFCM(),
	}
}

func (j *Job) CheckNewVideoJob() error {
	// チャンネルIDリストを取得
	cids, err := j.db.ChannelIDs()
	if err != nil {
		return err
	}

	// 過去5分間に投稿された動画を取得
	vids, err := j.yt.RssFeed(cids)
	if err != nil {
		return err
	}

	// 動画情報を取得
	videos, err := j.yt.Videos(vids)
	if err != nil {
		return err
	}

	// 歌ってみた動画ゲリラ対応
	notExistsVideos, err := j.db.NotExistsVideos(videos)
	if err != nil {
		return err
	}
	for _, v := range notExistsVideos {
		// 5分以内に公開される動画ではない場合
		if !j.yt.IsStartWithin5m(v) {
			continue
		}
		// 歌ってみた動画に含まれているキーワードが含まれていない　除外キーワードが含まれている場合
		if !j.yt.FindSongKeyword(v) || j.yt.FindIgnoreKeyword(v) {
			continue
		}

		tokens, err := j.db.getSongTokens()
		if err != nil {
			return err
		}
		err = j.fcm.Notification(
			"まもなく公開",
			tokens,
			&NotificationVideo{
				ID:        v.Id,
				Title:     v.Snippet.Title,
				Thumbnail: v.Snippet.Thumbnails.High.Url,
			},
		)
		if err != nil {
			return err
		}
	}

	// 歌みた動画か判別しづらい動画をメールに送信する
	for _, v := range videos {
		if j.yt.FindSongKeyword(v) {
			continue
		}
		if v.LiveStreamingDetails == nil {
			continue
		}
		if v.Snippet.LiveBroadcastContent != "upcoming" {
			continue
		}
		if v.ContentDetails.Duration == "P0D" {
			continue
		}
		// 特定のキーワードを含んでいる場合
		if j.yt.FindIgnoreKeyword(v) {
			continue
		}

		err := NewMail().Subject("歌みた動画判定").Id(v.Id).Title(v.Snippet.Title).Send()
		if err != nil {
			return err
		}
	}

	// 3回までリトライ　1秒後にリトライ
	err = retry.Do(
		func() error {
			// 動画情報をDBに登録
			// 登録済みの動画は無視
			err = j.db.SaveVideos(videos)
			if err != nil {
				return err
			}
			return nil
		},
		retry.Attempts(3),
		retry.Delay(1*time.Second),
	)
	if err != nil {
		return err
	}

	return nil
}

func (j *Job) SongVideoAnnounceJob() error {
	// 5分後にプレミア公開される歌みた動画を取得
	videos, err := j.db.songVideos5m()
	if err != nil {
		slog.Error("songVideos5m",
			slog.String("severity", "ERROR"),
			slog.String("message", err.Error()),
		)
		return err
	}
	if len(videos) == 0 {
		return nil
	}

	// FCMトークンを取得
	tokens, err := j.db.getSongTokens()
	if err != nil {
		slog.Error("fcmTokens",
			slog.String("severity", "ERROR"),
			slog.String("message", err.Error()),
		)
		return err
	}

	for _, v := range videos {
		err := NewMail().Subject("歌みた動画判定").Id(v.ID).Title(v.Title).Send()
		if err != nil {
			return err
		}

		// push通知
		err = j.fcm.Notification(
			"5分後に公開",
			tokens,
			&NotificationVideo{
				ID:        v.ID,
				Title:     v.Title,
				Thumbnail: v.Thumbnail,
			})
		if err != nil {
			return err
		}
	}
	return nil
}

// 公開済みの動画、Youtube上で削除された動画をDBから削除する
func (j *Job) DeleteVideoJob() error {
	ctx := context.Background()
	var vidList []string
	err := j.db.Service.NewSelect().Model((*Video)(nil)).Column("id").Scan(ctx, &vidList)
	if err != nil {
		return err
	}

	vlist, err := j.yt.Videos(vidList)
	if err != nil {
		return err
	}

	// 削除する動画ID
	var delVidList []string
	for _, v := range vlist {
		// 公開前、公開中は削除しない
		if v.Snippet.LiveBroadcastContent != "none" {
			continue
		}
		// プレミア公開、生放送が終了した動画は削除する
		if v.LiveStreamingDetails != nil {
			delVidList = append(delVidList, v.Id)
			continue
		}

		// プレミア公開、生放送ではなく、24時間以内に公開された動画は削除しない
		t, _ := time.Parse("2006-01-02T15:04:05Z", v.Snippet.PublishedAt)
		if time.Now().AddDate(0, 0, -1).Compare(t) < 0 {
			continue
		}
		delVidList = append(delVidList, v.Id)
	}

	// youtube上で削除された動画も削除
	var resVidList []string
	for _, v := range vlist {
		resVidList = append(resVidList, v.Id)
	}
	for _, id := range vidList {
		if !slices.Contains(resVidList, id) {
			delVidList = append(delVidList, id)
		}
	}

	_, err = j.db.Service.NewDelete().Model((*Video)(nil)).Where("id IN (?)", bun.In(delVidList)).Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

// キーワード告知
// func (j *Job) KeywordAnnounceJob() error {
// 	ctx := context.Background()
// 	now, _ := time.Parse(time.RFC3339, time.Now().UTC().Format("2006-01-02T15:04:00Z"))
// 	tAfter := now.Add(-20 * time.Minute)
// 	tBefore := now.Add(-10 * time.Minute)
// 	var videos []Video
// 	err := j.db.Service.NewSelect().Model(&videos).Where("? BETWEEN ? AND ?", bun.Ident("created_at"), tAfter, tBefore).Scan(ctx)
// 	if err != nil {
// 		return err
// 	}
// 	if len(videos) == 0 {
// 		return nil
// 	}

// 	// FCMトークンを取得
// 	keywordTextList, err := j.db.getKeywordTextList()
// 	if err != nil {
// 		slog.Error("getKeywordTextList",
// 			slog.String("severity", "ERROR"),
// 			slog.String("message", err.Error()),
// 		)
// 		return err
// 	}
// 	for _, keywordText := range keywordTextList {
// 		reg := ".*" + keywordText + ".*"
// 		for _, v := range videos {
// 			// キーワードに一致した場合
// 			if regexp.MustCompile(reg).Match([]byte(v.Title)) {
// 				err := j.fcm.TopicNotification(keywordText, &NotificationVideo{
// 					ID:        v.ID,
// 					Title:     v.Title,
// 					Thumbnail: v.Thumbnail,
// 				})
// 				if err != nil {
// 					slog.Error("TopicNotification",
// 						slog.String("severity", "ERROR"),
// 						slog.String("message", err.Error()),
// 					)
// 				}
// 				return err
// 			}
// 		}
// 	}

// 	return nil
// }
