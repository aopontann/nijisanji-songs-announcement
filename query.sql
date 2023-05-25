-- name: ListPlaylistID :many
SELECT replace(id, 'UC', 'UU') AS id FROM vtubers;

-- name: ListItemCount :many
SELECT replace(id, 'UC', 'UU') AS id, item_count FROM vtubers;

-- name: ListVtuberID :many
SELECT id FROM vtubers;

-- name: ExistsVideos :many
SELECT id FROM videos WHERE id IN (sqlc.slice('ids'));

-- name: CreateVideo :exec
INSERT IGNORE INTO videos(id, title, songConfirm, scheduled_start_time) VALUES(?,?,?,?);

-- name: UpdatePlaylistItemCount :exec
UPDATE vtubers SET item_count = ? WHERE id = ? AND item_count != ?;

-- name: ListVideoIdTitle :many
SELECT id, title FROM videos WHERE songConfirm = 1 AND scheduled_start_time BETWEEN ? AND ?;