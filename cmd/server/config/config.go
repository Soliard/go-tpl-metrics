package config

import (
	"flag"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	ServerHost string `env:"ADDRESS"`
	LogLevel   string `env:"LOG_LEVEL"`
}

func New() (*Config, error) {
	config := &Config{}

	flag.StringVar(&config.ServerHost, "a", "localhost:8080", "server addres")
	flag.StringVar(&config.LogLevel, "l", "warn", "log level")
	flag.Parse()

	err := env.Parse(config)
	if err != nil {
		return nil, err
	}

	return config, nil
}
