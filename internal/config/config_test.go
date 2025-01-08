package config_test

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"gitlab.com/hmajid2301/banterbus/internal/config"
)

func TestLoadConfig(t *testing.T) {
	t.Run("Should load config with default values", func(t *testing.T) {
		ctx := context.Background()
		actualCfg, err := config.LoadConfig(ctx)
		assert.NoError(t, err)

		expectedCfg := config.Config{
			App: config.App{
				Environment:   "production",
				LogLevel:      slog.LevelInfo,
				DefaultLocale: "en-GB",
			},
			Server: config.Server{
				Host: "0.0.0.0",
				Port: 8080,
			},
			DB: config.Database{
				URI: "postgresql://:@:5432/banterbus",
			},
			JWT: config.JWT{
				JWKSURL: "",
			},
			Timings: config.Timings{
				ShowQuestionScreenFor:   time.Second * 61,
				ShowVotingScreenFor:     time.Second * 31,
				AllReadyToNextScreenFor: time.Second * 2,
				ShowRevealScreenFor:     time.Second * 16,
				ShowScoreScreenFor:      time.Second * 15,
				ShowWinnerScreenFor:     time.Second * 15,
			},
			Scoring: config.Scoring{
				GuessFibber:        100,
				FibberEvadeCapture: 150,
			},
		}

		assert.Equal(t, expectedCfg, actualCfg)
	})

	t.Run("Should load config from environment values", func(t *testing.T) {
		ctx := context.Background()
		os.Setenv("BANTERBUS_DB_USERNAME", "banterbus")
		os.Setenv("BANTERBUS_DB_PASSWORD", "banterbus")
		os.Setenv("BANTERBUS_DB_HOST", "localhost")

		config, err := config.LoadConfig(ctx)
		assert.NoError(t, err)

		expectedURI := "postgresql://banterbus:banterbus@localhost:5432/banterbus"
		assert.Equal(t, expectedURI, config.DB.URI)
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
