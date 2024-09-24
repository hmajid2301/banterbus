package config

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	gap "github.com/muesli/go-app-paths"
	"github.com/sethvargo/go-envconfig"
)

type Config struct {
	DBFolder    string
	Environment string
	LogLevel    slog.Level
}

type ConfigIn struct {
	DBFolder    string `env:"BANTERBUS_DB_FOLDER"`
	Environment string `env:"BANTERBUS_ENVIRONMENT" default:"production"`
	LogLevel    string `env:"BANTERBUS_LOG_LEVEL"   default:"info"`
}

func LoadConfig(ctx context.Context) (Config, error) {
	var input ConfigIn
	if err := envconfig.Process(ctx, &input); err != nil {
		return Config{}, err
	}

	config := Config{
		DBFolder:    input.DBFolder,
		Environment: input.Environment,
		LogLevel:    parseLogLevel(input.LogLevel),
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
