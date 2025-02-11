// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: query.sql

package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

const addFibbingItRole = `-- name: AddFibbingItRole :one
INSERT INTO fibbing_it_player_roles (id, player_role, round_id, player_id) VALUES ($1, $2, $3, $4) RETURNING id, created_at, updated_at, player_role, round_id, player_id
`

type AddFibbingItRoleParams struct {
	ID         uuid.UUID
	PlayerRole string
	RoundID    uuid.UUID
	PlayerID   uuid.UUID
}

func (q *Queries) AddFibbingItRole(ctx context.Context, arg AddFibbingItRoleParams) (FibbingItPlayerRole, error) {
	row := q.db.QueryRow(ctx, addFibbingItRole,
		arg.ID,
		arg.PlayerRole,
		arg.RoundID,
		arg.PlayerID,
	)
	var i FibbingItPlayerRole
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.PlayerRole,
		&i.RoundID,
		&i.PlayerID,
	)
	return i, err
}

const addFibbingItRound = `-- name: AddFibbingItRound :one
INSERT INTO fibbing_it_rounds (id, round_type, round, fibber_question_id, normal_question_id, game_state_id) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, created_at, updated_at, round_type, round, fibber_question_id, normal_question_id, game_state_id
`

type AddFibbingItRoundParams struct {
	ID               uuid.UUID
	RoundType        string
	Round            int32
	FibberQuestionID uuid.UUID
	NormalQuestionID uuid.UUID
	GameStateID      uuid.UUID
}

func (q *Queries) AddFibbingItRound(ctx context.Context, arg AddFibbingItRoundParams) (FibbingItRound, error) {
	row := q.db.QueryRow(ctx, addFibbingItRound,
		arg.ID,
		arg.RoundType,
		arg.Round,
		arg.FibberQuestionID,
		arg.NormalQuestionID,
		arg.GameStateID,
	)
	var i FibbingItRound
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.RoundType,
		&i.Round,
		&i.FibberQuestionID,
		&i.NormalQuestionID,
		&i.GameStateID,
	)
	return i, err
}

const addFibbingItScore = `-- name: AddFibbingItScore :one
INSERT INTO fibbing_it_scores (id, player_id, score, round_id) VALUES ($1, $2, $3, $4) RETURNING id, created_at, updated_at, player_id, score, round_id
`

type AddFibbingItScoreParams struct {
	ID       uuid.UUID
	PlayerID uuid.UUID
	Score    int32
	RoundID  uuid.UUID
}

func (q *Queries) AddFibbingItScore(ctx context.Context, arg AddFibbingItScoreParams) (FibbingItScore, error) {
	row := q.db.QueryRow(ctx, addFibbingItScore,
		arg.ID,
		arg.PlayerID,
		arg.Score,
		arg.RoundID,
	)
	var i FibbingItScore
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.PlayerID,
		&i.Score,
		&i.RoundID,
	)
	return i, err
}

const addGameState = `-- name: AddGameState :one
INSERT INTO game_state (id, room_id, submit_deadline, state) VALUES ($1, $2, $3, $4) RETURNING id, created_at, updated_at, room_id, submit_deadline, state
`

type AddGameStateParams struct {
	ID             uuid.UUID
	RoomID         uuid.UUID
	SubmitDeadline pgtype.Timestamp
	State          string
}

func (q *Queries) AddGameState(ctx context.Context, arg AddGameStateParams) (GameState, error) {
	row := q.db.QueryRow(ctx, addGameState,
		arg.ID,
		arg.RoomID,
		arg.SubmitDeadline,
		arg.State,
	)
	var i GameState
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.RoomID,
		&i.SubmitDeadline,
		&i.State,
	)
	return i, err
}

const addGroup = `-- name: AddGroup :one
INSERT INTO questions_groups (id, group_name, group_type)
VALUES ($1, $2, $3)
RETURNING id, created_at, updated_at, group_name, group_type
`

type AddGroupParams struct {
	ID        uuid.UUID
	GroupName string
	GroupType string
}

func (q *Queries) AddGroup(ctx context.Context, arg AddGroupParams) (QuestionsGroup, error) {
	row := q.db.QueryRow(ctx, addGroup, arg.ID, arg.GroupName, arg.GroupType)
	var i QuestionsGroup
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.GroupName,
		&i.GroupType,
	)
	return i, err
}

const addPlayer = `-- name: AddPlayer :one
INSERT INTO players (id, avatar, nickname, locale) VALUES ($1, $2, $3, $4) RETURNING id, created_at, updated_at, avatar, nickname, is_ready, locale
`

type AddPlayerParams struct {
	ID       uuid.UUID
	Avatar   string
	Nickname string
	Locale   pgtype.Text
}

func (q *Queries) AddPlayer(ctx context.Context, arg AddPlayerParams) (Player, error) {
	row := q.db.QueryRow(ctx, addPlayer,
		arg.ID,
		arg.Avatar,
		arg.Nickname,
		arg.Locale,
	)
	var i Player
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Avatar,
		&i.Nickname,
		&i.IsReady,
		&i.Locale,
	)
	return i, err
}

const addQuestion = `-- name: AddQuestion :one
INSERT INTO questions (id, game_name, group_id, round_type) VALUES ($1, $2, $3, $4) RETURNING id, created_at, updated_at, game_name, round_type, enabled, group_id
`

type AddQuestionParams struct {
	ID        uuid.UUID
	GameName  string
	GroupID   uuid.UUID
	RoundType string
}

func (q *Queries) AddQuestion(ctx context.Context, arg AddQuestionParams) (Question, error) {
	row := q.db.QueryRow(ctx, addQuestion,
		arg.ID,
		arg.GameName,
		arg.GroupID,
		arg.RoundType,
	)
	var i Question
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.GameName,
		&i.RoundType,
		&i.Enabled,
		&i.GroupID,
	)
	return i, err
}

const addQuestionTranslation = `-- name: AddQuestionTranslation :one
INSERT INTO questions_i18n (id,  question, locale, question_id) VALUES ($1, $2, $3, $4) RETURNING id, created_at, updated_at, question, locale, question_id
`

type AddQuestionTranslationParams struct {
	ID         uuid.UUID
	Question   string
	Locale     string
	QuestionID uuid.UUID
}

func (q *Queries) AddQuestionTranslation(ctx context.Context, arg AddQuestionTranslationParams) (QuestionsI18n, error) {
	row := q.db.QueryRow(ctx, addQuestionTranslation,
		arg.ID,
		arg.Question,
		arg.Locale,
		arg.QuestionID,
	)
	var i QuestionsI18n
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Question,
		&i.Locale,
		&i.QuestionID,
	)
	return i, err
}

const addRoom = `-- name: AddRoom :one
INSERT INTO rooms (id, game_name, host_player, room_code, room_state) VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at, updated_at, game_name, host_player, room_state, room_code
`

type AddRoomParams struct {
	ID         uuid.UUID
	GameName   string
	HostPlayer uuid.UUID
	RoomCode   string
	RoomState  string
}

func (q *Queries) AddRoom(ctx context.Context, arg AddRoomParams) (Room, error) {
	row := q.db.QueryRow(ctx, addRoom,
		arg.ID,
		arg.GameName,
		arg.HostPlayer,
		arg.RoomCode,
		arg.RoomState,
	)
	var i Room
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.GameName,
		&i.HostPlayer,
		&i.RoomState,
		&i.RoomCode,
	)
	return i, err
}

const addRoomPlayer = `-- name: AddRoomPlayer :one
INSERT INTO rooms_players (room_id, player_id) VALUES ($1, $2) RETURNING room_id, player_id, created_at, updated_at
`

type AddRoomPlayerParams struct {
	RoomID   uuid.UUID
	PlayerID uuid.UUID
}

func (q *Queries) AddRoomPlayer(ctx context.Context, arg AddRoomPlayerParams) (RoomsPlayer, error) {
	row := q.db.QueryRow(ctx, addRoomPlayer, arg.RoomID, arg.PlayerID)
	var i RoomsPlayer
	err := row.Scan(
		&i.RoomID,
		&i.PlayerID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const disableQuestion = `-- name: DisableQuestion :one
UPDATE questions SET enabled = false WHERE id = $1 RETURNING id, created_at, updated_at, game_name, round_type, enabled, group_id
`

func (q *Queries) DisableQuestion(ctx context.Context, id uuid.UUID) (Question, error) {
	row := q.db.QueryRow(ctx, disableQuestion, id)
	var i Question
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.GameName,
		&i.RoundType,
		&i.Enabled,
		&i.GroupID,
	)
	return i, err
}

const enableQuestion = `-- name: EnableQuestion :one
UPDATE questions SET enabled = true WHERE id = $1 RETURNING id, created_at, updated_at, game_name, round_type, enabled, group_id
`

func (q *Queries) EnableQuestion(ctx context.Context, id uuid.UUID) (Question, error) {
	row := q.db.QueryRow(ctx, enableQuestion, id)
	var i Question
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.GameName,
		&i.RoundType,
		&i.Enabled,
		&i.GroupID,
	)
	return i, err
}

const getAllPlayerAnswerIsReady = `-- name: GetAllPlayerAnswerIsReady :one
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
)
`

func (q *Queries) GetAllPlayerAnswerIsReady(ctx context.Context, playerID uuid.UUID) (bool, error) {
	row := q.db.QueryRow(ctx, getAllPlayerAnswerIsReady, playerID)
	var all_players_ready bool
	err := row.Scan(&all_players_ready)
	return all_players_ready, err
}

const getAllPlayerByRoomCode = `-- name: GetAllPlayerByRoomCode :many
SELECT p.id, p.created_at, p.updated_at, p.avatar, p.nickname, p.is_ready, r.room_code, r.host_player
FROM players p
JOIN rooms_players rp ON p.id = rp.player_id
JOIN rooms r ON rp.room_id = r.id
WHERE rp.room_id = (
    SELECT r_inner.id
    FROM rooms r_inner
    WHERE r_inner.room_code = $1 AND (r_inner.room_state = 'CREATED' OR r_inner.room_state = 'PLAYING')
)
ORDER BY p.created_at
`

type GetAllPlayerByRoomCodeRow struct {
	ID         uuid.UUID
	CreatedAt  pgtype.Timestamp
	UpdatedAt  pgtype.Timestamp
	Avatar     string
	Nickname   string
	IsReady    pgtype.Bool
	RoomCode   string
	HostPlayer uuid.UUID
}

func (q *Queries) GetAllPlayerByRoomCode(ctx context.Context, roomCode string) ([]GetAllPlayerByRoomCodeRow, error) {
	rows, err := q.db.Query(ctx, getAllPlayerByRoomCode, roomCode)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetAllPlayerByRoomCodeRow
	for rows.Next() {
		var i GetAllPlayerByRoomCodeRow
		if err := rows.Scan(
			&i.ID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.Avatar,
			&i.Nickname,
			&i.IsReady,
			&i.RoomCode,
			&i.HostPlayer,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getAllPlayersByGameStateID = `-- name: GetAllPlayersByGameStateID :many
SELECT p.id, p.nickname, p.avatar, p.locale
FROM players p
JOIN rooms_players rp ON p.id = rp.player_id
JOIN game_state gs ON rp.room_id = gs.room_id
WHERE gs.id = $1
`

type GetAllPlayersByGameStateIDRow struct {
	ID       uuid.UUID
	Nickname string
	Avatar   string
	Locale   pgtype.Text
}

func (q *Queries) GetAllPlayersByGameStateID(ctx context.Context, id uuid.UUID) ([]GetAllPlayersByGameStateIDRow, error) {
	rows, err := q.db.Query(ctx, getAllPlayersByGameStateID, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetAllPlayersByGameStateIDRow
	for rows.Next() {
		var i GetAllPlayersByGameStateIDRow
		if err := rows.Scan(
			&i.ID,
			&i.Nickname,
			&i.Avatar,
			&i.Locale,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getAllPlayersInRoom = `-- name: GetAllPlayersInRoom :many
SELECT p.id, p.created_at, p.updated_at, p.avatar, p.nickname, p.is_ready, p.locale, r.room_code, r.host_player
FROM players p
JOIN rooms_players rp ON p.id = rp.player_id
JOIN rooms r ON rp.room_id = r.id
WHERE rp.room_id = (
    SELECT rp_inner.room_id
    FROM rooms_players rp_inner
    WHERE rp_inner.player_id = $1
)
`

type GetAllPlayersInRoomRow struct {
	ID         uuid.UUID
	CreatedAt  pgtype.Timestamp
	UpdatedAt  pgtype.Timestamp
	Avatar     string
	Nickname   string
	IsReady    pgtype.Bool
	Locale     pgtype.Text
	RoomCode   string
	HostPlayer uuid.UUID
}

func (q *Queries) GetAllPlayersInRoom(ctx context.Context, playerID uuid.UUID) ([]GetAllPlayersInRoomRow, error) {
	rows, err := q.db.Query(ctx, getAllPlayersInRoom, playerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetAllPlayersInRoomRow
	for rows.Next() {
		var i GetAllPlayersInRoomRow
		if err := rows.Scan(
			&i.ID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.Avatar,
			&i.Nickname,
			&i.IsReady,
			&i.Locale,
			&i.RoomCode,
			&i.HostPlayer,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getAllPlayersVotingIsReady = `-- name: GetAllPlayersVotingIsReady :one
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
)
`

func (q *Queries) GetAllPlayersVotingIsReady(ctx context.Context, playerID uuid.UUID) (bool, error) {
	row := q.db.QueryRow(ctx, getAllPlayersVotingIsReady, playerID)
	var all_players_ready bool
	err := row.Scan(&all_players_ready)
	return all_players_ready, err
}

const getAllVotesForRoundByGameStateID = `-- name: GetAllVotesForRoundByGameStateID :many
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
ORDER BY v.round_id DESC
`

type GetAllVotesForRoundByGameStateIDRow struct {
	VoterID          uuid.UUID
	VoterNickname    string
	VoterAvatar      string
	VotedForID       uuid.UUID
	VotedForNickname string
	FibberID         uuid.UUID
	FibberNickname   string
	RoundID          uuid.UUID
}

func (q *Queries) GetAllVotesForRoundByGameStateID(ctx context.Context, gameStateID uuid.UUID) ([]GetAllVotesForRoundByGameStateIDRow, error) {
	rows, err := q.db.Query(ctx, getAllVotesForRoundByGameStateID, gameStateID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetAllVotesForRoundByGameStateIDRow
	for rows.Next() {
		var i GetAllVotesForRoundByGameStateIDRow
		if err := rows.Scan(
			&i.VoterID,
			&i.VoterNickname,
			&i.VoterAvatar,
			&i.VotedForID,
			&i.VotedForNickname,
			&i.FibberID,
			&i.FibberNickname,
			&i.RoundID,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getCurrentQuestionByPlayerID = `-- name: GetCurrentQuestionByPlayerID :one
SELECT
    gs.id AS game_state_id,
    fr.round,
    fr.round_type,
    r.room_code,
    gs.submit_deadline,
    p.id AS player_id,
    p.nickname,
    fpr.player_role AS role,
    qi.question AS question,
    p.avatar,
    COALESCE(fia.answer, '') AS current_answer,
    COALESCE(fia.is_ready, FALSE) AS is_answer_ready
FROM players p
JOIN rooms_players rp ON p.id = rp.player_id
JOIN rooms r ON rp.room_id = r.id
JOIN game_state gs ON gs.room_id = r.id
JOIN fibbing_it_rounds fr ON fr.game_state_id = gs.id
LEFT JOIN questions q ON fr.normal_question_id = q.id
LEFT JOIN questions_i18n qi ON q.id = qi.question_id AND qi.locale = 'en-GB'
LEFT JOIN fibbing_it_player_roles fpr ON p.id = fpr.player_id AND fr.id = fpr.round_id
LEFT JOIN fibbing_it_answers fia ON p.id = fia.player_id AND fr.id = fia.round_id
WHERE p.id = $1
ORDER BY fr.round DESC
LIMIT 1
`

type GetCurrentQuestionByPlayerIDRow struct {
	GameStateID    uuid.UUID
	Round          int32
	RoundType      string
	RoomCode       string
	SubmitDeadline pgtype.Timestamp
	PlayerID       uuid.UUID
	Nickname       string
	Role           pgtype.Text
	Question       pgtype.Text
	Avatar         string
	CurrentAnswer  string
	IsAnswerReady  bool
}

func (q *Queries) GetCurrentQuestionByPlayerID(ctx context.Context, id uuid.UUID) (GetCurrentQuestionByPlayerIDRow, error) {
	row := q.db.QueryRow(ctx, getCurrentQuestionByPlayerID, id)
	var i GetCurrentQuestionByPlayerIDRow
	err := row.Scan(
		&i.GameStateID,
		&i.Round,
		&i.RoundType,
		&i.RoomCode,
		&i.SubmitDeadline,
		&i.PlayerID,
		&i.Nickname,
		&i.Role,
		&i.Question,
		&i.Avatar,
		&i.CurrentAnswer,
		&i.IsAnswerReady,
	)
	return i, err
}

const getFibberByRoundID = `-- name: GetFibberByRoundID :one
SELECT id, created_at, updated_at, player_role, round_id, player_id FROM fibbing_it_player_roles WHERE round_id=$1 and player_role='fibber'
`

func (q *Queries) GetFibberByRoundID(ctx context.Context, roundID uuid.UUID) (FibbingItPlayerRole, error) {
	row := q.db.QueryRow(ctx, getFibberByRoundID, roundID)
	var i FibbingItPlayerRole
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.PlayerRole,
		&i.RoundID,
		&i.PlayerID,
	)
	return i, err
}

const getGameState = `-- name: GetGameState :one
SELECT
    gs.id,
    gs.created_at,
    gs.updated_at,
    gs.room_id,
    gs.submit_deadline,
    gs.state
FROM game_state gs
WHERE gs.id = $1
`

func (q *Queries) GetGameState(ctx context.Context, id uuid.UUID) (GameState, error) {
	row := q.db.QueryRow(ctx, getGameState, id)
	var i GameState
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.RoomID,
		&i.SubmitDeadline,
		&i.State,
	)
	return i, err
}

const getGameStateByPlayerID = `-- name: GetGameStateByPlayerID :one
SELECT
    gs.id,
    gs.created_at,
    gs.updated_at,
    gs.room_id,
    gs.submit_deadline,
    gs.state
FROM game_state gs
JOIN rooms_players rp ON gs.room_id = rp.room_id
WHERE rp.player_id = $1
`

func (q *Queries) GetGameStateByPlayerID(ctx context.Context, playerID uuid.UUID) (GameState, error) {
	row := q.db.QueryRow(ctx, getGameStateByPlayerID, playerID)
	var i GameState
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.RoomID,
		&i.SubmitDeadline,
		&i.State,
	)
	return i, err
}

const getGroupByName = `-- name: GetGroupByName :one
SELECT
   id, created_at, updated_at, group_name, group_type
FROM
   questions_groups
WHERE
   group_name = $1
`

func (q *Queries) GetGroupByName(ctx context.Context, groupName string) (QuestionsGroup, error) {
	row := q.db.QueryRow(ctx, getGroupByName, groupName)
	var i QuestionsGroup
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.GroupName,
		&i.GroupType,
	)
	return i, err
}

const getGroups = `-- name: GetGroups :many
SELECT
   id, created_at, updated_at, group_name, group_type
FROM
   questions_groups
ORDER BY group_name DESC
`

func (q *Queries) GetGroups(ctx context.Context) ([]QuestionsGroup, error) {
	rows, err := q.db.Query(ctx, getGroups)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []QuestionsGroup
	for rows.Next() {
		var i QuestionsGroup
		if err := rows.Scan(
			&i.ID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.GroupName,
			&i.GroupType,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getLatestRoundByGameStateID = `-- name: GetLatestRoundByGameStateID :one
SELECT fir.id, fir.created_at, fir.updated_at, fir.round_type, fir.round, fir.fibber_question_id, fir.normal_question_id, fir.game_state_id, gs.submit_deadline
FROM fibbing_it_rounds fir
JOIN game_state gs ON fir.game_state_id = gs.id
WHERE gs.id = $1
ORDER BY fir.created_at DESC
LIMIT 1
`

type GetLatestRoundByGameStateIDRow struct {
	ID               uuid.UUID
	CreatedAt        pgtype.Timestamp
	UpdatedAt        pgtype.Timestamp
	RoundType        string
	Round            int32
	FibberQuestionID uuid.UUID
	NormalQuestionID uuid.UUID
	GameStateID      uuid.UUID
	SubmitDeadline   pgtype.Timestamp
}

func (q *Queries) GetLatestRoundByGameStateID(ctx context.Context, id uuid.UUID) (GetLatestRoundByGameStateIDRow, error) {
	row := q.db.QueryRow(ctx, getLatestRoundByGameStateID, id)
	var i GetLatestRoundByGameStateIDRow
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.RoundType,
		&i.Round,
		&i.FibberQuestionID,
		&i.NormalQuestionID,
		&i.GameStateID,
		&i.SubmitDeadline,
	)
	return i, err
}

const getLatestRoundByPlayerID = `-- name: GetLatestRoundByPlayerID :one
SELECT fir.id, fir.created_at, fir.updated_at, fir.round_type, fir.round, fir.fibber_question_id, fir.normal_question_id, fir.game_state_id, gs.submit_deadline
FROM fibbing_it_rounds fir
JOIN game_state gs ON fir.game_state_id = gs.id
JOIN rooms_players rp ON gs.room_id = rp.room_id
WHERE rp.player_id = $1
ORDER BY fir.created_at DESC
LIMIT 1
`

type GetLatestRoundByPlayerIDRow struct {
	ID               uuid.UUID
	CreatedAt        pgtype.Timestamp
	UpdatedAt        pgtype.Timestamp
	RoundType        string
	Round            int32
	FibberQuestionID uuid.UUID
	NormalQuestionID uuid.UUID
	GameStateID      uuid.UUID
	SubmitDeadline   pgtype.Timestamp
}

func (q *Queries) GetLatestRoundByPlayerID(ctx context.Context, playerID uuid.UUID) (GetLatestRoundByPlayerIDRow, error) {
	row := q.db.QueryRow(ctx, getLatestRoundByPlayerID, playerID)
	var i GetLatestRoundByPlayerIDRow
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.RoundType,
		&i.Round,
		&i.FibberQuestionID,
		&i.NormalQuestionID,
		&i.GameStateID,
		&i.SubmitDeadline,
	)
	return i, err
}

const getPlayerByID = `-- name: GetPlayerByID :one
SELECT id, created_at, updated_at, avatar, nickname, is_ready, locale FROM players WHERE id = $1
`

func (q *Queries) GetPlayerByID(ctx context.Context, id uuid.UUID) (Player, error) {
	row := q.db.QueryRow(ctx, getPlayerByID, id)
	var i Player
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Avatar,
		&i.Nickname,
		&i.IsReady,
		&i.Locale,
	)
	return i, err
}

const getQuestions = `-- name: GetQuestions :many
SELECT q.id, q.created_at, q.updated_at, q.game_name, q.round_type, q.enabled, q.group_id, qi.question, qi.locale, qg.group_name, qg.group_type
FROM questions q
JOIN questions_i18n qi ON q.id = qi.question_id
JOIN questions_groups qg ON q.group_id = qg.id
WHERE ($1::text = '' OR qi.locale = $1)
  AND ($2::text = '' OR q.round_type = $2)
  AND ($3::text = '' OR qg.group_name = $3)
  AND ($4::boolean IS NULL OR q.enabled = $4)
ORDER BY q.created_at DESC
LIMIT $5 OFFSET $6
`

type GetQuestionsParams struct {
	Column1 string
	Column2 string
	Column3 string
	Column4 bool
	Limit   int32
	Offset  int32
}

type GetQuestionsRow struct {
	ID        uuid.UUID
	CreatedAt pgtype.Timestamp
	UpdatedAt pgtype.Timestamp
	GameName  string
	RoundType string
	Enabled   pgtype.Bool
	GroupID   uuid.UUID
	Question  string
	Locale    string
	GroupName string
	GroupType string
}

func (q *Queries) GetQuestions(ctx context.Context, arg GetQuestionsParams) ([]GetQuestionsRow, error) {
	rows, err := q.db.Query(ctx, getQuestions,
		arg.Column1,
		arg.Column2,
		arg.Column3,
		arg.Column4,
		arg.Limit,
		arg.Offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetQuestionsRow
	for rows.Next() {
		var i GetQuestionsRow
		if err := rows.Scan(
			&i.ID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.GameName,
			&i.RoundType,
			&i.Enabled,
			&i.GroupID,
			&i.Question,
			&i.Locale,
			&i.GroupName,
			&i.GroupType,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getRandomQuestionByRound = `-- name: GetRandomQuestionByRound :many
SELECT
    qi.id, qi.created_at, qi.updated_at, qi.question, qi.locale, qi.question_id,
    random_question.group_id,
    random_question.id
FROM questions_i18n qi
JOIN (
    SELECT q.id, q.group_id
    FROM questions q
    WHERE q.game_name = $1
      AND q.round_type = $2
      AND q.enabled = TRUE
    ORDER BY RANDOM()
    LIMIT 1
) random_question ON qi.question_id = random_question.id
`

type GetRandomQuestionByRoundParams struct {
	GameName  string
	RoundType string
}

type GetRandomQuestionByRoundRow struct {
	ID         uuid.UUID
	CreatedAt  pgtype.Timestamp
	UpdatedAt  pgtype.Timestamp
	Question   string
	Locale     string
	QuestionID uuid.UUID
	GroupID    uuid.UUID
	ID_2       uuid.UUID
}

func (q *Queries) GetRandomQuestionByRound(ctx context.Context, arg GetRandomQuestionByRoundParams) ([]GetRandomQuestionByRoundRow, error) {
	rows, err := q.db.Query(ctx, getRandomQuestionByRound, arg.GameName, arg.RoundType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetRandomQuestionByRoundRow
	for rows.Next() {
		var i GetRandomQuestionByRoundRow
		if err := rows.Scan(
			&i.ID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.Question,
			&i.Locale,
			&i.QuestionID,
			&i.GroupID,
			&i.ID_2,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getRandomQuestionInGroup = `-- name: GetRandomQuestionInGroup :many
SELECT
    qi.id, qi.created_at, qi.updated_at, qi.question, qi.locale, qi.question_id,
    random_question.id
FROM questions_i18n qi
JOIN (
    SELECT q.id
    FROM questions q
    JOIN questions_groups qg ON q.group_id = qg.id
    WHERE qg.group_type = 'questions'
      AND q.group_id = $1
      AND q.enabled = TRUE
      AND q.id != $2
      AND q.round_type = $3
    ORDER BY RANDOM()
    LIMIT 1
) random_question ON qi.question_id = random_question.id
`

type GetRandomQuestionInGroupParams struct {
	GroupID   uuid.UUID
	ID        uuid.UUID
	RoundType string
}

type GetRandomQuestionInGroupRow struct {
	ID         uuid.UUID
	CreatedAt  pgtype.Timestamp
	UpdatedAt  pgtype.Timestamp
	Question   string
	Locale     string
	QuestionID uuid.UUID
	ID_2       uuid.UUID
}

func (q *Queries) GetRandomQuestionInGroup(ctx context.Context, arg GetRandomQuestionInGroupParams) ([]GetRandomQuestionInGroupRow, error) {
	rows, err := q.db.Query(ctx, getRandomQuestionInGroup, arg.GroupID, arg.ID, arg.RoundType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetRandomQuestionInGroupRow
	for rows.Next() {
		var i GetRandomQuestionInGroupRow
		if err := rows.Scan(
			&i.ID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.Question,
			&i.Locale,
			&i.QuestionID,
			&i.ID_2,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getRoomByCode = `-- name: GetRoomByCode :one
SELECT id, created_at, updated_at, game_name, host_player, room_state, room_code FROM rooms WHERE room_code = $1
`

func (q *Queries) GetRoomByCode(ctx context.Context, roomCode string) (Room, error) {
	row := q.db.QueryRow(ctx, getRoomByCode, roomCode)
	var i Room
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.GameName,
		&i.HostPlayer,
		&i.RoomState,
		&i.RoomCode,
	)
	return i, err
}

const getRoomByPlayerID = `-- name: GetRoomByPlayerID :one
SELECT r.id, r.created_at, r.updated_at, r.game_name, r.host_player, r.room_state, r.room_code FROM rooms r JOIN rooms_players rp ON r.id = rp.room_id WHERE rp.player_id = $1
`

func (q *Queries) GetRoomByPlayerID(ctx context.Context, playerID uuid.UUID) (Room, error) {
	row := q.db.QueryRow(ctx, getRoomByPlayerID, playerID)
	var i Room
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.GameName,
		&i.HostPlayer,
		&i.RoomState,
		&i.RoomCode,
	)
	return i, err
}

const getTotalScoresByGameStateID = `-- name: GetTotalScoresByGameStateID :many
SELECT
    s.player_id,
    p.avatar,
    p.nickname,
    SUM(s.score) AS total_score
FROM
    fibbing_it_scores s
JOIN
    fibbing_it_rounds r ON s.round_id = r.id
JOIN
    game_state gs ON r.game_state_id = gs.id
JOIN
    players p ON s.player_id = p.id
WHERE
    gs.id = $1 AND r.id != $2
GROUP BY
    s.player_id,
    p.avatar,
    p.nickname
`

type GetTotalScoresByGameStateIDParams struct {
	ID   uuid.UUID
	ID_2 uuid.UUID
}

type GetTotalScoresByGameStateIDRow struct {
	PlayerID   uuid.UUID
	Avatar     string
	Nickname   string
	TotalScore int64
}

func (q *Queries) GetTotalScoresByGameStateID(ctx context.Context, arg GetTotalScoresByGameStateIDParams) ([]GetTotalScoresByGameStateIDRow, error) {
	rows, err := q.db.Query(ctx, getTotalScoresByGameStateID, arg.ID, arg.ID_2)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetTotalScoresByGameStateIDRow
	for rows.Next() {
		var i GetTotalScoresByGameStateIDRow
		if err := rows.Scan(
			&i.PlayerID,
			&i.Avatar,
			&i.Nickname,
			&i.TotalScore,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getVotingState = `-- name: GetVotingState :many
SELECT
    fir.round AS round,
    gs.id AS game_state_id,
    qi.question,
    gs.submit_deadline,
    p.id AS player_id,
    p.nickname,
    p.avatar,
    COALESCE(COUNT(fv.id), 0) AS votes,
    fia.answer,
    fv.is_ready,
    fpr.player_role AS role
FROM fibbing_it_rounds fir
JOIN questions q ON fir.normal_question_id = q.id
JOIN questions_i18n qi ON q.id = qi.question_id AND qi.locale = 'en-GB'
JOIN game_state gs ON fir.game_state_id = gs.id
JOIN rooms_players rp ON rp.room_id = gs.room_id
JOIN players p ON p.id = rp.player_id
LEFT JOIN fibbing_it_answers fia ON fia.round_id = fir.id AND fia.player_id = p.id
LEFT JOIN fibbing_it_votes fv ON fv.round_id = fir.id AND fv.voted_for_player_id = p.id
LEFT JOIN fibbing_it_player_roles fpr ON p.id = fpr.player_id AND fir.id = fpr.round_id
WHERE fir.id = $1
GROUP BY
    fir.round,
    qi.question,
    gs.submit_deadline,
    p.id,
    p.nickname,
    p.avatar,
    fia.answer,
    fv.is_ready,
    fpr.player_role,
    gs.id
ORDER BY votes DESC, p.nickname
`

type GetVotingStateRow struct {
	Round          int32
	GameStateID    uuid.UUID
	Question       string
	SubmitDeadline pgtype.Timestamp
	PlayerID       uuid.UUID
	Nickname       string
	Avatar         string
	Votes          interface{}
	Answer         pgtype.Text
	IsReady        pgtype.Bool
	Role           pgtype.Text
}

func (q *Queries) GetVotingState(ctx context.Context, id uuid.UUID) ([]GetVotingStateRow, error) {
	rows, err := q.db.Query(ctx, getVotingState, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetVotingStateRow
	for rows.Next() {
		var i GetVotingStateRow
		if err := rows.Scan(
			&i.Round,
			&i.GameStateID,
			&i.Question,
			&i.SubmitDeadline,
			&i.PlayerID,
			&i.Nickname,
			&i.Avatar,
			&i.Votes,
			&i.Answer,
			&i.IsReady,
			&i.Role,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const removePlayerFromRoom = `-- name: RemovePlayerFromRoom :one
DELETE FROM rooms_players WHERE player_id = $1 RETURNING room_id, player_id, created_at, updated_at
`

func (q *Queries) RemovePlayerFromRoom(ctx context.Context, playerID uuid.UUID) (RoomsPlayer, error) {
	row := q.db.QueryRow(ctx, removePlayerFromRoom, playerID)
	var i RoomsPlayer
	err := row.Scan(
		&i.RoomID,
		&i.PlayerID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const toggleAnswerIsReady = `-- name: ToggleAnswerIsReady :one
UPDATE fibbing_it_answers SET is_ready = NOT is_ready WHERE player_id = $1 RETURNING id, created_at, updated_at, answer, player_id, round_id, is_ready
`

func (q *Queries) ToggleAnswerIsReady(ctx context.Context, playerID uuid.UUID) (FibbingItAnswer, error) {
	row := q.db.QueryRow(ctx, toggleAnswerIsReady, playerID)
	var i FibbingItAnswer
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Answer,
		&i.PlayerID,
		&i.RoundID,
		&i.IsReady,
	)
	return i, err
}

const togglePlayerIsReady = `-- name: TogglePlayerIsReady :one
UPDATE players SET is_ready = NOT is_ready WHERE id = $1 RETURNING id, created_at, updated_at, avatar, nickname, is_ready, locale
`

func (q *Queries) TogglePlayerIsReady(ctx context.Context, id uuid.UUID) (Player, error) {
	row := q.db.QueryRow(ctx, togglePlayerIsReady, id)
	var i Player
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Avatar,
		&i.Nickname,
		&i.IsReady,
		&i.Locale,
	)
	return i, err
}

const toggleVotingIsReady = `-- name: ToggleVotingIsReady :one
UPDATE fibbing_it_votes SET is_ready = NOT is_ready WHERE player_id = $1 RETURNING id, created_at, updated_at, player_id, voted_for_player_id, round_id, is_ready
`

func (q *Queries) ToggleVotingIsReady(ctx context.Context, playerID uuid.UUID) (FibbingItVote, error) {
	row := q.db.QueryRow(ctx, toggleVotingIsReady, playerID)
	var i FibbingItVote
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.PlayerID,
		&i.VotedForPlayerID,
		&i.RoundID,
		&i.IsReady,
	)
	return i, err
}

const updateAvatar = `-- name: UpdateAvatar :one
UPDATE players SET avatar = $1 WHERE id = $2 RETURNING id, created_at, updated_at, avatar, nickname, is_ready, locale
`

type UpdateAvatarParams struct {
	Avatar string
	ID     uuid.UUID
}

func (q *Queries) UpdateAvatar(ctx context.Context, arg UpdateAvatarParams) (Player, error) {
	row := q.db.QueryRow(ctx, updateAvatar, arg.Avatar, arg.ID)
	var i Player
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Avatar,
		&i.Nickname,
		&i.IsReady,
		&i.Locale,
	)
	return i, err
}

const updateGameState = `-- name: UpdateGameState :one
UPDATE game_state SET state = $1, submit_deadline = $2 WHERE id = $3 RETURNING id, created_at, updated_at, room_id, submit_deadline, state
`

type UpdateGameStateParams struct {
	State          string
	SubmitDeadline pgtype.Timestamp
	ID             uuid.UUID
}

func (q *Queries) UpdateGameState(ctx context.Context, arg UpdateGameStateParams) (GameState, error) {
	row := q.db.QueryRow(ctx, updateGameState, arg.State, arg.SubmitDeadline, arg.ID)
	var i GameState
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.RoomID,
		&i.SubmitDeadline,
		&i.State,
	)
	return i, err
}

const updateLocale = `-- name: UpdateLocale :one
UPDATE players SET locale = $1 WHERE id = $2 RETURNING id, created_at, updated_at, avatar, nickname, is_ready, locale
`

type UpdateLocaleParams struct {
	Locale pgtype.Text
	ID     uuid.UUID
}

func (q *Queries) UpdateLocale(ctx context.Context, arg UpdateLocaleParams) (Player, error) {
	row := q.db.QueryRow(ctx, updateLocale, arg.Locale, arg.ID)
	var i Player
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Avatar,
		&i.Nickname,
		&i.IsReady,
		&i.Locale,
	)
	return i, err
}

const updateNickname = `-- name: UpdateNickname :one
UPDATE players SET nickname = $1 WHERE id = $2 RETURNING id, created_at, updated_at, avatar, nickname, is_ready, locale
`

type UpdateNicknameParams struct {
	Nickname string
	ID       uuid.UUID
}

func (q *Queries) UpdateNickname(ctx context.Context, arg UpdateNicknameParams) (Player, error) {
	row := q.db.QueryRow(ctx, updateNickname, arg.Nickname, arg.ID)
	var i Player
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Avatar,
		&i.Nickname,
		&i.IsReady,
		&i.Locale,
	)
	return i, err
}

const updateRoomState = `-- name: UpdateRoomState :one
UPDATE rooms SET room_state = $1 WHERE id = $2 RETURNING id, created_at, updated_at, game_name, host_player, room_state, room_code
`

type UpdateRoomStateParams struct {
	RoomState string
	ID        uuid.UUID
}

func (q *Queries) UpdateRoomState(ctx context.Context, arg UpdateRoomStateParams) (Room, error) {
	row := q.db.QueryRow(ctx, updateRoomState, arg.RoomState, arg.ID)
	var i Room
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.GameName,
		&i.HostPlayer,
		&i.RoomState,
		&i.RoomCode,
	)
	return i, err
}

const upsertFibbingItAnswer = `-- name: UpsertFibbingItAnswer :one
INSERT INTO fibbing_it_answers (id, answer, round_id, player_id)
VALUES ($1, $2, $3, $4)
ON CONFLICT (player_id, round_id) DO UPDATE SET
    answer = EXCLUDED.answer,
    updated_at = CURRENT_TIMESTAMP
RETURNING id, created_at, updated_at, answer, player_id, round_id, is_ready
`

type UpsertFibbingItAnswerParams struct {
	ID       uuid.UUID
	Answer   string
	RoundID  uuid.UUID
	PlayerID uuid.UUID
}

func (q *Queries) UpsertFibbingItAnswer(ctx context.Context, arg UpsertFibbingItAnswerParams) (FibbingItAnswer, error) {
	row := q.db.QueryRow(ctx, upsertFibbingItAnswer,
		arg.ID,
		arg.Answer,
		arg.RoundID,
		arg.PlayerID,
	)
	var i FibbingItAnswer
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Answer,
		&i.PlayerID,
		&i.RoundID,
		&i.IsReady,
	)
	return i, err
}

const upsertFibbingItVote = `-- name: UpsertFibbingItVote :exec
INSERT INTO fibbing_it_votes (id, player_id, voted_for_player_id, round_id)
VALUES ($1, $2, $3, $4)
ON CONFLICT(player_id, round_id) DO UPDATE SET
    updated_at = CURRENT_TIMESTAMP,
    player_id = EXCLUDED.player_id,
    voted_for_player_id = EXCLUDED.voted_for_player_id,
    round_id = EXCLUDED.round_id
RETURNING id, created_at, updated_at, player_id, voted_for_player_id, round_id, is_ready
`

type UpsertFibbingItVoteParams struct {
	ID               uuid.UUID
	PlayerID         uuid.UUID
	VotedForPlayerID uuid.UUID
	RoundID          uuid.UUID
}

func (q *Queries) UpsertFibbingItVote(ctx context.Context, arg UpsertFibbingItVoteParams) error {
	_, err := q.db.Exec(ctx, upsertFibbingItVote,
		arg.ID,
		arg.PlayerID,
		arg.VotedForPlayerID,
		arg.RoundID,
	)
	return err
}
