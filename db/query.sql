-- name: AddRoom :one
INSERT INTO rooms (id, game_name, host_player, room_code) VALUES (?, ?, ?, ?) RETURNING *;

-- name: AddPlayer :one
INSERT INTO players (id, avatar, nickname) VALUES (?, ?, ?) RETURNING *;

-- name: AddRoomPlayer :one
INSERT INTO rooms_players (room_id, player_id) VALUES (?, ?) RETURNING *;

-- name: GetAllPlayersInRoom :many
SELECT p.*, r.room_code
FROM players p
JOIN rooms_players rp ON p.id = rp.player_id
JOIN rooms r ON rp.room_id = r.id
WHERE rp.room_id IN (
    SELECT room_id
    FROM rooms_players
    WHERE rp.player_id = ?
);

-- name: UpdateNickname :one
UPDATE players SET nickname = ? WHERE id = ? RETURNING *;
