package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.com/hmajid2301/banterbus/internal/config"
)

func TestLoadConfig(t *testing.T) {
	t.Run("Should load config with default values", func(t *testing.T) {
		config, err := config.LoadConfig()
		state := os.Getenv("XDG_DATA_HOME")
		configPath := filepath.Join(state, "banterbus")
		assert.Equal(t, configPath, config.DBFolder)

		assert.NoError(t, err)
	})

	t.Run("Should load config from environment values", func(t *testing.T) {
		os.Setenv("BANTERBUS_DB_FOLDER", "/home/test")
		config, err := config.LoadConfig()

		assert.NoError(t, err)
		assert.Equal(t, "/home/test", config.DBFolder)
	})
}
