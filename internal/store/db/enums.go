package db

// RoomState represents the current state of a room
// ENUM(Created, Playing, Paused, Finished, Abandoned)
type RoomState int

// FibbingItGameState represents the current state of a Fibbing It game
// ENUM(FibbingITQuestion, FibbingItVoting, FibbingItReveal, FibbingItScoring, FibbingItNewRound, FibbingItWinner)
type FibbingItGameState int
