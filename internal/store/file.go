package store

import (
	"context"
	"encoding/json"
	"fmt"
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

func (s *fileStorage) UpdateCounter(ctx context.Context, name string, value *int64) error {
	err := s.memory.UpdateCounter(ctx, name, value)
	if err != nil {
		return err
	}
	err = s.saveMemoryToFile()
	if err != nil {
		return err
	}
	return nil
}

func (s *fileStorage) UpdateGauge(ctx context.Context, name string, value *float64) error {
	err := s.memory.UpdateGauge(ctx, name, value)
	if err != nil {
		return err
	}
	err = s.saveMemoryToFile()
	if err != nil {
		return err
	}
	return nil
}

func (s *fileStorage) GetMetric(ctx context.Context, name string) (metric *models.Metrics, exists bool) {
	return s.memory.GetMetric(ctx, name)
}

func (s *fileStorage) GetAllMetrics(ctx context.Context) []models.Metrics {
	return s.memory.GetAllMetrics(ctx)
}

func restoreFromFile(filePath string) (map[string]*models.Metrics, error) {
	// в любом случае создаем директории и файл
	os.MkdirAll(filepath.Dir(filePath), 0755)
	file, err := os.OpenFile(filePath, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	b, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	metrics := map[string]*models.Metrics{}
	if string(b) == "" {
		return metrics, nil
	}
	err = json.Unmarshal(b, &metrics)
	if err != nil {
		return nil, fmt.Errorf("cant unmarshal (restore) data from storage file: %v", err)
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
