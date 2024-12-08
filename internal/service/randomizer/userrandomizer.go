package randomizer

import (
	"fmt"
	"math/rand/v2"

	"github.com/gomig/avatar"
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

func (UserRandomizer) GetAvatar() []byte {
	booleanRandNum := 2
	isMale := rand.IntN(booleanRandNum)
	avatar := avatar.NewPersonAvatar(isMale == 1)

	avatar.RandomizeHair()
	avatar.RandomizeSticker()
	avatar.RandomizeDress()
	avatar.RandomizePalette()
	avatar.RandomizeEye()
	avatar.RandomizeHairColor()
	avatar.RandomizeSkinColor()
	avatar.RandomizeFacialHair()
	avatar.RandomizeMouth()

	encodedAvatar := avatar.Base64()
	return []byte(encodedAvatar)
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
