package compressor

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"
)

func TestGzipMiddleware(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	tests := []struct {
		name                string
		acceptEncoding      string
		contentEncoding     string
		contentType         string
		responseBody        string
		expectCompression   bool
		expectDecompression bool
	}{
		{
			name:                "client supports gzip, server compresses",
			acceptEncoding:      "gzip",
			contentType:         "application/json",
			responseBody:        `{"test": "data"}`,
			expectCompression:   true,
			expectDecompression: false,
		},
		{
			name:                "client doesn't support gzip",
			acceptEncoding:      "",
			contentType:         "application/json",
			responseBody:        `{"test": "data"}`,
			expectCompression:   false,
			expectDecompression: false,
		},
		{
			name:                "client sends gzip data",
			acceptEncoding:      "gzip",
			contentEncoding:     "gzip",
			contentType:         "application/json",
			responseBody:        `{"test": "data"}`,
			expectCompression:   true,
			expectDecompression: true,
		},
		{
			name:                "non-compressible content type",
			acceptEncoding:      "gzip",
			contentType:         "text/plain",
			responseBody:        "plain text",
			expectCompression:   false,
			expectDecompression: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", tt.contentType)
				w.Write([]byte(tt.responseBody))
			})

			middleware := GzipMiddleware(logger)
			wrappedHandler := middleware(handler)

			// Подготавливаем тело запроса
			var requestBody io.Reader
			if tt.contentEncoding == "gzip" && tt.expectDecompression {
				// Сжимаем данные для отправки
				compressed, err := CompressData([]byte(tt.responseBody))
				if err != nil {
					t.Fatalf("Failed to compress test data: %v", err)
				}
				requestBody = bytes.NewBuffer(compressed)
			} else {
				requestBody = bytes.NewBufferString(tt.responseBody)
			}

			req := httptest.NewRequest("POST", "/", requestBody)
			if tt.acceptEncoding != "" {
				req.Header.Set("Accept-Encoding", tt.acceptEncoding)
			}
			if tt.contentEncoding != "" {
				req.Header.Set("Content-Encoding", tt.contentEncoding)
			}

			w := httptest.NewRecorder()

			// ВЫПОЛНЯЕМ ЗАПРОС
			wrappedHandler.ServeHTTP(w, req)

			// Check compression
			if tt.expectCompression {
				if w.Header().Get("Content-Encoding") != "gzip" {
					t.Errorf("Expected Content-Encoding: gzip, got %s", w.Header().Get("Content-Encoding"))
				}
			} else {
				if w.Header().Get("Content-Encoding") == "gzip" {
					t.Errorf("Unexpected Content-Encoding: gzip")
				}
			}

			// Check response body
			if tt.expectCompression {
				// Verify the response is actually compressed
				body := w.Body.Bytes()
				if len(body) > 0 {
					// Check if it's valid gzip data
					reader, err := gzip.NewReader(bytes.NewReader(body))
					if err != nil {
						t.Errorf("Response body is not valid gzip: %v", err)
					} else {
						reader.Close()
					}
				}
			}
		})
	}
}

func TestGzipMiddlewareWithInvalidCompressedData(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("response"))
	})

	middleware := GzipMiddleware(logger)
	wrappedHandler := middleware(handler)

	// Send invalid gzip data
	req := httptest.NewRequest("POST", "/", bytes.NewBufferString("invalid gzip data"))
	req.Header.Set("Content-Encoding", "gzip")

	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	// Should return 500 Internal Server Error
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestBufferWriter(t *testing.T) {
	originalW := httptest.NewRecorder()
	bw := &bufferWriter{
		ResponseWriter: originalW,
		headers:        make(http.Header),
		bodyBuf:        make([]byte, 0),
	}

	// Test WriteHeader - сохраняет статус код
	statusCode := http.StatusCreated
	bw.WriteHeader(statusCode)
	if bw.statusCode != statusCode {
		t.Errorf("WriteHeader() statusCode = %d, want %d", bw.statusCode, statusCode)
	}

	// Test Write - буферизует данные
	testData := []byte("test data")
	n, err := bw.Write(testData)
	if err != nil {
		t.Errorf("Write() error = %v", err)
	}
	if n != len(testData) {
		t.Errorf("Write() wrote %d bytes, want %d", n, len(testData))
	}
	if !bytes.Equal(bw.bodyBuf, testData) {
		t.Errorf("bodyBuf = %v, want %v", bw.bodyBuf, testData)
	}

	// Test Write multiple times - накапливает данные
	moreData := []byte(" more data")
	_, err = bw.Write(moreData)
	if err != nil {
		t.Errorf("Write() error = %v", err)
	}
	expectedData := append(testData, moreData...)
	if !bytes.Equal(bw.bodyBuf, expectedData) {
		t.Errorf("bodyBuf after second write = %v, want %v", bw.bodyBuf, expectedData)
	}
}
