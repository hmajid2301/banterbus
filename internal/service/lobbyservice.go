package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	sqlc "gitlab.com/hmajid2301/banterbus/internal/store/db"
)

type LobbyService struct {
	store      Storer
	randomizer Randomizer
}

type Randomizer interface {
	GetNickname() string
	GetAvatar() []byte
	GetRoomCode() string
	GetID() string
	GetFibberIndex(playersLen int) int
}

var ErrNicknameExists = errors.New("nickname already exists in room")

func NewLobbyService(store Storer, randomizer Randomizer) *LobbyService {
	return &LobbyService{store: store, randomizer: randomizer}
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

		if room.RoomState == sqlc.ROOMSTATE_FINISHED.String() || room.RoomState == sqlc.ROOMSTATE_ABANDONED.String() {
			break
		}
	}

	addPlayer := sqlc.AddPlayerParams{
		ID:       player.ID,
		Avatar:   player.Avatar,
		Nickname: player.Nickname,
	}

	roomID := r.randomizer.GetID()
	addRoom := sqlc.AddRoomParams{
		ID:         roomID,
		GameName:   gameName,
		RoomCode:   roomCode,
		RoomState:  sqlc.ROOMSTATE_CREATED.String(),
		HostPlayer: player.ID,
	}

	addRoomPlayer := sqlc.AddRoomPlayerParams{
		RoomID:   addRoom.ID,
		PlayerID: player.ID,
	}

	createRoom := sqlc.CreateRoomParams{
		Room:       addRoom,
		Player:     addPlayer,
		RoomPlayer: addRoomPlayer,
	}

	err := r.store.CreateRoom(ctx, createRoom)
	if err != nil {
		return Lobby{}, err
	}

	lobby := Lobby{
		Code: roomCode,
		Players: []LobbyPlayer{
			{
				ID:       player.ID,
				Nickname: player.Nickname,
				Avatar:   string(player.Avatar),
				IsReady:  false,
				IsHost:   true,
			},
		},
	}
	return lobby, nil
}

func (r *LobbyService) Join(ctx context.Context, roomCode string, playerID string, nickname string) (Lobby, error) {
	newPlayer := r.getNewPlayer(nickname, playerID)
	room, err := r.store.GetRoomByCode(ctx, roomCode)
	if err != nil {
		return Lobby{}, err
	}

	if room.RoomState != sqlc.ROOMSTATE_CREATED.String() {
		return Lobby{}, fmt.Errorf("room is not in CREATED state")
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

	addPlayer := sqlc.AddPlayerParams{
		ID:       newPlayer.ID,
		Avatar:   newPlayer.Avatar,
		Nickname: newPlayer.Nickname,
	}

	addRoomPlayer := sqlc.AddRoomPlayerParams{
		RoomID:   room.ID,
		PlayerID: newPlayer.ID,
	}

	addPlayerToRoom := sqlc.AddPlayerToRoomArgs{
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
	playerID string,
	playerNicknameToKick string,
) (Lobby, string, error) {
	room, err := r.store.GetRoomByCode(ctx, roomCode)
	if err != nil {
		return Lobby{}, "", err
	}

	if room.HostPlayer != playerID {
		return Lobby{}, "", fmt.Errorf("player is not the host of the room")
	}

	if room.RoomState != sqlc.ROOMSTATE_CREATED.String() {
		return Lobby{}, "", fmt.Errorf("room is not in CREATED state")
	}

	playersInRoom, err := r.store.GetAllPlayersInRoom(ctx, playerID)
	if err != nil {
		return Lobby{}, "", err
	}

	var playerToKickID string
	var removeIndex int
	for i, p := range playersInRoom {
		if p.Nickname == playerNicknameToKick {
			playerToKickID = p.ID
			removeIndex = i
			break
		}
	}

	if playerToKickID == "" {
		return Lobby{}, "", fmt.Errorf("player with nickname %s not found to kick", playerNicknameToKick)
	}

	playersInRoom = append(playersInRoom[:removeIndex], playersInRoom[removeIndex+1:]...)

	_, err = r.store.RemovePlayerFromRoom(ctx, playerToKickID)
	if err != nil {
		return Lobby{}, "", err
	}

	lobby := getLobbyPlayers(playersInRoom, roomCode)
	return lobby, playerToKickID, nil
}

func (r *LobbyService) Start(
	ctx context.Context,
	roomCode string,
	playerID string,
	deadline time.Time,
) (QuestionState, error) {
	room, err := r.store.GetRoomByCode(ctx, roomCode)
	if err != nil {
		return QuestionState{}, err
	}

	if room.HostPlayer != playerID {
		return QuestionState{}, fmt.Errorf("player is not the host of the room")
	}

	if room.RoomState != sqlc.ROOMSTATE_CREATED.String() {
		return QuestionState{}, fmt.Errorf("room is not in CREATED state")
	}

	playersInRoom, err := r.store.GetAllPlayersInRoom(ctx, playerID)
	if err != nil {
		return QuestionState{}, err
	}

	minimumPlayers := 2
	if len(playersInRoom) < minimumPlayers {
		return QuestionState{}, fmt.Errorf("not enough players to start the game")
	}

	for _, player := range playersInRoom {
		if !player.IsReady.Bool {
			return QuestionState{}, fmt.Errorf("not all players are ready: %s", player.ID)
		}
	}

	normalsQuestion, err := r.store.GetRandomQuestionByRound(ctx, sqlc.GetRandomQuestionByRoundParams{
		GameName:     room.GameName,
		LanguageCode: "en-GB",
		Round:        "free_form",
	})
	if err != nil {
		return QuestionState{}, err
	}

	fibberQuestion, err := r.store.GetRandomQuestionInGroup(ctx, sqlc.GetRandomQuestionInGroupParams{
		GroupID: normalsQuestion.GroupID,
		ID:      normalsQuestion.ID,
	})
	if err != nil {
		return QuestionState{}, err
	}

	players := []PlayerWithRole{}
	randomFibberLoc := r.randomizer.GetFibberIndex(len(playersInRoom))

	gameStateID := r.randomizer.GetID()
	err = r.store.StartGame(ctx, sqlc.StartGameArgs{
		GameStateID:       gameStateID,
		RoomID:            room.ID,
		NormalsQuestionID: normalsQuestion.ID,
		FibberQuestionID:  fibberQuestion.ID,
		Players:           playersInRoom,
		FibberLoc:         randomFibberLoc,
		Deadline:          deadline,
	})
	if err != nil {
		return QuestionState{}, err
	}

	for i, player := range playersInRoom {
		role := "normal"
		question := normalsQuestion.Question

		if i == randomFibberLoc {
			role = "fibber"
			question = fibberQuestion.Question
		}

		players = append(players, PlayerWithRole{
			ID:       player.ID,
			Nickname: player.Nickname,
			Avatar:   player.Avatar,
			Role:     role,
			Question: question,
		})
	}

	timeLeft := time.Until(deadline)

	gameState := QuestionState{
		GameStateID: gameStateID,
		Players:     players,
		Round:       1,
		RoundType:   "free_form",
		RoomCode:    roomCode,
		Deadline:    timeLeft,
	}
	return gameState, nil
}

func (r *LobbyService) getNewPlayer(playerNickname string, playerID string) NewPlayer {
	nickname := playerNickname
	if playerNickname == "" {
		nickname = r.randomizer.GetNickname()
	}

	avatar := r.randomizer.GetAvatar()
	newPlayer := NewPlayer{
		ID:       playerID,
		Nickname: nickname,
		Avatar:   avatar,
	}
	return newPlayer
}
