package config

import (
	"flag"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	ServerHost            string `env:"ADDRESS"`
	PollIntervalSeconds   int    `env:"POLL_INTERVAL"`
	ReportIntervalSeconds int    `env:"REPORT_INTERVAL"`
	LogLevel              string `env:"LOG_LEVEL"`
}

func New() (*Config, error) {
	config := &Config{}

	flag.StringVar(&config.ServerHost, "a", "localhost:8080", "server addres")
	flag.IntVar(&config.PollIntervalSeconds, "p", 2, "metrics poll interval is seconds")
	flag.IntVar(&config.ReportIntervalSeconds, "r", 10, "metrics send interval in seconds")
	flag.StringVar(&config.LogLevel, "l", "warn", "log level")
	flag.Parse()

	err := env.Parse(config)
	if err != nil {
		return nil, err
	}

	return config, nil
}
