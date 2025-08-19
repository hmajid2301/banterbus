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
	t.Parallel()

	t.Run("Should load config with default values", func(t *testing.T) {
		t.Parallel()

		// Save and clear environment variables that might interfere with defaults
		envVars := []string{
			"BANTERBUS_DB_USERNAME", "BANTERBUS_DB_PASSWORD", "BANTERBUS_DB_HOST",
			"BANTERBUS_DB_PORT", "BANTERBUS_DB_NAME", "BANTERBUS_REDIS_ADDRESS",
			"BANTERBUS_RETRIES", "BANTERBUS_BASE_DELAY_IN_MS", "BANTERBUS_ENVIRONMENT",
			"BANTERBUS_LOG_LEVEL", "BANTERBUS_WEBSERVER_HOST", "BANTERBUS_WEBSERVER_PORT",
			"BANTERBUS_DEFAULT_LOCALE", "BANTERBUS_AUTO_RECONNECT", "BANTERBUS_DISABLE_TELEMETRY",
			"BANTERBUS_JWKS_URL", "BANTERBUS_JWT_ADMIN_GROUP", "SHOW_QUESTION_SCREEN_FOR",
			"SHOW_VOTING_SCREEN_FOR", "ALL_READY_TO_NEXT_SCREEN_FOR", "SHOW_REVEAL_SCREEN_FOR",
			"SHOW_SCORE_SCREEN_FOR", "GUESS_FIBBER", "FIBBER_EVADE_CAPTURE",
		}

		originalValues := make(map[string]string)
		for _, envVar := range envVars {
			originalValues[envVar] = os.Getenv(envVar)
			os.Unsetenv(envVar)
		}

		t.Cleanup(func() {
			for envVar, originalValue := range originalValues {
				if originalValue != "" {
					os.Setenv(envVar, originalValue)
				} else {
					os.Unsetenv(envVar)
				}
			}
		})

		ctx := context.Background()
		actualCfg, err := config.LoadConfig(ctx)
		assert.NoError(t, err)

		expectedCfg := config.Config{
			App: config.App{
				Environment:   "production",
				LogLevel:      slog.LevelInfo,
				DefaultLocale: "en-GB",
				BaseDelay:     100 * time.Millisecond,
				Retries:       3,
			},
			Server: config.Server{
				Host: "0.0.0.0",
				Port: 8080,
			},
			DB: config.Database{
				URI: "postgresql://:@:5432/banterbus",
			},
			JWT: config.JWT{
				JWKSURL:    "",
				AdminGroup: "",
			},
			Timings: config.Timings{
				ShowQuestionScreenFor:   time.Second * 15,
				ShowVotingScreenFor:     time.Second * 120,
				AllReadyToNextScreenFor: time.Second * 2,
				ShowRevealScreenFor:     time.Second * 15,
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
		// WARNING: This test modifies global environment variables (os.Setenv).
		// Running in parallel with other tests that also modify environment variables
		// can lead to race conditions and flaky tests.
		// Consider refactoring to avoid global state or use a test-specific environment.

		// Save original env vars and restore them after the test
		originalDBUsername := os.Getenv("BANTERBUS_DB_USERNAME")
		originalDBPassword := os.Getenv("BANTERBUS_DB_PASSWORD")
		originalDBHost := os.Getenv("BANTERBUS_DB_HOST")
		originalDBName := os.Getenv("BANTERBUS_DB_NAME")
		t.Cleanup(func() {
			os.Setenv("BANTERBUS_DB_USERNAME", originalDBUsername)
			os.Setenv("BANTERBUS_DB_PASSWORD", originalDBPassword)
			os.Setenv("BANTERBUS_DB_HOST", originalDBHost)
			os.Setenv("BANTERBUS_DB_NAME", originalDBName)
		})

		ctx := context.Background()
		os.Setenv("BANTERBUS_DB_USERNAME", "banterbus")
		os.Setenv("BANTERBUS_DB_PASSWORD", "banterbus")
		os.Setenv("BANTERBUS_DB_HOST", "localhost")
		os.Setenv("BANTERBUS_DB_NAME", "banterbus")

		config, err := config.LoadConfig(ctx)
		assert.NoError(t, err)

		expectedURI := "postgresql://banterbus:banterbus@localhost:5432/banterbus"
		assert.Equal(t, expectedURI, config.DB.URI)
	})

	t.Run("Should default to info level logs", func(t *testing.T) {
		// WARNING: This test modifies global environment variables (os.Setenv).
		// Running in parallel with other tests that also modify environment variables
		// can lead to race conditions and flaky tests.
		// Consider refactoring to avoid global state or use a test-specific environment.

		originalLogLevel := os.Getenv("BANTERBUS_LOG_LEVEL")
		t.Cleanup(func() {
			os.Setenv("BANTERBUS_LOG_LEVEL", originalLogLevel)
		})

		ctx := context.Background()
		os.Setenv("BANTERBUS_LOG_LEVEL", "invalid_log")

		config, err := config.LoadConfig(ctx)
		assert.NoError(t, err)

		assert.Equal(t, slog.LevelInfo, config.App.LogLevel)
	})

	t.Run("Should throw error when invalid port", func(t *testing.T) {
		// WARNING: This test modifies global environment variables (os.Setenv).
		// Running in parallel with other tests that also modify environment variables
		// can lead to race conditions and flaky tests.
		// Consider refactoring to avoid global state or use a test-specific environment.

		originalPort := os.Getenv("BANTERBUS_WEBSERVER_PORT")
		t.Cleanup(func() {
			os.Setenv("BANTERBUS_WEBSERVER_PORT", originalPort)
		})

		ctx := context.Background()
		os.Setenv("BANTERBUS_WEBSERVER_PORT", "190000")

		_, err := config.LoadConfig(ctx)
		assert.Error(t, err)
	})

	t.Run("Should throw error when invalid ip", func(t *testing.T) {
		// WARNING: This test modifies global environment variables (os.Setenv).
		// Running in parallel with other tests that also modify environment variables
		// can lead to race conditions and flaky tests.
		// Consider refactoring to avoid global state or use a test-specific environment.

		originalHost := os.Getenv("BANTERBUS_WEBSERVER_HOST")
		t.Cleanup(func() {
			os.Setenv("BANTERBUS_WEBSERVER_HOST", originalHost)
		})

		ctx := context.Background()
		os.Setenv("BANTERBUS_WEBSERVER_HOST", "985646")

		_, err := config.LoadConfig(ctx)
		assert.Error(t, err)
	})
}
