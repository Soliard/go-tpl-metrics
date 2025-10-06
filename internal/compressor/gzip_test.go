package compressor

import (
	"bytes"
	"compress/gzip"
	"io"
	"testing"
)

func TestCompressData(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		wantErr bool
	}{
		{
			name:    "valid data",
			data:    []byte("test data for compression"),
			wantErr: false,
		},
		{
			name:    "empty data",
			data:    []byte(""),
			wantErr: false,
		},
		{
			name:    "large data",
			data:    bytes.Repeat([]byte("a"), 1000),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compressed, err := CompressData(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("CompressData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(compressed) == 0 {
					t.Errorf("CompressData() returned empty compressed data")
				}
			}
		})
	}
}

func TestUncompressData(t *testing.T) {
	originalData := []byte("test data for compression")
	compressed, err := CompressData(originalData)
	if err != nil {
		t.Fatalf("CompressData() error = %v", err)
	}

	tests := []struct {
		name     string
		data     []byte
		wantErr  bool
		expected []byte
	}{
		{
			name:     "valid compressed data",
			data:     compressed,
			wantErr:  false,
			expected: originalData,
		},
		{
			name:     "invalid compressed data",
			data:     []byte("invalid compressed data"),
			wantErr:  true,
			expected: nil,
		},
		{
			name:     "empty data",
			data:     []byte(""),
			wantErr:  true,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uncompressed, err := UncompressData(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("UncompressData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !bytes.Equal(uncompressed, tt.expected) {
				t.Errorf("UncompressData() = %v, want %v", uncompressed, tt.expected)
			}
		})
	}
}

func TestCompressUncompressRoundTrip(t *testing.T) {
	originalData := []byte("test data for round trip compression")

	compressed, err := CompressData(originalData)
	if err != nil {
		t.Fatalf("CompressData() error = %v", err)
	}

	uncompressed, err := UncompressData(compressed)
	if err != nil {
		t.Fatalf("UncompressData() error = %v", err)
	}

	if !bytes.Equal(originalData, uncompressed) {
		t.Errorf("Round trip failed: original = %v, uncompressed = %v", originalData, uncompressed)
	}
}

func TestCompressReader(t *testing.T) {
	originalData := []byte("test data for compression")

	// Compress the data
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	gz.Write(originalData)
	gz.Close()
	compressedData := buf.Bytes()

	// Create compressReader
	reader := io.NopCloser(bytes.NewReader(compressedData))
	cr, err := newCompressReader(reader)
	if err != nil {
		t.Fatalf("newCompressReader() error = %v", err)
	}
	defer cr.Close()

	// Test Read
	readData := make([]byte, len(originalData))
	n, err := cr.Read(readData)
	if err != nil && err != io.EOF {
		t.Errorf("Read() error = %v", err)
	}
	if n != len(originalData) {
		t.Errorf("Read() read %d bytes, want %d", n, len(originalData))
	}
	if !bytes.Equal(readData, originalData) {
		t.Errorf("Read() data = %v, want %v", readData, originalData)
	}
}

func TestCompressReaderClose(t *testing.T) {
	originalData := []byte("test data")

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	gz.Write(originalData)
	gz.Close()
	compressedData := buf.Bytes()

	reader := io.NopCloser(bytes.NewReader(compressedData))
	cr, err := newCompressReader(reader)
	if err != nil {
		t.Fatalf("newCompressReader() error = %v", err)
	}

	// Test Close
	err = cr.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}
}
