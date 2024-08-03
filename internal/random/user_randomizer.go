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
}

var adjectives = []string{
	"Happy",
	"Sad",
	"Fast",
	"Slow",
	"Big",
	"Small",
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
	isMale := rand.IntN(1)
	av := avatar.NewPersonAvatar(isMale != 0)
	av.RandomizePalette()
	av.RandomizeEye()
	av.RandomizeHairColor()
	av.RandomizeSkinColor()
	av.RandomizeFacialHair()
	av.RandomizeMouth()
	svg := av.SVG()
	return []byte(svg)
}
