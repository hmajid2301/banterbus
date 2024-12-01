package service

import "time"

type Lobby struct {
	Code    string
	Players []LobbyPlayer
}

type LobbyPlayer struct {
	ID       string
	Nickname string
	Avatar   string
	IsReady  bool
	IsHost   bool
}
type NewHostPlayer struct {
	ID       string
	Nickname string
}

type NewPlayer struct {
	ID       string
	Nickname string
	Avatar   []byte
}

type NewRoom struct {
	GameName string
}

type PlayerWithRole struct {
	ID       string
	Nickname string
	Role     string
	Question string
	Avatar   []byte
}

type QuestionState struct {
	GameStateID string
	Players     []PlayerWithRole
	Round       int
	RoundType   string
	RoomCode    string
}

type UpdateVotingState struct {
	Players     []PlayerWithRole
	GameStateID string
	Deadline    time.Time
	Round       int
}

type VotingState struct {
	Players  []PlayerWithVoting
	Question string
	Round    int
}

type PlayerWithVoting struct {
	ID       string
	Nickname string
	Avatar   string
	Votes    int
	Answer   string
}
