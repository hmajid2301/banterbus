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
	Avatar   string
}

type NewRoom struct {
	GameName string
}

type PlayerWithRole struct {
	ID              uuid.UUID
	Role            string
	Question        string
	IsAnswerReady   bool
	PossibleAnswers []string
	CurrentAnswer   string
}

// TODO: could just be a single player
type QuestionState struct {
	GameStateID uuid.UUID
	Players     []PlayerWithRole
	Round       int
	RoundType   string
	Deadline    time.Duration
}

type UpdateVotingState struct {
	GameStateID uuid.UUID
	Players     []PlayerWithRole
	Deadline    time.Time
	Round       int
}

type VotingState struct {
	Players     []PlayerWithVoting
	Question    string
	Round       int
	GameStateID uuid.UUID
	Deadline    time.Duration
}

type PlayerWithVoting struct {
	ID       uuid.UUID
	Nickname string
	Avatar   string
	Votes    int
	Answer   string
	IsReady  bool
	Role     string
}

type RevealRoleState struct {
	VotedForPlayerNickname string
	VotedForPlayerAvatar   string
	VotedForPlayerRole     string
	ShouldReveal           bool
	Deadline               time.Duration
	Round                  int
	RoundType              string
	PlayerIDs              []uuid.UUID
}

type ScoreState struct {
	Players     []PlayerWithScoring
	Deadline    time.Duration
	RoundType   string
	RoundNumber int
}

type PlayerWithScoring struct {
	ID       uuid.UUID
	Nickname string
	Avatar   string
	Score    int
}

type Scoring struct {
	GuessedFibber      int
	FibberEvadeCapture int
}

type Question struct {
	ID        string
	Text      string
	GroupName string
	Locale    string
	RoundType string
	Enabled   bool
}

type Group struct {
	ID   string
	Name string
}

type QuestionTranslation struct {
	Text   string
	Locale string
}

type WinnerState struct {
	Players []PlayerWithScoring
}
