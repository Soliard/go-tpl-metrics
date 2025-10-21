package config

import (
	"flag"
	"fmt"
	"os"

	"github.com/caarlos0/env/v6"
)

// ServerConfig содержит все настройки сервера метрик
type ServerConfig struct {
	ServerHost           string `env:"ADDRESS" json:"address"`               // адрес сервера
	LogLevel             string `env:"LOG_LEVEL" json:"log_level"`           // уровень логирования
	StoreIntervalSeconds int    `env:"STORE_INTERVAL" json:"store_interval"` // интервал сохранения
	FileStoragePath      string `env:"FILE_STORAGE_PATH" json:"store_file"`  // путь к файлу хранилища
	IsRestoreFromFile    bool   `env:"RESTORE" json:"restore"`               // восстанавливать из файла
	DatabaseDSN          string `env:"DATABASE_DSN" json:"database_dsn"`     // строка подключения к БД
	SignKey              string `env:"KEY" json:"sign_key"`                  // ключ для подписи данных
	CryptoKey            string `env:"CRYPTO_KEY" json:"crypto_key"`         // путь к приватному ключу
}

// ServerJSONConfig представляет структуру JSON конфигурации для сервера
type ServerJSONConfig struct {
	Address       string `json:"address"`        // аналог переменной окружения ADDRESS или флага -a
	Restore       bool   `json:"restore"`        // аналог переменной окружения RESTORE или флага -r
	StoreInterval string `json:"store_interval"` // аналог переменной окружения STORE_INTERVAL или флага -i
	StoreFile     string `json:"store_file"`     // аналог переменной окружения STORE_FILE или -f
	DatabaseDSN   string `json:"database_dsn"`   // аналог переменной окружения DATABASE_DSN или флага -d
	CryptoKey     string `json:"crypto_key"`     // аналог переменной окружения CRYPTO_KEY или флага -crypto-key
}

// NewServerConfig создает новую конфигурацию сервера.
// Приоритет: переменные окружения > флаги командной строки > JSON файл > значения по умолчанию
func NewServerConfig() (*ServerConfig, error) {
	config := &ServerConfig{}
	reader := &ConfigFileReader{}

	// 1. Читаем JSON конфигурацию (если указана)
	configFile := reader.GetConfigFilePath()
	if configFile != "" {
		if err := loadServerFromJSON(config, reader, configFile); err != nil {
			return nil, fmt.Errorf("failed to load config from JSON: %w", err)
		}
	}

	// 2. Затем парсим флаги командной строки (перезаписывают JSON)
	parseServerFlags(config)

	// 3. Наконец, парсим переменные окружения (самый высокий приоритет)
	err := env.Parse(config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

// loadServerFromJSON загружает конфигурацию сервера из JSON файла
func loadServerFromJSON(config *ServerConfig, reader *ConfigFileReader, filePath string) error {
	var jsonConfig ServerJSONConfig
	if err := reader.LoadJSONConfig(filePath, &jsonConfig); err != nil {
		return err
	}

	// Применяем значения из JSON
	config.ServerHost = jsonConfig.Address
	config.IsRestoreFromFile = jsonConfig.Restore
	config.FileStoragePath = jsonConfig.StoreFile
	config.DatabaseDSN = jsonConfig.DatabaseDSN
	config.CryptoKey = jsonConfig.CryptoKey

	// Парсим store_interval из строки в секунды
	if jsonConfig.StoreInterval != "" {
		seconds, err := reader.ParseDurationFromString(jsonConfig.StoreInterval)
		if err != nil {
			return err
		}
		config.StoreIntervalSeconds = seconds
	}

	return nil
}

// parseServerFlags парсит флаги командной строки для сервера
func parseServerFlags(config *ServerConfig) error {
	fs := flag.NewFlagSet("server", flag.ContinueOnError)

	fs.StringVar(&config.ServerHost, "a", config.ServerHost, "server address")
	fs.StringVar(&config.LogLevel, "l", "warn", "log level")
	fs.IntVar(&config.StoreIntervalSeconds, "i", config.StoreIntervalSeconds, "store data interval in seconds")
	fs.StringVar(&config.FileStoragePath, "f", config.FileStoragePath, "file storage name")
	fs.BoolVar(&config.IsRestoreFromFile, "r", config.IsRestoreFromFile, "is need to restore data from existed file")
	fs.StringVar(&config.DatabaseDSN, "d", config.DatabaseDSN, "database connection string")
	fs.StringVar(&config.SignKey, "k", "", "key will be used for signing and verifying data")
	fs.StringVar(&config.CryptoKey, "crypto-key", config.CryptoKey, "path to private PEM key for decryption")

	err := fs.Parse(os.Args[1:])
	return err
}
