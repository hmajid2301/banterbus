package randomizer

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserRandomizer(t *testing.T) {
	t.Parallel()

	randomizer := NewUserRandomizer()

	t.Run("Should create new user randomizer", func(t *testing.T) {
		t.Parallel()

		assert.IsType(t, UserRandomizer{}, randomizer)
	})

	t.Run("Should generate valid nickname", func(t *testing.T) {
		t.Parallel()

		nickname := randomizer.GetNickname()

		assert.NotEmpty(t, nickname)
		assert.True(t, len(nickname) > 5)

		// Should contain color, adjective, and animal
		hasColor := false
		for _, color := range colors {
			if contains(nickname, color) {
				hasColor = true
				break
			}
		}
		assert.True(t, hasColor, "Nickname should contain a color")

		hasAdjective := false
		for _, adj := range adjectives {
			if contains(nickname, adj) {
				hasAdjective = true
				break
			}
		}
		assert.True(t, hasAdjective, "Nickname should contain an adjective")

		hasAnimal := false
		for _, animal := range animals {
			if contains(nickname, animal) {
				hasAnimal = true
				break
			}
		}
		assert.True(t, hasAnimal, "Nickname should contain an animal")
	})

	t.Run("Should generate different nicknames", func(t *testing.T) {
		t.Parallel()

		nicknames := make(map[string]bool)
		duplicateCount := 0

		// Generate 100 nicknames and check for uniqueness
		for i := 0; i < 100; i++ {
			nickname := randomizer.GetNickname()
			if nicknames[nickname] {
				duplicateCount++
			}
			nicknames[nickname] = true
		}

		// Allow for some duplicates due to randomness but they should be rare
		assert.Less(t, duplicateCount, 10, "Too many duplicate nicknames generated")
	})

	t.Run("Should generate avatar with provided nickname", func(t *testing.T) {
		t.Parallel()

		nickname := "TestNickname"
		avatar := randomizer.GetAvatar(nickname)

		assert.Contains(t, avatar, "https://api.dicebear.com/9.x/bottts-neutral/svg")
		assert.Contains(t, avatar, "seed=TestNickname")
	})

	t.Run("Should generate avatar with random nickname when empty", func(t *testing.T) {
		t.Parallel()

		avatar := randomizer.GetAvatar("")

		assert.Contains(t, avatar, "https://api.dicebear.com/9.x/bottts-neutral/svg")
		assert.Contains(t, avatar, "seed=")
	})

	t.Run("Should generate valid room code", func(t *testing.T) {
		t.Parallel()

		code := randomizer.GetRoomCode()

		assert.Len(t, code, 5)
		match, err := regexp.MatchString("^[A-Z0-9]{5}$", code)
		assert.NoError(t, err)
		assert.True(t, match, "Room code should only contain uppercase letters and numbers")
	})

	t.Run("Should generate different room codes", func(t *testing.T) {
		t.Parallel()

		codes := make(map[string]bool)
		duplicateCount := 0

		// Generate 1000 codes and check for uniqueness
		for i := 0; i < 1000; i++ {
			code := randomizer.GetRoomCode()
			if codes[code] {
				duplicateCount++
			}
			codes[code] = true
		}

		// With 36^5 possible codes, duplicates should be very rare
		assert.Less(t, duplicateCount, 5, "Too many duplicate room codes generated")
	})

	t.Run("Should generate valid UUID", func(t *testing.T) {
		t.Parallel()

		id, err := randomizer.GetID()

		assert.NoError(t, err)
		assert.NotNil(t, id)
		assert.NotEmpty(t, id.String())
	})

	t.Run("Should generate different UUIDs", func(t *testing.T) {
		t.Parallel()

		ids := make(map[string]bool)

		for i := 0; i < 100; i++ {
			id, err := randomizer.GetID()
			assert.NoError(t, err)
			assert.False(t, ids[id.String()], "UUID should be unique")
			ids[id.String()] = true
		}
	})

	t.Run("Should generate valid fibber index", func(t *testing.T) {
		t.Parallel()

		tests := []int{1, 2, 5, 10, 100}

		for _, playerCount := range tests {
			index := randomizer.GetFibberIndex(playerCount)
			assert.GreaterOrEqual(t, index, 0)
			assert.Less(t, index, playerCount)
		}
	})

	t.Run("Should distribute fibber indices randomly", func(t *testing.T) {
		t.Parallel()

		playerCount := 4
		indices := make(map[int]int)

		// Generate many indices and check distribution
		for i := 0; i < 1000; i++ {
			index := randomizer.GetFibberIndex(playerCount)
			indices[index]++
		}

		// Each index should appear roughly 25% of the time
		for i := 0; i < playerCount; i++ {
			count := indices[i]
			assert.Greater(t, count, 200, "Index %d should appear frequently", i)
			assert.Less(t, count, 300, "Index %d should not appear too frequently", i)
		}
	})
}

// Helper function to check if string contains substring
func contains(str, substr string) bool {
	return len(str) >= len(substr) && (str == substr || len(substr) == 0 ||
		(len(str) > len(substr) && (str[:len(substr)] == substr || str[len(str)-len(substr):] == substr ||
			containsInMiddle(str, substr))))
}

func containsInMiddle(str, substr string) bool {
	for i := 1; i < len(str)-len(substr)+1; i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
