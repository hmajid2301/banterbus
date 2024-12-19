-- name: AddRoom :one
INSERT INTO rooms (id, game_name, host_player, room_code, room_state) VALUES ($1, $2, $3, $4, $5) RETURNING *;

-- name: AddPlayer :one
INSERT INTO players (id, avatar, nickname) VALUES ($1, $2, $3) RETURNING *;

-- name: AddRoomPlayer :one
INSERT INTO rooms_players (room_id, player_id) VALUES ($1, $2) RETURNING *;

-- name: RemovePlayerFromRoom :one
DELETE FROM rooms_players WHERE player_id = $1 RETURNING *;

-- name: UpdateRoomState :one
UPDATE rooms SET room_state = $1 WHERE id = $2 RETURNING *;

-- name: GetPlayerByID :one
SELECT * FROM players WHERE id = $1;

-- name: UpdateNickname :one
UPDATE players SET nickname = $1 WHERE id = $2 RETURNING *;

-- name: UpdateAvatar :one
UPDATE players SET avatar = $1 WHERE id = $2 RETURNING *;

-- name: TogglePlayerIsReady :one
UPDATE players SET is_ready = NOT is_ready WHERE id = $1 RETURNING *;

-- name: AddFibbingItRound :one
INSERT INTO fibbing_it_rounds (id, round_type, round, fibber_question_id, normal_question_id, game_state_id) VALUES ($1, $2, $3, $4, $5, $6) RETURNING *;

-- name: AddGameState :one
INSERT INTO game_state (id, room_id, submit_deadline, state) VALUES ($1, $2, $3, $4) RETURNING *;

-- name: UpdateGameState :one
UPDATE game_state SET state = $1, submit_deadline = $2 WHERE id = $3 RETURNING *;

-- name: AddFibbingItAnswer :one
INSERT INTO fibbing_it_answers (id, answer, round_id, player_id) VALUES ($1, $2, $3, $4) RETURNING *;

-- name: AddFibbingItRole :one
INSERT INTO fibbing_it_player_roles (id, player_role, round_id, player_id) VALUES ($1, $2, $3, $4) RETURNING *;

-- name: UpsertFibbingItVote :exec
INSERT INTO fibbing_it_votes (id, player_id, voted_for_player_id, round_id)
VALUES ($1, $2, $3, $4)
ON CONFLICT(player_id, round_id) DO UPDATE SET
    updated_at = EXCLUDED.updated_at,
    player_id = EXCLUDED.player_id,
    voted_for_player_id = EXCLUDED.voted_for_player_id,
    round_id = EXCLUDED.round_id
RETURNING *;

-- name: AddQuestion :one
INSERT INTO questions (id, game_name, round, question, language_code, group_id) VALUES ($1, $2, $3, $4, $5, $6) RETURNING *;

-- name: AddQuestionsGroup :one
INSERT INTO questions_groups (id, group_name, group_type) VALUES ($1, $2, $3) RETURNING *;

-- name: GetAllPlayersInRoom :many
SELECT p.id, p.created_at, p.updated_at, p.avatar, p.nickname, p.is_ready, r.room_code, r.host_player
FROM players p
JOIN rooms_players rp ON p.id = rp.player_id
JOIN rooms r ON rp.room_id = r.id
WHERE rp.room_id = (
    SELECT rp_inner.room_id
    FROM rooms_players rp_inner
    WHERE rp_inner.player_id = $1
);

-- name: GetAllPlayersByGameStateID :many
SELECT p.id, p.nickname, p.avatar
FROM players p
JOIN rooms_players rp ON p.id = rp.player_id
JOIN game_state gs ON rp.room_id = gs.room_id
WHERE gs.id = $1;

-- name: GetAllPlayerByRoomCode :many
SELECT p.id, p.created_at, p.updated_at, p.avatar, p.nickname, p.is_ready, r.room_code, r.host_player
FROM players p
JOIN rooms_players rp ON p.id = rp.player_id
JOIN rooms r ON rp.room_id = r.id
WHERE rp.room_id = (
    SELECT r_inner.id
    FROM rooms r_inner
    WHERE r_inner.room_code = $1 AND (r_inner.room_state = 'CREATED' OR r_inner.room_state = 'PLAYING')
)
ORDER BY p.created_at;

-- name: GetGameStateByPlayerID :one
SELECT
    gs.id,
    gs.created_at,
    gs.updated_at,
    gs.room_id,
    gs.submit_deadline,
    gs.state
FROM game_state gs
JOIN rooms_players rp ON gs.room_id = rp.room_id
WHERE rp.player_id = $1;

-- name: GetGameState :one
SELECT
    gs.id,
    gs.created_at,
    gs.updated_at,
    gs.room_id,
    gs.submit_deadline,
    gs.state
FROM game_state gs
WHERE gs.id = $1;

-- name: GetRoomByPlayerID :one
SELECT r.* FROM rooms r JOIN rooms_players rp ON r.id = rp.room_id WHERE rp.player_id = $1;

-- name: GetRoomByCode :one
SELECT * FROM rooms WHERE room_code = $1;

-- name: GetRandomQuestionByRound :one
SELECT * FROM questions WHERE game_name = $1 AND round = $2 AND language_code = $3 AND enabled = TRUE ORDER BY RANDOM() LIMIT 1;

-- name: GetRandomQuestionInGroup :one
SELECT *
FROM questions q
JOIN questions_groups qg ON q.group_id = qg.id
WHERE qg.group_type = 'questions'
  AND q.group_id = $1
  AND q.enabled = TRUE
  AND q.id != $2
ORDER BY RANDOM()
LIMIT 1;

-- name: GetLatestRoundByPlayerID :one
SELECT fir.*, gs.submit_deadline
FROM fibbing_it_rounds fir
JOIN game_state gs ON fir.game_state_id = gs.id
JOIN rooms_players rp ON gs.room_id = rp.room_id
WHERE rp.player_id = $1
ORDER BY fir.created_at DESC
LIMIT 1;

-- name: GetLatestRoundByGameStateID :one
SELECT fir.*, gs.submit_deadline
FROM fibbing_it_rounds fir
JOIN game_state gs ON fir.game_state_id = gs.id
WHERE gs.id = $1
ORDER BY fir.created_at DESC
LIMIT 1;

-- name: GetCurrentQuestionByPlayerID :one
SELECT
    gs.id AS game_state_id,
    fr.round,
    fr.round_type,
    r.room_code,
    gs.submit_deadline,
    p.id AS player_id,
    p.nickname,
    fpr.player_role AS role,
    COALESCE(fq1.question, fq2.question) AS question,
    p.avatar,
    COALESCE(fia.is_ready, FALSE) AS is_answer_ready
FROM players p
JOIN rooms_players rp ON p.id = rp.player_id
JOIN rooms r ON rp.room_id = r.id
JOIN game_state gs ON gs.room_id = r.id
JOIN fibbing_it_rounds fr ON fr.game_state_id = gs.id
LEFT JOIN questions fq1 ON fr.fibber_question_id = fq1.id
LEFT JOIN questions fq2 ON fr.normal_question_id = fq2.id
LEFT JOIN fibbing_it_player_roles fpr ON p.id = fpr.player_id AND fr.id = fpr.round_id
LEFT JOIN fibbing_it_answers fia ON p.id = fia.player_id AND fr.id = fia.round_id
WHERE p.id = $1
ORDER BY p.created_at
LIMIT 1;

-- name: GetVotingState :many
SELECT
    fir.round AS round,
    gs.id as game_state_id,
    q.question,
    gs.submit_deadline,
    p.id AS player_id,
    p.nickname,
    p.avatar,
    COALESCE(COUNT(fv.id), 0) AS votes,
    fia.answer,
    fv.is_ready,
    fpr.player_role AS role
FROM fibbing_it_rounds fir
JOIN questions q ON fir.fibber_question_id = q.id
JOIN game_state gs ON fir.game_state_id = gs.id
JOIN rooms_players rp ON rp.room_id = gs.room_id
JOIN players p ON p.id = rp.player_id
LEFT JOIN fibbing_it_answers fia ON fia.round_id = fir.id AND fia.player_id = p.id
LEFT JOIN fibbing_it_votes fv ON fv.round_id = fir.id AND fv.voted_for_player_id = p.id
LEFT JOIN fibbing_it_player_roles fpr ON p.id = fpr.player_id AND fir.id = fpr.round_id
WHERE fir.id = $1
GROUP BY
    fir.round,
    q.question,
    gs.submit_deadline,
    p.id,
    p.nickname,
    p.avatar,
    fia.answer,
    fv.is_ready,
    fpr.player_role,
    gs.id
ORDER BY votes DESC, p.nickname;

-- name: ToggleAnswerIsReady :one
UPDATE fibbing_it_answers SET is_ready = NOT is_ready WHERE player_id = $1 RETURNING *;

-- name: GetAllPlayerAnswerIsReady :one
SELECT
    COUNT(*) = SUM(CASE WHEN COALESCE(fa.is_ready, FALSE) THEN 1 ELSE 0 END) AS all_players_ready
FROM rooms_players rp
LEFT JOIN fibbing_it_answers fa ON fa.player_id = rp.player_id AND fa.round_id = (
    SELECT fir.id
    FROM fibbing_it_rounds fir
    WHERE fir.game_state_id = (
        SELECT gs.id
        FROM game_state gs
        WHERE gs.room_id = rp.room_id
        ORDER BY gs.created_at DESC
        LIMIT 1
    )
    ORDER BY fir.created_at DESC
    LIMIT 1
)
JOIN game_state gs ON gs.room_id = rp.room_id
WHERE rp.room_id = (
    SELECT room_id
    FROM rooms_players rp
    WHERE rp.player_id = $1
    LIMIT 1
);

-- name: ToggleVotingIsReady :one
UPDATE fibbing_it_votes SET is_ready = NOT is_ready WHERE player_id = $1 RETURNING *;

-- name: GetAllPlayersVotingIsReady :one
SELECT
    COUNT(*) = SUM(CASE WHEN COALESCE(fa.is_ready, FALSE) THEN 1 ELSE 0 END) AS all_players_ready
FROM rooms_players rp
LEFT JOIN fibbing_it_votes fa ON fa.player_id = rp.player_id AND fa.round_id = (
    SELECT fir.id
    FROM fibbing_it_rounds fir
    JOIN game_state gs ON fir.game_state_id = gs.id
    WHERE gs.room_id = rp.room_id
    ORDER BY fir.created_at DESC
    LIMIT 1
)
JOIN game_state gs ON gs.room_id = rp.room_id
WHERE rp.room_id = (
    SELECT room_id
    FROM rooms_players rp
    WHERE rp.player_id = $1
    LIMIT 1
);


-- name: AddFibbingItScore :one
INSERT INTO fibbing_it_scores (id, player_id, score, round_id) VALUES ($1, $2, $3, $4) RETURNING *;


-- name: GetAllVotesForRoundByGameStateID :many
SELECT
    v.player_id AS voter_id,
    p1.nickname AS voter_nickname,
    p1.avatar AS voter_avatar,
    v.voted_for_player_id AS voted_for_id,
    p2.nickname AS voted_for_nickname,
    r.player_id AS fibber_id,
    p3.nickname AS fibber_nickname,
    v.round_id
FROM fibbing_it_votes v
JOIN players p1 ON v.player_id = p1.id
JOIN players p2 ON v.voted_for_player_id = p2.id
JOIN fibbing_it_rounds fr ON v.round_id = fr.id
JOIN fibbing_it_player_roles r ON fr.id = r.round_id AND r.player_role = 'fibber'
JOIN players p3 ON r.player_id = p3.id
WHERE
    fr.game_state_id = $1
    AND fr.round_type = (
        SELECT round_type
        FROM fibbing_it_rounds
        WHERE game_state_id = $1
        ORDER BY round DESC
        LIMIT 1
    )
ORDER BY v.round_id DESC;
