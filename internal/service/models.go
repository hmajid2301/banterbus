package service

import (
	"time"

	"github.com/google/uuid"
)

type Lobby struct {
	Code    string
	Players []LobbyPlayer
}

type LobbyPlayer struct {
	ID       uuid.UUID
	Nickname string
	Avatar   string
	IsReady  bool
	IsHost   bool
}
type NewHostPlayer struct {
	ID       uuid.UUID
	Nickname string
}

type NewPlayer struct {
	ID       uuid.UUID
	Nickname string
	Avatar   []byte
}

type NewRoom struct {
	GameName string
}

type PlayerWithRole struct {
	ID            uuid.UUID
	Nickname      string
	Role          string
	Question      string
	Avatar        []byte
	IsAnswerReady bool
}

// TODO: could just be a single player
type QuestionState struct {
	GameStateID uuid.UUID
	Players     []PlayerWithRole
	Round       int
	RoundType   string
	RoomCode    string
	Deadline    time.Duration
}

type UpdateVotingState struct {
	GameStateID uuid.UUID
	Players     []PlayerWithRole
	Deadline    time.Time
	Round       int
}

type VotingState struct {
	Players  []PlayerWithVoting
	Question string
	Round    int
	Deadline time.Duration
}

type PlayerWithVoting struct {
	ID       uuid.UUID
	Nickname string
	Avatar   string
	Votes    int
	Answer   string
}
