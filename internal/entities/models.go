package entities

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

type LobbyPlayer struct {
	ID       string
	Nickname string
	Avatar   string
	IsReady  bool
	IsHost   bool
}

type Lobby struct {
	Code    string
	Players []LobbyPlayer
}

type PlayerWithRole struct {
	ID       string
	Nickname string
	Role     string
	Question string
	Avatar   []byte
}

type GameState struct {
	Players   []PlayerWithRole
	Round     int
	RoundType string
	RoomCode  string
}
