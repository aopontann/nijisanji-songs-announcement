package main

import (
	"strings"

	"github.com/rs/zerolog/log"
)

// プレイリストIDをキーとして動画の数を値とするマップ
type ItemCountResponse map[string]int64

// DBに保存されている全てのライバーのプレイリストのIDを取得する
// プレイリストにはアップロードされた動画全てが含まれている
func GetPlaylistsID() ([]string, error) {
	var id string
	var plist[]string
	rows, err := DB.Query("select id from vtubers")
	if err != nil {
		log.Error().Str("severity", "ERROR").Str("service", "get-playlists-id").Err(err).Msg("select playlist_id failed")
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&id)
		if err != nil {
			return nil, err
		}
		rep := strings.Replace(id, "UC", "UU", 1)
		plist = append(plist, rep)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return plist, nil
}

// DBに保存されているチャンネルIDとプレイリストに含まれている動画の数を取得する
func GetItemCount() (ItemCountResponse, error) {
	var (
		id string
		count int64
	)
	itemCount := ItemCountResponse{}
	rows, err := DB.Query("select id, item_count from vtubers")
	if err != nil {
		log.Error().Str("severity", "ERROR").Str("service", "get-item-count").Err(err).Msg("select id, item_count failed")
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&id, &count)
		if err != nil {
			return nil, err
		}
		itemCount[id] = count
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return itemCount, nil
}

// にじさんじライバーのチャンネルID一覧を取得する
func GetChannelIdList() ([]string, error) {
	var (
		channelId     string
		channelIdList []string
	)
	rows, err := DB.Query("select id from vtubers")
	if err != nil {
		log.Error().Str("severity", "ERROR").Err(err).Msg("select vtuber failed")
		return channelIdList, err
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&channelId)
		if err != nil {
			return channelIdList, err
		}
		channelIdList = append(channelIdList, channelId)
	}
	err = rows.Err()
	if err != nil {
		return channelIdList, err
	}
	return channelIdList, nil
}
