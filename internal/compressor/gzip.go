package compressor

import (
	"bytes"
	"compress/gzip"
)

func CompressData(data []byte) ([]byte, error) {
	compBuf := &bytes.Buffer{}
	gz, err := gzip.NewWriterLevel(compBuf, gzip.BestSpeed)
	if err != nil {
		return nil, err
	}

	if _, err := gz.Write(data); err != nil {
		return nil, err
	}

	if err := gz.Close(); err != nil {
		return nil, err
	}

	return compBuf.Bytes(), nil
}

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
