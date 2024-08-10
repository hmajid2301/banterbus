package config

import (
	"context"
	"fmt"
	"os"

	gap "github.com/muesli/go-app-paths"
	"github.com/sethvargo/go-envconfig"
)

type Config struct {
	DBFolder    string `env:"BANTERBUS_DB_FOLDER"`
	Environment string `env:"BANTERBUS_ENVIRONMENT" default:"dev"`
	LogLevel    string `env:"BANTERBUS_LOG_LEVEL" default:"info"`
}

func LoadConfig(ctx context.Context) (Config, error) {
	var config Config
	if err := envconfig.Process(ctx, &config); err != nil {
		return config, err
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
