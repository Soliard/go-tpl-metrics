package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// ConfigFileReader предоставляет общие функции для работы с конфигурационными файлами
type ConfigFileReader struct{}

// LoadJSONConfig загружает конфигурацию из JSON файла
func (r *ConfigFileReader) LoadJSONConfig(filePath string, target interface{}) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("cannot open config file %s: %w", filePath, err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("cannot decode JSON config: %w", err)
	}

	return nil
}

// GetConfigFilePath определяет путь к файлу конфигурации
// Проверяет флаг -c/-config, затем переменную окружения CONFIG
func (r *ConfigFileReader) GetConfigFilePath() string {
	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		if args[i] == "-c" || args[i] == "-config" {
			if i+1 < len(args) {
				return args[i+1]
			}
		}
	}

	return os.Getenv("CONFIG")
}

// ParseDurationFromString парсит строку времени в секунды
func (r *ConfigFileReader) ParseDurationFromString(durationStr string) (int, error) {
	if durationStr == "" {
		return 0, nil
	}

	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		return 0, fmt.Errorf("invalid duration format '%s': %w", durationStr, err)
	}

	return int(duration.Seconds()), nil
}
