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
		return Lobby{}, err
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
		return Lobby{}, playerToKickID, fmt.Errorf("player is not the host of the room")
	}

	if room.RoomState != db.Created.String() {
		return Lobby{}, playerToKickID, fmt.Errorf("room is not in CREATED state")
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
		return Lobby{}, playerToKickID, fmt.Errorf("player with nickname %s not found to kick", playerNicknameToKick)
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
		return QuestionState{}, fmt.Errorf("player is not the host of the room")
	}

	if room.RoomState != db.Created.String() {
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

	normalsQuestions, err := r.store.GetRandomQuestionByRound(ctx, db.GetRandomQuestionByRoundParams{
		GameName:  room.GameName,
		RoundType: "free_form",
	})
	if err != nil {
		return QuestionState{}, err
	}

	fibberQuestions, err := r.store.GetRandomQuestionInGroup(ctx, db.GetRandomQuestionInGroupParams{
		GroupID: normalsQuestions[0].GroupID,
		ID:      normalsQuestions[0].QuestionID,
	})
	if err != nil {
		return QuestionState{}, err
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
	nickname := playerNickname
	if playerNickname == "" {
		nickname = r.randomizer.GetNickname()
	}

	avatar := r.randomizer.GetAvatar(nickname)
	newPlayer := NewPlayer{
		ID:       playerID,
		Nickname: nickname,
		Avatar:   avatar,
	}
	return newPlayer
}
