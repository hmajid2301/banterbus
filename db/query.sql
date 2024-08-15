-- name: AddRoom :one
INSERT INTO rooms (id, game_name, host_player, room_code) VALUES (?, ?, ?, ?) RETURNING *;

-- name: AddPlayer :one
INSERT INTO players (id, avatar, nickname) VALUES (?, ?, ?) RETURNING *;

-- name: AddRoomPlayer :one
INSERT INTO rooms_players (room_id, player_id) VALUES (?, ?) RETURNING *;

-- name: GetAllPlayersInRoom :many
SELECT p.id, p.created_at, p.updated_at, p.avatar, p.nickname, r.room_code
FROM players p
JOIN rooms_players rp ON p.id = rp.player_id
JOIN rooms r ON rp.room_id = r.id
WHERE rp.room_id = (
    SELECT rp_inner.room_id
    FROM rooms_players rp_inner
    WHERE rp_inner.player_id = ?
);

-- name: GetRoomByCode :one
SELECT id FROM rooms WHERE room_code = ?;

-- name: UpdateNickname :one
UPDATE players SET nickname = ? WHERE id = ? RETURNING *;

-- name: UpdateAvatar :one
UPDATE players SET avatar = ? WHERE id = ? RETURNING *;
