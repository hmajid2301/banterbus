package config

import (
	"context"
	"fmt"
	"log/slog"
	"net"

	"github.com/invopop/ctxi18n/i18n"
	"github.com/sethvargo/go-envconfig"
)

// INFO: we need another struct for actual config values once we've passed the input ones
type Config struct {
	DB     Database
	Server Server
	Redis  Redis
	App    App
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

type App struct {
	Environment   string
	LogLevel      slog.Level
	DefaultLocale i18n.Code
}

type In struct {
	DBUsername    string `env:"BANTERBUS_DB_USERNAME"`
	DBPassword    string `env:"BANTERBUS_DB_PASSWORD"`
	DBHost        string `env:"BANTERBUS_DB_HOST"`
	DBPort        string `env:"BANTERBUS_DB_PORT, default=5432"`
	DBName        string `env:"BANTERBUS_DB_NAME, default=banterbus"`
	Environment   string `env:"BANTERBUS_ENVIRONMENT, default=production"`
	LogLevel      string `env:"BANTERBUS_LOG_LEVEL, default=info"`
	Host          string `env:"BANTERBUS_WEBSERVER_HOST, default=0.0.0.0"`
	Port          int    `env:"BANTERBUS_WEBSERVER_PORT, default=8080"`
	DefaultLocale string `env:"BANTERBUS_DEFAULT_LOCALE, default=en-GB"`
	RedisAddress  string `env:"BANTERBUS_REDIS_ADDRESS"`
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
		App: App{
			Environment:   input.Environment,
			LogLevel:      parseLogLevel(input.LogLevel),
			DefaultLocale: i18n.Code(input.DefaultLocale),
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
