// Package config предоставляет конфигурацию для агента сбора метрик.
// Поддерживает настройку через флаги командной строки и переменные окружения.
package config

import (
	"flag"

	"github.com/caarlos0/env/v6"
)

// Config предоставляет конфигурацию для агента метрик.
// Поддерживает настройку через флаги командной строки и переменные окружения.
type Config struct {
	ServerHost            string `env:"ADDRESS"`         // адрес сервера
	PollIntervalSeconds   int    `env:"POLL_INTERVAL"`   // интервал сбора метрик
	ReportIntervalSeconds int    `env:"REPORT_INTERVAL"` // интервал отправки метрик
	LogLevel              string `env:"LOG_LEVEL"`       // уровень логирования
	SignKey               string `env:"KEY"`             // ключ для подписи данных
	RequestsLimit         int    `env:"RATE_LIMIT"`      // лимит одновременных запросов
}

// New создает новую конфигурацию агента.
// Парсит флаги командной строки и переменные окружения.
func New() (*Config, error) {
	config := &Config{}

	flag.StringVar(&config.ServerHost, "a", "localhost:8080", "server addres")
	flag.IntVar(&config.PollIntervalSeconds, "p", 2, "metrics poll interval is seconds")
	flag.IntVar(&config.ReportIntervalSeconds, "r", 10, "metrics send interval in seconds")
	flag.StringVar(&config.LogLevel, "ll", "warn", "log level")
	flag.StringVar(&config.SignKey, "k", "", "key will be used for signing data from agent")
	flag.IntVar(&config.RequestsLimit, "l", 100, "server request rate limit")
	flag.Parse()

	err := env.Parse(config)
	if err != nil {
		return nil, err
	}

	return config, nil
}
