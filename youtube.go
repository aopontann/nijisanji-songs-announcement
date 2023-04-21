package main

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

type YouTubeCheckResponse struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Schedule  string `json:"schedule"`
	TwitterID string `json:"twitter_id"`
}

type YouTubeChannelsResponse struct {
	ID         string `json:"id"`
	PlaylistID string `json:"playlist_id"`
	VideoCount uint64 `json:"video_count"`
}

type YouTubePlaylistsResponse struct {
	ID         string `json:"id"`
	PlaylistID string `json:"playlist_id"`
	ItemCount  int64  `json:"item_count"`
}

type NewUploadChannelList struct {
	ID            string `json:"id"`
	NewVideoCount uint64 `json:"new_video_count"`
	OldVideoCount uint64 `json:"old_video_count"`
	PlaylistID    string `json:"playlist_id"`
}

type VideoIDList []string
type YTVRList []youtube.Video

// チャンネルIDと最新の動画数を格納する配列の型
type YTCRList []YouTubeChannelsResponse

// チャンネルIDとプレイリストID、プレイリストに含まれる動画数を格納する配列の型
type YTPRList []YouTubePlaylistsResponse

// 新しく動画をアップロードしたチャンネルIDと古い動画と新しい動画の数を格納する配列の型
type NewUpChList []NewUploadChannelList

var YouTubeService *youtube.Service

// Youtube Data API の 初期化処理
func YTNew() {
	var err error
	ctx := context.Background()
	YouTubeService, err = youtube.NewService(ctx, option.WithAPIKey(os.Getenv("YOUTUBE_API_KEY")))
	if err != nil {
		log.Fatal().Err(err).Msg("youtube.NewService create failed")
	}
}

// Youtube Data API Search List を叩いて検索結果の動画IDを取得
func Search() (VideoIDList, error) {
	// 動画検索範囲
	dtAfter := time.Now().UTC().Add(-120 * time.Minute).Format("2006-01-02T15:04:00Z")
	dtBefore := time.Now().UTC().Add(-60 * time.Minute).Format("2006-01-02T15:04:00Z")
	// 動画IDを格納する文字列型配列を宣言
	vid := make([]string, 0, 600)

	for _, q := range []string{"にじさんじ", "NIJISANJI"} {
		pt := ""
		for {
			// youtube data api search list にリクエストを送る
			call := YouTubeService.Search.List([]string{"id"}).MaxResults(50).Q(q).PublishedAfter(dtAfter).PublishedBefore(dtBefore).PageToken(pt)
			res, err := call.Do()
			if err != nil {
				log.Error().Str("severity", "ERROR").Err(err).Msg("search-list call error")
				return []string{}, err
			}

			for _, item := range res.Items {
				vid = append(vid, item.Id.VideoId)
			}

			log.Info().
				Str("severity", "INFO").
				Str("service", "youtube-search-list").
				Str("published", fmt.Sprintf("%s ~ %s", dtAfter, dtBefore)).
				Str("q", q).
				Str("pageInfo", fmt.Sprintf("perPage=%d total=%d nextPage=%s\n", res.PageInfo.ResultsPerPage, res.PageInfo.TotalResults, res.NextPageToken)).
				Strs("id", vid).
				Send()

			if res.NextPageToken == "" {
				break
			}
			pt = res.NextPageToken
		}
	}
	return vid, nil
}

// Youtube Data API から動画情報を取得
func (list VideoIDList) Video() (YTVRList, error) {
	var rlist []youtube.Video
	for i := 0; i*50 <= len(list); i++ {
		var id string
		if len(list) > 50*(i+1) {
			id = strings.Join(list[50*i:50*(i+1)], ",")
		} else {
			id = strings.Join(list[50*i:], ",")
		}
		call := YouTubeService.Videos.List([]string{"snippet", "contentDetails", "liveStreamingDetails"}).Id(id).MaxResults(50)
		res, err := call.Do()
		if err != nil {
			log.Error().Str("severity", "ERROR").Err(err).Msg("videos-list call error")
			return nil, err
		}

		for _, video := range res.Items {
			scheduledStartTime := "" // 例 2022-03-28T11:00:00Z
			if video.LiveStreamingDetails != nil {
				// "2022-03-28 11:00:00"形式に変換
				rep1 := strings.Replace(video.LiveStreamingDetails.ScheduledStartTime, "T", " ", 1)
				scheduledStartTime = strings.Replace(rep1, "Z", "", 1)
			}

			rlist = append(rlist, *video)

			log.Info().
				Str("severity", "INFO").
				Str("service", "youtube-video-list").
				Str("id", video.Id).
				Str("title", video.Snippet.Title).
				Str("duration", video.ContentDetails.Duration).
				Str("schedule", scheduledStartTime).
				Str("channel_id", video.Snippet.ChannelId).
				Send()
		}
	}
	return rlist, nil
}

// 歌ってみた動画かフィルターをかける処理
func (list YTVRList) Select() (YTVRList, error) {
	var rlist []youtube.Video
	// にじさんじライバーのチャンネルリストを取得
	channelIdList, err := GetChannelIdList()
	if err != nil {
		return nil, err
	}
	// 歌動画か判断する
	for _, video := range rlist {
		// プレミア公開する動画か
		if video.LiveStreamingDetails == nil {
			continue
		}
		// 放送が終了している場合
		if video.Snippet.LiveBroadcastContent == "none" {
			continue
		}
		// 動画の長さが9分59秒以下ではない場合
		if !regexp.MustCompile(`^PT([1-9]M[1-5]?[0-9]S|[1-5]?[0-9]S)`).Match([]byte(video.ContentDetails.Duration)) {
			continue
		}
		// 切り抜き動画である場合
		if regexp.MustCompile(`.*切り抜き.*`).Match([]byte(video.Snippet.Title)) {
			continue
		}
		// 動画タイトルに特定の文字が含まれているか
		// checkRes := TitleCheck(video.Title)
		// checkRes := true

		// にじさんじライバーのチャンネルで公開されたか
		if !NijisanjiCheck(channelIdList, video.Snippet.ChannelId) {
			continue
		}

		rlist = append(rlist, video)

		log.Info().
			Str("severity", "INFO").
			Str("service", "youtube-video-select").
			Str("id", video.Id).
			Str("title", video.Snippet.Title).
			Str("duration", video.ContentDetails.Duration).
			Str("schedule", video.LiveStreamingDetails.ScheduledStartTime).
			Send()
	}
	return rlist, nil
}

// 動画情報をDBに保存
func (list YTVRList) Save() error {
	tx, err := DB.Begin()
	if err != nil {
		log.Error().Str("severity", "ERROR").Err(err).Msg("DB.Begin error")
		return err
	}
	// DB準備
	stmt, err := tx.Prepare("INSERT IGNORE INTO videos(id, title, songConfirm, scheduled_start_time) VALUES(?,?,?,?)")
	if err != nil {
		log.Error().Str("severity", "ERROR").Err(err).Msg("DB.Prepare error")
		return err
	}

	for _, video := range list {
		// "2022-03-28 11:00:00"形式に変換
		rep1 := strings.Replace(video.LiveStreamingDetails.ScheduledStartTime, "T", " ", 1)
		scheduledStartTime := strings.Replace(rep1, "Z", "", 1)
		// DBに動画情報を保存
		_, err := stmt.Exec(video.Id, video.Snippet.Title, true, scheduledStartTime)
		if err != nil {
			log.Error().Str("severity", "ERROR").Err(err).Msg("Save videos failed")
			if tx.Rollback() != nil {
				log.Error().Str("severity", "ERROR").Err(err).Msg("Rollback error")
			}
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Error().Str("severity", "ERROR").Err(err).Msg("tx.Commit error")
		return err
	}
	return nil
}

// 動画IDからAPIを叩いて動画情報を取得し、DBに保存されている動画情報と異なるデータがある場合、新しい情報に上書きする
// 上書きした場合やデータを削除した場合は true を返す
func CheckVideo(vid string) (bool, error) {
	var title string
	var scheTime string
	// エラー表示に使うカスタムログの設定
	clog := log.Error().Str("severity", "ERROR").Str("service", "youtube-check-video")
	// Youtube Data API にリクエスト送信して動画情報を取得
	call := YouTubeService.Videos.List([]string{"snippet", "contentDetails", "liveStreamingDetails"}).Id(vid).MaxResults(1)
	res, err := call.Do()
	if err != nil {
		clog.Err(err).Msg("videos-list call error")
		return false, err
	}
	// 動画が削除されていた場合、DBに保存されている動画情報も削除する
	if len(res.Items) == 0 {
		_, err := DB.Exec("DELETE FROM videos WHERE id = ?", vid)
		if err != nil {
			clog.Err(err).Msg("delete video error")
			return false, err
		}
		return true, nil
	}
	// DBから動画情報を取得
	err = DB.QueryRow("SELECT title, scheduled_start_time FROM videos WHERE id = ?", vid).Scan(&title, &scheTime)
	if err != nil {
		clog.Err(err).Msg("select video error")
		return false, err
	}
	// DBに保存されている動画情報がAPIから取得したデータと異なる場合
	if title != res.Items[0].Snippet.Title || scheTime != res.Items[0].LiveStreamingDetails.ScheduledStartTime {
		// 新しい動画情報に上書きする
		_, err := DB.Exec("UPDATE videos SET title = ?, scheduled_start_time = ? WHERE id = ?", title, scheTime, vid)
		if err != nil {
			clog.Err(err).Msg("Failed to overwrite")
			return false, err
		}
		return true, nil
	}
	// DBに保存されている動画情報がAPIから取得したデータと同じ場合
	return false, nil
}

// チャンネルIDとプレイリストIDとチャンネルにアップロードされた動画の数を取得
func Channels() (YTCRList, error) {
	var cResList []YouTubeChannelsResponse
	// にじさんじライバーのチャンネルリストを取得
	cList, err := GetChannelIdList()
	if err != nil {
		return nil, err
	}

	for i := 0; i*50 <= len(cList); i++ {
		var id string
		if len(cList) > 50*(i+1) {
			id = strings.Join(cList[50*i:50*(i+1)], ",")
		} else {
			id = strings.Join(cList[50*i:], ",")
		}
		call := YouTubeService.Channels.List([]string{"statistics", "contentDetails"}).MaxResults(50).Id(id)
		res, err := call.Do()
		if err != nil {
			log.Error().Str("severity", "ERROR").Err(err).Msg("channels-list call error")
			return []YouTubeChannelsResponse{}, err
		}

		for _, item := range res.Items {
			cResList = append(cResList, YouTubeChannelsResponse{ID: item.Id, VideoCount: item.Statistics.VideoCount, PlaylistID: item.ContentDetails.RelatedPlaylists.Uploads})
		}
	}
	return cResList, nil
}

// 新しく取得したチャンネルの動画数をDBに保存
func (clist YTCRList) Save() error {
	tx, err := DB.Begin()
	if err != nil {
		log.Error().Str("severity", "ERROR").Err(err).Msg("DB.Begin error")
		return err
	}
	// 動画が削除されて動画数が減っていても、上書きする
	stmt, err := tx.Prepare("UPDATE vtubers SET video_count = ? WHERE id = ? AND video_count != ?")
	if err != nil {
		log.Error().Str("severity", "ERROR").Err(err).Msg("DB.Prepare error")
		return err
	}

	for _, list := range clist {
		_, err := stmt.Exec(list.VideoCount, list.ID, list.VideoCount)
		if err != nil {
			log.Error().Str("severity", "ERROR").Err(err).Msg("Save video_count failed")
			if tx.Rollback() != nil {
				log.Error().Str("severity", "ERROR").Err(err).Msg("Rollback error")
			}
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Error().Str("severity", "ERROR").Err(err).Msg("tx.Commit error")
		return err
	}

	return nil
}

// 新しく動画をアップロードしたチャンネルIDと古い動画数と新しい動画の数を返す
func (newlist YTCRList) CheckUpload() (NewUpChList, error) {
	// 新しく動画をアップロードしたチャンネルIDを格納する変数
	var diffCList NewUpChList
	// DBに保存されているチャンネルIDとそのチャンネルの動画本数の一覧を取得 map[string] uint64
	vvcList, err := GetVideoCountList()
	if err != nil {
		return nil, err
	}

	for _, newCh := range newlist {
		// DBに保存されている動画の本数より、新しく取得した動画の本数が多い場合　動画が削除されて数が減っている場合は返さない
		if vvcList[newCh.ID] < newCh.VideoCount {
			log.Info().
				Str("severity", "INFO").
				Str("service", "youtube-check-upload").
				Str("ChannelId", newCh.ID).
				Uint64("NewVideoCount", newCh.VideoCount).
				Uint64("OldVideoCount", vvcList[newCh.ID]).
				Send()
			diffCList = append(diffCList, NewUploadChannelList{ID: newCh.ID, NewVideoCount: newCh.VideoCount, OldVideoCount: vvcList[newCh.ID], PlaylistID: newCh.PlaylistID})
		}
	}
	return diffCList, nil
}

// 新しくアップロードされた動画のIDを取得
func (list NewUpChList) GetNewVideoId() (VideoIDList, error) {
	// 動画IDを格納する文字列型配列を宣言
	vid := make([]string, 0, 600)

	for _, ch := range list {
		// 取得した動画IDをログに出力するための変数
		var rid []string
		call := YouTubeService.PlaylistItems.List([]string{"snippet", "contentDetails"}).PlaylistId(ch.PlaylistID).MaxResults(int64(ch.NewVideoCount - ch.OldVideoCount))
		res, err := call.Do()
		if err != nil {
			log.Error().Str("severity", "ERROR").Err(err).Msg("playlistitems-list call error")
			return []string{}, err
		}

		for _, item := range res.Items {
			rid = append(rid, item.ContentDetails.VideoId)
			vid = append(vid, item.ContentDetails.VideoId)
		}

		log.Info().
			Str("severity", "INFO").
			Str("service", "youtube-playlistitems-list").
			Str("ChannelId", ch.ID).
			Str("PlaylistId", ch.PlaylistID).
			Strs("videoId", rid).
			Send()
	}
	return vid, nil
}

// チャンネルのアップロードされた動画を含むプレイリストに含まれている動画の数を取得
func Playlists() (YTPRList, error) {
	var pResList []YouTubePlaylistsResponse
	// にじさんじライバーのチャンネルリストを取得
	plist, err := GetPlaylistsID()
	if err != nil {
		return nil, err
	}

	for i := 0; i*50 <= len(plist); i++ {
		var id string
		if len(plist) > 50*(i+1) {
			id = strings.Join(plist[50*i:50*(i+1)], ",")
		} else {
			id = strings.Join(plist[50*i:], ",")
		}
		call := YouTubeService.Playlists.List([]string{"snippet", "contentDetails"}).MaxResults(50).Id(id)
		res, err := call.Do()
		if err != nil {
			log.Error().Str("severity", "ERROR").Err(err).Msg("playlists-list call error")
			return []YouTubePlaylistsResponse{}, err
		}

		for _, item := range res.Items {
			pResList = append(pResList, YouTubePlaylistsResponse{ID: item.Snippet.ChannelId, PlaylistID: item.Id, ItemCount: item.ContentDetails.ItemCount})
		}
	}
	return pResList, nil
}

// DBに保存されているプレイリストの動画数とAPIから取得したプレイリストの動画数を比較し、動画数が変わっているプレイリストを返す
func (plist YTPRList) Select() (YTPRList, error) {
	var selectedList YTPRList
	dblist, err := GetItemCount()
	if err != nil {
		log.Error().Str("severity", "ERROR").Err(err).Msg("GetItemCount() error")
		return nil, err
	}

	for _, list := range plist {
		if dblist[list.PlaylistID] != list.ItemCount {
			selectedList = append(selectedList, list)
		}
	}
	return selectedList, nil
}

// プレイリストに含まれている動画の数を更新する
func (plist YTPRList) Save() error {
	tx, err := DB.Begin()
	if err != nil {
		log.Error().Str("severity", "ERROR").Err(err).Msg("DB.Begin error")
		return err
	}
	// 動画が削除されて動画数が減っていても、上書きする
	stmt, err := tx.Prepare("UPDATE vtubers SET item_count = ? WHERE id = ? AND item_count != ?")
	if err != nil {
		log.Error().Str("severity", "ERROR").Err(err).Msg("DB.Prepare error")
		return err
	}

	for _, list := range plist {
		_, err := stmt.Exec(list.ItemCount, list.ID, list.ItemCount)
		if err != nil {
			log.Error().Str("severity", "ERROR").Err(err).Msg("Save item_count failed")
			if tx.Rollback() != nil {
				log.Error().Str("severity", "ERROR").Err(err).Msg("Rollback error")
			}
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Error().Str("severity", "ERROR").Err(err).Msg("tx.Commit error")
		return err
	}

	return nil
}

// プレイリストに含まれている動画IDを取得する
func (plist YTPRList) Items() (VideoIDList, error) {
	// 動画IDを格納する文字列型配列を宣言
	vid := make([]string, 0, 600)

	for _, list := range plist {
		// 取得した動画IDをログに出力するための変数
		var rid []string
		call := YouTubeService.PlaylistItems.List([]string{"snippet"}).PlaylistId(list.PlaylistID).MaxResults(3)
		res, err := call.Do()
		if err != nil {
			log.Error().Str("severity", "ERROR").Err(err).Msg("playlistitems-list call error")
			return []string{}, err
		}

		for _, item := range res.Items {
			rid = append(rid, item.Snippet.ResourceId.VideoId)
			vid = append(vid, item.Snippet.ResourceId.VideoId)
		}

		log.Info().
			Str("severity", "INFO").
			Str("service", "youtube-playlistitems-list").
			Str("ChannelId", list.ID).
			Str("PlaylistId", list.PlaylistID).
			Strs("videoId", rid).
			Send()
	}
	return vid, nil
}
