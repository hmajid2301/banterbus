-- name: AddRoom :one
INSERT INTO rooms (game_name, host_player, room_code) VALUES (?, ?, ?) RETURNING *;

-- name: AddPlayer :one
INSERT INTO players (avatar, nickname, latest_session_id) VALUES (?, ?, ?) RETURNING *;

-- name: AddRoomPlayer :one
INSERT INTO rooms_players (room_id, player_id) VALUES (?, ?) RETURNING *;
