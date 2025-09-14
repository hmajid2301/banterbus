package config_test

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/contrib/processors/minsev"

	"gitlab.com/banterbus/banterbus/internal/config"
)

func TestLoadConfig(t *testing.T) {
	t.Parallel()

	t.Run("Should load config with default values", func(t *testing.T) {
		t.Parallel()

		envVars := []string{
			"BANTERBUS_DB_USERNAME", "BANTERBUS_DB_PASSWORD", "BANTERBUS_DB_HOST",
			"BANTERBUS_DB_PORT", "BANTERBUS_DB_NAME", "BANTERBUS_REDIS_ADDRESS",
			"BANTERBUS_RETRIES", "BANTERBUS_BASE_DELAY_IN_MS", "BANTERBUS_ENVIRONMENT",
			"BANTERBUS_LOG_LEVEL", "BANTERBUS_WEBSERVER_HOST", "BANTERBUS_WEBSERVER_PORT",
			"BANTERBUS_DEFAULT_LOCALE", "BANTERBUS_DEFAULT_GAME", "BANTERBUS_MAX_ROUNDS",
			"BANTERBUS_AUTO_RECONNECT", "BANTERBUS_DISABLE_TELEMETRY",
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

		ctx := t.Context()
		actualCfg, err := config.LoadConfig(ctx)
		assert.NoError(t, err)

		expectedCfg := config.Config{
			App: config.App{
				Environment:   "production",
				LogLevel:      minsev.SeverityInfo,
				DefaultLocale: "en-GB",
				DefaultGame:   "fibbing_it",
				MaxRounds:     3,
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
				ShowVotingScreenFor:     time.Second * 60,
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
}
