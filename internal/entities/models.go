package entities

type NewPlayer struct {
	Nickname  string
	Avatar    []byte
	SessionID int64
}

type NewRoom struct {
	GameName string
	RoomCode string
}
