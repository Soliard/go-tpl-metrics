package signer

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/Soliard/go-tpl-metrics/internal/logger"
	"go.uber.org/zap"
)

type SignedResponseWriter struct {
	http.ResponseWriter
	bodyBuffer *bytes.Buffer
	statusCode int
}

func (r *SignedResponseWriter) Write(b []byte) (int, error) {
	return r.bodyBuffer.Write(b)
}

func (r *SignedResponseWriter) WriteHeader(statusCode int) {
	r.statusCode = statusCode
}

func SignResponseMiddleware(key []byte, fallbackLogger *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := logger.LoggerFromCtx(r.Context(), fallbackLogger)
			if !SignKeyExists(key) {
				next.ServeHTTP(w, r)
				return
			}

			sw := &SignedResponseWriter{
				ResponseWriter: w,
				bodyBuffer:     &bytes.Buffer{},
			}
			logger.Info("response will be signed")
			next.ServeHTTP(sw, r)

			signature := Sign(sw.bodyBuffer.Bytes(), key)
			w.Header().Set("HashSHA256", EncodeSign(signature))
			if sw.statusCode != 0 {
				w.WriteHeader(sw.statusCode)
			}

			w.Write(sw.bodyBuffer.Bytes())
		})
	}
}

func VerifySignatureMiddleware(key []byte, fallbackLogger *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := logger.LoggerFromCtx(r.Context(), fallbackLogger)
			if !SignKeyExists(key) {
				next.ServeHTTP(w, r)
				return
			}

			signatureHex := r.Header.Get("HashSHA256")
			if signatureHex == "" {
				logger.Debug("expected signature, but empty")
				logger.Warn("recieved request without sign")
				if r.Header.Get("Content-Type") == "application/json" {
					w.Header().Set("Content-Type", "application/json")
				}
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{"error": "missing signature"})
				return
			}
			signature, err := DecodeSign(signatureHex)
			if err != nil {
				logger.Warn("invalid signature format")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{"error": "invalid signature format"})
				return
			}
			body, err := io.ReadAll(r.Body)
			if err != nil {
				logger.Warn("failed to read body")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{"error": "failed to read body"})
				return
			}
			// восстановление тела
			r.Body = io.NopCloser(bytes.NewBuffer(body))

			if !Verify(body, key, signature) {
				logger.Warn("invalid signature")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{"error": "invalid signature"})
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
