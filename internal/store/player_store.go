package store

import (
	"context"
	"database/sql"
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
			rbErr := tx.Rollback()
			if rbErr != nil {
				err = fmt.Errorf(
					"failed to rollback: %w; while handling this error: %w",
					rbErr,
					err,
				)
			}
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
			rbErr := tx.Rollback()
			if rbErr != nil {
				err = fmt.Errorf(
					"failed to rollback: %w; while handling this error: %w",
					rbErr,
					err,
				)
			}
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
			rbErr := tx.Rollback()
			if rbErr != nil {
				err = fmt.Errorf(
					"failed to rollback: %w; while handling this error: %w",
					rbErr,
					err,
				)
			}
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
