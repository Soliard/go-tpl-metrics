package config

import (
	"flag"

	"github.com/Soliard/go-tpl-metrics/internal/logger"
	"github.com/caarlos0/env/v6"
)

type Config struct {
	ServerHost string `env:"ADDRESS"`
}

func New(logger *logger.Logger) *Config {
	config := &Config{}

	flag.StringVar(&config.ServerHost, "a", "localhost:8080", "server addres")
	flag.Parse()

	logger.Info("Server config after flags: ", config)

	err := env.Parse(config)
	if err != nil {
		logger.Error("Cannot parse config from env: ", err)
	}

	logger.Info("Server config after env vars: ", config)
	return config
}
