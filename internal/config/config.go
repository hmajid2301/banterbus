package config

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"

	gap "github.com/muesli/go-app-paths"
	"github.com/sethvargo/go-envconfig"
)

// INFO: we need another struct for actual config values once we've passed the input ones
type Config struct {
	DBFolder string
	Server   Server
	App      App
}

type Server struct {
	Host string
	Port int
}

type App struct {
	Environment string
	LogLevel    slog.Level
}

type ConfigIn struct {
	DBFolder    string `env:"BANTERBUS_DB_FOLDER"`
	Environment string `env:"BANTERBUS_ENVIRONMENT, default=production"`
	LogLevel    string `env:"BANTERBUS_LOG_LEVEL,   default=info"`
	Host        string `env:"BANTERBUS_WEBSERVER_HOST, default=0.0.0.0"`
	Port        int    `env:"BANTERBUS_WEBSERVER_PORT, default=8080"`
}

func LoadConfig(ctx context.Context) (Config, error) {
	var input ConfigIn
	if err := envconfig.Process(ctx, &input); err != nil {
		return Config{}, err
	}

	err := validateServerConfig(input)
	if err != nil {
		return Config{}, err
	}

	config := Config{
		DBFolder: input.DBFolder,
		Server: Server{
			Host: input.Host,
			Port: input.Port,
		},
		App: App{
			Environment: input.Environment,
			LogLevel:    parseLogLevel(input.LogLevel),
		},
	}

	if config.DBFolder == "" {
		scope := gap.NewScope(gap.User, "banterbus")
		dirs, err := scope.DataDirs()
		if err != nil {
			return config, fmt.Errorf("unable to get data directory: %w", err)
		}

		dbFolder, _ := os.UserHomeDir()
		if len(dirs) > 0 {
			dbFolder = dirs[0]
		}
		config.DBFolder = dbFolder
	}

	return config, nil
}

func validateServerConfig(cfg ConfigIn) error {
	if cfg.Port < 1 || cfg.Port > 65535 {
		return fmt.Errorf("expected port to be between 1 and 65535 but received: %d", cfg.Port)
	}

	hostIp := net.ParseIP(cfg.Host)
	if hostIp == nil {
		return fmt.Errorf("expected valid IPv4 address but received: %v", hostIp)
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
