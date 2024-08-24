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

type Player struct {
	ID       string
	Nickname string
	Avatar   string
	IsReady  bool
}

type Room struct {
	Code    string
	Players []Player
}
