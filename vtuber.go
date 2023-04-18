package main

import "github.com/rs/zerolog/log"

// プレイリストIDをキーとして動画の数を値とするマップ
type ItemCountResponse map[string]int64

// チャンネルのアップロードされた動画を含むプレイリストのIDを取得する
func GetPlaylistsID() ([]string, error) {
	var pid string
	var plist[]string
	rows, err := DB.Query("select playlist_id from vtubers")
	if err != nil {
		log.Error().Str("severity", "ERROR").Str("service", "get-playlists-id").Err(err).Msg("select playlist_id failed")
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&pid)
		if err != nil {
			return nil, err
		}
		plist = append(plist, pid)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return plist, nil
}

// DBに保存されているプレイリストIDとプレイリストに含まれている動画の数を取得する
func GetItemCount() (ItemCountResponse, error) {
	var (
		pid string
		count int64
	)
	itemCount := ItemCountResponse{}
	rows, err := DB.Query("select playlist_id, item_count from vtubers")
	if err != nil {
		log.Error().Str("severity", "ERROR").Str("service", "get-item-count").Err(err).Msg("select playlist_id failed")
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&pid, &count)
		if err != nil {
			return nil, err
		}
		itemCount[pid] = count
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return itemCount, nil
}