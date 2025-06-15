package main

import (
	"flag"
	"fmt"

	"github.com/Soliard/go-tpl-metrics/internal/logger"
	"github.com/Soliard/go-tpl-metrics/internal/server"
	"github.com/caarlos0/env/v6"
)

func ParseFlags() server.Config {
	config := server.Config{}

	flag.StringVar(&config.ServerHost, "a", "localhost:8080", "server addres")
	flag.Parse()

	logger.LogConfig("server", config)

	err := env.Parse(&config)
	if err != nil {
		logger.LogError("server", fmt.Errorf("cannot parse config from env: %w", err))
	}

	logger.LogConfig("server", config)
	return config
}
