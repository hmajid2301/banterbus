package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/invopop/ctxi18n"
	"github.com/jackc/pgx/v5/pgtype"

	"gitlab.com/banterbus/banterbus/internal/store/db"
)

type LobbyService struct {
	store         Storer
	randomizer    Randomizer
	defaultLocale string
}

var ErrNicknameExists = errors.New("nickname already exists in room")
var ErrPlayerAlreadyInRoom = errors.New("player is already in the room")
var ErrPlayerNotInGame = errors.New("player is not currently in any game")

func NewLobbyService(store Storer, randomizer Randomizer, defaultLocale string) *LobbyService {
	return &LobbyService{store: store, randomizer: randomizer, defaultLocale: defaultLocale}
}

func (r *LobbyService) Create(
	ctx context.Context,
	gameName string,
	newHostPlayer NewHostPlayer,
) (LobbyCreationResult, error) {
	var newPlayerID uuid.UUID
	var err error

	if newHostPlayer.ID != uuid.Nil {
		newPlayerID = newHostPlayer.ID
	} else {
		newPlayerID, err = r.randomizer.GetID()
		if err != nil {
			return LobbyCreationResult{}, fmt.Errorf("failed to generate new player ID: %w", err)
		}
	}
	player := r.getNewPlayer(newHostPlayer.Nickname, newPlayerID)

	var roomCode string
	for {
		roomCode = r.randomizer.GetRoomCode()
		room, err := r.store.GetRoomByCode(ctx, roomCode)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				break
			}

			return LobbyCreationResult{}, err
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

	roomID, err := r.randomizer.GetID()
	if err != nil {
		return LobbyCreationResult{}, err
	}
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

	err = r.store.CreateRoom(ctx, createRoom)
	if err != nil {
		return LobbyCreationResult{}, fmt.Errorf("failed to create room: %w", err)
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
	return LobbyCreationResult{
		Lobby:       lobby,
		NewPlayerID: newPlayerID,
	}, nil
}

func (r *LobbyService) Join(
	ctx context.Context,
	roomCode string,
	playerID uuid.UUID,
	nickname string,
) (LobbyJoinResult, error) {
	var newPlayerID uuid.UUID
	var err error

	if playerID != uuid.Nil {
		newPlayerID = playerID
	} else {
		newPlayerID, err = r.randomizer.GetID()
		if err != nil {
			return LobbyJoinResult{}, fmt.Errorf("failed to generate new player ID: %w", err)
		}
	}

	// Check if player is already in a room
	existingRoom, err := r.store.GetRoomByPlayerID(ctx, newPlayerID)
	if err == nil && existingRoom.RoomCode != "" {
		return LobbyJoinResult{}, ErrPlayerAlreadyInRoom
	}

	newPlayer := r.getNewPlayer(nickname, newPlayerID)
	room, err := r.store.GetRoomByCode(ctx, roomCode)
	if err != nil {
		return LobbyJoinResult{}, err
	}

	if room.RoomState != db.Created.String() {
		return LobbyJoinResult{}, errors.New("room is not in CREATED state")
	}

	playersInRoom, err := r.store.GetAllPlayerByRoomCode(ctx, roomCode)
	if err != nil {
		return LobbyJoinResult{}, err
	}

	for _, p := range playersInRoom {
		if p.Nickname == nickname {
			return LobbyJoinResult{}, ErrNicknameExists
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
		return LobbyJoinResult{}, err
	}

	// TODO: could use information above to work out players in room
	players, err := r.store.GetAllPlayersInRoom(ctx, newPlayer.ID)
	if err != nil {
		return LobbyJoinResult{}, err
	}

	lobby := getLobbyPlayers(players, roomCode)
	return LobbyJoinResult{
		Lobby:       lobby,
		NewPlayerID: newPlayerID,
	}, nil
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
		return Lobby{}, playerToKickID, errors.New("player is not the host of the room")
	}

	if room.RoomState != db.Created.String() {
		return Lobby{}, playerToKickID, errors.New("room is not in CREATED state")
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
		return Lobby{}, playerToKickID, errors.New(
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
	var gameStateID uuid.UUID
	room, err := r.store.GetRoomByCode(ctx, roomCode)
	if err != nil {
		return QuestionState{}, err
	}

	if room.HostPlayer != playerID {
		return QuestionState{}, errors.New("player is not the host of the room")
	}

	if room.RoomState != db.Created.String() {
		return QuestionState{}, errors.New("room is not in CREATED state")
	}

	playersInRoom, err := r.store.GetAllPlayersInRoom(ctx, playerID)
	if err != nil {
		return QuestionState{}, err
	}

	minimumPlayers := 2
	if len(playersInRoom) < minimumPlayers {
		return QuestionState{}, errors.New("not enough players to start the game")
	}

	for _, player := range playersInRoom {
		if !player.IsReady.Bool {
			return QuestionState{}, fmt.Errorf("not all players are ready: %s", player.ID)
		}
	}

	normalsQuestions, fibberQuestions, err := getQuestions(ctx, r.store, room.GameName, RoundTypeFreeForm)
	if err != nil {
		return QuestionState{}, errors.New(err.Error())
	}

	randomFibberLoc := r.randomizer.GetFibberIndex(len(playersInRoom))

	gameStateID, err = r.randomizer.GetID()
	if err != nil {
		return QuestionState{}, err
	}
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
		RoundType:   RoundTypeFreeForm,
		Deadline:    timeLeft,
	}
	return gameState, nil
}

func (r *LobbyService) GetRoomState(ctx context.Context, playerID uuid.UUID) (db.RoomState, error) {
	room, err := r.store.GetRoomByPlayerID(ctx, playerID)
	if err != nil {
		if err.Error() == "sql: no rows in result set" || err.Error() == "no rows in result set" {
			return db.Created, ErrPlayerNotInGame
		}
		return db.Created, err
	}

	roomState, err := db.RoomStateFromString(room.RoomState)
	return roomState, err
}

func (r *LobbyService) GetLobby(ctx context.Context, playerID uuid.UUID) (Lobby, error) {
	players, err := r.store.GetAllPlayersInRoom(ctx, playerID)
	if err != nil {
		if err.Error() == "sql: no rows in result set" || err.Error() == "no rows in result set" {
			return Lobby{}, ErrPlayerNotInGame
		}
		return Lobby{}, err
	}

	if len(players) == 0 {
		return Lobby{}, errors.New("no players found in room")
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
		GroupType:          "",
		GroupID:            normalsQuestions[0].GroupID,
		ExcludedQuestionID: normalsQuestions[0].QuestionID,
		RoundType:          roundType,
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
				GroupType:          "",
				GroupID:            newNormals[0].GroupID,
				ExcludedQuestionID: newNormals[0].QuestionID,
				RoundType:          roundType,
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
