package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/invopop/ctxi18n"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/mdobak/go-xerrors"

	"gitlab.com/hmajid2301/banterbus/internal/store/db"
)

type LobbyService struct {
	store         Storer
	randomizer    Randomizer
	defaultLocale string
}

var ErrNicknameExists = errors.New("nickname already exists in room")
var ErrPlayerAlreadyInRoom = errors.New("player is already in the room")

func NewLobbyService(store Storer, randomizer Randomizer, defaultLocale string) *LobbyService {
	return &LobbyService{store: store, randomizer: randomizer, defaultLocale: defaultLocale}
}

func (r *LobbyService) Create(ctx context.Context, gameName string, newHostPlayer NewHostPlayer) (Lobby, error) {
	player := r.getNewPlayer(newHostPlayer.Nickname, newHostPlayer.ID)

	var roomCode string
	for {
		roomCode = r.randomizer.GetRoomCode()
		room, err := r.store.GetRoomByCode(ctx, roomCode)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				break
			}

			return Lobby{}, err
		}

		if room.RoomState == db.Finished.String() || room.RoomState == db.Abandoned.String() {
			break
		}
	}

	locale := ctxi18n.Locale(ctx).Code().String()
	addPlayer := db.AddPlayerParams{
		ID:       player.ID,
		Avatar:   player.Avatar,
		Nickname: player.Nickname,
		Locale:   pgtype.Text{String: locale},
	}

	roomID := r.randomizer.GetID()
	addRoom := db.AddRoomParams{
		ID:         roomID,
		GameName:   gameName,
		RoomCode:   roomCode,
		RoomState:  db.Created.String(),
		HostPlayer: player.ID,
	}

	addRoomPlayer := db.AddRoomPlayerParams{
		RoomID:   addRoom.ID,
		PlayerID: player.ID,
	}

	createRoom := db.CreateRoomArgs{
		Room:       addRoom,
		Player:     addPlayer,
		RoomPlayer: addRoomPlayer,
	}

	err := r.store.CreateRoom(ctx, createRoom)
	if err != nil {
		return Lobby{}, xerrors.Append(fmt.Errorf("failed to create room"), err)
	}

	lobby := Lobby{
		Code: roomCode,
		Players: []LobbyPlayer{
			{
				ID:       player.ID,
				Nickname: player.Nickname,
				Avatar:   player.Avatar,
				IsReady:  false,
				IsHost:   true,
			},
		},
	}
	return lobby, nil
}

func (r *LobbyService) Join(ctx context.Context, roomCode string, playerID uuid.UUID, nickname string) (Lobby, error) {
	room, err := r.store.GetRoomByPlayerID(ctx, playerID)
	if err == nil {
		if room.RoomCode == roomCode {
			return Lobby{}, ErrPlayerAlreadyInRoom
		}
	}

	newPlayer := r.getNewPlayer(nickname, playerID)
	room, err = r.store.GetRoomByCode(ctx, roomCode)
	if err != nil {
		return Lobby{}, err
	}

	if room.RoomState != db.Created.String() {
		return Lobby{}, xerrors.New("room is not in CREATED state")
	}

	playersInRoom, err := r.store.GetAllPlayerByRoomCode(ctx, roomCode)
	if err != nil {
		return Lobby{}, err
	}

	for _, p := range playersInRoom {
		if p.Nickname == nickname {
			return Lobby{}, ErrNicknameExists
		}
	}

	locale := ctxi18n.Locale(ctx).Code().String()
	addPlayer := db.AddPlayerParams{
		ID:       newPlayer.ID,
		Avatar:   newPlayer.Avatar,
		Nickname: newPlayer.Nickname,
		Locale:   pgtype.Text{String: locale},
	}

	addRoomPlayer := db.AddRoomPlayerParams{
		RoomID:   room.ID,
		PlayerID: newPlayer.ID,
	}

	addPlayerToRoom := db.AddPlayerToRoomArgs{
		Player:     addPlayer,
		RoomPlayer: addRoomPlayer,
	}

	err = r.store.AddPlayerToRoom(ctx, addPlayerToRoom)
	if err != nil {
		return Lobby{}, err
	}

	// TODO: could use information above to work out players in room
	players, err := r.store.GetAllPlayersInRoom(ctx, newPlayer.ID)
	if err != nil {
		return Lobby{}, err
	}

	lobby := getLobbyPlayers(players, roomCode)
	return lobby, nil
}

func (r *LobbyService) KickPlayer(
	ctx context.Context,
	roomCode string,
	playerID uuid.UUID,
	playerNicknameToKick string,
) (Lobby, uuid.UUID, error) {
	var playerToKickID uuid.UUID
	room, err := r.store.GetRoomByCode(ctx, roomCode)
	if err != nil {
		return Lobby{}, playerToKickID, err
	}

	if room.HostPlayer != playerID {
		return Lobby{}, playerToKickID, xerrors.New("player is not the host of the room")
	}

	if room.RoomState != db.Created.String() {
		return Lobby{}, playerToKickID, xerrors.New("room is not in CREATED state")
	}

	playersInRoom, err := r.store.GetAllPlayersInRoom(ctx, playerID)
	if err != nil {
		return Lobby{}, playerToKickID, err
	}

	var removeIndex int
	for i, p := range playersInRoom {
		if p.Nickname == playerNicknameToKick {
			playerToKickID = p.ID
			removeIndex = i
			break
		}
	}

	if playerToKickID == uuid.Nil {
		return Lobby{}, playerToKickID, xerrors.New(
			fmt.Sprintf("player with nickname %s not found", playerNicknameToKick),
		)
	}

	playersInRoom = append(playersInRoom[:removeIndex], playersInRoom[removeIndex+1:]...)

	_, err = r.store.RemovePlayerFromRoom(ctx, playerToKickID)
	if err != nil {
		return Lobby{}, playerToKickID, err
	}

	lobby := getLobbyPlayers(playersInRoom, roomCode)
	return lobby, playerToKickID, nil
}

func (r *LobbyService) Start(
	ctx context.Context,
	roomCode string,
	playerID uuid.UUID,
	deadline time.Time,
) (QuestionState, error) {
	room, err := r.store.GetRoomByCode(ctx, roomCode)
	if err != nil {
		return QuestionState{}, err
	}

	if room.HostPlayer != playerID {
		return QuestionState{}, xerrors.New("player is not the host of the room")
	}

	if room.RoomState != db.Created.String() {
		return QuestionState{}, xerrors.New("room is not in CREATED state")
	}

	playersInRoom, err := r.store.GetAllPlayersInRoom(ctx, playerID)
	if err != nil {
		return QuestionState{}, err
	}

	minimumPlayers := 2
	if len(playersInRoom) < minimumPlayers {
		return QuestionState{}, xerrors.New("not enough players to start the game")
	}

	for _, player := range playersInRoom {
		if !player.IsReady.Bool {
			return QuestionState{}, xerrors.New("not all players are ready: %s", player.ID)
		}
	}

	normalsQuestions, fibberQuestions, err := getQuestions(ctx, r.store, room.GameName, "free_form")
	if err != nil {
		return QuestionState{}, xerrors.New(err.Error())
	}

	randomFibberLoc := r.randomizer.GetFibberIndex(len(playersInRoom))

	gameStateID := r.randomizer.GetID()
	err = r.store.StartGame(ctx, db.StartGameArgs{
		GameStateID:       gameStateID,
		RoomID:            room.ID,
		NormalsQuestionID: normalsQuestions[0].QuestionID,
		FibberQuestionID:  fibberQuestions[0].QuestionID,
		Players:           playersInRoom,
		FibberLoc:         randomFibberLoc,
		Deadline:          deadline,
	})
	if err != nil {
		return QuestionState{}, err
	}

	players := []PlayerWithRole{}
	for i, player := range playersInRoom {
		role := NormalRole

		var question string
		for _, localeQuestion := range normalsQuestions {
			if localeQuestion.Locale == player.Locale.String {
				question = localeQuestion.Question
			} else if question == "" && localeQuestion.Locale == r.defaultLocale {
				question = localeQuestion.Question
			}
		}

		if i == randomFibberLoc {
			question = ""
			role = FibberRole
			for _, localeQuestion := range fibberQuestions {
				if localeQuestion.Locale == player.Locale.String {
					question = localeQuestion.Question
				} else if question == "" && localeQuestion.Locale == r.defaultLocale {
					question = localeQuestion.Question
				}
			}
		}

		players = append(players, PlayerWithRole{
			ID:            player.ID,
			Role:          role,
			Question:      question,
			IsAnswerReady: false,
		})
	}

	timeLeft := time.Until(deadline)

	gameState := QuestionState{
		GameStateID: gameStateID,
		Players:     players,
		Round:       1,
		RoundType:   "free_form",
		Deadline:    timeLeft,
	}
	return gameState, nil
}

func (r *LobbyService) GetRoomState(ctx context.Context, playerID uuid.UUID) (db.RoomState, error) {
	room, err := r.store.GetRoomByPlayerID(ctx, playerID)
	if err != nil {
		return db.Created, err
	}

	roomState, err := db.RoomStateFromString(room.RoomState)
	return roomState, err
}

func (r *LobbyService) GetLobby(ctx context.Context, playerID uuid.UUID) (Lobby, error) {
	players, err := r.store.GetAllPlayersInRoom(ctx, playerID)
	if err != nil {
		return Lobby{}, err
	}

	room := getLobbyPlayers(players, players[0].RoomCode)
	return room, err
}

func (r *LobbyService) getNewPlayer(playerNickname string, playerID uuid.UUID) NewPlayer {
	avatar := r.randomizer.GetAvatar(playerNickname)
	newPlayer := NewPlayer{
		ID:       playerID,
		Nickname: playerNickname,
		Avatar:   avatar,
	}
	return newPlayer
}

// TODO: add unit tests for this
func getQuestions(
	ctx context.Context,
	store Storer,
	gameName string,
	roundType string,
) ([]db.GetRandomQuestionByRoundRow, []db.GetRandomQuestionInGroupRow, error) {
	maxRetries := 3

	normalsQuestions, err := store.GetRandomQuestionByRound(ctx, db.GetRandomQuestionByRoundParams{
		GameName:  gameName,
		RoundType: roundType,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get normal questions: %w", err)
	}

	if len(normalsQuestions) == 0 {
		return nil, nil, fmt.Errorf("no normal questions found")
	}

	fibberQuestions, err := store.GetRandomQuestionInGroup(ctx, db.GetRandomQuestionInGroupParams{
		GroupID:   normalsQuestions[0].GroupID,
		ID:        normalsQuestions[0].QuestionID,
		RoundType: roundType,
	})

	if err == sql.ErrNoRows {
		for i := 0; i < maxRetries; i++ {
			newNormals, err := store.GetRandomQuestionByRound(ctx, db.GetRandomQuestionByRoundParams{
				GameName:  gameName,
				RoundType: roundType,
			})
			if err != nil {
				return nil, nil, fmt.Errorf("retry %d: failed to get normal questions: %w", i+1, err)
			}
			if len(newNormals) == 0 {
				continue
			}

			newFibber, err := store.GetRandomQuestionInGroup(ctx, db.GetRandomQuestionInGroupParams{
				GroupID:   newNormals[0].GroupID,
				ID:        newNormals[0].QuestionID,
				RoundType: roundType,
			})

			if err == nil {
				normalsQuestions = newNormals
				fibberQuestions = newFibber
				break
			} else if err == sql.ErrNoRows {
				continue
			}

			return nil, nil, fmt.Errorf("retry %d: failed to get fibber questions: %w", i+1, err)
		}

		if len(fibberQuestions) == 0 {
			return nil, nil, fmt.Errorf("no fibber questions found after %d retries", maxRetries)
		}
	} else if err != nil {
		return nil, nil, fmt.Errorf("initial fibber question fetch failed: %w", err)
	}

	return normalsQuestions, fibberQuestions, nil
}
