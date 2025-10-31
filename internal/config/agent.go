package config

import (
	"flag"
	"fmt"
	"os"

	"github.com/caarlos0/env/v6"
)

// AgentConfig предоставляет конфигурацию для агента метрик
type AgentConfig struct {
	ServerHost            string `env:"ADDRESS" json:"address"`                 // адрес сервера
	PollIntervalSeconds   int    `env:"POLL_INTERVAL" json:"poll_interval"`     // интервал сбора метрик
	ReportIntervalSeconds int    `env:"REPORT_INTERVAL" json:"report_interval"` // интервал отправки метрик
	LogLevel              string `env:"LOG_LEVEL" json:"log_level"`             // уровень логирования
	SignKey               string `env:"KEY" json:"sign_key"`                    // ключ для подписи данных
	RequestsLimit         int    `env:"RATE_LIMIT" json:"rate_limit"`           // лимит одновременных запросов
	CryptoKey             string `env:"CRYPTO_KEY" json:"crypto_key"`           // путь к публичному ключу для шифрования
}

// AgentJSONConfig представляет структуру JSON конфигурации для агента
type AgentJSONConfig struct {
	Address        string `json:"address"`         // аналог переменной окружения ADDRESS или флага -a
	ReportInterval string `json:"report_interval"` // аналог переменной окружения REPORT_INTERVAL или флага -r
	PollInterval   string `json:"poll_interval"`   // аналог переменной окружения POLL_INTERVAL или флага -p
	CryptoKey      string `json:"crypto_key"`      // аналог переменной окружения CRYPTO_KEY или флага -crypto-key
}

func fillAgentDefaults(c *AgentConfig) {
	if c.ServerHost == "" {
		c.ServerHost = "localhost:8080"
	}
	if c.PollIntervalSeconds < 1 {
		c.PollIntervalSeconds = 2
	}
	if c.ReportIntervalSeconds < 1 {
		c.ReportIntervalSeconds = 10
	}
	if c.LogLevel == "" {
		c.LogLevel = "warn"
	}
	if c.RequestsLimit == 0 {
		c.RequestsLimit = 100
	}
}

// NewAgentConfig создает новую конфигурацию агента.
// Приоритет: переменные окружения > флаги командной строки > JSON файл > значения по умолчанию
func NewAgentConfig() (*AgentConfig, error) {
	config := &AgentConfig{}
	reader := &ConfigFileReader{}

	// 1. Сначала читаем JSON конфигурацию (если указана)
	configFile := reader.GetConfigFilePath()
	if configFile != "" {
		if err := loadAgentFromJSON(config, reader, configFile); err != nil {
			return nil, fmt.Errorf("failed to load config from JSON: %w", err)
		}
	}

	// 2. Затем парсим флаги командной строки (перезаписывают JSON)
	parseAgentFlags(config)

	// 3. Наконец, парсим переменные окружения (самый высокий приоритет)
	err := env.Parse(config)
	if err != nil {
		return nil, err
	}

	fillAgentDefaults(config)

	return config, nil
}

// loadAgentFromJSON загружает конфигурацию агента из JSON файла
func loadAgentFromJSON(config *AgentConfig, reader *ConfigFileReader, filePath string) error {
	var jsonConfig AgentJSONConfig
	if err := reader.LoadJSONConfig(filePath, &jsonConfig); err != nil {
		return err
	}

	// Применяем значения из JSON
	config.ServerHost = jsonConfig.Address
	config.CryptoKey = jsonConfig.CryptoKey

	// Парсим интервалы из строк в секунды
	if jsonConfig.ReportInterval != "" {
		seconds, err := reader.ParseDurationFromString(jsonConfig.ReportInterval)
		if err != nil {
			return err
		}
		config.ReportIntervalSeconds = seconds
	}

	if jsonConfig.PollInterval != "" {
		seconds, err := reader.ParseDurationFromString(jsonConfig.PollInterval)
		if err != nil {
			return err
		}
		config.PollIntervalSeconds = seconds
	}

	return nil
}

// parseAgentFlags парсит флаги командной строки для агента
func parseAgentFlags(config *AgentConfig) error {
	fs := flag.NewFlagSet("agent", flag.ContinueOnError)

	fs.StringVar(&config.ServerHost, "a", config.ServerHost, "server address")
	fs.IntVar(&config.PollIntervalSeconds, "p", config.PollIntervalSeconds, "metrics poll interval in seconds")
	fs.IntVar(&config.ReportIntervalSeconds, "r", config.ReportIntervalSeconds, "metrics send interval in seconds")
	fs.StringVar(&config.LogLevel, "ll", config.LogLevel, "log level")
	fs.StringVar(&config.SignKey, "k", config.SignKey, "key will be used for signing data from agent")
	fs.IntVar(&config.RequestsLimit, "l", config.RequestsLimit, "server request rate limit")
	fs.StringVar(&config.CryptoKey, "crypto-key", config.CryptoKey, "path to public PEM key for encryption")

	err := fs.Parse(os.Args[1:])
	return err
}
