package config

import (
	"flag"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	ServerHost           string `env:"ADDRESS"`
	LogLevel             string `env:"LOG_LEVEL"`
	StoreIntervalSeconds int    `env:"STORE_INTERVAL"`
	FileStoragePath      string `env:"FILE_STORAGE_PATH"`
	IsRestoreFromFile    bool   `env:"RESTORE"`
	DatabaseDSN          string `env:"DATABASE_DSN"`
}

func New() (*Config, error) {
	config := &Config{}

	flag.StringVar(&config.ServerHost, "a", "localhost:8080", "server addres")
	flag.StringVar(&config.LogLevel, "l", "warn", "log level")
	flag.IntVar(&config.StoreIntervalSeconds, "i", 0, "store data interval in seconds") //not used, storing every update
	flag.StringVar(&config.FileStoragePath, "f", "", "file storage name")               //FileStorage\\default.txt
	flag.BoolVar(&config.IsRestoreFromFile, "r", false, "is need to restore data from existed f file")
	flag.StringVar(&config.DatabaseDSN, "d", "postgres://postgres:postgres@localhost:5432/gotplmetrics", "database connection string") //postgres://postgres:postgres@localhost:5432/gotplmetrics
	flag.Parse()

	err := env.Parse(config)
	if err != nil {
		return nil, err
	}

	return config, nil
}
