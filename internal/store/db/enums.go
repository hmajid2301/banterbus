package db

import "errors"

type RoomState int

const (
	Created RoomState = iota
	Playing
	Paused
	Finished
	Abandoned
)

func (rs RoomState) String() string {
	return [...]string{"CREATED", "PLAYING", "PAUSED", "FINISHED", "ABANDONED"}[rs]
}

func RoomStateFromString(s string) (RoomState, error) {
	stringToRoomState := map[string]RoomState{
		"CREATED":   Created,
		"PLAYING":   Playing,
		"PAUSED":    Paused,
		"FINISHED":  Finished,
		"ABANDONED": Abandoned,
	}

	if rs, ok := stringToRoomState[s]; ok {
		return rs, nil
	}
	return 0, errors.New("invalid RoomState string")
}

type FibbingItGameState int

const (
	FibbingITQuestion FibbingItGameState = iota
	FibbingItVoting
	FibbingItRevealRole
	FibbingItScoring
	FibbingItWinner
)

func GameStateFromString(s string) (FibbingItGameState, error) {
	stringToGameState := map[string]FibbingItGameState{
		"FIBBING_IT_QUESTION":    FibbingITQuestion,
		"FIBBING_IT_VOTING":      FibbingItVoting,
		"FIBBING_IT_REVEAL_ROLE": FibbingItRevealRole,
		"FIBBING_IT_SCORING":     FibbingItScoring,
		"FIBBING_IT_WINNER":      FibbingItWinner,
	}

	if rs, ok := stringToGameState[s]; ok {
		return rs, nil
	}
	return 0, errors.New("invalid FibbingItGameState string")
}

func (gs FibbingItGameState) String() string {
	return [...]string{
		"FIBBING_IT_QUESTION",
		"FIBBING_IT_VOTING",
		"FIBBING_IT_REVEAL_ROLE",
		"FIBBING_IT_SCORING",
		"FIBBING_IT_WINNER",
	}[gs]
}
