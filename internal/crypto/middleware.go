// internal/crypto/middleware.go
package crypto

import (
	"bytes"
	"crypto/rsa"
	"encoding/json"
	"io"
	"net/http"

	"github.com/Soliard/go-tpl-metrics/internal/logger"
	"go.uber.org/zap"
)

// DecryptMiddleware создает middleware для расшифровки входящих запросов
func DecryptMiddleware(privateKey *rsa.PrivateKey, fallbackLogger *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := logger.LoggerFromCtx(r.Context(), fallbackLogger)

			if privateKey == nil {
				// Шифрование отключено
				next.ServeHTTP(w, r)
				return
			}

			// Читаем зашифрованное тело запроса
			encryptedBody, err := io.ReadAll(r.Body)
			if err != nil {
				logger.Error("failed to read encrypted body", zap.Error(err))
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{"error": "failed to read encrypted body"})
				return
			}

			// Расшифровываем данные
			decryptedBody, err := DecryptHybrid(encryptedBody, privateKey)
			if err != nil {
				logger.Error("failed to decrypt body", zap.Error(err))
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{"error": "failed to decrypt body"})
				return
			}

			// Восстанавливаем тело запроса
			r.Body = io.NopCloser(bytes.NewBuffer(decryptedBody))
			logger.Info("request body decrypted successfully")

			next.ServeHTTP(w, r)
		})
	}
}
