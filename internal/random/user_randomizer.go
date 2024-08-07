package random

import (
	"fmt"
	"math/rand/v2"

	"github.com/gomig/avatar"
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
}

var animals = []string{
	"Dog",
	"Cat",
	"Lion",
	"Tiger",
	"Bear",
	"Elephant",
}

func (UserRandomizer) GetNickname() string {
	color := colors[rand.IntN(len(colors))]
	adjective := adjectives[rand.IntN(len(adjectives))]
	animal := animals[rand.IntN(len(animals))]
	return fmt.Sprintf("%s%s%s", color, adjective, animal)
}

func (UserRandomizer) GetAvatar() []byte {
	isMale := rand.IntN(2)
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
