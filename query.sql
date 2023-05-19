-- name: ListPlaylistID :many
SELECT replace(id, 'UC', 'UU') AS id FROM vtubers;

-- name: ListItemCount :many
SELECT replace(id, 'UC', 'UU') AS id, item_count FROM vtubers;