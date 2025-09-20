package config

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/invopop/ctxi18n/i18n"
	"github.com/sethvargo/go-envconfig"
	"go.opentelemetry.io/contrib/processors/minsev"
)

// INFO: we need another struct for actual config values once we've passed the input ones
type Config struct {
	DB        Database
	Server    Server
	Redis     Redis
	App       App
	JWT       JWT
	Timings   Timings
	Scoring   Scoring
	Telemetry Telemetry
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
	JWKSURL    string
	AdminGroup string
}

type App struct {
	Environment   string
	LogLevel      minsev.Severity
	DefaultLocale i18n.Code
	DefaultGame   string
	MaxRounds     int
	AutoReconnect bool
	Retries       int
	BaseDelay     time.Duration
}

type Timings struct {
	ShowQuestionScreenFor   time.Duration
	ShowVotingScreenFor     time.Duration
	AllReadyToNextScreenFor time.Duration
	ShowRevealScreenFor     time.Duration
	ShowScoreScreenFor      time.Duration
	ShowWinnerScreenFor     time.Duration
}

type Scoring struct {
	GuessFibber        int
	FibberEvadeCapture int
}

type Telemetry struct {
	OAuth2ClientID     string
	OAuth2ClientSecret string
	OAuth2TokenURL     string
	OAuth2Issuer       string
	OAuth2Scopes       string
}

type In struct {
	DBUsername string `env:"BANTERBUS_DB_USERNAME"`
	DBPassword string `env:"BANTERBUS_DB_PASSWORD"`
	DBHost     string `env:"BANTERBUS_DB_HOST"`
	DBPort     string `env:"BANTERBUS_DB_PORT, default=5432"`
	DBName     string `env:"BANTERBUS_DB_NAME, default=banterbus"`

	RedisAddress string `env:"BANTERBUS_REDIS_ADDRESS"`

	Retries   int `env:"BANTERBUS_RETRIES, default=3"`
	BaseDelay int `env:"BANTERBUS_BASE_DELAY_IN_MS, default=100"`

	Environment   string `env:"BANTERBUS_ENVIRONMENT, default=production"`
	LogLevel      string `env:"BANTERBUS_LOG_LEVEL, default=info"`
	Host          string `env:"BANTERBUS_WEBSERVER_HOST, default=0.0.0.0"`
	Port          int    `env:"BANTERBUS_WEBSERVER_PORT, default=8080"`
	DefaultLocale string `env:"BANTERBUS_DEFAULT_LOCALE, default=en-GB"`
	DefaultGame   string `env:"BANTERBUS_DEFAULT_GAME, default=fibbing_it"`
	MaxRounds     int    `env:"BANTERBUS_MAX_ROUNDS, default=3"`
	AutoReconnect bool   `env:"BANTERBUS_AUTO_RECONNECT, default=false"`

	JWKSURL    string `env:"BANTERBUS_JWKS_URL"`
	AdminGroup string `env:"BANTERBUS_JWT_ADMIN_GROUP"`

	ShowQuestionScreenFor   time.Duration `env:"SHOW_QUESTION_SCREEN_FOR, default=15s"`
	ShowVotingScreenFor     time.Duration `env:"SHOW_VOTING_SCREEN_FOR, default=60s"`
	AllReadyToNextScreenFor time.Duration `env:"ALL_READY_TO_NEXT_SCREEN_FOR, default=2s"`
	ShowRevealScreenFor     time.Duration `env:"SHOW_REVEAL_SCREEN_FOR, default=15s"`
	ShowScoreScreenFor      time.Duration `env:"SHOW_SCORE_SCREEN_FOR, default=15s"`
	ShowWinnerScoreFor      time.Duration `env:"SHOW_SCORE_SCREEN_FOR, default=15s"`

	GuessFibber        int `env:"GUESS_FIBBER, default=100"`
	FibberEvadeCapture int `env:"FIBBER_EVADE_CAPTURE, default=150"`

	// OAuth2 configuration for OTEL exporters
	OTELOAuth2ClientID     string `env:"OTEL_OAUTH2_CLIENT_ID"`
	OTELOAuth2ClientSecret string `env:"OTEL_OAUTH2_CLIENT_SECRET"`
	OTELOAuth2TokenURL     string `env:"OTEL_OAUTH2_TOKEN_URL"`
	OTELOAuth2Issuer       string `env:"OTEL_OAUTH2_ISSUER"`
	OTELOAuth2Scopes       string `env:"OTEL_OAUTH2_SCOPES"`
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

	// Create a proper URL with encoded userinfo to handle special characters from Bao
	userinfo := url.UserPassword(input.DBUsername, input.DBPassword)

	uri := fmt.Sprintf(
		"postgresql://%s@%s:%s/%s",
		userinfo.String(),
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
			JWKSURL:    input.JWKSURL,
			AdminGroup: input.AdminGroup,
		},
		App: App{
			Environment:   input.Environment,
			LogLevel:      parseLogLevel(input.LogLevel),
			DefaultLocale: i18n.Code(input.DefaultLocale),
			DefaultGame:   input.DefaultGame,
			MaxRounds:     input.MaxRounds,
			AutoReconnect: input.AutoReconnect,
			BaseDelay:     time.Millisecond * time.Duration(input.BaseDelay),
			Retries:       input.Retries,
		},
		Timings: Timings{
			ShowQuestionScreenFor:   input.ShowQuestionScreenFor,
			ShowVotingScreenFor:     input.ShowVotingScreenFor,
			AllReadyToNextScreenFor: input.AllReadyToNextScreenFor,
			ShowRevealScreenFor:     input.ShowRevealScreenFor,
			ShowScoreScreenFor:      input.ShowScoreScreenFor,
			ShowWinnerScreenFor:     input.ShowWinnerScoreFor,
		},
		Scoring: Scoring{
			GuessFibber:        input.GuessFibber,
			FibberEvadeCapture: input.FibberEvadeCapture,
		},
		Telemetry: Telemetry{
			OAuth2ClientID:     input.OTELOAuth2ClientID,
			OAuth2ClientSecret: input.OTELOAuth2ClientSecret,
			OAuth2TokenURL:     input.OTELOAuth2TokenURL,
			OAuth2Issuer:       input.OTELOAuth2Issuer,
			OAuth2Scopes:       input.OTELOAuth2Scopes,
		},
	}

	return config, nil
}

func validateServerConfig(cfg In) error {
	if cfg.Port < 1 || cfg.Port > 65535 {
		return fmt.Errorf("expected port to be between 1 and 65535 but received: %d", cfg.Port)
	}

	hostIP := net.ParseIP(cfg.Host)
	if hostIP == nil {
		return fmt.Errorf("expected valid IPv4 address but received: %v", hostIP)
	}

	return nil
}

func parseLogLevel(logLevel string) minsev.Severity {
	switch strings.ToLower(logLevel) {
	case "debug":
		return minsev.SeverityDebug
	case "info":
		return minsev.SeverityInfo
	case "warn":
		return minsev.SeverityWarn
	case "error":
		return minsev.SeverityError
	default:
		return minsev.SeverityInfo
	}
}
