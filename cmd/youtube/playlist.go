package youtube

import (
	"context"
	"strings"

	"github.com/rs/zerolog/log"
)

type YouTubePlaylistsResponse struct {
	ID        string `json:"id"`
	ItemCount int64  `json:"item_count"`
}

// プレイリストIDとキー、プレイリストに含まれている動画数を値とした連想配列を返す
func (yt *Youtube) Playlists() (map[string]int64, error) {
	list := make(map[string]int64, 500)
	ctx := context.Background()
	plist, err := yt.queries.ListPlaylistID(ctx)
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
		call := yt.service.Playlists.List([]string{"snippet", "contentDetails"}).MaxResults(50).Id(id)
		res, err := call.Do()
		if err != nil {
			log.Error().Str("severity", "ERROR").Err(err).Msg("playlists-list call error")
			return nil, err
		}

		for _, item := range res.Items {
			list[item.Id] = item.ContentDetails.ItemCount
		}
	}
	return list, nil
}

// DBに保存されているプレイリストの動画数とAPIから取得したプレイリストの動画数を比較し、動画数が変わっているプレイリストIDを返す
func (yt *Youtube) CheckItemCount(list map[string]int64) ([]string, error) {
	// 動画数が変わっているプレイリストIDを入れる配列
	var clist []string
	ctx := context.Background()
	itemCountList, err := yt.queries.ListItemCount(ctx)
	if err != nil {
		return nil, err
	}

	for _, row := range itemCountList {
		value, isThere := list[row.ID]
		if isThere && value != int64(row.ItemCount) {
			clist = append(clist, row.ID)
		}
	}

	return clist, nil
}
