package entities

type CreateRoomPlayer struct {
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

type Player struct {
	ID       string
	Nickname string
	Avatar   string
}

type Room struct {
	Code    string
	Players []Player
}
