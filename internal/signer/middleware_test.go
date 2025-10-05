package signer

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"
)

func TestSignResponseMiddleware(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	tests := []struct {
		name           string
		key            []byte
		responseBody   string
		expectedHeader string
	}{
		{
			name:           "with valid key",
			key:            []byte("secret"),
			responseBody:   "test response",
			expectedHeader: "HashSHA256",
		},
		{
			name:           "with empty key",
			key:            []byte(""),
			responseBody:   "test response",
			expectedHeader: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(tt.responseBody))
			})

			middleware := SignResponseMiddleware(tt.key, logger)
			wrappedHandler := middleware(handler)

			req := httptest.NewRequest("GET", "/", nil)
			w := httptest.NewRecorder()

			wrappedHandler.ServeHTTP(w, req)

			if tt.expectedHeader != "" {
				if w.Header().Get(tt.expectedHeader) == "" {
					t.Errorf("Expected header %s not found", tt.expectedHeader)
				}
			} else {
				if w.Header().Get("HashSHA256") != "" {
					t.Errorf("Unexpected HashSHA256 header found")
				}
			}
		})
	}
}

func TestVerifySignatureMiddleware(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	key := []byte("secret")
	data := []byte("test data")
	signature := Sign(data, key)
	signatureHex := EncodeSign(signature)

	tests := []struct {
		name           string
		key            []byte
		requestBody    string
		signature      string
		expectedStatus int
	}{
		{
			name:           "valid signature",
			key:            key,
			requestBody:    string(data),
			signature:      signatureHex,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing signature",
			key:            key,
			requestBody:    string(data),
			signature:      "",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid signature format",
			key:            key,
			requestBody:    string(data),
			signature:      "invalid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "wrong signature",
			key:            key,
			requestBody:    string(data),
			signature:      EncodeSign([]byte("wrong")),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "empty key",
			key:            []byte(""),
			requestBody:    string(data),
			signature:      signatureHex,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			middleware := VerifySignatureMiddleware(tt.key, logger)
			wrappedHandler := middleware(handler)

			req := httptest.NewRequest("POST", "/", bytes.NewBufferString(tt.requestBody))
			if tt.signature != "" {
				req.Header.Set("HashSHA256", tt.signature)
			}
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()

			wrappedHandler.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestSignedResponseWriter(t *testing.T) {
	body := []byte("test response")
	statusCode := http.StatusOK

	w := httptest.NewRecorder()

	sw := &SignedResponseWriter{
		ResponseWriter: w,
		bodyBuffer:     &bytes.Buffer{},
	}

	// Test Write
	n, err := sw.Write(body)
	if err != nil {
		t.Errorf("Write() error = %v", err)
	}
	if n != len(body) {
		t.Errorf("Write() wrote %d bytes, want %d", n, len(body))
	}

	// Test WriteHeader
	sw.WriteHeader(statusCode)
	if sw.statusCode != statusCode {
		t.Errorf("WriteHeader() statusCode = %d, want %d", sw.statusCode, statusCode)
	}

	// Test body buffer
	if !bytes.Equal(sw.bodyBuffer.Bytes(), body) {
		t.Errorf("bodyBuffer = %v, want %v", sw.bodyBuffer.Bytes(), body)
	}
}
