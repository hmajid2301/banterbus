package config_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.com/hmajid2301/banterbus/internal/config"
)

func TestLoadConfig(t *testing.T) {
	t.Run("Should load config with default values", func(t *testing.T) {
		ctx := context.Background()
		config, err := config.LoadConfig(ctx)
		state := os.Getenv("XDG_DATA_HOME")
		configPath := filepath.Join(state, "banterbus")
		assert.Equal(t, configPath, config.DBFolder)

		assert.NoError(t, err)
	})

	t.Run("Should load config from environment values", func(t *testing.T) {
		ctx := context.Background()
		os.Setenv("BANTERBUS_DB_FOLDER", "/home/test")
		config, err := config.LoadConfig(ctx)
		fmt.Println(os.Getenv("BANTERBUS_DB_FOLDER"))

		assert.NoError(t, err)
		assert.Equal(t, "/home/test", config.DBFolder)
	})
}
