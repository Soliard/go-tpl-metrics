package config

import (
	"flag"

	"github.com/Soliard/go-tpl-metrics/internal/logger"
	"github.com/caarlos0/env/v6"
)

type Config struct {
	ServerHost            string `env:"ADDRESS"`
	PollIntervalSeconds   int    `env:"POLL_INTERVAL"`
	ReportIntervalSeconds int    `env:"REPORT_INTERVAL"`
}

func New(logger *logger.Logger) *Config {
	config := &Config{}

	flag.StringVar(&config.ServerHost, "a", "localhost:8080", "server addres")
	flag.IntVar(&config.PollIntervalSeconds, "p", 2, "metrics poll interval is seconds")
	flag.IntVar(&config.ReportIntervalSeconds, "r", 10, "metrics send interval in seconds")
	flag.Parse()

	logger.Info("Agent config after flags: ", config)

	err := env.Parse(config)
	if err != nil {
		logger.Error("Cannot parse config from env: ", err)
	}

	logger.Info("Agent config after env vars: ", config)
	return config
}
