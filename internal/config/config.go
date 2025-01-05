package config

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"time"

	"github.com/invopop/ctxi18n/i18n"
	"github.com/mdobak/go-xerrors"
	"github.com/sethvargo/go-envconfig"
)

// INFO: we need another struct for actual config values once we've passed the input ones
type Config struct {
	DB      Database
	Server  Server
	Redis   Redis
	App     App
	JWT     JWT
	Timings Timings
	Scoring Scoring
}

type Database struct {
	URI string
}

type Server struct {
	Host string
	Port int
}

type Redis struct {
	Address string
}

type JWT struct {
	JWSURL string
}

type App struct {
	Environment   string
	LogLevel      slog.Level
	DefaultLocale i18n.Code
	AutoReconnect bool
}

type Timings struct {
	ShowQuestionScreenFor   time.Duration
	ShowVotingScreenFor     time.Duration
	AllReadyToNextScreenFor time.Duration
	ShowRevealScreenFor     time.Duration
	ShowScoreScreenFor      time.Duration
}

type Scoring struct {
	GuessFibber        int
	FibberEvadeCapture int
}

type In struct {
	DBUsername   string `env:"BANTERBUS_DB_USERNAME"`
	DBPassword   string `env:"BANTERBUS_DB_PASSWORD"`
	DBHost       string `env:"BANTERBUS_DB_HOST"`
	DBPort       string `env:"BANTERBUS_DB_PORT, default=5432"`
	DBName       string `env:"BANTERBUS_DB_NAME, default=banterbus"`
	RedisAddress string `env:"BANTERBUS_REDIS_ADDRESS"`

	Environment   string `env:"BANTERBUS_ENVIRONMENT, default=production"`
	LogLevel      string `env:"BANTERBUS_LOG_LEVEL, default=info"`
	Host          string `env:"BANTERBUS_WEBSERVER_HOST, default=0.0.0.0"`
	Port          int    `env:"BANTERBUS_WEBSERVER_PORT, default=8080"`
	DefaultLocale string `env:"BANTERBUS_DEFAULT_LOCALE, default=en-GB"`
	AutoReconnect bool   `env:"BANTERBUS_AUTO_RECONNECT, default=false"`

	JWKSURL string `env:"BANTERBUS_JWKS_URL"`

	ShowQuestionScreenFor   time.Duration `env:"SHOW_QUESTION_SCREEN_FOR, default=61s"`
	ShowVotingScreenFor     time.Duration `env:"SHOW_VOTING_SCREEN_FOR, default=31s"`
	AllReadyToNextScreenFor time.Duration `env:"ALL_READY_TO_NEXT_SCREEN_FOR, default=2s"`
	ShowRevealScreenFor     time.Duration `env:"SHOW_REVEAL_SCREEN_FOR, default=16s"`
	ShowScoreScreenFor      time.Duration `env:"SHOW_SCORE_SCREEN_FOR, default=15s"`

	GuessFibber        int `env:"GUESS_FIBBER, default=100"`
	FibberEvadeCapture int `env:"FIBBER_EVADE_CAPTURE, default=150"`
}

func LoadConfig(ctx context.Context) (Config, error) {
	var input In
	if err := envconfig.Process(ctx, &input); err != nil {
		return Config{}, err
	}

	err := validateServerConfig(input)
	if err != nil {
		return Config{}, err
	}

	uri := fmt.Sprintf(
		"postgresql://%s:%s@%s:%s/%s",
		input.DBUsername,
		input.DBPassword,
		input.DBHost,
		input.DBPort,
		input.DBName,
	)

	config := Config{
		DB: Database{
			URI: uri,
		},
		Server: Server{
			Host: input.Host,
			Port: input.Port,
		},
		Redis: Redis{
			Address: input.RedisAddress,
		},
		JWT: JWT{
			input.JWKSURL,
		},
		App: App{
			Environment:   input.Environment,
			LogLevel:      parseLogLevel(input.LogLevel),
			DefaultLocale: i18n.Code(input.DefaultLocale),
			AutoReconnect: input.AutoReconnect,
		},
		Timings: Timings{
			ShowQuestionScreenFor:   input.ShowQuestionScreenFor,
			ShowVotingScreenFor:     input.ShowVotingScreenFor,
			AllReadyToNextScreenFor: input.AllReadyToNextScreenFor,
			ShowRevealScreenFor:     input.ShowRevealScreenFor,
			ShowScoreScreenFor:      input.ShowScoreScreenFor,
		},
		Scoring: Scoring{
			GuessFibber:        input.GuessFibber,
			FibberEvadeCapture: input.FibberEvadeCapture,
		},
	}

	return config, nil
}

func validateServerConfig(cfg In) error {
	if cfg.Port < 1 || cfg.Port > 65535 {
		return xerrors.New("expected port to be between 1 and 65535 but received: %d", cfg.Port)
	}

	hostIP := net.ParseIP(cfg.Host)
	if hostIP == nil {
		return xerrors.New("expected valid IPv4 address but received: %v", hostIP)
	}

	return nil
}

func parseLogLevel(logLevel string) slog.Level {
	switch logLevel {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
