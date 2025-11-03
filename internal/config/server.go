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
	GRPCServerHost       string `env:"GRPC_ADDRESS" json:"grpc_address"`     // адрес gRPC сервера
	LogLevel             string `env:"LOG_LEVEL" json:"log_level"`           // уровень логирования
	StoreIntervalSeconds int    `env:"STORE_INTERVAL" json:"store_interval"` // интервал сохранения
	FileStoragePath      string `env:"FILE_STORAGE_PATH" json:"store_file"`  // путь к файлу хранилища
	IsRestoreFromFile    bool   `env:"RESTORE" json:"restore"`               // восстанавливать из файла
	DatabaseDSN          string `env:"DATABASE_DSN" json:"database_dsn"`     // строка подключения к БД
	SignKey              string `env:"KEY" json:"sign_key"`                  // ключ для подписи данных
	CryptoKey            string `env:"CRYPTO_KEY" json:"crypto_key"`         // путь к приватному ключу
	TrustedSubnet        string `env:"TRUSTED_SUBNET" json:"trusted_subnet"` // доверенная подсеть (CIDR)
}

// ServerJSONConfig представляет структуру JSON конфигурации для сервера
type ServerJSONConfig struct {
	Address       string `json:"address"`        // аналог переменной окружения ADDRESS или флага -a
	GRPCAddress   string `json:"grpc_address"`   // адрес gRPC сервера
	Restore       bool   `json:"restore"`        // аналог переменной окружения RESTORE или флага -r
	StoreInterval string `json:"store_interval"` // аналог переменной окружения STORE_INTERVAL или флага -i
	StoreFile     string `json:"store_file"`     // аналог переменной окружения STORE_FILE или -f
	DatabaseDSN   string `json:"database_dsn"`   // аналог переменной окружения DATABASE_DSN или флага -d
	CryptoKey     string `json:"crypto_key"`     // аналог переменной окружения CRYPTO_KEY или флага -crypto-key
	TrustedSubnet string `json:"trusted_subnet"` // аналог TRUSTED_SUBNET или флага -t
}

func fillServerDefaults(c *ServerConfig) {
	if c.ServerHost == "" {
		c.ServerHost = "localhost:8080"
	}
	if c.LogLevel == "" {
		c.LogLevel = "warn"
	}
	if c.StoreIntervalSeconds == 0 {
		c.StoreIntervalSeconds = 5
	}
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

	fillServerDefaults(config)

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
	config.GRPCServerHost = jsonConfig.GRPCAddress
	config.IsRestoreFromFile = jsonConfig.Restore
	config.FileStoragePath = jsonConfig.StoreFile
	config.DatabaseDSN = jsonConfig.DatabaseDSN
	config.CryptoKey = jsonConfig.CryptoKey
	config.TrustedSubnet = jsonConfig.TrustedSubnet

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
	fs.StringVar(&config.GRPCServerHost, "ga", config.GRPCServerHost, "grpc server address")
	fs.StringVar(&config.LogLevel, "l", config.LogLevel, "log level")
	fs.IntVar(&config.StoreIntervalSeconds, "i", config.StoreIntervalSeconds, "store data interval in seconds")
	fs.StringVar(&config.FileStoragePath, "f", config.FileStoragePath, "file storage name")
	fs.BoolVar(&config.IsRestoreFromFile, "r", config.IsRestoreFromFile, "is need to restore data from existed file")
	fs.StringVar(&config.DatabaseDSN, "d", config.DatabaseDSN, "database connection string")
	fs.StringVar(&config.SignKey, "k", config.SignKey, "key will be used for signing and verifying data")
	fs.StringVar(&config.CryptoKey, "crypto-key", config.CryptoKey, "path to private PEM key for decryption")
	fs.StringVar(&config.TrustedSubnet, "t", config.TrustedSubnet, "trusted subnet in CIDR (e.g. 192.168.1.0/24)")

	err := fs.Parse(os.Args[1:])
	return err
}
