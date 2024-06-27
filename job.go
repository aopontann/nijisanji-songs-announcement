package nsa

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"regexp"
	"slices"
	"strings"
	"time"

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
	// DBに登録されているプレイリストの動画数を取得
	oldPlaylists, err := j.db.Playlists()
	if err != nil {
		return err
	}

	// プレイリストIDリスト
	var plist []string
	for pid := range oldPlaylists {
		plist = append(plist, pid)
	}

	// Youtube Data API から最新のプレイリストの動画数を取得
	newPlaylists, err := j.yt.Playlists(plist)
	if err != nil {
		return err
	}

	// 動画数が変化しているプレイリストIDを取得
	var changedPlaylistID []string
	for pid, itemCount := range oldPlaylists {
		if itemCount != newPlaylists[pid] {
			changedPlaylistID = append(changedPlaylistID, pid)
		}
	}

	// 新しくアップロードされた動画IDを取得
	vidList, err := j.yt.PlaylistItems(changedPlaylistID)
	if err != nil {
		return err
	}

	// 動画情報を取得
	videos, err := j.yt.Videos(vidList)
	if err != nil {
		return err
	}

	// フィルター
	// filtedVideos := j.yt.FilterVideos(videos)

	// トランザクション開始
	ctx := context.Background()
	tx, err := j.db.Service.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}

	// DBのプレイリスト動画数を更新
	err = j.db.UpdatePlaylistItem(tx, newPlaylists)
	if err != nil {
		return err
	}

	// 動画情報をDBに登録
	// 登録済みの動画は無視
	err = j.db.SaveVideos(tx, videos)
	if err != nil {
		return err
	}

	// コミット
	err = tx.Commit()
	if err != nil {
		return err
	}

	// 歌みた動画か判別しづらい動画をメールに送信する
	for _, v := range videos {
		for _, word := range getSongWordList() {
			if strings.Contains(strings.ToLower(v.Snippet.Title), strings.ToLower(word)) {
				continue
			}
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
		err := SendMail("歌みた動画判定", v.Id)
		if err != nil {
			return err
		}
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
		err = SendMail("検証 5分後に公開", fmt.Sprintf("https://www.youtube.com/watch?v=%s", v.ID))
		if err != nil {
			return err
		}

		// push通知
		err := j.fcm.SongNotification(v, tokens)
		if err != nil {
			return err
		}

		// ツイート
		// err = tw.Id(v.ID).Title(v.Title).Tweet()
		// if err != nil {
		// 	http.Error(w, err.Error(), http.StatusInternalServerError)
		// 	return
		// }

		// Missky post
		// err = mk.Post(v.ID, v.Title)
		// if err != nil {
		// 	http.Error(w, err.Error(), http.StatusInternalServerError)
		// 	return
		// }
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
func (j *Job) KeywordAnnounceJob() error {
	ctx := context.Background()
	now, _ := time.Parse(time.RFC3339, time.Now().UTC().Format("2006-01-02T15:04:00Z"))
	tAfter := now.Add(-20 * time.Minute)
	tBefore := now.Add(-10 * time.Minute)
	var videos []Video
	err := j.db.Service.NewSelect().Model(&videos).Where("? BETWEEN ? AND ?", bun.Ident("created_at"), tAfter, tBefore).Scan(ctx)
	if err != nil {
		return err
	}
	if len(videos) == 0 {
		return nil
	}

	// FCMトークンを取得
	keywordTextList, err := j.db.getKeywordTextList()
	if err != nil {
		slog.Error("getKeywordTextList",
			slog.String("severity", "ERROR"),
			slog.String("message", err.Error()),
		)
		return err
	}
	for _, keywordText := range keywordTextList {
		reg := ".*" + keywordText + ".*"
		for _, v := range videos {
			// キーワードに一致した場合
			if regexp.MustCompile(reg).Match([]byte(v.Title)) {
				err := j.fcm.KeywordNotification(v, keywordText)
				if err != nil {
					slog.Error("KeywordNotification",
						slog.String("severity", "ERROR"),
						slog.String("message", err.Error()),
					)
				}
				return err
			}
		}
	}

	return nil
}

func (j *Job) UpdatePlaylistItemJob() error {
	// DBに登録されているプレイリストの動画数を取得
	oldPlaylists, err := j.db.Playlists()
	if err != nil {
		return err
	}

	// プレイリストIDリスト
	var plist []string
	for pid := range oldPlaylists {
		plist = append(plist, pid)
	}

	// Youtube Data API から最新のプレイリストの動画数を取得
	newPlaylists, err := j.yt.Playlists(plist)
	if err != nil {
		return err
	}

	// トランザクション開始
	ctx := context.Background()
	tx, err := j.db.Service.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}

	// DBのプレイリスト動画数を更新
	err = j.db.UpdatePlaylistItem(tx, newPlaylists)
	if err != nil {
		return err
	}

	// コミット
	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}
