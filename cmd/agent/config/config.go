package config

import (
	"flag"
	"fmt"

	"github.com/Soliard/go-tpl-metrics/internal/logger"
	"github.com/caarlos0/env/v6"
)

type Config struct {
	ServerHost            string `env:"ADDRESS"`
	PollIntervalSeconds   int    `env:"POLL_INTERVAL"`
	ReportIntervalSeconds int    `env:"REPORT_INTERVAL"`
}

func New() *Config {
	config := &Config{}

	flag.StringVar(&config.ServerHost, "a", "localhost:8080", "server addres")
	flag.IntVar(&config.PollIntervalSeconds, "p", 2, "metrics poll interval is seconds")
	flag.IntVar(&config.ReportIntervalSeconds, "r", 10, "metrics send interval in seconds")
	flag.Parse()

	logger.LogConfig("agent", config)

	err := env.Parse(config)
	if err != nil {
		logger.LogError("agent", fmt.Errorf("cannot parse config from env: %w", err))
	}

	logger.LogConfig("agent", config)
	return config
}
