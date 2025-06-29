package store

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"

	"github.com/Soliard/go-tpl-metrics/models"
)

type fileStorage struct {
	memory   *memStorage
	filePath string
}

func NewFileStorage(filePath string, isRestore bool) (Storage, error) {
	metrics := map[string]*models.Metrics{}
	if isRestore {
		metricsFromFile, err := restoreFromFile(filePath)
		if err != nil {
			return nil, err
		}
		metrics = metricsFromFile
	}

	return &fileStorage{
		memory:   &memStorage{metrics: metrics},
		filePath: filePath,
	}, nil
}

func (s *fileStorage) UpdateCounter(name string, value *int64) error {
	err := s.memory.UpdateCounter(name, value)
	if err != nil {
		return err
	}
	err = s.saveMemoryToFile()
	if err != nil {
		return err
	}
	return nil
}

func (s *fileStorage) UpdateGauge(name string, value *float64) error {
	err := s.memory.UpdateGauge(name, value)
	if err != nil {
		return err
	}
	err = s.saveMemoryToFile()
	if err != nil {
		return err
	}
	return nil
}

func (s *fileStorage) GetMetric(name string) (metric *models.Metrics, exists bool) {
	return s.memory.GetMetric(name)
}

func (s *fileStorage) GetAllMetrics() []models.Metrics {
	return s.memory.GetAllMetrics()
}

func restoreFromFile(filePath string) (map[string]*models.Metrics, error) {
	file, err := os.OpenFile(filePath, os.O_RDONLY, 0666)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	b, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	metrics := map[string]*models.Metrics{}
	err = json.Unmarshal(b, &metrics)
	if err != nil {
		return nil, err
	}
	return metrics, nil

}

func (s *fileStorage) saveMemoryToFile() error {
	os.MkdirAll(filepath.Dir(s.filePath), 0755)
	file, err := os.OpenFile(s.filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := json.MarshalIndent(s.memory.metrics, "", "  ")
	if err != nil {
		return err
	}
	_, err = file.Write(data)
	if err != nil {
		return err
	}
	return nil
}
