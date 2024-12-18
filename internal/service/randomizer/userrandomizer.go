package randomizer

import (
	"fmt"
	"math/rand/v2"

	"github.com/google/uuid"
)

type UserRandomizer struct{}

func NewUserRandomizer() UserRandomizer {
	return UserRandomizer{}
}

var colors = []string{
	"Red",
	"Blue",
	"Green",
	"Yellow",
	"Purple",
	"Tangerine",
	"Mandarin",
	"Pink",
	"Turquoise",
	"Magenta",
	"Crimson",
	"Amber",
	"Violet",
	"Indigo",
}

var adjectives = []string{
	"Happy",
	"Sad",
	"Fast",
	"Slow",
	"Big",
	"Small",
	"Joyous",
	"Smelly",
	"Brave",
	"Curious",
	"Fierce",
	"Lazy",
	"Playful",
	"Grumpy",
	"Cheerful",
	"Sleepy",
	"Excited",
	"Angry",
	"Friendly",
	"Mysterious",
}

var animals = []string{
	"Dog",
	"Cat",
	"Lion",
	"Tiger",
	"Bear",
	"Elephant",
	"Giraffe",
	"Zebra",
	"Kangaroo",
	"Panda",
	"Koala",
	"Penguin",
	"Monkey",
	"Rabbit",
	"Fox",
	"Wolf",
	"Deer",
	"Otter",
	"Raccoon",
	"Leopard",
}

func (UserRandomizer) GetNickname() string {
	color := colors[rand.IntN(len(colors))]
	adjective := adjectives[rand.IntN(len(adjectives))]
	animal := animals[rand.IntN(len(animals))]
	return fmt.Sprintf("%s%s%s", color, adjective, animal)
}

func (u UserRandomizer) GetAvatar(nickname string) string {
	if nickname == "" {
		nickname = u.GetNickname()
	}

	return fmt.Sprintf("https://api.dicebear.com/9.x/bottts-neutral/svg?radius=20&seed=%s", nickname)
}

func (UserRandomizer) GetRoomCode() string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	codeLength := 5
	codeByte := make([]byte, codeLength)
	for i := range codeByte {
		codeByte[i] = charset[rand.IntN(len(charset))]
	}
	code := string(codeByte)
	return code
}

func (UserRandomizer) GetID() uuid.UUID {
	return uuid.Must(uuid.NewV7())
}

func (UserRandomizer) GetFibberIndex(playersLen int) int {
	return rand.IntN(playersLen)
}
