-- name: AddRoom :one
INSERT INTO rooms (id, game_name, host_player, room_code, room_state)  VALUES (?, ?, ?, ?, ?) RETURNING *;

-- name: AddPlayer :one
INSERT INTO players (id, avatar, nickname) VALUES (?, ?, ?) RETURNING *;

-- name: AddRoomPlayer :one
INSERT INTO rooms_players (room_id, player_id) VALUES (?, ?) RETURNING *;

-- name: GetAllPlayersInRoom :many
SELECT p.id, p.created_at, p.updated_at, p.avatar, p.nickname, p.is_ready, r.room_code
FROM players p
JOIN rooms_players rp ON p.id = rp.player_id
JOIN rooms r ON rp.room_id = r.id
WHERE rp.room_id = (
    SELECT rp_inner.room_id
    FROM rooms_players rp_inner
    WHERE rp_inner.player_id = ?
);

-- name: GetRoomByPlayerID :one
SELECT r.* FROM rooms r JOIN rooms_players rp ON r.id = rp.room_id WHERE rp.player_id = ?;

-- name: GetRoomByCode :one
SELECT * FROM rooms WHERE room_code = ?;

-- name: GetPlayerByID :one
SELECT * FROM players WHERE id = ?;

-- name: UpdateNickname :one
UPDATE players SET nickname = ? WHERE id = ? RETURNING *;

-- name: UpdateAvatar :one
UPDATE players SET avatar = ? WHERE id = ? RETURNING *;

-- name: UpdateIsReady :one
UPDATE players SET is_ready = ? WHERE id = ? RETURNING *;
