package random

import "math/rand/v2"

type RoomRandomizer struct{}

func NewRoomRandomizer() RoomRandomizer {
	return RoomRandomizer{}
}

func (RoomRandomizer) GetRoomCode() string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	codeByte := make([]byte, 5)
	for i := range codeByte {
		codeByte[i] = charset[rand.IntN(len(charset))]
	}
	code := string(codeByte)
	return code
}
