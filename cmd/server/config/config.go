// Package config предоставляет конфигурацию для сервера метрик.
// Поддерживает настройку через флаги командной строки и переменные окружения.
package config

import (
	"flag"

	"github.com/caarlos0/env/v6"
)

// Config содержит все настройки сервера метрик
type Config struct {
	ServerHost           string `env:"ADDRESS"`           // адрес сервера
	LogLevel             string `env:"LOG_LEVEL"`         // уровень логирования
	StoreIntervalSeconds int    `env:"STORE_INTERVAL"`    // интервал сохранения (не используется)
	FileStoragePath      string `env:"FILE_STORAGE_PATH"` // путь к файлу хранилища
	IsRestoreFromFile    bool   `env:"RESTORE"`           // восстанавливать из файла при запуске
	DatabaseDSN          string `env:"DATABASE_DSN"`      // строка подключения к БД
	SignKey              string `env:"KEY"`               // ключ для подписи данных
}

// New создает новую конфигурацию сервера.
// Парсит флаги командной строки и переменные окружения.
func New() (*Config, error) {
	config := &Config{}

	flag.StringVar(&config.ServerHost, "a", "localhost:8080", "server addres")
	flag.StringVar(&config.LogLevel, "l", "warn", "log level")
	flag.IntVar(&config.StoreIntervalSeconds, "i", 0, "store data interval in seconds") //not used, storing every update
	flag.StringVar(&config.FileStoragePath, "f", "", "file storage name")               //FileStorage\\default.txt
	flag.BoolVar(&config.IsRestoreFromFile, "r", false, "is need to restore data from existed f file")
	//postgres://postgres:postgres@localhost:5432/gotplmetrics?sslmode=disable
	flag.StringVar(&config.DatabaseDSN, "d", "postgres://postgres:postgres@localhost:5432/gotplmetrics?sslmode=disable", "database connection string")
	flag.StringVar(&config.SignKey, "k", "", "key will be used for signing and verifying data")
	flag.Parse()

	err := env.Parse(config)
	if err != nil {
		return nil, err
	}

	return config, nil
}
