package config_test

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.com/hmajid2301/banterbus/internal/config"
)

func TestLoadConfig(t *testing.T) {
	t.Run("Should load config with default values", func(t *testing.T) {
		ctx := context.Background()
		actualCfg, err := config.LoadConfig(ctx)
		assert.NoError(t, err)

		state := os.Getenv("XDG_DATA_HOME")
		configPath := filepath.Join(state, "banterbus")
		expectedCfg := config.Config{
			DBFolder: configPath,
			App: config.App{
				Environment: "production",
				LogLevel:    slog.LevelInfo,
			},
			Server: config.Server{
				Host: "0.0.0.0",
				Port: 8080,
			},
		}

		assert.Equal(t, expectedCfg, actualCfg)

	})

	t.Run("Should load config from environment values", func(t *testing.T) {
		ctx := context.Background()
		os.Setenv("BANTERBUS_DB_FOLDER", "/home/test")
		config, err := config.LoadConfig(ctx)

		assert.NoError(t, err)
		assert.Equal(t, "/home/test", config.DBFolder)
	})

	t.Run("Should default to info level logs", func(t *testing.T) {
		ctx := context.Background()
		os.Setenv("BANTERBUS_LOG_LEVEL", "invalid_log")

		config, err := config.LoadConfig(ctx)
		assert.NoError(t, err)

		assert.Equal(t, slog.LevelInfo, config.App.LogLevel)
	})

	t.Run("Should throw error when invalid port", func(t *testing.T) {
		ctx := context.Background()
		os.Setenv("BANTERBUS_WEBSERVER_PORT", "190000")

		_, err := config.LoadConfig(ctx)
		assert.Error(t, err)
	})

	t.Run("Should throw error when invalid ip", func(t *testing.T) {
		ctx := context.Background()
		os.Setenv("BANTERBUS_WEBSERVER_HOST", "985646")

		_, err := config.LoadConfig(ctx)
		assert.Error(t, err)
	})
}
