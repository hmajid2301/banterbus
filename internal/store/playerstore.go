package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"gitlab.com/hmajid2301/banterbus/internal/entities"
	sqlc "gitlab.com/hmajid2301/banterbus/internal/store/db"
)

func (s Store) UpdateAvatar(
	ctx context.Context,
	avatar []byte,
	playerID string,
) (players []sqlc.GetAllPlayersInRoomRow, err error) {
	tx, err := s.db.Begin()
	if err != nil {
		return players, err
	}

	defer func() {
		if err != nil {
			err = errors.Join(err, tx.Rollback())
		}
	}()

	room, err := s.queries.WithTx(tx).GetRoomByPlayerID(ctx, playerID)
	if err != nil {
		return players, err
	}

	if room.RoomState != CREATED.String() {
		return players, fmt.Errorf("room is not in CREATED state")
	}

	_, err = s.queries.WithTx(tx).UpdateAvatar(ctx, sqlc.UpdateAvatarParams{
		Avatar: avatar,
		ID:     playerID,
	})
	if err != nil {
		return players, err
	}

	players, err = s.queries.WithTx(tx).GetAllPlayersInRoom(ctx, playerID)
	if err != nil {
		return players, err
	}

	return players, tx.Commit()
}

func (s Store) UpdateNickname(
	ctx context.Context,
	nickname string,
	playerID string,
) (players []sqlc.GetAllPlayersInRoomRow, err error) {
	tx, err := s.db.Begin()
	if err != nil {
		return players, err
	}

	defer func() {
		if err != nil {
			err = errors.Join(err, tx.Rollback())
		}
	}()

	room, err := s.queries.WithTx(tx).GetRoomByPlayerID(ctx, playerID)
	if err != nil {
		return players, err
	}

	if room.RoomState != CREATED.String() {
		return players, fmt.Errorf("room is not in CREATED state")
	}

	playersInRoom, err := s.queries.WithTx(tx).GetAllPlayersInRoom(ctx, playerID)
	if err != nil {
		return players, err
	}

	for _, p := range playersInRoom {
		if p.Nickname == nickname {
			return players, entities.ErrNicknameExists
		}
	}

	_, err = s.queries.WithTx(tx).UpdateNickname(ctx, sqlc.UpdateNicknameParams{
		Nickname: nickname,
		ID:       playerID,
	})
	if err != nil {
		return players, err
	}

	players, err = s.queries.WithTx(tx).GetAllPlayersInRoom(ctx, playerID)
	if err != nil {
		return players, err
	}

	return players, tx.Commit()
}

func (s Store) ToggleIsReady(
	ctx context.Context,
	playerID string,
) (players []sqlc.GetAllPlayersInRoomRow, err error) {
	tx, err := s.db.Begin()
	if err != nil {
		return players, err
	}

	defer func() {
		if err != nil {
			err = errors.Join(err, tx.Rollback())
		}
	}()

	room, err := s.queries.WithTx(tx).GetRoomByPlayerID(ctx, playerID)
	if err != nil {
		return players, err
	}

	if room.RoomState != CREATED.String() {
		return players, fmt.Errorf("room is not in CREATED state")
	}

	player, err := s.queries.WithTx(tx).GetPlayerByID(ctx, playerID)
	if err != nil {
		return players, err
	}

	_, err = s.queries.WithTx(tx).UpdateIsReady(ctx, sqlc.UpdateIsReadyParams{
		ID:      playerID,
		IsReady: sql.NullBool{Bool: !player.IsReady.Bool, Valid: true},
	})
	if err != nil {
		return players, err
	}

	players, err = s.queries.WithTx(tx).GetAllPlayersInRoom(ctx, playerID)
	if err != nil {
		return players, err
	}

	return players, tx.Commit()
}

func (s Store) GetRoomState(ctx context.Context, playerID string) (RoomState, error) {
	room, err := s.queries.GetRoomByPlayerID(ctx, playerID)
	if err != nil {
		return CREATED, err
	}

	var roomState RoomState
	switch room.RoomState {
	case "CREATED":
		roomState = CREATED
	case "PLAYING":
		roomState = PLAYING
	case "PAUSED":
		roomState = PAUSED
	case "FINISHED":
		roomState = FINISHED
	case "ABANDONED":
		roomState = ABANDONED
	default:
		return CREATED, fmt.Errorf("unknown room state: %s", room.RoomState)
	}

	return roomState, nil
}

func (s Store) GetLobbyByPlayerID(
	ctx context.Context,
	playerID string,
) (players []sqlc.GetAllPlayersInRoomRow, err error) {
	players, err = s.queries.GetAllPlayersInRoom(ctx, playerID)
	if err != nil {
		return players, err
	}

	if len(players) == 0 {
		return players, fmt.Errorf("no players found in lobby")
	}

	return players, nil
}

// TODO: rename this to get questions
func (s Store) GetGameStateByPlayerID(ctx context.Context, playerID string) (gameState entities.GameState, err error) {
	g, err := s.queries.GetGameStateByPlayerID(ctx, playerID)
	if err != nil {
		return gameState, err
	}

	question := g.NormalQuestion.String
	if g.FibberQuestion.Valid {
		question = g.FibberQuestion.String
	}

	players := []entities.PlayerWithRole{
		{
			ID:       g.ID,
			Nickname: g.Nickname,
			Role:     g.PlayerRole.String,
			Avatar:   g.Avatar,
			Question: question,
		},
	}
	gameState = entities.GameState{
		Players:   players,
		Round:     int(g.Round),
		RoundType: g.RoundType,
		RoomCode:  g.RoomCode,
	}

	return gameState, err
}
