// Package compressor предоставляет утилиты для сжатия и распаковки данных.
// Использует gzip алгоритм с пулом объектов для оптимизации производительности.
package compressor

import (
	"bytes"
	"compress/gzip"
	"io"
	"sync"
)

var (
	// gzipWriterPool пул gzip.Writer для переиспользования объектов
	gzipWriterPool = sync.Pool{
		New: func() interface{} {
			gz, _ := gzip.NewWriterLevel(io.Discard, gzip.BestSpeed)
			return gz
		},
	}

	// bufferPool пул bytes.Buffer для переиспользования буферов
	bufferPool = sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}
)

// CompressData сжимает данные используя gzip алгоритм.
// Использует пулы объектов для оптимизации производительности.
func CompressData(data []byte) ([]byte, error) {
	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufferPool.Put(buf)

	gz := gzipWriterPool.Get().(*gzip.Writer)
	gz.Reset(buf)
	defer gzipWriterPool.Put(gz)

	if _, err := gz.Write(data); err != nil {
		return nil, err
	}

	if err := gz.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// UncompressData распаковывает данные сжатые gzip алгоритмом.
func UncompressData(data []byte) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	defer gz.Close()

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(gz); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
